package parser

import (
	"testing"

	"github.com/owenrumney/bwql/internal/ast"
)

func TestParseSelectStar(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins")
	if err != nil {
		t.Fatal(err)
	}
	sel, ok := stmt.(*ast.SelectStatement)
	if !ok {
		t.Fatal("expected SelectStatement")
	}
	if sel.Table != "logins" {
		t.Errorf("expected table 'logins', got %q", sel.Table)
	}
	if len(sel.Columns) != 1 || sel.Columns[0].Name != "*" {
		t.Errorf("expected single * column, got %v", sel.Columns)
	}
}

func TestParseSelectColumns(t *testing.T) {
	stmt, err := Parse("SELECT name, username, uri FROM logins")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	if len(sel.Columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(sel.Columns))
	}
	expected := []string{"name", "username", "uri"}
	for i, col := range sel.Columns {
		if col.Name != expected[i] {
			t.Errorf("column[%d]: expected %q, got %q", i, expected[i], col.Name)
		}
	}
}

func TestParseWhereIsNull(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE totp IS NULL")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	isNull, ok := sel.Where.(*ast.IsNullExpr)
	if !ok {
		t.Fatal("expected IsNullExpr")
	}
	if isNull.Column != "totp" {
		t.Errorf("expected column 'totp', got %q", isNull.Column)
	}
	if isNull.Not {
		t.Error("expected Not=false")
	}
}

func TestParseWhereIsNotNull(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE totp IS NOT NULL")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	isNull, ok := sel.Where.(*ast.IsNullExpr)
	if !ok {
		t.Fatal("expected IsNullExpr")
	}
	if !isNull.Not {
		t.Error("expected Not=true")
	}
}

func TestParseWhereComparison(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE password_age > 365")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	cmp, ok := sel.Where.(*ast.ComparisonExpr)
	if !ok {
		t.Fatal("expected ComparisonExpr")
	}
	if cmp.Left != "password_age" {
		t.Errorf("expected left 'password_age', got %q", cmp.Left)
	}
	if cmp.Operator != ">" {
		t.Errorf("expected operator '>', got %q", cmp.Operator)
	}
	if cmp.Right.String() != "365" {
		t.Errorf("expected right '365', got %q", cmp.Right.String())
	}
}

func TestParseWhereLike(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE name LIKE '%github%'")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	cmp, ok := sel.Where.(*ast.ComparisonExpr)
	if !ok {
		t.Fatal("expected ComparisonExpr")
	}
	if cmp.Operator != "LIKE" {
		t.Errorf("expected LIKE, got %q", cmp.Operator)
	}
	if cmp.Right.String() != "%github%" {
		t.Errorf("expected '%%github%%', got %q", cmp.Right.String())
	}
}

func TestParseWhereAnd(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE totp IS NULL AND password_age > 365")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	and, ok := sel.Where.(*ast.AndExpr)
	if !ok {
		t.Fatal("expected AndExpr")
	}
	if _, ok := and.Left.(*ast.IsNullExpr); !ok {
		t.Error("expected left to be IsNullExpr")
	}
	if _, ok := and.Right.(*ast.ComparisonExpr); !ok {
		t.Error("expected right to be ComparisonExpr")
	}
}

func TestParseOrderBy(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins ORDER BY password_age DESC")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	if len(sel.OrderBy) != 1 {
		t.Fatalf("expected 1 order by clause, got %d", len(sel.OrderBy))
	}
	if sel.OrderBy[0].Column != "password_age" {
		t.Errorf("expected column 'password_age', got %q", sel.OrderBy[0].Column)
	}
	if !sel.OrderBy[0].Desc {
		t.Error("expected Desc=true")
	}
}

func TestParseLimit(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins LIMIT 10")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	if sel.Limit == nil {
		t.Fatal("expected Limit to be set")
	}
	if *sel.Limit != 10 {
		t.Errorf("expected Limit=10, got %d", *sel.Limit)
	}
}

func TestParseSemicolon(t *testing.T) {
	_, err := Parse("SELECT * FROM logins;")
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseCaseInsensitive(t *testing.T) {
	_, err := Parse("select * from logins where totp is null")
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseWhereIn(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE folder IN ('Work', 'Personal')")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	in, ok := sel.Where.(*ast.InExpr)
	if !ok {
		t.Fatal("expected InExpr")
	}
	if in.Column != "folder" {
		t.Errorf("expected column 'folder', got %q", in.Column)
	}
	if len(in.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(in.Values))
	}
}

func TestParseWhereNotLike(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE name NOT LIKE '%test%'")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	cmp, ok := sel.Where.(*ast.ComparisonExpr)
	if !ok {
		t.Fatal("expected ComparisonExpr")
	}
	if cmp.Operator != "NOT LIKE" {
		t.Errorf("expected NOT LIKE, got %q", cmp.Operator)
	}
}

func TestParseWhereNotIn(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE folder NOT IN ('Work')")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	in, ok := sel.Where.(*ast.InExpr)
	if !ok {
		t.Fatal("expected InExpr")
	}
	if !in.Not {
		t.Error("expected Not=true")
	}
}

func TestParseWhereOr(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE name = 'a' OR name = 'b'")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	_, ok := sel.Where.(*ast.OrExpr)
	if !ok {
		t.Fatal("expected OrExpr")
	}
}

func TestParseWhereParens(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins WHERE (name = 'a' OR name = 'b') AND totp IS NULL")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	and, ok := sel.Where.(*ast.AndExpr)
	if !ok {
		t.Fatal("expected AndExpr")
	}
	_, ok = and.Left.(*ast.ParenExpr)
	if !ok {
		t.Fatal("expected left to be ParenExpr")
	}
}

func TestParseInsert(t *testing.T) {
	stmt, err := Parse("INSERT INTO folders (name) VALUES ('Work')")
	if err != nil {
		t.Fatal(err)
	}
	ins, ok := stmt.(*ast.InsertStatement)
	if !ok {
		t.Fatal("expected InsertStatement")
	}
	if ins.Table != "folders" {
		t.Errorf("expected table 'folders', got %q", ins.Table)
	}
	if len(ins.Columns) != 1 || ins.Columns[0] != "name" {
		t.Errorf("expected columns [name], got %v", ins.Columns)
	}
	if len(ins.Values) != 1 || ins.Values[0].String() != "Work" {
		t.Errorf("expected values [Work], got %v", ins.Values)
	}
}

func TestParseUpdate(t *testing.T) {
	stmt, err := Parse("UPDATE logins SET password = 'new' WHERE name = 'GitHub'")
	if err != nil {
		t.Fatal(err)
	}
	upd, ok := stmt.(*ast.UpdateStatement)
	if !ok {
		t.Fatal("expected UpdateStatement")
	}
	if upd.Table != "logins" {
		t.Errorf("expected table 'logins', got %q", upd.Table)
	}
	if len(upd.Set) != 1 {
		t.Fatalf("expected 1 SET clause, got %d", len(upd.Set))
	}
	if upd.Set[0].Column != "password" {
		t.Errorf("expected column 'password', got %q", upd.Set[0].Column)
	}
	if upd.Where == nil {
		t.Fatal("expected WHERE clause")
	}
}

func TestParseDelete(t *testing.T) {
	stmt, err := Parse("DELETE FROM logins WHERE name = 'old'")
	if err != nil {
		t.Fatal(err)
	}
	del, ok := stmt.(*ast.DeleteStatement)
	if !ok {
		t.Fatal("expected DeleteStatement")
	}
	if del.Table != "logins" {
		t.Errorf("expected table 'logins', got %q", del.Table)
	}
	if del.Where == nil {
		t.Fatal("expected WHERE clause")
	}
}

func TestParseTrailingContent(t *testing.T) {
	_, err := Parse("SELECT * FROM logins DROP TABLE logins")
	if err == nil {
		t.Fatal("expected error for trailing content")
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseMultipleOrderBy(t *testing.T) {
	stmt, err := Parse("SELECT * FROM logins ORDER BY folder ASC, name DESC")
	if err != nil {
		t.Fatal(err)
	}
	sel := stmt.(*ast.SelectStatement)
	if len(sel.OrderBy) != 2 {
		t.Fatalf("expected 2 ORDER BY clauses, got %d", len(sel.OrderBy))
	}
	if sel.OrderBy[0].Desc {
		t.Error("expected first clause ASC")
	}
	if !sel.OrderBy[1].Desc {
		t.Error("expected second clause DESC")
	}
}
