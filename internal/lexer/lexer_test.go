package lexer

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []TokenType
	}{
		{
			name:  "simple select",
			input: "SELECT * FROM logins",
			expect: []TokenType{
				TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenEOF,
			},
		},
		{
			name:  "select with columns",
			input: "SELECT name, username FROM logins",
			expect: []TokenType{
				TokenSelect, TokenIdent, TokenComma, TokenIdent, TokenFrom, TokenIdent, TokenEOF,
			},
		},
		{
			name:  "where with IS NULL",
			input: "SELECT * FROM logins WHERE totp IS NULL",
			expect: []TokenType{
				TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere,
				TokenIdent, TokenIs, TokenNull, TokenEOF,
			},
		},
		{
			name:  "where with comparison",
			input: "SELECT * FROM logins WHERE password_age > 365",
			expect: []TokenType{
				TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere,
				TokenIdent, TokenGt, TokenNumber, TokenEOF,
			},
		},
		{
			name:  "where with string",
			input: "SELECT * FROM logins WHERE name LIKE '%github%'",
			expect: []TokenType{
				TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere,
				TokenIdent, TokenLike, TokenString, TokenEOF,
			},
		},
		{
			name:  "order by desc",
			input: "SELECT * FROM logins ORDER BY password_age DESC",
			expect: []TokenType{
				TokenSelect, TokenStar, TokenFrom, TokenIdent,
				TokenOrder, TokenBy, TokenIdent, TokenDesc, TokenEOF,
			},
		},
		{
			name:  "operators",
			input: "a = 1 AND b != 2 AND c <> 3 AND d <= 4 AND e >= 5",
			expect: []TokenType{
				TokenIdent, TokenEq, TokenNumber, TokenAnd,
				TokenIdent, TokenNeq, TokenNumber, TokenAnd,
				TokenIdent, TokenNeq, TokenNumber, TokenAnd,
				TokenIdent, TokenLte, TokenNumber, TokenAnd,
				TokenIdent, TokenGte, TokenNumber, TokenEOF,
			},
		},
		{
			name:  "with semicolon",
			input: "SELECT * FROM logins;",
			expect: []TokenType{
				TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenSemicolon, TokenEOF,
			},
		},
		{
			name:  "case insensitive keywords",
			input: "select * from logins where name is null",
			expect: []TokenType{
				TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere,
				TokenIdent, TokenIs, TokenNull, TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()

			if len(tokens) != len(tt.expect) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expect), len(tokens))
			}

			for i, tok := range tokens {
				if tok.Type != tt.expect[i] {
					t.Errorf("token[%d]: expected type %d, got %d (literal: %q)", i, tt.expect[i], tok.Type, tok.Literal)
				}
			}
		})
	}
}

func TestStringLiteral(t *testing.T) {
	l := New("'hello world'")
	tok := l.NextToken()
	if tok.Type != TokenString {
		t.Fatalf("expected TokenString, got %d", tok.Type)
	}
	if tok.Literal != "hello world" {
		t.Fatalf("expected 'hello world', got %q", tok.Literal)
	}
}

func TestUnterminatedString(t *testing.T) {
	l := New("'hello")
	tok := l.NextToken()
	if tok.Type != TokenIllegal {
		t.Fatalf("expected TokenIllegal for unterminated string, got %d", tok.Type)
	}
}

func TestEscapedString(t *testing.T) {
	l := New(`'it\'s here'`)
	tok := l.NextToken()
	if tok.Type != TokenString {
		t.Fatalf("expected TokenString, got %d", tok.Type)
	}
	if tok.Literal != "it's here" {
		t.Fatalf("expected \"it's here\", got %q", tok.Literal)
	}
}

func TestEmptyInput(t *testing.T) {
	l := New("")
	tok := l.NextToken()
	if tok.Type != TokenEOF {
		t.Fatalf("expected TokenEOF, got %d", tok.Type)
	}
}

func TestIllegalCharacter(t *testing.T) {
	l := New("@")
	tok := l.NextToken()
	if tok.Type != TokenIllegal {
		t.Fatalf("expected TokenIllegal, got %d", tok.Type)
	}
}
