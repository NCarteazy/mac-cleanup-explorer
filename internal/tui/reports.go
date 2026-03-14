package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
	"github.com/nick/mac-cleanup-explorer/internal/scanner"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// navigateToDetailMsg requests the app to show a detail view for an item.
type navigateToDetailMsg struct {
	item analyzer.ReportItem
}

// navigateBackMsg requests the app to go back to the previous view.
type navigateBackMsg struct{}

// sidebarItem represents a report entry in the sidebar.
type sidebarItem struct {
	name      string
	count     int
	totalSize int64
}

// reportsModel is the two-pane report browser.
type reportsModel struct {
	reports    map[string][]analyzer.ReportItem
	reportList []analyzer.Report
	scanResult *scanner.ScanResult

	sidebarItems  []sidebarItem
	sidebarCursor int
	activePane    int // 0 = sidebar, 1 = table

	table tableModel

	// For space treemap drill-down
	navStack    []*scanner.FileNode
	currentNode *scanner.FileNode

	width, height int
}

// newReportsModel creates a report browser initialized with scan data.
func newReportsModel(
	reports map[string][]analyzer.ReportItem,
	reportList []analyzer.Report,
	scanResult *scanner.ScanResult,
	width, height int,
) reportsModel {
	m := reportsModel{
		reports:    reports,
		reportList: reportList,
		scanResult: scanResult,
		width:      width,
		height:     height,
	}

	// Build sidebar items
	m.sidebarItems = m.buildSidebarItems()

	// Initialize table with first report if available
	if len(m.sidebarItems) > 0 {
		m.table = m.buildTable(0)
	}

	return m
}

// buildSidebarItems creates sidebar entries from the report list.
func (m reportsModel) buildSidebarItems() []sidebarItem {
	var items []sidebarItem
	for _, r := range m.reportList {
		reportItems := m.reports[r.Name()]
		var totalSize int64
		for _, item := range reportItems {
			totalSize += item.Size
		}
		items = append(items, sidebarItem{
			name:      r.Description(),
			count:     len(reportItems),
			totalSize: totalSize,
		})
	}
	return items
}

// reportNameForIndex returns the report key for a sidebar index.
func (m reportsModel) reportNameForIndex(idx int) string {
	if idx < 0 || idx >= len(m.reportList) {
		return ""
	}
	return m.reportList[idx].Name()
}

// buildTable creates a table model for the report at the given sidebar index.
func (m reportsModel) buildTable(sidebarIdx int) tableModel {
	reportName := m.reportNameForIndex(sidebarIdx)
	items := m.reports[reportName]

	cols, rows := m.columnsAndRows(reportName, items)

	tbl := newTableModel(cols, rows)

	// Calculate available table height: total height minus breadcrumbs (1), status bar (1), borders (4)
	tableHeight := m.height - 6
	if tableHeight < 5 {
		tableHeight = 5
	}
	sidebarWidth := m.sidebarWidth()
	tableWidth := m.width - sidebarWidth - 6 // account for panel borders/padding
	if tableWidth < 40 {
		tableWidth = 40
	}
	tbl.SetSize(tableWidth, tableHeight)

	return tbl
}

// columnsAndRows returns the columns and row data for a given report type.
func (m reportsModel) columnsAndRows(reportName string, items []analyzer.ReportItem) ([]column, [][]string) {
	switch reportName {
	case "space":
		return m.spaceColumnsAndRows(items)
	case "large_files":
		cols := []column{
			{Name: "Path", Width: 40, Align: lipgloss.Left},
			{Name: "Size", Width: 12, Align: lipgloss.Right},
			{Name: "Modified", Width: 18, Align: lipgloss.Left},
			{Name: "Type", Width: 16, Align: lipgloss.Left},
		}
		rows := make([][]string, len(items))
		for i, item := range items {
			rows[i] = []string{
				shortenPath(item.Path),
				theme.FormatSize(item.Size),
				item.LastMod,
				item.Description,
			}
		}
		return cols, rows
	case "stale":
		cols := []column{
			{Name: "Path", Width: 40, Align: lipgloss.Left},
			{Name: "Size", Width: 12, Align: lipgloss.Right},
			{Name: "Last Accessed", Width: 18, Align: lipgloss.Left},
			{Name: "Category", Width: 18, Align: lipgloss.Left},
		}
		rows := make([][]string, len(items))
		for i, item := range items {
			rows[i] = []string{
				shortenPath(item.Path),
				theme.FormatSize(item.Size),
				item.LastAccess,
				item.Category,
			}
		}
		return cols, rows
	case "caches":
		cols := []column{
			{Name: "Path", Width: 44, Align: lipgloss.Left},
			{Name: "Size", Width: 12, Align: lipgloss.Right},
			{Name: "Category", Width: 22, Align: lipgloss.Left},
		}
		rows := make([][]string, len(items))
		for i, item := range items {
			rows[i] = []string{
				shortenPath(item.Path),
				theme.FormatSize(item.Size),
				item.Category,
			}
		}
		return cols, rows
	case "devbloat":
		cols := []column{
			{Name: "Path", Width: 44, Align: lipgloss.Left},
			{Name: "Size", Width: 12, Align: lipgloss.Right},
			{Name: "Type", Width: 22, Align: lipgloss.Left},
		}
		rows := make([][]string, len(items))
		for i, item := range items {
			rows[i] = []string{
				shortenPath(item.Path),
				theme.FormatSize(item.Size),
				item.Category,
			}
		}
		return cols, rows
	case "leftovers":
		cols := []column{
			{Name: "Path", Width: 44, Align: lipgloss.Left},
			{Name: "Size", Width: 12, Align: lipgloss.Right},
			{Name: "App Name", Width: 22, Align: lipgloss.Left},
		}
		rows := make([][]string, len(items))
		for i, item := range items {
			rows[i] = []string{
				shortenPath(item.Path),
				theme.FormatSize(item.Size),
				item.Description,
			}
		}
		return cols, rows
	case "duplicates":
		cols := []column{
			{Name: "Path", Width: 40, Align: lipgloss.Left},
			{Name: "Size", Width: 12, Align: lipgloss.Right},
			{Name: "Copies", Width: 26, Align: lipgloss.Left},
		}
		rows := make([][]string, len(items))
		for i, item := range items {
			rows[i] = []string{
				shortenPath(item.Path),
				theme.FormatSize(item.Size),
				item.Description,
			}
		}
		return cols, rows
	default:
		// Generic fallback
		cols := []column{
			{Name: "Path", Width: 50, Align: lipgloss.Left},
			{Name: "Size", Width: 12, Align: lipgloss.Right},
			{Name: "Category", Width: 20, Align: lipgloss.Left},
		}
		rows := make([][]string, len(items))
		for i, item := range items {
			rows[i] = []string{
				shortenPath(item.Path),
				theme.FormatSize(item.Size),
				item.Category,
			}
		}
		return cols, rows
	}
}

// spaceColumnsAndRows returns columns/rows for the space treemap.
// If currentNode is set (drill-down), it shows children of that node.
func (m reportsModel) spaceColumnsAndRows(items []analyzer.ReportItem) ([]column, [][]string) {
	cols := []column{
		{Name: "Name", Width: 30, Align: lipgloss.Left},
		{Name: "Size", Width: 12, Align: lipgloss.Right},
		{Name: "%", Width: 8, Align: lipgloss.Right},
		{Name: "Files", Width: 10, Align: lipgloss.Right},
		{Name: "Category", Width: 18, Align: lipgloss.Left},
	}

	if m.currentNode != nil {
		// Drill-down mode: show children of currentNode
		totalSize := m.currentNode.Size
		if totalSize == 0 {
			totalSize = 1
		}
		children := m.currentNode.Children
		rows := make([][]string, len(children))
		for i, child := range children {
			pct := float64(child.Size) / float64(totalSize) * 100
			fileCount := fmt.Sprintf("%d", child.FileCount)
			cat := ""
			if child.IsDir {
				cat = "[dir]"
			} else {
				cat = filepath.Ext(child.Name)
			}
			rows[i] = []string{
				child.Name,
				theme.FormatSize(child.Size),
				fmt.Sprintf("%.1f%%", pct),
				fileCount,
				cat,
			}
		}
		return cols, rows
	}

	// Top-level: use report items
	var totalSize int64
	for _, item := range items {
		totalSize += item.Size
	}
	if totalSize == 0 {
		totalSize = 1
	}

	rows := make([][]string, len(items))
	for i, item := range items {
		pct := float64(item.Size) / float64(totalSize) * 100
		name := filepath.Base(item.Path)
		rows[i] = []string{
			name,
			theme.FormatSize(item.Size),
			fmt.Sprintf("%.1f%%", pct),
			fmt.Sprintf("%d", item.FileCount),
			item.Category,
		}
	}
	return cols, rows
}

// Update handles messages for the reports model.
func (m reportsModel) Update(msg tea.Msg) (reportsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table = m.buildTable(m.sidebarCursor)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.activePane = 1 - m.activePane // toggle 0 <-> 1
			return m, nil

		case "esc", "backspace":
			if m.activePane == 1 {
				// In table pane
				reportName := m.reportNameForIndex(m.sidebarCursor)
				if reportName == "space" && m.currentNode != nil {
					// Pop navigation stack
					if len(m.navStack) > 0 {
						m.currentNode = m.navStack[len(m.navStack)-1]
						m.navStack = m.navStack[:len(m.navStack)-1]
					} else {
						m.currentNode = nil
					}
					m.table = m.buildTable(m.sidebarCursor)
					return m, nil
				}
				// Switch to sidebar
				m.activePane = 0
				return m, nil
			}
			// In sidebar: navigate back to dashboard
			return m, func() tea.Msg { return navigateBackMsg{} }

		case "enter":
			if m.activePane == 0 {
				// Select report from sidebar -> update table
				m.currentNode = nil
				m.navStack = nil
				m.table = m.buildTable(m.sidebarCursor)
				m.activePane = 1
				return m, nil
			}
			// In table: enter on a row
			reportName := m.reportNameForIndex(m.sidebarCursor)
			if reportName == "space" {
				return m.handleSpaceDrillDown()
			}
			// For other reports, navigate to detail
			items := m.reports[reportName]
			rowIdx := m.table.SelectedRow()
			if rowIdx >= 0 && rowIdx < len(items) {
				item := items[rowIdx]
				return m, func() tea.Msg { return navigateToDetailMsg{item: item} }
			}
			return m, nil

		case "up", "k", "down", "j", "home", "g", "end", "G", "s":
			if m.activePane == 0 {
				// Sidebar navigation
				switch msg.String() {
				case "up", "k":
					if m.sidebarCursor > 0 {
						m.sidebarCursor--
					}
				case "down", "j":
					if m.sidebarCursor < len(m.sidebarItems)-1 {
						m.sidebarCursor++
					}
				case "home", "g":
					m.sidebarCursor = 0
				case "end", "G":
					if len(m.sidebarItems) > 0 {
						m.sidebarCursor = len(m.sidebarItems) - 1
					}
				}
				return m, nil
			}
			// Table navigation
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// handleSpaceDrillDown handles entering a directory in the space treemap.
func (m reportsModel) handleSpaceDrillDown() (reportsModel, tea.Cmd) {
	rowIdx := m.table.SelectedRow()

	if m.currentNode != nil {
		// Currently drilled-down: drill into a child
		children := m.currentNode.Children
		if rowIdx >= 0 && rowIdx < len(children) {
			child := children[rowIdx]
			if child.IsDir && len(child.Children) > 0 {
				m.navStack = append(m.navStack, m.currentNode)
				m.currentNode = child
				m.table = m.buildTable(m.sidebarCursor)
			}
		}
		return m, nil
	}

	// Top-level: drill into a root child
	if m.scanResult != nil && m.scanResult.Root != nil {
		items := m.reports["space"]
		if rowIdx >= 0 && rowIdx < len(items) {
			item := items[rowIdx]
			// Find the FileNode matching this item
			for _, child := range m.scanResult.Root.Children {
				if child.Path == item.Path && child.IsDir && len(child.Children) > 0 {
					m.navStack = append(m.navStack, nil) // nil means top-level
					m.currentNode = child
					m.table = m.buildTable(m.sidebarCursor)
					break
				}
			}
		}
	}
	return m, nil
}

// View renders the two-pane reports browser.
func (m reportsModel) View() string {
	if m.width <= 0 {
		return "Reports loading..."
	}

	// Breadcrumbs
	breadcrumbs := m.renderBreadcrumbs()

	// Status bar
	statusBar := m.renderStatusBar()

	// Calculate available height for panes
	contentHeight := m.height - 3 // breadcrumbs + status bar + spacing
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Sidebar
	sidebarWidth := m.sidebarWidth()
	sidebar := m.renderSidebar(sidebarWidth, contentHeight)

	// Table pane
	tableWidth := m.width - sidebarWidth - 2 // gap between panes
	if tableWidth < 20 {
		tableWidth = 20
	}
	tablePane := m.renderTablePane(tableWidth, contentHeight)

	// Join panes horizontally
	panes := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", tablePane)

	// Join everything vertically
	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumbs,
		panes,
		statusBar,
	)
}

// renderBreadcrumbs renders the breadcrumb bar at the top.
func (m reportsModel) renderBreadcrumbs() string {
	parts := []string{"Dashboard", "Reports"}

	if m.sidebarCursor < len(m.sidebarItems) {
		parts = append(parts, m.sidebarItems[m.sidebarCursor].name)
	}

	// Add space treemap drill-down path
	reportName := m.reportNameForIndex(m.sidebarCursor)
	if reportName == "space" && m.currentNode != nil {
		// Show the full drill-down path
		var pathParts []string
		for _, node := range m.navStack {
			if node != nil {
				pathParts = append(pathParts, node.Name)
			}
		}
		pathParts = append(pathParts, m.currentNode.Name)
		for _, p := range pathParts {
			parts = append(parts, p)
		}
	}

	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor)).
		Render(" > ")

	styledParts := make([]string, len(parts))
	for i, p := range parts {
		styledParts[i] = theme.BreadcrumbStyle.Render(p)
	}

	return strings.Join(styledParts, sep)
}

// renderStatusBar renders the help/status bar at the bottom.
func (m reportsModel) renderStatusBar() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.OverlayColor))

	sep := sepStyle.Render(" | ")

	hints := []string{
		keyStyle.Render("tab") + " " + descStyle.Render("Switch Pane"),
		keyStyle.Render("enter") + " " + descStyle.Render("Select"),
		keyStyle.Render("esc") + " " + descStyle.Render("Back"),
		keyStyle.Render("s") + " " + descStyle.Render("Sort"),
		keyStyle.Render("?") + " " + descStyle.Render("Help"),
	}

	bar := theme.StatusBarStyle.Copy().
		Width(m.width).
		Align(lipgloss.Center).
		Render(strings.Join(hints, sep))

	return bar
}

// sidebarWidth returns the width of the sidebar pane.
func (m reportsModel) sidebarWidth() int {
	w := m.width / 4
	if w < 28 {
		w = 28
	}
	if w > 40 {
		w = 40
	}
	return w
}

// renderSidebar renders the left sidebar pane.
func (m reportsModel) renderSidebar(width, height int) string {
	var rows []string

	title := theme.TitleStyle.Render("Reports")
	rows = append(rows, title, "")

	for i, item := range m.sidebarItems {
		nameStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextColor))
		countStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.MutedColor))

		if i == m.sidebarCursor {
			nameStyle = theme.SelectedRowStyle.Copy()
			countStyle = theme.SelectedRowStyle.Copy()
		}

		// Truncate name to fit
		maxNameWidth := width - 8
		if maxNameWidth < 10 {
			maxNameWidth = 10
		}
		name := truncate(item.name, maxNameWidth)

		line := nameStyle.Render(name)
		countText := fmt.Sprintf(" (%d)", item.count)
		line += countStyle.Render(countText)

		rows = append(rows, line)

		// Show total size under each item
		sizeText := theme.FormatSize(item.totalSize)
		sizeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.MutedColor))
		rows = append(rows, sizeStyle.Render("  "+sizeText))
	}

	content := strings.Join(rows, "\n")

	panelStyle := theme.PanelStyle.Copy()
	if m.activePane == 0 {
		panelStyle = theme.ActivePanelStyle.Copy()
	}

	return panelStyle.
		Width(width).
		Height(height).
		Render(content)
}

// renderTablePane renders the right table pane.
func (m reportsModel) renderTablePane(width, height int) string {
	// Update table dimensions for rendering
	tableHeight := height - 4 // account for panel padding
	if tableHeight < 3 {
		tableHeight = 3
	}
	m.table.SetSize(width-6, tableHeight)

	var content string
	if len(m.table.rows) == 0 {
		content = theme.MutedStyle.Render("No items in this report.")
	} else {
		content = m.table.View()
	}

	panelStyle := theme.PanelStyle.Copy()
	if m.activePane == 1 {
		panelStyle = theme.ActivePanelStyle.Copy()
	}

	return panelStyle.
		Width(width).
		Height(height).
		Render(content)
}

// shortenPath shortens a file path for table display.
func shortenPath(path string) string {
	// Try to shorten home directory
	home := "~"
	if strings.Contains(path, "/Users/") {
		parts := strings.SplitN(path, "/Users/", 2)
		if len(parts) == 2 {
			userParts := strings.SplitN(parts[1], "/", 2)
			if len(userParts) == 2 {
				path = home + "/" + userParts[1]
			}
		}
	}
	return path
}
