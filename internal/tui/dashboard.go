package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
	"github.com/nick/mac-cleanup-explorer/internal/scanner"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// categoryEntry holds aggregated data for a single dashboard category.
type categoryEntry struct {
	name  string
	size  int64
	ratio float64
}

// dashboardModel is the Bubble Tea model for the dashboard view.
type dashboardModel struct {
	scanResult *scanner.ScanResult
	reports    map[string][]analyzer.ReportItem
	width      int
	height     int
	scrollY    int
	categories []categoryEntry
}

// categoryOrder defines the canonical display order.
var categoryOrder = []string{
	"Applications",
	"System & App Data",
	"Documents",
	"Downloads",
	"Media",
	"Developer",
	"Trash",
	"Other",
}

// newDashboardModel creates and initializes a dashboardModel.
func newDashboardModel(sr *scanner.ScanResult, width, height int) dashboardModel {
	m := dashboardModel{
		scanResult: sr,
		width:      width,
		height:     height,
	}
	m.categories = m.computeCategories()
	return m
}

func (m dashboardModel) Init() tea.Cmd {
	return nil
}

func (m dashboardModel) Update(msg tea.Msg) (dashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "down", "j":
			m.scrollY++
		}
	}
	return m, nil
}

func (m dashboardModel) View() string {
	if m.scanResult == nil {
		return "No scan data."
	}

	maxWidth := m.width
	if maxWidth <= 0 {
		maxWidth = 80
	}
	// Content area: leave a margin on each side
	contentWidth := maxWidth - 4
	if contentWidth < 40 {
		contentWidth = 40
	}
	if contentWidth > 100 {
		contentWidth = 100
	}

	sections := []string{
		m.renderSummaryCard(contentWidth),
		"",
		m.renderCategoryBreakdown(contentWidth),
		"",
		m.renderQuickInsights(contentWidth),
		"",
		m.renderNavHint(maxWidth),
	}

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	// Simple scroll: split into lines and slice
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Available height is the terminal height
	availHeight := m.height
	if availHeight <= 0 {
		availHeight = 40
	}

	// Clamp scroll
	maxScroll := totalLines - availHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollY > maxScroll {
		m.scrollY = maxScroll
	}

	start := m.scrollY
	end := start + availHeight
	if end > totalLines {
		end = totalLines
	}

	visible := strings.Join(lines[start:end], "\n")

	// Center horizontally
	return lipgloss.Place(maxWidth, availHeight, lipgloss.Center, lipgloss.Top, visible)
}

// ---------- Summary Card ----------

func (m dashboardModel) renderSummaryCard(width int) string {
	sr := m.scanResult

	title := theme.TitleStyle.Render("  Scan Summary")

	scanPath := sr.Root.Path
	maxPathLen := width - 20
	if maxPathLen < 20 {
		maxPathLen = 20
	}
	if len(scanPath) > maxPathLen {
		scanPath = "..." + scanPath[len(scanPath)-(maxPathLen-3):]
	}

	labelStyle := theme.MutedStyle.Copy()
	valueStyle := theme.TextStyle.Copy()

	// Build rows
	rows := []string{
		fmt.Sprintf("  %s  %s",
			labelStyle.Render("Path:"),
			valueStyle.Render(scanPath)),
		fmt.Sprintf("  %s  %s",
			labelStyle.Render("Total Size:"),
			theme.SuccessStyle.Render(theme.FormatSize(sr.Root.Size))),
		fmt.Sprintf("  %s  %s    %s  %s    %s  %s",
			labelStyle.Render("Files:"),
			valueStyle.Render(humanize.Comma(sr.TotalFiles)),
			labelStyle.Render("Dirs:"),
			valueStyle.Render(humanize.Comma(sr.TotalDirs)),
			labelStyle.Render("Duration:"),
			valueStyle.Render(sr.Duration.Round(1e6).String())),
	}

	if len(sr.Errors) > 0 {
		errText := fmt.Sprintf("%d errors during scan", len(sr.Errors))
		rows = append(rows, fmt.Sprintf("  %s  %s",
			labelStyle.Render("Errors:"),
			theme.DangerStyle.Render(errText)))
	}

	body := strings.Join(rows, "\n")

	panel := theme.PanelStyle.Copy().
		Width(width).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))

	return panel
}

// ---------- Category Breakdown ----------

func (m dashboardModel) computeCategories() []categoryEntry {
	if m.scanResult == nil || m.scanResult.Root == nil {
		return nil
	}

	root := m.scanResult.Root
	totalSize := root.Size
	if totalSize == 0 {
		totalSize = 1
	}

	// Aggregate children into categories using the same logic as space report
	catSizes := make(map[string]int64)
	for _, child := range root.Children {
		cat := categorizeDirName(child.Name)
		catSizes[cat] += child.Size
	}

	// Build entries in canonical order
	var entries []categoryEntry
	for _, name := range categoryOrder {
		sz, ok := catSizes[name]
		if !ok || sz == 0 {
			continue
		}
		entries = append(entries, categoryEntry{
			name:  name,
			size:  sz,
			ratio: float64(sz) / float64(totalSize),
		})
	}

	// Sort by size descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].size > entries[j].size
	})

	return entries
}

// categorizeDirName maps a top-level directory name to a dashboard category.
func categorizeDirName(name string) string {
	lower := strings.ToLower(name)
	switch {
	case lower == "library":
		return "System & App Data"
	case lower == "applications":
		return "Applications"
	case lower == "documents":
		return "Documents"
	case lower == "downloads":
		return "Downloads"
	case lower == "pictures" || lower == "movies" || lower == "music":
		return "Media"
	case lower == "developer" || lower == "projects" || lower == "code":
		return "Developer"
	case lower == ".trash":
		return "Trash"
	default:
		return "Other"
	}
}

func (m dashboardModel) renderCategoryBreakdown(width int) string {
	if len(m.categories) == 0 {
		return ""
	}

	title := theme.TitleStyle.Render("  Category Breakdown")

	barMax := 25
	// Adjust bar width for narrow terminals
	if width < 70 {
		barMax = 15
	}

	// Column widths
	nameWidth := 20
	sizeWidth := 10
	pctWidth := 7

	var rows []string
	for _, cat := range m.categories {
		pct := cat.ratio * 100

		nameStr := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextColor)).
			Width(nameWidth).
			Render(cat.name)

		bar := theme.SizeBar(width, cat.ratio, barMax)

		sizeStr := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextColor)).
			Width(sizeWidth).
			Align(lipgloss.Right).
			Render(theme.FormatSize(cat.size))

		pctColor := theme.SizeBarColor(cat.ratio)
		pctStr := lipgloss.NewStyle().
			Foreground(lipgloss.Color(pctColor)).
			Width(pctWidth).
			Align(lipgloss.Right).
			Render(fmt.Sprintf("%.1f%%", pct))

		row := fmt.Sprintf("  %s  %s  %s  %s", nameStr, bar, sizeStr, pctStr)
		rows = append(rows, row)
	}

	body := strings.Join(rows, "\n")

	panel := theme.PanelStyle.Copy().
		Width(width).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))

	return panel
}

// ---------- Quick Insights ----------

func (m dashboardModel) renderQuickInsights(width int) string {
	if m.reports == nil {
		return theme.MutedStyle.Render("  Analyzing reports...")
	}

	title := theme.TitleStyle.Render("  Quick Insights")

	// Accent colors for different insight bullets
	type insight struct {
		icon  string
		text  string
		color string
	}

	var insights []insight

	// 1. Items over 1 GB
	largeItems := m.reports["large_files"]
	var countOver1GB int
	for _, item := range largeItems {
		if item.Size >= 1024*1024*1024 {
			countOver1GB++
		}
	}
	if countOver1GB > 0 {
		insights = append(insights, insight{
			icon:  "\u25cf",
			text:  fmt.Sprintf("%d item(s) over 1 GB", countOver1GB),
			color: theme.DangerColor,
		})
	}

	// 2. Total cache size
	cacheItems := m.reports["caches"]
	var cacheTotal int64
	for _, item := range cacheItems {
		cacheTotal += item.Size
	}
	if cacheTotal > 0 {
		cacheGB := float64(cacheTotal) / (1024 * 1024 * 1024)
		insights = append(insights, insight{
			icon:  "\u25cf",
			text:  fmt.Sprintf("%.1f GB in caches", cacheGB),
			color: theme.WarningColor,
		})
	}

	// 3. Developer tools bloat
	devItems := m.reports["devbloat"]
	var devTotal int64
	for _, item := range devItems {
		devTotal += item.Size
	}
	if devTotal > 0 {
		devGB := float64(devTotal) / (1024 * 1024 * 1024)
		insights = append(insights, insight{
			icon:  "\u25cf",
			text:  fmt.Sprintf("%.1f GB in developer tools", devGB),
			color: theme.SecondaryColor,
		})
	}

	// 4. Stale items
	staleItems := m.reports["stale"]
	if len(staleItems) > 0 {
		var staleTotal int64
		for _, item := range staleItems {
			staleTotal += item.Size
		}
		staleGB := float64(staleTotal) / (1024 * 1024 * 1024)
		insights = append(insights, insight{
			icon:  "\u25cf",
			text:  fmt.Sprintf("%d stale items not accessed in 6+ months (%.1f GB)", len(staleItems), staleGB),
			color: theme.PrimaryColor,
		})
	}

	if len(insights) == 0 {
		noInsights := theme.MutedStyle.Render("  No notable findings.")
		panel := theme.PanelStyle.Copy().
			Width(width).
			Render(lipgloss.JoinVertical(lipgloss.Left, title, "", noInsights))
		return panel
	}

	var rows []string
	for _, ins := range insights {
		bullet := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ins.color)).
			Render(ins.icon)
		text := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextColor)).
			Render(ins.text)
		rows = append(rows, fmt.Sprintf("  %s  %s", bullet, text))
	}

	body := strings.Join(rows, "\n")

	panel := theme.PanelStyle.Copy().
		Width(width).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))

	return panel
}

// ---------- Navigation Hint ----------

func (m dashboardModel) renderNavHint(fullWidth int) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.OverlayColor))

	sep := sepStyle.Render("  |  ")

	hint := strings.Join([]string{
		keyStyle.Render("enter") + " " + descStyle.Render("Browse Reports"),
		keyStyle.Render("e") + " " + descStyle.Render("Export"),
		keyStyle.Render("q") + " " + descStyle.Render("Quit"),
		keyStyle.Render("?") + " " + descStyle.Render("Help"),
	}, sep)

	bar := theme.StatusBarStyle.Copy().
		Width(fullWidth).
		Align(lipgloss.Center).
		Render(hint)

	return bar
}
