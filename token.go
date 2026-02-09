package quickql

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
)

const (
	EOF = lexer.EOF
)

const (
	ILLEGAL lexer.TokenType = iota

	// WS = Whitespace.
	WS

	// LPAREN is "(".
	LPAREN
	// RPAREN is ")".
	RPAREN

	// STRING is a quoted string.
	STRING
	// IDENT is non-quoted value.
	IDENT
	// VALUE is a numeric constant e.g. 7, -7, -77.77.
	VALUE
	// OP is any operator (e.g. =, ~, !=).
	OP
	// KEYWORD is a reserved word (AND, OR, NOT).
	KEYWORD
)

func Symbols() map[string]lexer.TokenType {
	lexTokens := map[string]lexer.TokenType{
		"ILLEGAL": ILLEGAL,
		"EOF":     EOF,
		"WS":      WS,
		"LPAREN":  LPAREN,
		"RPAREN":  RPAREN,
		"STRING":  STRING,
		"IDENT":   IDENT,
		"VALUE":   VALUE,
		"KEYWORD": KEYWORD,
	}

	return lexTokens
}

func TokenName(token lexer.Token) string {
	switch token.Type {
	case EOF:
		return "EOF"
	case WS:
		return "WS"
	case RPAREN:
		return "RPAREN"
	case LPAREN:
		return "LPAREN"
	case ILLEGAL:
		return "ILLEGAL"
	case STRING:
		return "STRING"
	case IDENT:
		return "IDENT"
	case VALUE:
		return "VALUE"
	case OP:
		return "OP"
	case KEYWORD:
		return "KEYWORD"
	}

	panic("unknown token type")
}

func DebugOutput(lexTokens []lexer.Token) []string {
	output := make([]string, len(lexTokens))
	for i, token := range lexTokens {
		output[i] = fmt.Sprintf("Token{%s, %q}", TokenName(token), token.Value)
	}

	return output
}
