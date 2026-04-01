package table

import (
	"fmt"
	"strings"
)

type Mode int

const (
	ModeTable    Mode = iota // standard table output
	ModeExpanded             // \x expanded display
	ModeAuto                 // auto-select based on terminal width
)

func Render(columns []string, rows []map[string]string, mode Mode) string {
	if len(columns) == 0 {
		return ""
	}

	rows = sanitizeRows(rows)

	switch mode {
	case ModeExpanded:
		return renderExpanded(columns, rows)
	case ModeAuto:
		if tableWidth(columns, rows) > terminalWidth() {
			return renderExpanded(columns, rows)
		}
		return renderTable(columns, rows)
	default:
		return renderTable(columns, rows)
	}
}

func renderTable(columns []string, rows []map[string]string) string {
	widths := columnWidths(columns, rows)

	var b strings.Builder

	writeSeparator(&b, columns, widths)
	writeRow(&b, columns, widths, nil)
	writeSeparator(&b, columns, widths)

	for _, row := range rows {
		writeRow(&b, columns, widths, row)
	}

	if len(rows) > 0 {
		writeSeparator(&b, columns, widths)
	}

	fmt.Fprintf(&b, "(%d rows)\n", len(rows))

	return b.String()
}

func renderExpanded(columns []string, rows []map[string]string) string {
	labelWidth := 0
	for _, col := range columns {
		if len(col) > labelWidth {
			labelWidth = len(col)
		}
	}

	var b strings.Builder

	for i, row := range rows {
		header := fmt.Sprintf("-[ RECORD %d ]", i+1)
		padLen := labelWidth + 3 + 20 - len(header)
		if padLen < 1 {
			padLen = 1
		}
		b.WriteString(header)
		b.WriteString(strings.Repeat("-", padLen))
		b.WriteByte('\n')

		for _, col := range columns {
			val := row[strings.ToLower(col)]
			fmt.Fprintf(&b, "%-*s | %s\n", labelWidth, strings.ToUpper(col), val)
		}
	}

	fmt.Fprintf(&b, "(%d rows)\n", len(rows))

	return b.String()
}

func columnWidths(columns []string, rows []map[string]string) map[string]int {
	widths := make(map[string]int)
	for _, col := range columns {
		widths[col] = len(col)
	}
	for _, row := range rows {
		for _, col := range columns {
			val := row[strings.ToLower(col)]
			if len(val) > widths[col] {
				widths[col] = len(val)
			}
		}
	}

	maxWidth := 50
	for col, w := range widths {
		if w > maxWidth {
			widths[col] = maxWidth
		}
	}
	return widths
}

func tableWidth(columns []string, rows []map[string]string) int {
	widths := columnWidths(columns, rows)
	total := 1
	for _, col := range columns {
		total += widths[col] + 3
	}
	return total
}

func writeSeparator(b *strings.Builder, columns []string, widths map[string]int) {
	b.WriteByte('+')
	for _, col := range columns {
		b.WriteString(strings.Repeat("-", widths[col]+2))
		b.WriteByte('+')
	}
	b.WriteByte('\n')
}

func writeRow(b *strings.Builder, columns []string, widths map[string]int, row map[string]string) {
	b.WriteByte('|')
	for _, col := range columns {
		w := widths[col]
		var val string
		if row == nil {
			val = strings.ToUpper(col)
		} else {
			val = row[strings.ToLower(col)]
		}
		if len(val) > w {
			if w > 3 {
				val = val[:w-3] + "..."
			} else {
				val = val[:w]
			}
		}
		fmt.Fprintf(b, " %-*s |", w, val)
	}
	b.WriteByte('\n')
}

func sanitizeRows(rows []map[string]string) []map[string]string {
	clean := make([]map[string]string, len(rows))
	for i, row := range rows {
		cleanRow := make(map[string]string, len(row))
		for k, v := range row {
			cleanRow[k] = sanitize(v)
		}
		clean[i] = cleanRow
	}
	return clean
}

func sanitize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\x1b' {
			// Skip entire ANSI escape sequence: ESC [ ... final_byte
			i++
			if i < len(s) && s[i] == '[' {
				i++
				for i < len(s) && s[i] >= 0x20 && s[i] <= 0x3F {
					i++
				}
			}
			continue
		}
		switch {
		case ch == '\n' || ch == '\r':
			b.WriteByte(' ')
		case ch < 0x20:
			continue
		default:
			b.WriteByte(ch)
		}
	}
	return b.String()
}

