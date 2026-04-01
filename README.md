# bwql

A SQL query interface for your [Bitwarden](https://bitwarden.com/) vault. Query, filter, and manage your vault items using familiar SQL syntax — directly from the terminal.

![demo](demo.gif)

## Security

bwql is built for people who care about the security of their credentials. The design reflects that:

**Zero external dependencies.** The entire project is built using only the Go standard library. No third-party packages. No transitive dependencies. No `go.sum` file. Nothing to be compromised in a supply chain attack. You can verify this yourself:

```bash
go mod graph
# github.com/owenrumney/bwql go@1.26.1  (stdlib only)
```

**Nothing is stored.** bwql does not write anything to disk — no config files, no logs, no cache, no history file. Your vault data exists only in memory for the duration of your session. When you exit, it's gone.

**No network access.** bwql itself makes no network calls. It shells out to the official `bw` CLI, which handles all communication with Bitwarden servers. bwql never sees your master password or session token in any form other than what the `bw` CLI provides.

**Fully auditable.** The codebase is small and straightforward. Read it. Every line of code that touches your data is in `internal/`.

## Install

```bash
go install github.com/owenrumney/bwql/cmd/bwql@latest
```

Or build from source:

```bash
git clone https://github.com/owenrumney/bwql.git
cd bwql
go build -o bwql ./cmd/bwql/
```

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Bitwarden CLI](https://bitwarden.com/help/cli/) (`bw`) installed and available in your PATH

## Usage

### Connect to your vault

```bash
# Option 1: Use BW_SESSION environment variable
export BW_SESSION=$(bw unlock --raw)
bwql

# Option 2: Pass session directly
bwql --session $(bw unlock --raw)
```

### Demo mode

Try bwql without connecting to a real vault:

```bash
bwql --demo
```

This loads a realistic mock vault with 42 items across multiple folders — useful for exploring the query syntax or recording demos.

### Tables

| Table        | Description                                      |
|--------------|--------------------------------------------------|
| `logins`     | Login items — username, password, URI, TOTP, etc |
| `cards`      | Payment cards                                    |
| `notes`      | Secure notes                                     |
| `identities` | Identity items                                   |
| `folders`    | Vault folders (supports mutations)               |
| `items`      | All items in a summary view                      |

### Login columns

| Column          | Description                                        |
|-----------------|----------------------------------------------------|
| `id`            | Bitwarden item ID                                  |
| `name`          | Item name                                          |
| `username`      | Login username                                     |
| `password`      | Login password                                     |
| `uri`           | Associated URIs                                    |
| `totp`          | TOTP secret (empty = no 2FA configured)            |
| `folder`        | Folder name                                        |
| `favorite`      | Whether the item is favorited                      |
| `password_age`  | Days since the password was last changed           |
| `revision_date` | Date the item was last modified                    |
| `created`       | Date the item was created                          |
| `notes`         | Item notes                                         |

### Query examples

```sql
-- Find logins without 2FA
SELECT name, username FROM logins WHERE totp IS NULL;

-- Find old passwords
SELECT name, username, password_age FROM logins
  WHERE password_age > 365
  ORDER BY password_age DESC;

-- Search by name
SELECT * FROM logins WHERE name LIKE '%github%';

-- Filter by folder
SELECT name, username FROM logins WHERE folder = 'Work' AND totp IS NULL;

-- List all cards
SELECT * FROM cards;

-- Search notes
SELECT * FROM notes WHERE name LIKE '%aws%';

-- View folders
SELECT * FROM folders;
```

### Folder mutations

```sql
-- Create a folder
INSERT INTO folders (name) VALUES ('New Folder');

-- Rename a folder
UPDATE folders SET name = 'Engineering' WHERE name = 'Work';

-- Delete a folder
DELETE FROM folders WHERE name = 'Unused';
```

### Display modes

bwql supports expanded display like `psql`:

```
bwql> \x on
Expanded display is on.

bwql> SELECT * FROM logins WHERE name = 'GitHub Enterprise';
-[ RECORD 1 ]------------------------------
ID            | abc-123-def
NAME          | GitHub Enterprise
USERNAME      | charlie.brown@acme.com
PASSWORD      | Kj8$mNp2!xQw9#Lz
URI           | https://github.acme.com
TOTP          | JBSWY3DPEHPK3PXP
FOLDER        | Work
FAVORITE      | true
PASSWORD_AGE  | 15
REVISION_DATE | 2026-03-17
CREATED       | 2025-02-15
NOTES         |
(1 rows)
```

| Command     | Effect                                                |
|-------------|-------------------------------------------------------|
| `\x`        | Cycle through: off, on, auto                          |
| `\x on`     | Always use expanded (vertical) display                |
| `\x off`    | Always use table display                              |
| `\x auto`   | Use expanded when the table exceeds terminal width    |

The default mode is `auto`.

### WHERE clause

Supports the following operators and expressions:

| Expression                | Example                                  |
|---------------------------|------------------------------------------|
| `=`, `!=`, `<>` | `WHERE name = 'GitHub'`                  |
| `<`, `>`, `<=`, `>=`     | `WHERE password_age > 365`               |
| `LIKE`                    | `WHERE name LIKE '%aws%'`                |
| `IS NULL` / `IS NOT NULL` | `WHERE totp IS NULL`                     |
| `IN (...)`                | `WHERE folder IN ('Work', 'Personal')`   |
| `AND` / `OR`             | `WHERE totp IS NULL AND folder = 'Work'` |
| `NOT`                     | `WHERE NOT favorite = 'true'`            |
| Parentheses               | `WHERE (a = '1' OR b = '2') AND c = '3'`|

### REPL features

- **Up/Down arrows** — command history
- **Left/Right arrows** — cursor movement
- **Ctrl-A / Ctrl-E** — jump to start / end of line
- **Ctrl-U** — clear line
- **Ctrl-L** — clear screen
- **Ctrl-D** — exit

## How it works

1. On startup, bwql calls `bw status` to check authentication
2. It loads your vault via `bw list items` and `bw list folders`
3. Items are parsed into an in-memory table model
4. Your SQL queries are parsed by a hand-rolled recursive descent parser
5. The query engine evaluates the AST against the in-memory data
6. Results are rendered as a table (or expanded display) in the terminal
7. Mutations shell out to `bw create`, `bw edit`, or `bw delete`
8. When you exit, everything is discarded — nothing is persisted

## Contributing

Pull requests are welcome. That said, bwql interacts directly with people's password vaults, so I'll review contributions carefully and cautiously. The bar for merging is high — not because contributions aren't valued, but because maintaining trust means being thorough.

If you're thinking of making a larger change, open an issue first so we can discuss the approach.

## License

MIT
