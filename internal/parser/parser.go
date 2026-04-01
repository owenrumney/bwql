package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/owenrumney/bwql/internal/ast"
	"github.com/owenrumney/bwql/internal/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func Parse(input string) (ast.Statement, error) {
	l := lexer.New(input)
	tokens := l.Tokenize()
	p := New(tokens)
	return p.Parse()
}

func (p *Parser) Parse() (ast.Statement, error) {
	tok := p.current()
	var stmt ast.Statement
	var err error

	switch tok.Type {
	case lexer.TokenSelect:
		stmt, err = p.parseSelect()
	case lexer.TokenInsert:
		stmt, err = p.parseInsert()
	case lexer.TokenUpdate:
		stmt, err = p.parseUpdate()
	case lexer.TokenDelete:
		stmt, err = p.parseDelete()
	default:
		return nil, fmt.Errorf("unexpected token %q at position %d, expected SELECT, INSERT, UPDATE, or DELETE", tok.Literal, tok.Pos)
	}

	if err != nil {
		return nil, err
	}

	if p.current().Type != lexer.TokenEOF {
		return nil, fmt.Errorf("unexpected token %q after statement", p.current().Literal)
	}

	return stmt, nil
}

func (p *Parser) parseSelect() (*ast.SelectStatement, error) {
	p.advance()

	columns, err := p.parseColumns()
	if err != nil {
		return nil, err
	}

	if !p.expectAndAdvance(lexer.TokenFrom) {
		return nil, fmt.Errorf("expected FROM, got %q", p.current().Literal)
	}

	table, err := p.parseTableName()
	if err != nil {
		return nil, err
	}

	stmt := &ast.SelectStatement{
		Columns: columns,
		Table:   table,
	}

	if p.current().Type == lexer.TokenWhere {
		p.advance()
		where, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	if p.current().Type == lexer.TokenOrder {
		p.advance()
		if !p.expectAndAdvance(lexer.TokenBy) {
			return nil, fmt.Errorf("expected BY after ORDER, got %q", p.current().Literal)
		}
		orderBy, err := p.parseOrderBy()
		if err != nil {
			return nil, err
		}
		stmt.OrderBy = orderBy
	}

	if p.current().Type == lexer.TokenLimit {
		p.advance()
		if p.current().Type != lexer.TokenNumber {
			return nil, fmt.Errorf("expected number after LIMIT, got %q", p.current().Literal)
		}
		n, err := strconv.Atoi(p.current().Literal)
		if err != nil {
			return nil, fmt.Errorf("invalid LIMIT value: %s", p.current().Literal)
		}
		stmt.Limit = &n
		p.advance()
	}

	// optional semicolon
	if p.current().Type == lexer.TokenSemicolon {
		p.advance()
	}

	return stmt, nil
}

func (p *Parser) parseColumns() ([]ast.Column, error) {
	var columns []ast.Column

	if p.current().Type == lexer.TokenStar {
		columns = append(columns, ast.Column{Name: "*"})
		p.advance()
		return columns, nil
	}

	for {
		col, err := p.parseColumn()
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)

		if p.current().Type != lexer.TokenComma {
			break
		}
		p.advance()
	}

	return columns, nil
}

func (p *Parser) parseColumn() (ast.Column, error) {
	if p.current().Type != lexer.TokenIdent {
		return ast.Column{}, fmt.Errorf("expected column name, got %q", p.current().Literal)
	}

	col := ast.Column{Name: p.current().Literal}
	p.advance()

	if p.current().Type == lexer.TokenAs {
		p.advance()
		if p.current().Type != lexer.TokenIdent {
			return ast.Column{}, fmt.Errorf("expected alias after AS, got %q", p.current().Literal)
		}
		col.Alias = p.current().Literal
		p.advance()
	}

	return col, nil
}

func (p *Parser) parseTableName() (string, error) {
	if p.current().Type != lexer.TokenIdent {
		return "", fmt.Errorf("expected table name, got %q", p.current().Literal)
	}
	name := strings.ToLower(p.current().Literal)
	p.advance()
	return name, nil
}

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (ast.Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.current().Type == lexer.TokenOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.OrExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseAnd() (ast.Expr, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for p.current().Type == lexer.TokenAnd {
		p.advance()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &ast.AndExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	if p.current().Type == lexer.TokenNot {
		p.advance()
		expr, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ast.NotExpr{Expr: expr}, nil
	}

	if p.current().Type == lexer.TokenLParen {
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.current().Type != lexer.TokenRParen {
			return nil, fmt.Errorf("expected ), got %q", p.current().Literal)
		}
		p.advance()
		return &ast.ParenExpr{Expr: expr}, nil
	}

	if p.current().Type != lexer.TokenIdent {
		return nil, fmt.Errorf("expected column name, got %q", p.current().Literal)
	}

	column := p.current().Literal
	p.advance()

	// IS [NOT] NULL
	if p.current().Type == lexer.TokenIs {
		p.advance()
		not := false
		if p.current().Type == lexer.TokenNot {
			not = true
			p.advance()
		}
		if p.current().Type != lexer.TokenNull {
			return nil, fmt.Errorf("expected NULL after IS%s, got %q", notStr(not), p.current().Literal)
		}
		p.advance()
		return &ast.IsNullExpr{Column: column, Not: not}, nil
	}

	// [NOT] IN (...) / [NOT] LIKE
	if p.current().Type == lexer.TokenNot {
		p.advance()
		switch p.current().Type {
		case lexer.TokenIn:
			p.advance()
			values, err := p.parseValueList()
			if err != nil {
				return nil, err
			}
			return &ast.InExpr{Column: column, Values: values, Not: true}, nil
		case lexer.TokenLike:
			p.advance()
			val, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			return &ast.ComparisonExpr{Left: column, Operator: "NOT LIKE", Right: val}, nil
		default:
			return nil, fmt.Errorf("expected IN or LIKE after NOT, got %q", p.current().Literal)
		}
	}

	if p.current().Type == lexer.TokenIn {
		p.advance()
		values, err := p.parseValueList()
		if err != nil {
			return nil, err
		}
		return &ast.InExpr{Column: column, Values: values, Not: false}, nil
	}

	if p.current().Type == lexer.TokenLike {
		p.advance()
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return &ast.ComparisonExpr{Left: column, Operator: "LIKE", Right: val}, nil
	}

	// Comparison operators
	op, err := p.parseOperator()
	if err != nil {
		return nil, err
	}
	val, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	return &ast.ComparisonExpr{Left: column, Operator: op, Right: val}, nil
}

func (p *Parser) parseOperator() (string, error) {
	tok := p.current()
	switch tok.Type {
	case lexer.TokenEq:
		p.advance()
		return "=", nil
	case lexer.TokenNeq:
		p.advance()
		return "!=", nil
	case lexer.TokenLt:
		p.advance()
		return "<", nil
	case lexer.TokenGt:
		p.advance()
		return ">", nil
	case lexer.TokenLte:
		p.advance()
		return "<=", nil
	case lexer.TokenGte:
		p.advance()
		return ">=", nil
	default:
		return "", fmt.Errorf("expected operator, got %q", tok.Literal)
	}
}

func (p *Parser) parseValue() (ast.Value, error) {
	tok := p.current()
	switch tok.Type {
	case lexer.TokenString:
		p.advance()
		return &ast.StringValue{Val: tok.Literal}, nil
	case lexer.TokenNumber:
		p.advance()
		return &ast.NumberValue{Val: tok.Literal}, nil
	case lexer.TokenTrue:
		p.advance()
		return &ast.BoolValue{Val: true}, nil
	case lexer.TokenFalse:
		p.advance()
		return &ast.BoolValue{Val: false}, nil
	default:
		return nil, fmt.Errorf("expected value, got %q", tok.Literal)
	}
}

func (p *Parser) parseValueList() ([]ast.Value, error) {
	if p.current().Type != lexer.TokenLParen {
		return nil, fmt.Errorf("expected (, got %q", p.current().Literal)
	}
	p.advance()

	var values []ast.Value
	for {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		if p.current().Type != lexer.TokenComma {
			break
		}
		p.advance()
	}

	if p.current().Type != lexer.TokenRParen {
		return nil, fmt.Errorf("expected ), got %q", p.current().Literal)
	}
	p.advance()

	return values, nil
}

func (p *Parser) parseOrderBy() ([]ast.OrderByClause, error) {
	var clauses []ast.OrderByClause
	for {
		if p.current().Type != lexer.TokenIdent {
			return nil, fmt.Errorf("expected column name in ORDER BY, got %q", p.current().Literal)
		}
		clause := ast.OrderByClause{Column: p.current().Literal}
		p.advance()

		if p.current().Type == lexer.TokenAsc {
			p.advance()
		} else if p.current().Type == lexer.TokenDesc {
			clause.Desc = true
			p.advance()
		}

		clauses = append(clauses, clause)

		if p.current().Type != lexer.TokenComma {
			break
		}
		p.advance()
	}
	return clauses, nil
}

// INSERT INTO <table> (col, ...) VALUES (val, ...)
// INSERT INTO <table> VALUES (val, ...)
func (p *Parser) parseInsert() (*ast.InsertStatement, error) {
	p.advance()

	if !p.expectAndAdvance(lexer.TokenInto) {
		return nil, fmt.Errorf("expected INTO after INSERT, got %q", p.current().Literal)
	}

	table, err := p.parseTableName()
	if err != nil {
		return nil, err
	}

	stmt := &ast.InsertStatement{Table: table}

	// Optional column list
	if p.current().Type == lexer.TokenLParen {
		p.advance()
		for {
			if p.current().Type != lexer.TokenIdent {
				return nil, fmt.Errorf("expected column name, got %q", p.current().Literal)
			}
			stmt.Columns = append(stmt.Columns, p.current().Literal)
			p.advance()
			if p.current().Type != lexer.TokenComma {
				break
			}
			p.advance()
		}
		if p.current().Type != lexer.TokenRParen {
			return nil, fmt.Errorf("expected ), got %q", p.current().Literal)
		}
		p.advance()
	}

	if !p.expectAndAdvance(lexer.TokenValues) {
		return nil, fmt.Errorf("expected VALUES, got %q", p.current().Literal)
	}

	values, err := p.parseValueList()
	if err != nil {
		return nil, err
	}
	stmt.Values = values

	if p.current().Type == lexer.TokenSemicolon {
		p.advance()
	}

	return stmt, nil
}

// UPDATE <table> SET col = val, ... [WHERE ...]
func (p *Parser) parseUpdate() (*ast.UpdateStatement, error) {
	p.advance()

	table, err := p.parseTableName()
	if err != nil {
		return nil, err
	}

	if !p.expectAndAdvance(lexer.TokenSet) {
		return nil, fmt.Errorf("expected SET, got %q", p.current().Literal)
	}

	stmt := &ast.UpdateStatement{Table: table}

	for {
		if p.current().Type != lexer.TokenIdent {
			return nil, fmt.Errorf("expected column name in SET, got %q", p.current().Literal)
		}
		col := p.current().Literal
		p.advance()

		if !p.expectAndAdvance(lexer.TokenEq) {
			return nil, fmt.Errorf("expected = after column name, got %q", p.current().Literal)
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		stmt.Set = append(stmt.Set, ast.SetClause{Column: col, Value: val})

		if p.current().Type != lexer.TokenComma {
			break
		}
		p.advance()
	}

	if p.current().Type == lexer.TokenWhere {
		p.advance()
		where, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	if p.current().Type == lexer.TokenSemicolon {
		p.advance()
	}

	return stmt, nil
}

// DELETE FROM <table> [WHERE ...]
func (p *Parser) parseDelete() (*ast.DeleteStatement, error) {
	p.advance()

	if !p.expectAndAdvance(lexer.TokenFrom) {
		return nil, fmt.Errorf("expected FROM after DELETE, got %q", p.current().Literal)
	}

	table, err := p.parseTableName()
	if err != nil {
		return nil, err
	}

	stmt := &ast.DeleteStatement{Table: table}

	if p.current().Type == lexer.TokenWhere {
		p.advance()
		where, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	if p.current().Type == lexer.TokenSemicolon {
		p.advance()
	}

	return stmt, nil
}

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	p.pos++
}

func (p *Parser) expectAndAdvance(t lexer.TokenType) bool {
	if p.current().Type != t {
		return false
	}
	p.advance()
	return true
}

func notStr(not bool) string {
	if not {
		return " NOT"
	}
	return ""
}
