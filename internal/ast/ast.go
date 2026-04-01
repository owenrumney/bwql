package ast

type Statement interface {
	statementNode()
}

type SelectStatement struct {
	Columns  []Column
	Table    string
	Where    Expr
	OrderBy  []OrderByClause
	Limit    *int
}

func (s *SelectStatement) statementNode() {}

type InsertStatement struct {
	Table   string
	Columns []string
	Values  []Value
}

func (s *InsertStatement) statementNode() {}

type UpdateStatement struct {
	Table string
	Set   []SetClause
	Where Expr
}

func (s *UpdateStatement) statementNode() {}

type SetClause struct {
	Column string
	Value  Value
}

type DeleteStatement struct {
	Table string
	Where Expr
}

func (s *DeleteStatement) statementNode() {}

type Column struct {
	Name  string // "*" for select all
	Alias string
}

type OrderByClause struct {
	Column string
	Desc   bool
}

// Expressions for WHERE clauses

type Expr interface {
	exprNode()
}

type ComparisonExpr struct {
	Left     string
	Operator string // =, !=, <, >, <=, >=, LIKE
	Right    Value
}

func (e *ComparisonExpr) exprNode() {}

type IsNullExpr struct {
	Column string
	Not    bool // IS NOT NULL
}

func (e *IsNullExpr) exprNode() {}

type AndExpr struct {
	Left  Expr
	Right Expr
}

func (e *AndExpr) exprNode() {}

type OrExpr struct {
	Left  Expr
	Right Expr
}

func (e *OrExpr) exprNode() {}

type NotExpr struct {
	Expr Expr
}

func (e *NotExpr) exprNode() {}

type InExpr struct {
	Column string
	Values []Value
	Not    bool
}

func (e *InExpr) exprNode() {}

type ParenExpr struct {
	Expr Expr
}

func (e *ParenExpr) exprNode() {}

// Values

type Value interface {
	valueNode()
	String() string
}

type StringValue struct {
	Val string
}

func (v *StringValue) valueNode()      {}
func (v *StringValue) String() string { return v.Val }

type NumberValue struct {
	Val string
}

func (v *NumberValue) valueNode()      {}
func (v *NumberValue) String() string { return v.Val }

type BoolValue struct {
	Val bool
}

func (v *BoolValue) valueNode() {}
func (v *BoolValue) String() string {
	if v.Val {
		return "true"
	}
	return "false"
}
