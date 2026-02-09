package quickql

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/alecthomas/participle/v2/lexer"
)

type BaseLexer struct{}

type Position struct {
	Line   int
	Column int
}

type Lexer struct {
	filename string
	reader   *bufio.Reader
	pos      Position
	nextPos  Position
	ch       rune
}

var (
	_ lexer.Definition = (*BaseLexer)(nil)
	_ lexer.Lexer      = (*Lexer)(nil)
)

// LexerError is the general error type for the lexer.
type LexerError struct {
	Pos lexer.Position
	Msg string
}

func (e LexerError) Error() string {
	var position string
	if e.Pos.Filename != "" {
		position = fmt.Sprintf("%s:%d:%d", e.Pos.Filename, e.Pos.Line, e.Pos.Column)
	} else {
		position = fmt.Sprintf("%d:%d", e.Pos.Line, e.Pos.Column)
	}

	return fmt.Sprintf("[%s] %s", position, e.Msg)
}

var queryLexer = NewLexer()

func NewLexer() *BaseLexer {
	return &BaseLexer{}
}

func (base *BaseLexer) Symbols() map[string]lexer.TokenType {
	return Symbols()
}

// Lex returns the interface type.
func (base *BaseLexer) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	l := &Lexer{
		filename: filename,
		reader:   bufio.NewReader(r),
		nextPos:  Position{Line: 1, Column: 1},
	}
	// load the first character
	l.next()

	return l, nil
}

// Next converts to participle's lexer interface.
func (l *Lexer) Next() (lexer.Token, error) {
	p, t, value := l.Scan()
	pos := lexer.Position{
		Filename: l.filename,
		Line:     p.Line,
		Column:   p.Column,
	}

	switch t {
	case ILLEGAL:
		return lexer.Token{}, LexerError{Pos: pos, Msg: value}
	case EOF:
		return lexer.EOFToken(pos), nil
	}

	return lexer.Token{
		Type:  t,
		Value: value,
		Pos: lexer.Position{
			Filename: l.filename,
			Line:     p.Line,
			Column:   p.Column,
		},
	}, nil
}

func (l *Lexer) Scan() (Position, lexer.TokenType, string) {
	return l.scan()
}

// Read an string value and decide if it's a VALUE or IDENTIFIER
// the reason for differntiation is that "*foo < -12" (VALUE LT VALUE) is not valid
// but "foo < -12" (IDENT LT VALUE) is valid and this is a good way to discrimitate
// on the leading keyword.
func (l *Lexer) readIdentifer(ch rune, buffer []rune) (string, lexer.TokenType) {
	buffer = append(buffer, ch)
	tok := IDENT
	if isDigit(ch) {
		tok = VALUE
	}

	ch = l.peek()

	for l.ch != 0 && !isWhitespace(l.ch) && !isOperator(l.ch) {
		buffer = append(buffer, ch)
		if !isAlpha(ch) && ch != '.' {
			tok = VALUE
		}

		l.next()
		ch = l.peek()
	}

	val := string(buffer)
	if tok == IDENT {
		switch val {
		case "AND", "and", "OR", "or", "NOT", "not":
			tok = KEYWORD
		}
	}

	return val, tok
}

// Scan the input and return the next token all value tokens should be normalized
// to a standard keyword (e.g. `:` is `=` and AND will always be uppercase "AND").
func (l *Lexer) scan() (Position, lexer.TokenType, string) {
	// skip whitespace
	for isWhitespace(l.ch) {
		l.next()
	}

	pos := l.pos

	if l.ch == 0 {
		return pos, EOF, ""
	}

	var tok lexer.TokenType
	var val string

	switch ch := l.next(); ch {
	// handle string literals
	case '\'', '"':
		tok, val = l.parseString(ch)

	// grouping
	case '(':
		val = "("
		tok = LPAREN
	case ')':
		val = ")"
		tok = RPAREN

	// // IN operations
	// case '[':
	// 	val = "["
	// case ']':
	// 	val = "]"
	// case ',':
	// 	val = ","

	// handle operators
	case '~':
		val = "~"
		tok = OP
	case '!':
		switch l.ch {
		case '=':
			l.next()
			val = "!="
			tok = OP
		case '~':
			l.next()
			val = "!~"
			tok = OP
		default:
			tok = ILLEGAL
			val = "operator ! must be followed by = or ~"
		}
	case ':':
		// this is from the `key:value` syntax, convert to a regular =
		val = "="
		tok = OP
	case '=':
		if l.consumeIf('=') {
			tok = ILLEGAL
			val = "double == not allowed"
		} else {
			val = "="
			tok = OP
		}
	case '<':
		if l.consumeIf('=') {
			val = "<="
		} else {
			val = "<"
		}
		tok = OP
	case '>':
		if l.consumeIf('=') {
			val = ">="
		} else {
			val = ">"
		}
		tok = OP

	// `-` can be either a NOT or a negative number
	case '-':
		if isDigit(l.peek()) {
			chars := make([]rune, 0, 32)
			val, tok = l.readIdentifer(ch, chars)
		} else {
			val = "NOT"
			tok = KEYWORD
		}
	// pretty much everything else is an identifier or value
	default:
		// most identifiers are small, so pre-allocate a bit of space
		chars := make([]rune, 0, 32)

		val, tok = l.readIdentifer(ch, chars)
	}

	return pos, tok, val
}

// peek at the next character to be consumed.
func (l *Lexer) peek() rune {
	return l.ch
}

// consume the next character, return 0 on EOF.
func (l *Lexer) next() rune {
	l.pos = l.nextPos

	cur := l.ch

	ch, _, err := l.reader.ReadRune()
	if errors.Is(err, io.EOF) {
		if l.ch != 0 {
			l.ch = 0
			l.nextPos.Column++
		}

		return cur
	}

	if ch == '\n' {
		l.nextPos.Line++
		l.nextPos.Column = 1
	} else if ch != '\r' {
		l.nextPos.Column++
	}

	l.ch = ch

	return cur
}

// helper to handle differentiation of > and >=.
func (l *Lexer) consumeIf(ch rune) bool {
	if l.peek() == ch {
		l.next()

		return true
	}

	return false
}

// quote is the matching quote e.g. ' or ".
func (l *Lexer) parseString(quote rune) (lexer.TokenType, string) {
	// strings tend to be short
	chars := make([]rune, 0, 32)
	for ch := l.next(); ch != quote; ch = l.next() {
		switch ch {
		case '\\':
			// consume the quote, now grab the real value
			ch = l.next()
			if ch == 0 {
				return ILLEGAL, "no character after escape"
			}
			chars = append(chars, ch)
		case 0:
			return ILLEGAL, "unterminated string"
		default:
			chars = append(chars, ch)
		}
	}

	return STRING, string(chars)
}

// basic digit check.
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// handle the operators.
func isOperator(ch rune) bool {
	return ch == '=' || ch == '<' || ch == '>' || ch == '!' || ch == ':' || ch == '~' || ch == '(' || ch == ')'
}
