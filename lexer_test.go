package quickql

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestBasicLexer(t *testing.T) {
	cases := []struct {
		input string
	}{
		{input: "hello=world"},
		{input: "hello = world"},
		{input: "hello:world"},
		{input: "hello<world"},
		{input: "hello = [world]"},
		{input: "hello != [world]"},
		{input: "hello:66e839032fb119d31dc9c968"},
		{input: `foo bar`},
		{input: `"foo bar"`},
		{input: `"foo \"bar\""`},
		{input: `foo OR bar`},
		{input: "-10"},
		{input: "-10.7"},
		{input: "10.7"},
		{input: "hello:world\nfoo:bar"},
		{input: "-hello:world"},
		{input: `name = "Bob" and (age > 20 or age < 5)`},
		{input: "="},
		{input: "~"},
		{input: "foo ~ bar"},
		// identifiers may contain digits, `_` and `.` after the first character
		{input: "field1"},
		{input: "user_2.name"},
		{input: "_private"},
		{input: "obj.prop"},
		// non-ASCII letters are letters, including decomposed (NFD) ones
		{input: "prénom"},
		{input: "имя"},
		{input: "prénom"},
		// a leading digit or symbol is a VALUE
		{input: "1field"},
		{input: "*foo"},
		{input: "*"},
		{input: ".foo"},
		{input: "[world]"},
		{input: "🇨🇦"},
		{input: "\xff"},
		{input: "́x"},
		// other symbols inside a word make it a VALUE
		{input: "foo*"},
		{input: "foo-bar"},
		// keywords are only recognised in all-upper or all-lower case
		{input: "AND and OR or NOT not"},
		{input: "And Or Not"},
		{input: "ANDROID"},
		// operators
		{input: "a<=b"},
		{input: "a>=b"},
		{input: "a!~b"},
		{input: "a>-5"},
		{input: "a<b<c"},
		{input: "(a)"},
		// `-` not followed by a digit is NOT
		{input: "-a"},
		{input: "- a"},
		{input: "-"},
		{input: "--5"},
		{input: "-(a)"},
		// strings
		{input: `""`},
		{input: `''`},
		{input: `'it\'s'`},
		{input: `"say 'hi'"`},
		{input: `"a\\b"`},
		{input: `"a\nb"`}, // escapes are identity only: `\n` is the letter n
		{input: `"a b" c`},
		// whitespace
		{input: "a\tb\r\nc\n"},
		{input: " \t\r\n "},
		{input: "a b"}, // non-breaking space is not whitespace
	}

	queryLexer := NewLexer()

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			l, err := queryLexer.Lex("", strings.NewReader(c.input))
			require.NoError(t, err)

			tokens, err := lexer.ConsumeAll(l)
			require.NoError(t, err)

			snaps.MatchSnapshot(t, DebugOutput(tokens))
		})
	}
}

func TestFailLexer(t *testing.T) {
	cases := []struct {
		input string
	}{
		{input: `"foo`},
		{input: `'test\`},
		{input: `foo!bar`},
		{input: `foo==bar`},
		{input: `"foo\"`},
		{input: `'foo"`},
		{input: "a\n  b=\"x"},
		{input: "!"},
		{input: "foo!"},
		{input: "a ! = b"},
		// NUL is not allowed anywhere in the input
		{input: "foo\x00bar"},
		{input: "\x00"},
		{input: `"a` + "\x00" + `b"`},
		{input: "a b\x00"},
	}

	queryLexer := NewLexer()

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			l, err := queryLexer.Lex("", strings.NewReader(c.input))
			require.NoError(t, err)

			_, err = lexer.ConsumeAll(l)
			var lexErr LexerError
			require.ErrorAs(t, err, &lexErr)

			snaps.MatchSnapshot(t, err.Error())
		})
	}
}

func TestLexerErrorFilename(t *testing.T) {
	l, err := NewLexer().Lex("query.txt", strings.NewReader("a\nb=\"x"))
	require.NoError(t, err)

	_, err = lexer.ConsumeAll(l)
	require.EqualError(t, err, "[query.txt:2:3] unterminated string")
}

var errBoom = errors.New("boom")

func TestLexerReaderError(t *testing.T) {
	cases := map[string]io.Reader{
		"first read": iotest.ErrReader(errBoom),
		"mid input":  io.MultiReader(strings.NewReader("abc "), iotest.ErrReader(errBoom)),
		"mid token":  io.MultiReader(strings.NewReader("abc"), iotest.ErrReader(errBoom)),
		"in string":  io.MultiReader(strings.NewReader(`"abc`), iotest.ErrReader(errBoom)),
	}

	for name, r := range cases {
		t.Run(name, func(t *testing.T) {
			l, err := NewLexer().Lex("", r)
			require.NoError(t, err)

			_, err = lexer.ConsumeAll(l)
			require.ErrorIs(t, err, errBoom)
		})
	}
}

func TestLexerPositions(t *testing.T) {
	l, err := NewLexer().Lex("f", strings.NewReader("ab\n  cd = \"x y\"\r\n(e)"))
	require.NoError(t, err)

	tokens, err := lexer.ConsumeAll(l)
	require.NoError(t, err)

	output := make([]string, len(tokens))
	for i, tok := range tokens {
		output[i] = fmt.Sprintf("%s %s %q", tok.Pos, TokenName(tok), tok.Value)
	}

	snaps.MatchSnapshot(t, output)
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

func TestSimple(t *testing.T) {
	input := `obj.prop=world`
	// input := `(fish!=food)`
	// input := `-hello:world`
	l, err := queryLexer.Lex("", strings.NewReader(input))
	require.NoError(t, err)

	tokens, err := lexer.ConsumeAll(l)
	require.NoError(t, err)
	require.NotEmpty(t, tokens)
}
