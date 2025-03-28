package quickql

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
	}{
		{input: ``},
		{input: `one two three`},
		{input: `François`},
		{input: `one 🇨🇦 three`},
		{input: `hello:world`},
		// basic operators
		{input: `hello=world`},
		{input: `country=🇨🇦`},
		{input: `hello = world`},
		{input: `hello!=world`},
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
		// {input: `foo:*bar*`},
		// {input: `foo:bar*`},
		{input: `hello:world OR super:cool`},
		{input: `hello:world or cool`},
		{input: `world OR super:cool`},
		{input: `hello:66e839032fb119d31dc9c968`},
		{input: `hello:"super cool"`},
		// {input: `-hello:77`},
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

			snaps.MatchSnapshot(t, litter.Sdump(got))
		})
	}
}

func TestSmoke(t *testing.T) {
	// input := `cat!=dog (fish!=food) bar`
	// input := `(fish!=food)`
	// input := "cat!=dog ( foo>1 AND bar<2 and cat=fish) OR ( baz<=3 AND qux>=4 )"
	// input := " foo>1 AND bar < 2 "
	// input := `hello="world\\fish"`
	input := `bar:*back`

	got, trace, err := ParseDebug(input)
	if err != nil {
		t.Log("\n" + trace)
	}
	require.NoError(t, err, input)

	snaps.MatchSnapshot(t, litter.Sdump(got))
}
