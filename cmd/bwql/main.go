package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/owenrumney/bwql/internal/bw"
	"github.com/owenrumney/bwql/internal/demo"
	"github.com/owenrumney/bwql/internal/engine"
	"github.com/owenrumney/bwql/internal/parser"
	"github.com/owenrumney/bwql/internal/repl"
	"github.com/owenrumney/bwql/internal/table"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	sessionFlag := flag.String("session", "", "Bitwarden session key (overrides BW_SESSION env var)")
	demoFlag := flag.Bool("demo", false, "Run with a demo vault (no Bitwarden CLI required)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("bwql %s (%s) built %s\n", version, commit, date)
		return
	}

	var eng *engine.Engine

	session := *sessionFlag
	if session == "" {
		session = os.Getenv("BW_SESSION")
	}

	if *demoFlag || session == "demo" {
		eng = loadDemo()
	} else {
		eng = loadVault(session)
	}

	fmt.Println("Tables: logins, cards, notes, identities, folders, items")
	fmt.Println("Type 'help' for examples, 'exit' or Ctrl-D to quit.")
	fmt.Println()

	displayMode := table.ModeAuto

	rl := repl.New("bwql> ")
	for {
		input, err := rl.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
			break
		}
		if strings.EqualFold(input, "help") {
			printHelp()
			continue
		}
		if strings.EqualFold(input, "tables") {
			printTables()
			continue
		}

		if handled, newMode := handleDisplayMode(input, displayMode); handled {
			displayMode = newMode
			continue
		}

		stmt, err := parser.Parse(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %s\n", err)
			continue
		}

		result, err := eng.Execute(stmt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			continue
		}

		rows := make([]map[string]string, len(result.Rows))
		for i, r := range result.Rows {
			rows[i] = map[string]string(r)
		}
		fmt.Print(table.Render(result.Columns, rows, displayMode))
	}
}

func loadDemo() *engine.Engine {
	items := demo.Items()
	folders := demo.Folders()
	client := demo.NewClient()

	eng := engine.New(items, folders, client)
	printSummary("demo vault", items)
	return eng
}

func loadVault(session string) *engine.Engine {
	client := bw.NewClient(session)

	status, err := client.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\nIs the Bitwarden CLI installed?\n", err)
		os.Exit(1)
	}

	switch status.Status {
	case "unauthenticated":
		fmt.Fprintln(os.Stderr, "error: not logged in to Bitwarden. Run 'bw login' first.")
		os.Exit(1)
	case "locked":
		if session == "" {
			fmt.Fprintln(os.Stderr, "error: vault is locked. Run 'bw unlock' and export BW_SESSION, or set BW_SESSION env var.")
			os.Exit(1)
		}
	}

	fmt.Println("Loading vault...")
	items, err := client.ListItems()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading items: %s\n", err)
		os.Exit(1)
	}

	folders, err := client.ListFolders()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading folders: %s\n", err)
		os.Exit(1)
	}

	eng := engine.New(items, folders, client)
	printSummary("vault", items)
	return eng
}

func printSummary(source string, items []bw.Item) {
	var logins, notes, cards, identities int
	for _, item := range items {
		switch item.Type {
		case bw.ItemTypeLogin:
			logins++
		case bw.ItemTypeSecureNote:
			notes++
		case bw.ItemTypeCard:
			cards++
		case bw.ItemTypeIdentity:
			identities++
		}
	}
	fmt.Printf("Loaded %d items from %s (%d logins, %d notes, %d cards, %d identities)\n",
		len(items), source, logins, notes, cards, identities)
}

func printTables() {
	fmt.Println("  logins     - Login items (username, password, uri, totp, etc.)")
	fmt.Println("  cards      - Card items (cardholder, number, brand, etc.)")
	fmt.Println("  notes      - Secure notes")
	fmt.Println("  identities - Identity items")
	fmt.Println("  folders    - Folders (supports INSERT, UPDATE, DELETE)")
	fmt.Println("  items      - All items (summary view)")
}

func handleDisplayMode(input string, current table.Mode) (bool, table.Mode) {
	lower := strings.ToLower(strings.TrimSpace(input))
	switch lower {
	case `\x`:
		// Toggle: table -> expanded -> auto -> table
		switch current {
		case table.ModeTable:
			fmt.Println("Expanded display is on.")
			return true, table.ModeExpanded
		case table.ModeExpanded:
			fmt.Println("Expanded display is used automatically.")
			return true, table.ModeAuto
		default:
			fmt.Println("Expanded display is off.")
			return true, table.ModeTable
		}
	case `\x on`:
		fmt.Println("Expanded display is on.")
		return true, table.ModeExpanded
	case `\x off`:
		fmt.Println("Expanded display is off.")
		return true, table.ModeTable
	case `\x auto`:
		fmt.Println("Expanded display is used automatically.")
		return true, table.ModeAuto
	}
	return false, current
}

func printHelp() {
	_, _ = os.Stdout.WriteString(`Commands:
  tables              List available tables
  \x                  Toggle expanded display (off -> on -> auto)
  \x on|off|auto      Set display mode explicitly
  help                Show this help
  exit / quit         Exit bwql

Query examples:
  SELECT * FROM logins;
  SELECT name, username FROM logins WHERE totp IS NULL;
  SELECT * FROM logins WHERE password_age > 365 ORDER BY password_age DESC;
  SELECT * FROM logins WHERE name LIKE '%github%';
  SELECT * FROM logins WHERE folder = 'Work' AND totp IS NULL;
  SELECT * FROM cards;
  SELECT * FROM notes WHERE name LIKE '%aws%';
  SELECT * FROM identities;
  SELECT * FROM folders;

Folder mutations:
  INSERT INTO folders (name) VALUES ('New Folder');
  UPDATE folders SET name = 'Renamed' WHERE name = 'Old';
  DELETE FROM folders WHERE name = 'Unused';
`)
}
