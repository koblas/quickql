package quickql

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
	}{
		{input: ``},
		{input: `    `},
		{input: `one two three`},
		{input: `François`},
		{input: `one 🇨🇦 three`},
		{input: `hello:world`},
		// basic operators
		{input: `hello=world`},
		{input: `country=🇨🇦`},
		{input: `hello = world`},
		{input: `hello!=world`},
		{input: `obj.prop=world`},
		{input: `hello != world`},
		{input: `hello<world`},
		{input: `hello<=world`},
		{input: `hello>world`},
		{input: `hello>=world`},
		{input: `hello ~ world`},
		{input: `hello !~ world`},
		{input: `hello`},
		{input: `hello world`},
		{input: `hello:world AND super:cool`},
		{input: `hello:world super:cool`},
		{input: `hello:world AND super:cool AND really:cool`},
		{input: `hello:world super:cool really:cool`},
		{input: `hello:null`},
		{input: `foo:*bar`},
		{input: `foo:*bar*`},
		{input: `foo:bar*`},
		{input: `hello:world OR super:cool`},
		{input: `hello:world or cool`},
		{input: `world OR super:cool`},
		{input: `hello:66e839032fb119d31dc9c968`},
		{input: `hello:"super cool"`},
		{input: `-hello:77`},
		{input: `hello:66`},
		{input: `hello:66.2`},
		{input: `foo>1 AND bar<2`},
		{input: `panic NOT ever`},
		{input: `repo:has.commit.after(yesterday)`}, // not really supported, but testing the parser
		{input: `hello>=3 OR world<=4.7`},
		{input: `(hello>=3) OR (world<=4.7)`},
		{input: `name = "Bill"`},  // this is an expression.
		{input: `name eq "Bill"`}, // this is a series of tokens.
		{input: `name = "Bob" and (age > 20 or age < 5)`},
		{input: `name = "Bob" and (age>20 or age<5)`},
		{input: "cat!=dog ( foo>1 AND bar<2 and cat=fish) OR ( baz<=3 AND qux>=4 )"},
		{input: `hello >  77`},
		{input: `hello > -77`},
		{input: `hello > -77.003`},
		{input: `hello > -77.003.123`},
		{input: `NOT (foo OR bar)`},
		{input: `NOT NOT foo`}, //nolint:dupword
		{input: `((foo = bar))`},
		{input: `"AND"`},
		{input: `foo = "hello world" bar`},
		{input: `a OR b AND c OR d`},
		{input: `a..b = c`},
		{input: `hello\"world`},
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
		// {input: `class = [0, 100000] AND subclass = 1001`},
		// {input: `type = [ '0','1' ] and subType = [ '1007' ] and createdAt gte '2024-12-07T00:16:00+05:30' and createdAt lte '2024-12-08T00:16:59+05:30'`},
		// {input: `name ne "Bob"`},
		// {input: `createdAt gt '2024-12-07T00:16:00+05:30' and createdAt lt '2024-12-08T00:16:59+05:30'`},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, trace, err := ParseDebug(tc.input)
			if err != nil {
				t.Log("\n" + trace)
			}
			require.NoError(t, err)

			// snaps.MatchSnapshot(t, litter.Sdump(got))
			snaps.MatchSnapshot(t, got.AsSExpr())
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		input string
	}{
		{input: `~`},
		{input: `=`},
		{input: `NOT`},
		{input: `OR`},
		{input: `OR foo`},
		{input: `(foo`},
		{input: `foo)`},
		{input: `foo=`},
		// a field must be an identifier
		{input: `1field=x`},
		{input: `*foo=x`},
		{input: `"foo"=bar`},
		{input: `and=1`},
		{input: `foo-bar=x`},
		// a value can't be a keyword or a negation
		{input: `foo=and`},
		{input: `foo=-bar`},
		{input: `foo=(bar)`},
		// operators don't chain
		{input: `foo=bar=baz`},
		{input: `a<b<c`},
		{input: `foo:`},
		{input: `foo::bar`},
		// dangling / doubled keywords
		{input: `AND`},
		{input: `a AND`},
		{input: `a OR`},
		{input: `AND a`},
		{input: `a AND AND b`}, //nolint:dupword
		{input: `a OR OR b`},   //nolint:dupword
		{input: `-`},
		{input: `a NOT`},
		// grouping
		{input: `()`},
		{input: `(a))`},
		{input: `((a)`},
		{input: `)`},
		{input: `(`},
		// lexer errors
		{input: `foo!=`},
		{input: `a==b`},
		{input: `"unterminated`},
		{input: "foo\x00bar"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, trace, err := ParseDebug(tc.input)
			if err != nil {
				t.Log("\n" + trace)
			}
			require.Error(t, err)

			snaps.MatchSnapshot(t, err.Error())
		})
	}
}

func TestParseErrorTypes(t *testing.T) {
	var lexErr LexerError

	// lexer failures surface as a LexerError through Parse's wrapping
	_, err := Parse(`foo!bar`)
	require.ErrorAs(t, err, &lexErr)
	require.Equal(t, 4, lexErr.Pos.Column)

	_, err = Parse("foo\x00bar")
	require.ErrorAs(t, err, &lexErr)

	// grammar failures do not
	_, err = Parse(`a AND`)
	require.Error(t, err)
	require.NotErrorAs(t, err, &lexErr)
}

func TestParseEmpty(t *testing.T) {
	for _, input := range []string{"", " ", "\t\r\n"} {
		got, err := Parse(input)
		require.NoError(t, err)
		require.Nil(t, got.LogicExpr)
		require.Empty(t, got.AsSExpr())
	}
}

func TestParseDeepNesting(t *testing.T) {
	input := strings.Repeat("(", 200) + "a" + strings.Repeat(")", 200)

	got, err := Parse(input)
	require.NoError(t, err)
	require.Equal(t, "(value a)", got.AsSExpr())
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

func TestSmoke(t *testing.T) {
	// input := `cat!=dog (fish!=food) bar`
	// input := `(fish!=food)`
	// input := "cat!=dog ( foo>1 AND bar<2 and cat=fish) OR ( baz<=3 AND qux>=4 )"
	// input := " foo>1 AND bar < 2 "
	// input := `hello="world\\fish"`
	input := `obj.prop=world`

	got, trace, err := ParseDebug(input)
	if err != nil || true {
		t.Log("\n" + trace)
	}
	require.NoError(t, err, input)

	snaps.MatchSnapshot(t, litter.Sdump(got))
}
