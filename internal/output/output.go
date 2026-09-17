package output

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// PrintTable prints data as an aligned table to stdout.
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print separator
	sep := make([]string, len(headers))
	for i := range sep {
		sep[i] = strings.Repeat("-", maxLen(headers[i], getMaxColWidth(rows, i)))
	}
	fmt.Fprintln(w, strings.Join(sep, "\t"))

	// Print rows
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

func maxLen(a string, b int) int {
	la := len([]rune(a))
	if la > b {
		return la
	}
	return b
}

func getMaxColWidth(rows [][]string, col int) int {
	max := 0
	for _, row := range rows {
		if col < len(row) {
			l := len([]rune(row[col]))
			if l > max {
				max = l
			}
		}
	}
	return max
}

// PrintJSON prints data as indented JSON to stdout.
func PrintJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false) // 保留 < > & 原样输出 (校验 SQL 含 <field> 等标签)
	return enc.Encode(v)
}

// PrintCSV prints data as CSV to stdout.
func PrintCSV(headers []string, rows [][]string) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(headers); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	w.Flush()
	return w.Error()
}

// PrintCSVFromMaps prints a slice of maps as CSV.
func PrintCSVFromMaps(headers []string, rows [][]string) error {
	buf := bufio.NewWriter(os.Stdout)
	defer buf.Flush()

	// Manual CSV writing to avoid quoting all fields
	for i, h := range headers {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(escapeCSV(h))
	}
	buf.WriteString("\r\n")

	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(escapeCSV(cell))
		}
		buf.WriteString("\r\n")
	}
	return nil
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\r\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}
