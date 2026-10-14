package quickql

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"testing/iotest"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lexTypes returns the token names (without the trailing EOF) for input.
func lexTypes(t *testing.T, input string) []string {
	t.Helper()

	l, err := NewLexer().Lex("", strings.NewReader(input))
	require.NoError(t, err)

	tokens, err := lexer.ConsumeAll(l)
	require.NoError(t, err)

	names := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Type == EOF {
			continue
		}
		names = append(names, TokenName(tok)+":"+tok.Value)
	}

	return names
}

func TestLexerIdentifierClassification(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		// identifiers may contain digits, `_` and `.` after the first character
		{input: "field1", want: []string{"IDENT:field1"}},
		{input: "user_2.name", want: []string{"IDENT:user_2.name"}},
		{input: "_private", want: []string{"IDENT:_private"}},
		{input: "obj.prop", want: []string{"IDENT:obj.prop"}},
		// non-ASCII letters are letters
		{input: "prénom", want: []string{"IDENT:prénom"}},
		{input: "François", want: []string{"IDENT:François"}},
		{input: "имя", want: []string{"IDENT:имя"}},
		{input: "prénom", want: []string{"IDENT:prénom"}},
		{input: "́x", want: []string{"VALUE:́x"}},
		// a leading digit or symbol is a VALUE
		{input: "1field", want: []string{"VALUE:1field"}},
		{input: "*foo", want: []string{"VALUE:*foo"}},
		{input: "*", want: []string{"VALUE:*"}},
		{input: ".foo", want: []string{"VALUE:.foo"}},
		{input: "[world]", want: []string{"VALUE:[world]"}},
		{input: "🇨🇦", want: []string{"VALUE:🇨🇦"}},
		{input: "\xff", want: []string{"VALUE:�"}},
		// other symbols inside a word make it a VALUE
		{input: "foo*", want: []string{"VALUE:foo*"}},
		{input: "foo-bar", want: []string{"VALUE:foo-bar"}},
		// numbers
		{input: "10", want: []string{"VALUE:10"}},
		{input: "-10", want: []string{"VALUE:-10"}},
		{input: "-10.5", want: []string{"VALUE:-10.5"}},
		{input: "66e839032fb119d31dc9c968", want: []string{"VALUE:66e839032fb119d31dc9c968"}},
		// keywords are only recognised in all-upper or all-lower case
		{input: "AND and OR or NOT not", want: []string{
			"KEYWORD:AND", "KEYWORD:and", "KEYWORD:OR", "KEYWORD:or", "KEYWORD:NOT", "KEYWORD:not",
		}},
		{input: "And Or Not", want: []string{"IDENT:And", "IDENT:Or", "IDENT:Not"}},
		{input: "ANDROID", want: []string{"IDENT:ANDROID"}},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			require.Equal(t, c.want, lexTypes(t, c.input))
		})
	}
}

func TestLexerOperatorsAndPunctuation(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{input: "a:b", want: []string{"IDENT:a", "OP:=", "IDENT:b"}},
		{input: "a<=b", want: []string{"IDENT:a", "OP:<=", "IDENT:b"}},
		{input: "a>=b", want: []string{"IDENT:a", "OP:>=", "IDENT:b"}},
		{input: "a!~b", want: []string{"IDENT:a", "OP:!~", "IDENT:b"}},
		{input: "a>-5", want: []string{"IDENT:a", "OP:>", "VALUE:-5"}},
		{input: "a<b<c", want: []string{"IDENT:a", "OP:<", "IDENT:b", "OP:<", "IDENT:c"}},
		{input: "(a)", want: []string{"LPAREN:(", "IDENT:a", "RPAREN:)"}},
		// `-` not followed by a digit is NOT
		{input: "-a", want: []string{"KEYWORD:NOT", "IDENT:a"}},
		{input: "- a", want: []string{"KEYWORD:NOT", "IDENT:a"}},
		{input: "-", want: []string{"KEYWORD:NOT"}},
		{input: "--5", want: []string{"KEYWORD:NOT", "VALUE:-5"}},
		{input: "-(a)", want: []string{"KEYWORD:NOT", "LPAREN:(", "IDENT:a", "RPAREN:)"}},
		// strings
		{input: `""`, want: []string{"STRING:"}},
		{input: `''`, want: []string{"STRING:"}},
		{input: `'it\'s'`, want: []string{"STRING:it's"}},
		{input: `"say 'hi'"`, want: []string{"STRING:say 'hi'"}},
		{input: `"a\\b"`, want: []string{`STRING:a\b`}},
		// escapes are identity only: `\n` is the letter n
		{input: `"a\nb"`, want: []string{"STRING:anb"}},
		{input: `"a b" c`, want: []string{"STRING:a b", "IDENT:c"}},
		// whitespace variants
		{input: "a\tb\r\nc\n", want: []string{"IDENT:a", "IDENT:b", "IDENT:c"}},
		{input: " \t\r\n ", want: []string{}},
		// non-breaking space is not whitespace
		{input: "a b", want: []string{"VALUE:a b"}},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			require.Equal(t, c.want, lexTypes(t, c.input))
		})
	}
}

func TestLexerErrors(t *testing.T) {
	cases := []struct {
		input string
		pos   string
		msg   string
	}{
		{input: `"foo`, pos: "1:1", msg: "unterminated string"},
		{input: `"foo\"`, pos: "1:1", msg: "unterminated string"},
		{input: `'foo"`, pos: "1:1", msg: "unterminated string"},
		{input: `"foo\`, pos: "1:1", msg: "no character after escape"},
		{input: "a\n  b=\"x", pos: "2:5", msg: "unterminated string"},
		{input: "!", pos: "1:1", msg: "operator ! must be followed by = or ~"},
		{input: "foo!", pos: "1:4", msg: "operator ! must be followed by = or ~"},
		{input: "a ! = b", pos: "1:3", msg: "operator ! must be followed by = or ~"},
		{input: "a==b", pos: "1:2", msg: "double == not allowed"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			l, err := NewLexer().Lex("", strings.NewReader(c.input))
			require.NoError(t, err)

			_, err = lexer.ConsumeAll(l)
			require.Error(t, err)
			require.Equal(t, "["+c.pos+"] "+c.msg, err.Error())
		})
	}
}

func TestLexerErrorFilename(t *testing.T) {
	l, err := NewLexer().Lex("query.txt", strings.NewReader("a\nb=\"x"))
	require.NoError(t, err)

	_, err = lexer.ConsumeAll(l)
	require.EqualError(t, err, "[query.txt:2:3] unterminated string")
}

func TestLexerPositions(t *testing.T) {
	l, err := NewLexer().Lex("f", strings.NewReader("ab\n  cd = \"x y\"\r\n(e)"))
	require.NoError(t, err)

	tokens, err := lexer.ConsumeAll(l)
	require.NoError(t, err)

	type tp struct {
		Value  string
		Line   int
		Column int
	}
	got := make([]tp, 0, len(tokens))
	for _, tok := range tokens {
		require.Equal(t, "f", tok.Pos.Filename)
		got = append(got, tp{tok.Value, tok.Pos.Line, tok.Pos.Column})
	}

	require.Equal(t, []tp{
		{"ab", 1, 1},
		{"cd", 2, 3},
		{"=", 2, 6},
		{"x y", 2, 8},
		{"(", 3, 1},
		{"e", 3, 2},
		{")", 3, 3},
		{"", 3, 4},
	}, got)
}

func TestLexerEOFIsSticky(t *testing.T) {
	l, err := NewLexer().Lex("", strings.NewReader("a"))
	require.NoError(t, err)

	tok, err := l.Next()
	require.NoError(t, err)
	require.Equal(t, "a", tok.Value)

	for range 3 {
		tok, err = l.Next()
		require.NoError(t, err)
		require.True(t, tok.EOF())
	}
}

func TestSymbolsCoverAllTokenTypes(t *testing.T) {
	symbols := Symbols()
	for _, tt := range []lexer.TokenType{ILLEGAL, EOF, WS, LPAREN, RPAREN, STRING, IDENT, VALUE, OP, KEYWORD} {
		var tok lexer.Token
		tok.Type = tt
		name := TokenName(tok)
		require.Contains(t, symbols, name)
		require.Equal(t, tt, symbols[name])
	}
}

func TestParseEdgeCases(t *testing.T) {
	cases := []struct {
		input string
	}{
		// identifiers with digits / non-ASCII letters can be fields
		{input: `field1=x`},
		{input: `user_2.name:bob`},
		{input: `prénom=Zoé`},
		{input: `a1 > 5 AND b2 < 6`},
		// values
		{input: `foo=""`},
		{input: `foo=-5`},
		{input: `foo>=-0.5`},
		{input: `foo=*`},
		{input: `foo = "a=b"`},
		{input: `foo = "AND"`},
		// whitespace
		{input: "\ta\r\n\tb\n"},
		{input: "  a  "},
		// keywords
		{input: `a and b`},
		{input: `a or b`},
		{input: `not a`},
		{input: `a And b`},
		{input: `ANDROID`},
		// NOT / `-`
		{input: `-a`},
		{input: `- a`},
		{input: `-(a OR b)`},
		{input: `NOT a=1`},
		{input: `--5`},
		{input: `a -5`},
		{input: `-a -b`},
		// precedence: OR binds tighter than (implicit) AND
		{input: `a OR b c`},
		{input: `a b OR c`},
		{input: `a OR b OR c`},
		{input: `NOT a OR b`},
		// grouping
		{input: `(a)`},
		{input: `(a)(b)`},
		{input: `(a OR b)c`},
		{input: `(a b) OR c`},
		{input: strings.Repeat("(", 200) + "a" + strings.Repeat(")", 200)},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got, err := Parse(c.input)
			require.NoError(t, err)

			snaps.MatchSnapshot(t, got.AsSExpr())
		})
	}
}

func TestParseEdgeCaseErrors(t *testing.T) {
	cases := []string{
		// a field must be an identifier
		`1field=x`,
		`*foo=x`,
		`"foo"=bar`,
		`and=1`,
		`foo-bar=x`,
		// a value can't be a keyword or a negation
		`foo=and`,
		`foo=-bar`,
		`foo=(bar)`,
		// operators don't chain
		`foo=bar=baz`,
		`a<b<c`,
		`foo:`,
		`foo::bar`,
		// dangling / doubled keywords
		`AND`,
		`a AND`,
		`a OR`,
		`AND a`,
		`a AND AND b`, //nolint:dupword
		`a OR OR b`,   //nolint:dupword
		`-`,
		`a NOT`,
		// grouping
		`()`,
		`(a))`,
		`((a)`,
		`)`,
		`(`,
		// lexer errors
		`foo!=`,
		`a==b`,
		`"unterminated`,
	}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			_, err := Parse(input)
			require.Error(t, err)
		})
	}
}

func TestLexerRejectsNUL(t *testing.T) {
	for _, input := range []string{"foo\x00bar", "\x00", `"a` + "\x00" + `b"`, "a b\x00"} {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			l, err := NewLexer().Lex("", strings.NewReader(input))
			require.NoError(t, err)

			_, err = lexer.ConsumeAll(l)
			var lexErr LexerError
			require.ErrorAs(t, err, &lexErr)
			require.Equal(t, "NUL byte not allowed", lexErr.Msg)
			require.Equal(t, strings.IndexByte(input, 0)+1, lexErr.Pos.Column)

			_, err = Parse(input)
			require.ErrorAs(t, err, &lexErr)
		})
	}
}

var errBoom = errors.New("boom")

func TestLexerReaderError(t *testing.T) {
	boom := errBoom

	cases := map[string]io.Reader{
		"first read": iotest.ErrReader(boom),
		"mid input":  io.MultiReader(strings.NewReader("abc "), iotest.ErrReader(boom)),
		"mid token":  io.MultiReader(strings.NewReader("abc"), iotest.ErrReader(boom)),
		"in string":  io.MultiReader(strings.NewReader(`"abc`), iotest.ErrReader(boom)),
	}

	for name, r := range cases {
		t.Run(name, func(t *testing.T) {
			l, err := NewLexer().Lex("", r)
			require.NoError(t, err)

			_, err = lexer.ConsumeAll(l)
			require.ErrorIs(t, err, boom)
		})
	}
}

func TestParseErrorTypes(t *testing.T) {
	var lexErr LexerError

	// lexer failures surface as a LexerError through Parse's wrapping
	_, err := Parse(`foo!bar`)
	require.ErrorAs(t, err, &lexErr)
	require.Equal(t, 4, lexErr.Pos.Column)

	_, err = Parse(`"open`)
	require.ErrorAs(t, err, &lexErr)

	// grammar failures do not
	_, err = Parse(`a AND`)
	require.Error(t, err)
	require.NotErrorAs(t, err, &lexErr)

	// ParseDebug reports the same errors and a trace
	_, trace, err := ParseDebug(`a AND`)
	require.Error(t, err)
	require.NotEmpty(t, trace)
}

func TestParseEmpty(t *testing.T) {
	for _, input := range []string{"", " ", "\t\r\n"} {
		got, err := Parse(input)
		require.NoError(t, err)
		require.Nil(t, got.LogicExpr)
		require.Empty(t, got.AsSExpr())
	}
}

func TestParseConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			input := fmt.Sprintf(`f%d = %d OR (g AND NOT h)`, i, i)
			got, err := Parse(input)
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf(`(or (= f%d "%d") (and (value g) (NOT (value h))))`, i, i), got.AsSExpr())
		}()
	}
	wg.Wait()
}
