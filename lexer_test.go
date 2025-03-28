package quickql

import (
	"strings"
	"testing"

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
	}

	queryLexer := NewLexer()

	symbols := queryLexer.Symbols()
	names := map[lexer.TokenType]string{}
	for k, v := range symbols {
		names[v] = k
	}

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
	}

	queryLexer := NewLexer()

	symbols := queryLexer.Symbols()
	names := map[lexer.TokenType]string{}
	for k, v := range symbols {
		names[v] = k
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			l, err := queryLexer.Lex("", strings.NewReader(c.input))
			require.NoError(t, err)

			_, err = lexer.ConsumeAll(l)
			require.Error(t, err)
		})
	}
}

func TestSimple(t *testing.T) {
	input := `(fish!=food)`
	// input := `-hello:world`
	l, err := queryLexer.Lex("", strings.NewReader(input))
	require.NoError(t, err)

	tokens, err := lexer.ConsumeAll(l)
	require.NoError(t, err)
	require.NotEmpty(t, tokens)
}
