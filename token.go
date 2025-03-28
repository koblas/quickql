package quickql

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
)

const (
	EOF                     = lexer.EOF
	ILLEGAL lexer.TokenType = iota

	// Whitespace.
	WS

	// Operators.
	LPAREN
	RPAREN

	STRING
	IDENT
	VALUE
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
