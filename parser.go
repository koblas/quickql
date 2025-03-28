package quickql

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
)

type Expr struct {
	LogicExpr *OrLogicExpr `parser:"@@"`
}

type OrLogicExpr struct {
	Expr []*AndLogicExpr `parser:"@@ (('OR'|'or') @@)*"`
}

type AndLogicExpr struct {
	Expr []primary `parser:"@@ ((('AND'|'and') @@) | @@)*"`
}

type primary interface{ isPrimary() }

type PNot struct {
	primary
	Expr primary `parser:"('NOT'|'not') @@"`
}

type PParen struct {
	primary
	Expr *Expr `parser:"'(' @@ ')'"`
}

type PExpr struct {
	primary
	// FieldExpr           <- field:Identifier _ op:CmpOp _ value:Value             { return parseFieldExpression(field, op, value) }
	Field *Identifier `parser:"@@"`
	Op    string      `parser:"@('<=' | '<' | '>' | '>=' | ':' | '=' | '~' | '!=' | '!~')"`
	Value *Value      `parser:"@@"`
}

type PValue struct {
	primary
	// BoolFieldExpr       <- field:Identifier                                      { return parseBoolFieldExpr(field) }
	Expr *Value `parser:"@@"`
}

type Value struct {
	VString     string      `parser:"(@STRING"`
	VValue      []string    `parser:"| @VALUE"`
	VIdentifier *Identifier `parser:"| @@ )"`
}

type EValue struct {
	Value []string `parser:"@AlphaNumeric | (@Char+ @AlphaNumeric?)"`
}

type Identifier struct {
	// Identifier          <- AlphaNumeric ("." AlphaNumeric)*                      { return Identifier(c.text), nil }
	Value []string `parser:"@IDENT ('.' @IDENT)*"`
}

var parser = participle.MustBuild[Expr](
	participle.Lexer(queryLexer),
	participle.UseLookahead(2),
	participle.Union[primary](PParen{}, PNot{}, PExpr{}, PValue{}),
)

func Parse(q string) (Expr, error) {
	if q == "" {
		return Expr{}, nil
	}

	expr, err := parser.ParseString("", q)
	if err != nil {
		return Expr{}, fmt.Errorf("parse string: %w", err)
	}

	return *expr, nil
}

func ParseDebug(q string) (Expr, string, error) {
	if q == "" {
		return Expr{}, "", nil
	}

	var buf strings.Builder
	expr, err := parser.ParseString("", q, participle.Trace(&buf))
	if err != nil {
		return Expr{}, buf.String(), err
	}

	return *expr, buf.String(), nil
}
