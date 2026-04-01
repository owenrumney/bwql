package lexer

import (
	"strings"
	"unicode"
)

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Pos = l.pos

	switch l.ch {
	case '*':
		tok = Token{Type: TokenStar, Literal: "*", Pos: l.pos}
	case ',':
		tok = Token{Type: TokenComma, Literal: ",", Pos: l.pos}
	case '(':
		tok = Token{Type: TokenLParen, Literal: "(", Pos: l.pos}
	case ')':
		tok = Token{Type: TokenRParen, Literal: ")", Pos: l.pos}
	case '=':
		tok = Token{Type: TokenEq, Literal: "=", Pos: l.pos}
	case ';':
		tok = Token{Type: TokenSemicolon, Literal: ";", Pos: l.pos}
	case '<':
		switch l.peekChar() {
		case '=':
			l.readChar()
			tok = Token{Type: TokenLte, Literal: "<=", Pos: l.pos - 1}
		case '>':
			l.readChar()
			tok = Token{Type: TokenNeq, Literal: "<>", Pos: l.pos - 1}
		default:
			tok = Token{Type: TokenLt, Literal: "<", Pos: l.pos}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: TokenGte, Literal: ">=", Pos: l.pos - 1}
		} else {
			tok = Token{Type: TokenGt, Literal: ">", Pos: l.pos}
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: TokenNeq, Literal: "!=", Pos: l.pos - 1}
		} else {
			tok = Token{Type: TokenIllegal, Literal: string(l.ch), Pos: l.pos}
		}
	case '\'':
		startPos := l.pos
		str, terminated := l.readString()
		if !terminated {
			return Token{Type: TokenIllegal, Literal: "unterminated string", Pos: startPos}
		}
		tok = Token{Type: TokenString, Literal: str, Pos: startPos}
		return tok
	case 0:
		tok = Token{Type: TokenEOF, Literal: "", Pos: l.pos}
		return tok
	default:
		if isLetter(l.ch) {
			startPos := l.pos
			ident := l.readIdentifier()
			upper := strings.ToUpper(ident)
			tokType := LookupKeyword(upper)
			return Token{Type: tokType, Literal: ident, Pos: startPos}
		}
		if isDigit(l.ch) {
			startPos := l.pos
			num := l.readNumber()
			return Token{Type: TokenNumber, Literal: num, Pos: startPos}
		}
		tok = Token{Type: TokenIllegal, Literal: string(l.ch), Pos: l.pos}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() string {
	pos := l.pos
	dots := 0
	for isDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
			dots++
		}
		l.readChar()
	}
	num := l.input[pos:l.pos]
	if dots > 1 || num == "." || num[len(num)-1] == '.' {
		// Malformed, but return what we consumed — parser will handle it
		return num
	}
	return num
}

func (l *Lexer) readString() (string, bool) {
	l.readChar()
	var b strings.Builder
	for l.ch != '\'' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			if l.ch == 0 {
				return b.String(), false
			}
		}
		b.WriteByte(l.ch)
		l.readChar()
	}
	if l.ch == 0 {
		return b.String(), false
	}
	l.readChar()
	return b.String(), true
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
