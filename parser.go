package quickql

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
)

/*
type Expr struct {
	// Expr                <- _ e:OrExpr _                                          { return e, nil }
	Expr *OrExpr `parser:"WS? @@ WS?"`
}

type OrExpr struct {
	// OrExpr              <- left:AndExpr rest:(_ ( OrOp ) _ AndExpr)*             { return parseBooleanExpression(left, rest) }
	Expr []*AndExpr `parser:"@@ (WS ('OR'|'or') WS @@)*"`
	// Expr *AndExpr `parser:"@@"`
}

type AndExpr struct {
	// AndExpr             <- left:NotExpr rest:(_ ( op:AndOp ) _ NotExpr)*         { return parseBooleanExpression(left, rest) }
	Expr []primary `parser:"@@ (WS (('AND'|'and') WS)? @@)*"`
}

type primary interface{ isPrimary() }

type PNot struct {
	primary
	Expr primary `parser:"('NOT'|'not') WS @@"`
}

type PParen struct {
	primary
	// ParenExpr           <- '(' _ expr:Expr _ ')'                                 { return expr.(Expr), nil }
	Expr *Expr `parser:"'(' @@ ')'"`
}

type PField struct {
	primary
	// FieldExpr           <- field:Identifier _ op:CmpOp _ value:Value             { return parseFieldExpression(field, op, value) }
	Field *Identifier `parser:"@@"`
	Op    string      `parser:"WS? @('<=' | '<' | '>' | '>=' | ':' | '=' | '~' | '!=' | '!~') WS?"`
	Value *Value      `parser:"@@"`
}

type PValue struct {
	primary
	// BoolFieldExpr       <- field:Identifier                                      { return parseBoolFieldExpr(field) }
	Expr *Value `parser:"@@"`
}

type Value struct {
	// Value               <- OneOfExpr / String / Number / Boolean / Identifier
	VString     *EString    `parser:"(@@"`
	VValue      string      `parser:"| @Number @AlphaNumeric?"`
	VIdentifier *Identifier `parser:"| @@ )"`
}
	/*/

// type ENumber struct {
// 	Value string `parser:"@Number (!?AlphaNumeric)"`
// }

// func (obj *ENumber) Parse(lex *lexer.PeekingLexer) error {
// token := lex.Peek()

// fmt.Println(lex.Peek())

// return participle.NextMatch
// }

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
