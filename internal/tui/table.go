package tui

import (
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// column defines a single table column.
type column struct {
	Name  string
	Width int
	Align lipgloss.Position // lipgloss.Left or lipgloss.Right
}

// tableModel is a reusable scrollable table component.
type tableModel struct {
	columns  []column
	rows     [][]string // each row is a slice of cell values
	cursor   int
	offset   int // scroll offset
	height   int // visible rows (excluding header)
	width    int
	sortCol  int
	sortDesc bool
}

// newTableModel creates a new table with columns and initial rows.
func newTableModel(columns []column, rows [][]string) tableModel {
	return tableModel{
		columns:  columns,
		rows:     rows,
		sortCol:  -1,
		sortDesc: true,
	}
}

// Update handles keyboard navigation for the table.
func (t tableModel) Update(msg tea.Msg) (tableModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
			}
			// Scroll up if cursor is above visible area
			if t.cursor < t.offset {
				t.offset = t.cursor
			}
		case "down", "j":
			if t.cursor < len(t.rows)-1 {
				t.cursor++
			}
			// Scroll down if cursor is below visible area
			visibleRows := t.visibleRowCount()
			if t.cursor >= t.offset+visibleRows {
				t.offset = t.cursor - visibleRows + 1
			}
		case "s":
			// Cycle sort column
			t.sortCol++
			if t.sortCol >= len(t.columns) {
				t.sortCol = 0
				t.sortDesc = !t.sortDesc
			}
			t.sortRows()
		case "home", "g":
			t.cursor = 0
			t.offset = 0
		case "end", "G":
			if len(t.rows) > 0 {
				t.cursor = len(t.rows) - 1
				visibleRows := t.visibleRowCount()
				if t.cursor >= visibleRows {
					t.offset = t.cursor - visibleRows + 1
				}
			}
		}
	}
	return t, nil
}

// View renders the table with header and visible rows.
func (t tableModel) View() string {
	if len(t.columns) == 0 {
		return ""
	}

	var b strings.Builder

	// Render header
	headerCells := make([]string, len(t.columns))
	for i, col := range t.columns {
		name := col.Name
		// Add sort indicator
		if i == t.sortCol {
			if t.sortDesc {
				name += " v"
			} else {
				name += " ^"
			}
		}
		headerCells[i] = theme.TableHeaderStyle.Copy().
			Width(col.Width).
			Align(col.Align).
			Render(truncate(name, col.Width))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	b.WriteString("\n")

	// Render visible rows
	visibleRows := t.visibleRowCount()
	end := t.offset + visibleRows
	if end > len(t.rows) {
		end = len(t.rows)
	}

	altRowStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.SurfaceColor))

	for i := t.offset; i < end; i++ {
		row := t.rows[i]
		cells := make([]string, len(t.columns))
		for j, col := range t.columns {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			cellStyle := lipgloss.NewStyle().
				Width(col.Width).
				Align(col.Align)

			if i == t.cursor {
				cellStyle = theme.SelectedRowStyle.Copy().
					Width(col.Width).
					Align(col.Align)
			} else if (i-t.offset)%2 == 1 {
				cellStyle = altRowStyle.Copy().
					Width(col.Width).
					Align(col.Align)
			}

			cells[j] = cellStyle.Render(truncate(val, col.Width))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// SelectedRow returns the current cursor position.
func (t tableModel) SelectedRow() int {
	return t.cursor
}

// SetRows replaces the row data and resets cursor/offset if needed.
func (t *tableModel) SetRows(rows [][]string) {
	t.rows = rows
	if t.cursor >= len(rows) {
		t.cursor = 0
	}
	t.offset = 0
}

// SetSize updates the table dimensions.
func (t *tableModel) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// visibleRowCount returns how many rows can be displayed.
func (t tableModel) visibleRowCount() int {
	if t.height <= 1 {
		return 10 // default fallback
	}
	return t.height
}

// sortRows sorts the table rows by the current sort column.
func (t *tableModel) sortRows() {
	if t.sortCol < 0 || t.sortCol >= len(t.columns) {
		return
	}
	col := t.sortCol
	desc := t.sortDesc
	sort.SliceStable(t.rows, func(i, j int) bool {
		a := ""
		b := ""
		if col < len(t.rows[i]) {
			a = t.rows[i][col]
		}
		if col < len(t.rows[j]) {
			b = t.rows[j][col]
		}
		// Try numeric comparison first (for sizes, counts, percentages)
		aNum, aErr := parseNumeric(a)
		bNum, bErr := parseNumeric(b)
		if aErr == nil && bErr == nil {
			if desc {
				return aNum > bNum
			}
			return aNum < bNum
		}
		// Fall back to string comparison
		if desc {
			return a > b
		}
		return a < b
	})
}

// parseNumeric tries to parse a string as a number, stripping common suffixes.
func parseNumeric(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	s = strings.ReplaceAll(s, ",", "")

	// Handle humanized sizes like "1.5 GB", "200 MB", "45 kB"
	multipliers := map[string]float64{
		"B":  1,
		"kB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}
	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, " "+suffix) {
			numStr := strings.TrimSuffix(s, " "+suffix)
			numStr = strings.TrimSpace(numStr)
			f, err := strconv.ParseFloat(numStr, 64)
			if err == nil {
				return f * mult, nil
			}
		}
	}

	return strconv.ParseFloat(s, 64)
}

// truncate truncates a string to fit within width, adding ellipsis if needed.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(s) <= width {
		return s
	}
	if width <= 1 {
		return "\u2026"
	}
	return s[:width-1] + "\u2026"
}
