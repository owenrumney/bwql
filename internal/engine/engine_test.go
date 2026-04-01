package engine

import (
	"testing"
	"time"

	"github.com/owenrumney/bwql/internal/bw"
	"github.com/owenrumney/bwql/internal/parser"
)

type mockClient struct{}

func (m *mockClient) CreateFolder(name string) (*bw.Folder, error) {
	return &bw.Folder{ID: "new-id", Name: name}, nil
}
func (m *mockClient) EditFolder(id, name string) (*bw.Folder, error) {
	return &bw.Folder{ID: id, Name: name}, nil
}
func (m *mockClient) DeleteFolder(_ string) error { return nil }
func (m *mockClient) EditItem(item *bw.Item) (*bw.Item, error) {
	item.RevisionDate = time.Now()
	return item, nil
}
func (m *mockClient) DeleteItem(_ string) error { return nil }

func strPtr(s string) *string { return &s }

func testEngine() *Engine {
	now := time.Now()
	oldDate := now.AddDate(-2, 0, 0)
	recentDate := now.AddDate(0, -1, 0)

	items := []bw.Item{
		{
			ID:           "1",
			Type:         bw.ItemTypeLogin,
			Name:         "GitHub",
			FolderID:     strPtr("f1"),
			Favorite:     true,
			RevisionDate: recentDate,
			CreationDate: oldDate,
			Login: &bw.Login{
				Username:             strPtr("user@example.com"),
				Password:             strPtr("password123"),
				URIs:                 []bw.URI{{URI: "https://github.com"}},
				TOTP:                 strPtr("JBSWY3DPEHPK3PXP"),
				PasswordRevisionDate: &recentDate,
			},
		},
		{
			ID:           "2",
			Type:         bw.ItemTypeLogin,
			Name:         "AWS Console",
			FolderID:     strPtr("f1"),
			RevisionDate: oldDate,
			CreationDate: oldDate,
			Login: &bw.Login{
				Username:             strPtr("admin"),
				Password:             strPtr("oldpassword"),
				URIs:                 []bw.URI{{URI: "https://aws.amazon.com"}},
				PasswordRevisionDate: &oldDate,
			},
		},
		{
			ID:           "3",
			Type:         bw.ItemTypeLogin,
			Name:         "Random Site",
			RevisionDate: recentDate,
			CreationDate: recentDate,
			Login: &bw.Login{
				Username: strPtr("me"),
				Password: strPtr("pass"),
				URIs:     []bw.URI{{URI: "https://example.com"}, {URI: "https://www.example.com"}},
			},
		},
		{
			ID:           "4",
			Type:         bw.ItemTypeSecureNote,
			Name:         "AWS Keys",
			Notes:        strPtr("access_key=AKIA..."),
			RevisionDate: recentDate,
			CreationDate: recentDate,
		},
		{
			ID:           "5",
			Type:         bw.ItemTypeCard,
			Name:         "Visa",
			RevisionDate: recentDate,
			CreationDate: recentDate,
			Card: &bw.Card{
				CardholderName: strPtr("Owen Rumney"),
				Brand:          strPtr("visa"),
				Number:         strPtr("4111111111111111"),
				ExpMonth:       strPtr("12"),
				ExpYear:        strPtr("2027"),
			},
		},
	}

	folders := []bw.Folder{
		{ID: "f1", Name: "Work"},
	}

	return New(items, folders, &mockClient{})
}

func TestSelectAllLogins(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 3 {
		t.Fatalf("expected 3 logins, got %d", len(result.Rows))
	}
}

func TestSelectWhereTotpIsNull(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT name, username FROM logins WHERE totp IS NULL")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 logins without TOTP, got %d", len(result.Rows))
	}
	if len(result.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result.Columns))
	}
}

func TestSelectWherePasswordAgeGt365(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT name, password_age FROM logins WHERE password_age > 365")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 login with old password, got %d", len(result.Rows))
	}
	if result.Rows[0]["name"] != "AWS Console" {
		t.Errorf("expected 'AWS Console', got %q", result.Rows[0]["name"])
	}
}

func TestSelectWhereLike(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins WHERE name LIKE '%github%'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Rows))
	}
}

func TestSelectOrderByDesc(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT name, password_age FROM logins ORDER BY password_age DESC")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) < 2 {
		t.Fatal("expected at least 2 rows")
	}
	if result.Rows[0]["name"] != "AWS Console" {
		t.Errorf("expected 'AWS Console' first, got %q", result.Rows[0]["name"])
	}
}

func TestSelectLimit(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins LIMIT 1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
}

func TestSelectFromNotes(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM notes")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 note, got %d", len(result.Rows))
	}
}

func TestSelectFromCards(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM cards")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 card, got %d", len(result.Rows))
	}
}

func TestSelectWhereFolder(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins WHERE folder = 'Work'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 logins in Work folder, got %d", len(result.Rows))
	}
}

func TestSelectWhereAnd(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins WHERE totp IS NULL AND folder = 'Work'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 login, got %d", len(result.Rows))
	}
	if result.Rows[0]["name"] != "AWS Console" {
		t.Errorf("expected 'AWS Console', got %q", result.Rows[0]["name"])
	}
}

func TestUnknownTable(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM bogus")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for unknown table")
	}
}

func TestSelectFromItems(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM items")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 5 {
		t.Fatalf("expected 5 items, got %d", len(result.Rows))
	}
}

func TestSelectFromFolders(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM folders")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(result.Rows))
	}
	if result.Rows[0]["name"] != "Work" {
		t.Errorf("expected 'Work', got %q", result.Rows[0]["name"])
	}
}

func TestInsertFolder(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("INSERT INTO folders (name) VALUES ('Personal')")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
	if result.Rows[0]["name"] != "Personal" {
		t.Errorf("expected 'Personal', got %q", result.Rows[0]["name"])
	}

	sel, _ := parser.Parse("SELECT * FROM folders")
	r, _ := eng.Execute(sel)
	if len(r.Rows) != 2 {
		t.Fatalf("expected 2 folders after insert, got %d", len(r.Rows))
	}
}

func TestInsertFolderShortForm(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("INSERT INTO folders VALUES ('Personal')")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Rows[0]["name"] != "Personal" {
		t.Errorf("expected 'Personal', got %q", result.Rows[0]["name"])
	}
}

func TestUpdateFolder(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE folders SET name = 'Engineering' WHERE name = 'Work'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 updated row, got %d", len(result.Rows))
	}
	if result.Rows[0]["name"] != "Engineering" {
		t.Errorf("expected 'Engineering', got %q", result.Rows[0]["name"])
	}
}

func TestUpdateFolderRequiresWhere(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE folders SET name = 'Oops'")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for UPDATE without WHERE")
	}
}

func TestDeleteFolder(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("DELETE FROM folders WHERE name = 'Work'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 deleted row, got %d", len(result.Rows))
	}

	sel, _ := parser.Parse("SELECT * FROM folders")
	r, _ := eng.Execute(sel)
	if len(r.Rows) != 0 {
		t.Fatalf("expected 0 folders after delete, got %d", len(r.Rows))
	}
}

func TestDeleteFolderRequiresWhere(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("DELETE FROM folders")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for DELETE without WHERE")
	}
}

func TestUpdateLoginPassword(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE logins SET password = 'newpass123' WHERE name = 'GitHub'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 updated row, got %d", len(result.Rows))
	}
	if result.Rows[0]["password"] != "newpass123" {
		t.Errorf("expected password 'newpass123', got %q", result.Rows[0]["password"])
	}

	sel, _ := parser.Parse("SELECT password FROM logins WHERE name = 'GitHub'")
	r, _ := eng.Execute(sel)
	if len(r.Rows) != 1 || r.Rows[0]["password"] != "newpass123" {
		t.Error("password change not persisted in engine")
	}
}

func TestUpdateLoginUsername(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE logins SET username = 'newuser' WHERE name = 'GitHub'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Rows[0]["username"] != "newuser" {
		t.Errorf("expected username 'newuser', got %q", result.Rows[0]["username"])
	}
}

func TestUpdateLoginMultipleColumns(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE logins SET username = 'newuser', password = 'newpass' WHERE name = 'GitHub'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Rows[0]["username"] != "newuser" {
		t.Errorf("expected username 'newuser', got %q", result.Rows[0]["username"])
	}
	if result.Rows[0]["password"] != "newpass" {
		t.Errorf("expected password 'newpass', got %q", result.Rows[0]["password"])
	}
}

func TestUpdateLoginFolder(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE logins SET folder = 'Work' WHERE name = 'Random Site'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if result.Rows[0]["folder"] != "Work" {
		t.Errorf("expected folder 'Work', got %q", result.Rows[0]["folder"])
	}
}

func TestUpdateLoginFolderNotFound(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE logins SET folder = 'NonExistent' WHERE name = 'GitHub'")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for non-existent folder")
	}
}

func TestUpdateLoginInvalidColumn(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE logins SET bogus = 'val' WHERE name = 'GitHub'")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for invalid column")
	}
}

func TestUpdateLoginRequiresWhere(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE logins SET password = 'oops'")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for UPDATE without WHERE")
	}
}

func TestDeleteLogin(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("DELETE FROM logins WHERE name = 'Random Site'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 deleted row, got %d", len(result.Rows))
	}
	if result.Rows[0]["name"] != "Random Site" {
		t.Errorf("expected 'Random Site', got %q", result.Rows[0]["name"])
	}

	sel, _ := parser.Parse("SELECT * FROM logins")
	r, _ := eng.Execute(sel)
	if len(r.Rows) != 2 {
		t.Fatalf("expected 2 logins after delete, got %d", len(r.Rows))
	}
}

func TestDeleteLoginNoMatch(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("DELETE FROM logins WHERE name = 'Does Not Exist'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 0 {
		t.Fatalf("expected 0 deleted rows, got %d", len(result.Rows))
	}
}

func TestDeleteLoginRequiresWhere(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("DELETE FROM logins")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for DELETE without WHERE")
	}
}

func TestSelectWhereOr(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins WHERE name = 'GitHub' OR name = 'AWS Console'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestSelectWhereNotLike(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins WHERE name NOT LIKE '%GitHub%'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows (non-GitHub), got %d", len(result.Rows))
	}
}

func TestSelectWhereIn(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins WHERE name IN ('GitHub', 'AWS Console')")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestSelectWhereParens(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins WHERE (name = 'GitHub' OR name = 'AWS Console') AND folder = 'Work'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestSelectWhereLikeUnderscore(t *testing.T) {
	eng := testEngine()
	// _ should match exactly one character: "GitHub" doesn't match "G_tHub"
	stmt, err := parser.Parse("SELECT * FROM logins WHERE name LIKE 'G_tHub'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	// "GitHub" -> 'g_thub' matches 'github' with _ matching 'i'
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
}

func TestSelectWhereLikeUnderscoreWithPercent(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("SELECT * FROM logins WHERE name LIKE '%_ub'")
	if err != nil {
		t.Fatal(err)
	}
	result, err := eng.Execute(stmt)
	if err != nil {
		t.Fatal(err)
	}
	// "GitHub" ends with "Hub", _ matches 'H', so '%_ub' matches
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row (GitHub), got %d", len(result.Rows))
	}
}

func TestDeleteFromUnsupportedTable(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("DELETE FROM cards WHERE name = 'x'")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for unsupported table")
	}
}

func TestUpdateUnsupportedTable(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("UPDATE cards SET name = 'x' WHERE name = 'y'")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for unsupported table")
	}
}

func TestInsertFolderUnknownColumn(t *testing.T) {
	eng := testEngine()
	stmt, err := parser.Parse("INSERT INTO folders (name, bogus) VALUES ('x', 'y')")
	if err != nil {
		t.Fatal(err)
	}
	_, err = eng.Execute(stmt)
	if err == nil {
		t.Fatal("expected error for unknown column")
	}
}
