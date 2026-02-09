package quickql

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
)

type Expr struct {
	LogicExpr *AndLogicExpr `parser:"@@ | EOF"`
}

func (obj *Expr) AsSExpr() string {
	if obj.LogicExpr == nil {
		return ""
	}

	return obj.LogicExpr.asSExpr()
}

type OrLogicExpr struct {
	Expr []primary `parser:"@@ (('OR'|'or') @@)*"`
}

func (obj *OrLogicExpr) asSExpr() string {
	if len(obj.Expr) == 1 {
		return obj.Expr[0].asSExpr()
	}

	var exprs []string
	for _, e := range obj.Expr {
		exprs = append(exprs, e.asSExpr())
	}

	return fmt.Sprintf("(or %s)", strings.Join(exprs, " "))
}

type AndLogicExpr struct {
	Expr []OrLogicExpr `parser:"@@ ((('AND'|'and') @@) | @@)*"`
	// Expr []OrLogicExpr `parser:"@@ (('AND'|'and') @@)*"`
}

func (obj *AndLogicExpr) asSExpr() string {
	if len(obj.Expr) == 1 {
		return obj.Expr[0].asSExpr()
	}

	var exprs []string
	for _, e := range obj.Expr {
		exprs = append(exprs, e.asSExpr())
	}

	return fmt.Sprintf("(and %s)", strings.Join(exprs, " "))
}

type primary interface{ asSExpr() string }

type PNot struct {
	Expr primary `parser:"('NOT'|'not') @@"`
}

func (obj PNot) asSExpr() string {
	return fmt.Sprintf("(NOT %s)", obj.Expr.asSExpr())
}

type PParen struct {
	Expr *Expr `parser:"'(' @@ ')'"`
}

func (obj PParen) asSExpr() string {
	return obj.Expr.AsSExpr()
}

type PExpr struct {
	Field Identifier `parser:"@@"`
	Op    string     `parser:"@('<=' | '<' | '>' | '>=' | '=' | '~' | '!=' | '!~')"`
	Value Value      `parser:"@@"`
}

func (obj PExpr) asSExpr() string {
	return fmt.Sprintf("(%s %s %q)", obj.Op, obj.Field.String(), obj.Value.String())
}

type PValue struct {
	Expr *Value `parser:"@@"`
}

func (obj PValue) asSExpr() string {
	return fmt.Sprintf("(value %s)", obj.Expr.String())
}

type Value struct {
	VString     *string     `parser:"(@STRING"`
	VValue      *string     `parser:"| @VALUE"`
	VIdentifier *Identifier `parser:"| @@ )"`
}

func (obj *Value) String() string {
	if obj.VString != nil {
		return *obj.VString
	}
	if obj.VValue != nil {
		return *obj.VValue
	}
	if obj.VIdentifier != nil {
		return obj.VIdentifier.String()
	}

	panic("unknown value")
}

type Identifier struct {
	Value string `parser:"@IDENT"`
}

func (obj *Identifier) String() string {
	return obj.Value
}

//
//

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
