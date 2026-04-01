package engine

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/owenrumney/bwql/internal/ast"
	"github.com/owenrumney/bwql/internal/bw"
)

type Row map[string]string

type Result struct {
	Columns []string
	Rows    []Row
}

type BWClient interface {
	CreateFolder(name string) (*bw.Folder, error)
	EditFolder(id, name string) (*bw.Folder, error)
	DeleteFolder(id string) error
	EditItem(item *bw.Item) (*bw.Item, error)
	DeleteItem(id string) error
}

type Engine struct {
	items      []bw.Item
	folderList []bw.Folder
	folders    map[string]string // id -> name
	client     BWClient
}

func New(items []bw.Item, folders []bw.Folder, client BWClient) *Engine {
	folderMap := make(map[string]string)
	for _, f := range folders {
		folderMap[f.ID] = f.Name
	}
	return &Engine{items: items, folderList: folders, folders: folderMap, client: client}
}

func (e *Engine) Execute(stmt ast.Statement) (*Result, error) {
	switch s := stmt.(type) {
	case *ast.SelectStatement:
		return e.executeSelect(s)
	case *ast.InsertStatement:
		return e.executeInsert(s)
	case *ast.UpdateStatement:
		return e.executeUpdate(s)
	case *ast.DeleteStatement:
		return e.executeDelete(s)
	default:
		return nil, fmt.Errorf("unsupported statement type")
	}
}

func (e *Engine) replaceFolder(updated bw.Folder) {
	for i, f := range e.folderList {
		if f.ID == updated.ID {
			e.folderList[i] = updated
			return
		}
	}
}

func (e *Engine) removeFolders(ids map[string]bool) {
	var remaining []bw.Folder
	for _, f := range e.folderList {
		if !ids[f.ID] {
			remaining = append(remaining, f)
		}
	}
	e.folderList = remaining
}

func (e *Engine) refreshFolderMap() {
	e.folders = make(map[string]string)
	for _, f := range e.folderList {
		e.folders[f.ID] = f.Name
	}
}

func (e *Engine) executeSelect(stmt *ast.SelectStatement) (*Result, error) {
	rows, allColumns, err := e.getTableRows(stmt.Table)
	if err != nil {
		return nil, err
	}

	if stmt.Where != nil {
		var filtered []Row
		for _, row := range rows {
			match, err := evalExpr(stmt.Where, row)
			if err != nil {
				return nil, fmt.Errorf("evaluating WHERE: %w", err)
			}
			if match {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}

	if len(stmt.OrderBy) > 0 {
		sort.SliceStable(rows, func(i, j int) bool {
			for _, ob := range stmt.OrderBy {
				col := strings.ToLower(ob.Column)
				vi := rows[i][col]
				vj := rows[j][col]

				cmp := compareValues(vi, vj)
				if cmp == 0 {
					continue
				}
				if ob.Desc {
					return cmp > 0
				}
				return cmp < 0
			}
			return false
		})
	}

	if stmt.Limit != nil && *stmt.Limit < len(rows) {
		rows = rows[:*stmt.Limit]
	}

	columns := resolveColumns(stmt.Columns, allColumns)

	return &Result{Columns: columns, Rows: rows}, nil
}

func (e *Engine) getTableRows(table string) ([]Row, []string, error) {
	switch table {
	case "logins", "login":
		return e.loginsTable()
	case "cards", "card":
		return e.cardsTable()
	case "notes", "note":
		return e.notesTable()
	case "identities", "identity":
		return e.identitiesTable()
	case "items", "item":
		return e.allItemsTable()
	case "folders", "folder":
		return e.foldersTable()
	default:
		return nil, nil, fmt.Errorf("unknown table %q (valid: logins, cards, notes, identities, folders, items)", table)
	}
}

var loginColumns = []string{"id", "name", "username", "password", "uri", "totp", "folder", "favorite", "password_age", "revision_date", "created", "notes"}

func (e *Engine) loginsTable() ([]Row, []string, error) {
	var rows []Row
	for _, item := range e.items {
		if item.Type != bw.ItemTypeLogin || item.Login == nil {
			continue
		}
		row := Row{
			"id":            item.ID,
			"name":          item.Name,
			"username":      deref(item.Login.Username),
			"password":      deref(item.Login.Password),
			"uri":           joinURIs(item.Login.URIs),
			"totp":          deref(item.Login.TOTP),
			"folder":        e.folderName(item.FolderID),
			"favorite":      boolStr(item.Favorite),
			"password_age":  passwordAge(item.Login.PasswordRevisionDate, item.RevisionDate),
			"revision_date": item.RevisionDate.Format("2006-01-02"),
			"created":       item.CreationDate.Format("2006-01-02"),
			"notes":         deref(item.Notes),
		}
		rows = append(rows, row)
	}
	return rows, loginColumns, nil
}

var cardColumns = []string{"id", "name", "cardholder", "brand", "number", "exp_month", "exp_year", "code", "folder", "revision_date", "notes"}

func (e *Engine) cardsTable() ([]Row, []string, error) {
	var rows []Row
	for _, item := range e.items {
		if item.Type != bw.ItemTypeCard || item.Card == nil {
			continue
		}
		row := Row{
			"id":            item.ID,
			"name":          item.Name,
			"cardholder":    deref(item.Card.CardholderName),
			"brand":         deref(item.Card.Brand),
			"number":        deref(item.Card.Number),
			"exp_month":     deref(item.Card.ExpMonth),
			"exp_year":      deref(item.Card.ExpYear),
			"code":          deref(item.Card.Code),
			"folder":        e.folderName(item.FolderID),
			"revision_date": item.RevisionDate.Format("2006-01-02"),
			"notes":         deref(item.Notes),
		}
		rows = append(rows, row)
	}
	return rows, cardColumns, nil
}

var noteColumns = []string{"id", "name", "notes", "folder", "revision_date", "created"}

func (e *Engine) notesTable() ([]Row, []string, error) {
	var rows []Row
	for _, item := range e.items {
		if item.Type != bw.ItemTypeSecureNote {
			continue
		}
		row := Row{
			"id":            item.ID,
			"name":          item.Name,
			"notes":         deref(item.Notes),
			"folder":        e.folderName(item.FolderID),
			"revision_date": item.RevisionDate.Format("2006-01-02"),
			"created":       item.CreationDate.Format("2006-01-02"),
		}
		rows = append(rows, row)
	}
	return rows, noteColumns, nil
}

var identityColumns = []string{"id", "name", "first_name", "last_name", "email", "phone", "company", "username", "folder", "revision_date"}

func (e *Engine) identitiesTable() ([]Row, []string, error) {
	var rows []Row
	for _, item := range e.items {
		if item.Type != bw.ItemTypeIdentity || item.Identity == nil {
			continue
		}
		row := Row{
			"id":            item.ID,
			"name":          item.Name,
			"first_name":    deref(item.Identity.FirstName),
			"last_name":     deref(item.Identity.LastName),
			"email":         deref(item.Identity.Email),
			"phone":         deref(item.Identity.Phone),
			"company":       deref(item.Identity.Company),
			"username":      deref(item.Identity.Username),
			"folder":        e.folderName(item.FolderID),
			"revision_date": item.RevisionDate.Format("2006-01-02"),
		}
		rows = append(rows, row)
	}
	return rows, identityColumns, nil
}

var allItemColumns = []string{"id", "type", "name", "folder", "favorite", "revision_date", "created", "notes"}

func (e *Engine) allItemsTable() ([]Row, []string, error) {
	rows := make([]Row, 0, len(e.items))
	for _, item := range e.items {
		row := Row{
			"id":            item.ID,
			"type":          itemTypeName(item.Type),
			"name":          item.Name,
			"folder":        e.folderName(item.FolderID),
			"favorite":      boolStr(item.Favorite),
			"revision_date": item.RevisionDate.Format("2006-01-02"),
			"created":       item.CreationDate.Format("2006-01-02"),
			"notes":         deref(item.Notes),
		}
		rows = append(rows, row)
	}
	return rows, allItemColumns, nil
}

var folderColumns = []string{"id", "name"}

func (e *Engine) foldersTable() ([]Row, []string, error) {
	rows := make([]Row, 0, len(e.folderList))
	for _, f := range e.folderList {
		row := Row{
			"id":   f.ID,
			"name": f.Name,
		}
		rows = append(rows, row)
	}
	return rows, folderColumns, nil
}

func (e *Engine) executeInsert(stmt *ast.InsertStatement) (*Result, error) {
	switch stmt.Table {
	case "folders", "folder":
		return e.insertFolder(stmt)
	default:
		return nil, fmt.Errorf("INSERT not yet supported for table %q", stmt.Table)
	}
}

func (e *Engine) insertFolder(stmt *ast.InsertStatement) (*Result, error) {
	var name string

	switch {
	case len(stmt.Columns) > 0:
		for i, col := range stmt.Columns {
			if strings.EqualFold(col, "name") {
				if i >= len(stmt.Values) {
					return nil, fmt.Errorf("missing value for column 'name'")
				}
				name = stmt.Values[i].String()
			} else {
				return nil, fmt.Errorf("unknown column %q for folders table", col)
			}
		}
	case len(stmt.Values) == 1:
		name = stmt.Values[0].String()
	default:
		return nil, fmt.Errorf("INSERT INTO folders expects (name) VALUES ('folder name')")
	}

	if name == "" {
		return nil, fmt.Errorf("folder name cannot be empty")
	}

	folder, err := e.client.CreateFolder(name)
	if err != nil {
		return nil, err
	}

	e.folderList = append(e.folderList, *folder)
	e.refreshFolderMap()

	return &Result{
		Columns: folderColumns,
		Rows:    []Row{{"id": folder.ID, "name": folder.Name}},
	}, nil
}

func (e *Engine) executeUpdate(stmt *ast.UpdateStatement) (*Result, error) {
	switch stmt.Table {
	case "folders", "folder":
		return e.updateFolders(stmt)
	case "logins", "login":
		return e.updateLogins(stmt)
	default:
		return nil, fmt.Errorf("UPDATE not yet supported for table %q", stmt.Table)
	}
}

func (e *Engine) updateFolders(stmt *ast.UpdateStatement) (*Result, error) {
	var newName string
	for _, s := range stmt.Set {
		if strings.EqualFold(s.Column, "name") {
			newName = s.Value.String()
		} else {
			return nil, fmt.Errorf("cannot update column %q on folders", s.Column)
		}
	}
	if newName == "" {
		return nil, fmt.Errorf("UPDATE folders requires SET name = 'new name'")
	}

	rows, _, err := e.foldersTable()
	if err != nil {
		return nil, err
	}

	if stmt.Where == nil {
		return nil, fmt.Errorf("UPDATE folders requires a WHERE clause")
	}

	var updated []Row
	for _, row := range rows {
		match, err := evalExpr(stmt.Where, row)
		if err != nil {
			return nil, fmt.Errorf("evaluating WHERE: %w", err)
		}
		if !match {
			continue
		}

		folder, err := e.client.EditFolder(row["id"], newName)
		if err != nil {
			e.refreshFolderMap()
			return nil, fmt.Errorf("updated %d of %d+ matching folders before error: %w", len(updated), len(updated)+1, err)
		}

		e.replaceFolder(*folder)
		updated = append(updated, Row{"id": folder.ID, "name": folder.Name})
	}

	e.refreshFolderMap()

	return &Result{
		Columns: folderColumns,
		Rows:    updated,
	}, nil
}

func (e *Engine) executeDelete(stmt *ast.DeleteStatement) (*Result, error) {
	switch stmt.Table {
	case "folders", "folder":
		return e.deleteFolders(stmt)
	case "logins", "login":
		return e.deleteLogins(stmt)
	default:
		return nil, fmt.Errorf("DELETE not yet supported for table %q", stmt.Table)
	}
}

func (e *Engine) deleteFolders(stmt *ast.DeleteStatement) (*Result, error) {
	if stmt.Where == nil {
		return nil, fmt.Errorf("DELETE FROM folders requires a WHERE clause")
	}

	rows, _, err := e.foldersTable()
	if err != nil {
		return nil, err
	}

	var deleted []Row
	deletedIDs := make(map[string]bool)
	for _, row := range rows {
		match, err := evalExpr(stmt.Where, row)
		if err != nil {
			return nil, fmt.Errorf("evaluating WHERE: %w", err)
		}
		if match {
			if err := e.client.DeleteFolder(row["id"]); err != nil {
				e.removeFolders(deletedIDs)
				e.refreshFolderMap()
				return nil, fmt.Errorf("deleted %d of %d+ matching folders before error: %w", len(deleted), len(deleted)+1, err)
			}
			deletedIDs[row["id"]] = true
			deleted = append(deleted, row)
		}
	}

	e.removeFolders(deletedIDs)
	e.refreshFolderMap()

	return &Result{
		Columns: folderColumns,
		Rows:    deleted,
	}, nil
}

func (e *Engine) updateLogins(stmt *ast.UpdateStatement) (*Result, error) {
	if stmt.Where == nil {
		return nil, fmt.Errorf("UPDATE logins requires a WHERE clause")
	}

	rows, _, err := e.loginsTable()
	if err != nil {
		return nil, err
	}

	var updated []Row
	for _, row := range rows {
		match, err := evalExpr(stmt.Where, row)
		if err != nil {
			return nil, fmt.Errorf("evaluating WHERE: %w", err)
		}
		if !match {
			continue
		}

		item := e.findItem(row["id"])
		if item == nil {
			return nil, fmt.Errorf("item %s not found", row["id"])
		}

		for _, s := range stmt.Set {
			col := strings.ToLower(s.Column)
			val := s.Value.String()
			if err := applyLoginUpdate(item, col, val, e.folders); err != nil {
				return nil, err
			}
		}

		editedItem, err := e.client.EditItem(item)
		if err != nil {
			return nil, fmt.Errorf("updated %d of %d+ matching logins before error: %w", len(updated), len(updated)+1, err)
		}

		e.replaceItem(*editedItem)
		updated = append(updated, e.itemToLoginRow(*editedItem))
	}

	return &Result{
		Columns: loginColumns,
		Rows:    updated,
	}, nil
}

func applyLoginUpdate(item *bw.Item, col, val string, folders map[string]string) error {
	switch col {
	case "name":
		item.Name = val
	case "username":
		if item.Login == nil {
			return fmt.Errorf("item is not a login")
		}
		item.Login.Username = &val
	case "password":
		if item.Login == nil {
			return fmt.Errorf("item is not a login")
		}
		item.Login.Password = &val
		now := time.Now()
		item.Login.PasswordRevisionDate = &now
	case "uri":
		if item.Login == nil {
			return fmt.Errorf("item is not a login")
		}
		item.Login.URIs = []bw.URI{{URI: val}}
	case "totp":
		if item.Login == nil {
			return fmt.Errorf("item is not a login")
		}
		if val == "" {
			item.Login.TOTP = nil
		} else {
			item.Login.TOTP = &val
		}
	case "folder":
		folderID := folderIDByName(folders, val)
		if folderID == "" && val != "" {
			return fmt.Errorf("folder %q not found", val)
		}
		if val == "" {
			item.FolderID = nil
		} else {
			item.FolderID = &folderID
		}
	case "favorite":
		item.Favorite = strings.EqualFold(val, "true")
	case "notes":
		if val == "" {
			item.Notes = nil
		} else {
			item.Notes = &val
		}
	default:
		return fmt.Errorf("cannot update column %q on logins", col)
	}
	return nil
}

func folderIDByName(folders map[string]string, name string) string {
	for id, n := range folders {
		if strings.EqualFold(n, name) {
			return id
		}
	}
	return ""
}

func (e *Engine) deleteLogins(stmt *ast.DeleteStatement) (*Result, error) {
	if stmt.Where == nil {
		return nil, fmt.Errorf("DELETE FROM logins requires a WHERE clause")
	}

	rows, _, err := e.loginsTable()
	if err != nil {
		return nil, err
	}

	var deleted []Row
	for _, row := range rows {
		match, err := evalExpr(stmt.Where, row)
		if err != nil {
			return nil, fmt.Errorf("evaluating WHERE: %w", err)
		}
		if !match {
			continue
		}

		if err := e.client.DeleteItem(row["id"]); err != nil {
			return nil, fmt.Errorf("deleted %d of %d+ matching logins before error: %w", len(deleted), len(deleted)+1, err)
		}

		e.removeItem(row["id"])
		deleted = append(deleted, row)
	}

	return &Result{
		Columns: loginColumns,
		Rows:    deleted,
	}, nil
}

func (e *Engine) findItem(id string) *bw.Item {
	for i := range e.items {
		if e.items[i].ID == id {
			// Return a copy so mutations don't affect the engine state
			// until we explicitly replace it
			item := e.items[i]
			return &item
		}
	}
	return nil
}

func (e *Engine) replaceItem(updated bw.Item) {
	for i, item := range e.items {
		if item.ID == updated.ID {
			e.items[i] = updated
			return
		}
	}
}

func (e *Engine) removeItem(id string) {
	for i, item := range e.items {
		if item.ID == id {
			e.items = append(e.items[:i], e.items[i+1:]...)
			return
		}
	}
}

func (e *Engine) itemToLoginRow(item bw.Item) Row {
	return Row{
		"id":            item.ID,
		"name":          item.Name,
		"username":      deref(item.Login.Username),
		"password":      deref(item.Login.Password),
		"uri":           joinURIs(item.Login.URIs),
		"totp":          deref(item.Login.TOTP),
		"folder":        e.folderName(item.FolderID),
		"favorite":      boolStr(item.Favorite),
		"password_age":  passwordAge(item.Login.PasswordRevisionDate, item.RevisionDate),
		"revision_date": item.RevisionDate.Format("2006-01-02"),
		"created":       item.CreationDate.Format("2006-01-02"),
		"notes":         deref(item.Notes),
	}
}

func resolveColumns(requested []ast.Column, all []string) []string {
	if len(requested) == 1 && requested[0].Name == "*" {
		return all
	}
	var cols []string
	for _, c := range requested {
		if c.Alias != "" {
			cols = append(cols, c.Alias)
		} else {
			cols = append(cols, c.Name)
		}
	}
	return cols
}

func evalExpr(expr ast.Expr, row Row) (bool, error) {
	switch e := expr.(type) {
	case *ast.ComparisonExpr:
		col := strings.ToLower(e.Left)
		val := row[col]
		right := e.Right.String()

		switch e.Operator {
		case "=":
			return strings.EqualFold(val, right), nil
		case "!=":
			return !strings.EqualFold(val, right), nil
		case "LIKE":
			return matchLike(val, right), nil
		case "NOT LIKE":
			return !matchLike(val, right), nil
		case "<", ">", "<=", ">=":
			return compareOp(val, right, e.Operator)
		}
		return false, fmt.Errorf("unknown operator %q", e.Operator)

	case *ast.IsNullExpr:
		col := strings.ToLower(e.Column)
		val := row[col]
		isNull := val == ""
		if e.Not {
			return !isNull, nil
		}
		return isNull, nil

	case *ast.AndExpr:
		left, err := evalExpr(e.Left, row)
		if err != nil {
			return false, err
		}
		if !left {
			return false, nil
		}
		return evalExpr(e.Right, row)

	case *ast.OrExpr:
		left, err := evalExpr(e.Left, row)
		if err != nil {
			return false, err
		}
		if left {
			return true, nil
		}
		return evalExpr(e.Right, row)

	case *ast.NotExpr:
		result, err := evalExpr(e.Expr, row)
		if err != nil {
			return false, err
		}
		return !result, nil

	case *ast.InExpr:
		col := strings.ToLower(e.Column)
		val := row[col]
		found := false
		for _, v := range e.Values {
			if strings.EqualFold(val, v.String()) {
				found = true
				break
			}
		}
		if e.Not {
			return !found, nil
		}
		return found, nil

	case *ast.ParenExpr:
		return evalExpr(e.Expr, row)
	}

	return false, fmt.Errorf("unknown expression type")
}

func matchLike(val, pattern string) bool {
	val = strings.ToLower(val)
	pattern = strings.ToLower(pattern)

	// Dynamic programming approach for SQL LIKE matching
	// % matches any sequence of characters, _ matches exactly one
	v, p := len(val), len(pattern)
	// dp[i][j] = val[:i] matches pattern[:j]
	// Use two rows to save memory
	prev := make([]bool, p+1)
	curr := make([]bool, p+1)

	prev[0] = true
	for j := 1; j <= p; j++ {
		if pattern[j-1] == '%' {
			prev[j] = prev[j-1]
		}
	}

	for i := 1; i <= v; i++ {
		curr[0] = false
		for j := 1; j <= p; j++ {
			switch pattern[j-1] {
			case '%':
				curr[j] = curr[j-1] || prev[j]
			case '_':
				curr[j] = prev[j-1]
			default:
				curr[j] = prev[j-1] && val[i-1] == pattern[j-1]
			}
		}
		prev, curr = curr, prev
		for j := range curr {
			curr[j] = false
		}
	}

	return prev[p]
}

func compareOp(a, b, op string) (bool, error) {
	cmp := compareValues(a, b)
	switch op {
	case "<":
		return cmp < 0, nil
	case ">":
		return cmp > 0, nil
	case "<=":
		return cmp <= 0, nil
	case ">=":
		return cmp >= 0, nil
	}
	return false, fmt.Errorf("unknown operator %q", op)
}

func compareValues(a, b string) int {
	na, errA := strconv.ParseFloat(a, 64)
	nb, errB := strconv.ParseFloat(b, 64)
	if errA == nil && errB == nil {
		if na < nb {
			return -1
		}
		if na > nb {
			return 1
		}
		return 0
	}
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func joinURIs(uris []bw.URI) string {
	parts := make([]string, 0, len(uris))
	for _, u := range uris {
		parts = append(parts, u.URI)
	}
	return strings.Join(parts, ", ")
}

func (e *Engine) folderName(id *string) string {
	if id == nil {
		return ""
	}
	if name, ok := e.folders[*id]; ok {
		return name
	}
	return ""
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func passwordAge(revDate *time.Time, fallback time.Time) string {
	ref := fallback
	if revDate != nil {
		ref = *revDate
	}
	days := int(math.Floor(time.Since(ref).Hours() / 24))
	return strconv.Itoa(days)
}

func itemTypeName(t int) string {
	switch t {
	case bw.ItemTypeLogin:
		return "login"
	case bw.ItemTypeSecureNote:
		return "note"
	case bw.ItemTypeCard:
		return "card"
	case bw.ItemTypeIdentity:
		return "identity"
	default:
		return "unknown"
	}
}
