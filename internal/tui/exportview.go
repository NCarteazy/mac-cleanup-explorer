package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
	"github.com/nick/mac-cleanup-explorer/internal/export"
	"github.com/nick/mac-cleanup-explorer/internal/scanner"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// navigateToExportMsg requests the app to show the export panel.
type navigateToExportMsg struct{}

// exportModel is the export panel overlay.
type exportModel struct {
	reports     map[string][]analyzer.ReportItem
	reportList  []analyzer.Report
	selected    map[string]bool // which reports are selected
	cursor      int
	sortedNames []string // report names in consistent order

	preview       string // formatted markdown preview
	previewScroll int

	activePane int // 0=checkbox list, 1=preview
	width      int
	height     int

	// System info for export header
	sysInfo export.SystemInfo

	// Flash message
	flashMsg   string
	flashTimer int

	// Where we came from
	previousView viewState
}

// newExportModel creates a new export panel model.
func newExportModel(
	reports map[string][]analyzer.ReportItem,
	reportList []analyzer.Report,
	scanResult *scanner.ScanResult,
	scanPath string,
	width, height int,
	previousView viewState,
) exportModel {
	// Build sorted names from report list
	sortedNames := make([]string, 0, len(reportList))
	for _, r := range reportList {
		sortedNames = append(sortedNames, r.Name())
	}
	sort.Strings(sortedNames)

	// Start with all reports selected
	selected := make(map[string]bool, len(sortedNames))
	for _, name := range sortedNames {
		selected[name] = true
	}

	sysInfo := gatherSystemInfo(scanResult, scanPath)

	m := exportModel{
		reports:      reports,
		reportList:   reportList,
		selected:     selected,
		sortedNames:  sortedNames,
		width:        width,
		height:       height,
		sysInfo:      sysInfo,
		previousView: previousView,
	}

	m.regeneratePreview()
	return m
}

// gatherSystemInfo populates SystemInfo from the scan result.
func gatherSystemInfo(scanResult *scanner.ScanResult, scanPath string) export.SystemInfo {
	info := export.SystemInfo{
		ScanScope: scanPath,
	}

	if scanResult != nil {
		info.ScanTime = scanResult.Duration.Round(time.Millisecond).String()
		if scanResult.Root != nil {
			info.DiskSize = theme.FormatSize(scanResult.Root.Size)
		}
	}

	return info
}

// regeneratePreview rebuilds the preview text based on selected reports.
func (m *exportModel) regeneratePreview() {
	selectedReports := make(map[string][]analyzer.ReportItem)
	for name, isSelected := range m.selected {
		if isSelected {
			if items, ok := m.reports[name]; ok {
				selectedReports[name] = items
			}
		}
	}

	if len(selectedReports) == 0 {
		m.preview = "(No reports selected)"
		return
	}

	m.preview = export.FormatMultipleReports(selectedReports, m.sysInfo)
}

// reportDescription returns the human-readable description for a report name.
func (m exportModel) reportDescription(name string) string {
	for _, r := range m.reportList {
		if r.Name() == name {
			return r.Description()
		}
	}
	return name
}

// Update handles messages for the export model.
func (m exportModel) Update(msg tea.Msg) (exportModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case flashTickMsg:
		m.flashTimer--
		if m.flashTimer <= 0 {
			m.flashMsg = ""
			return m, nil
		}
		return m, flashTickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return navigateBackMsg{} }

		case "tab":
			m.activePane = 1 - m.activePane
			return m, nil

		case " ":
			if m.activePane == 0 && m.cursor >= 0 && m.cursor < len(m.sortedNames) {
				name := m.sortedNames[m.cursor]
				m.selected[name] = !m.selected[name]
				m.regeneratePreview()
				m.previewScroll = 0
			}
			return m, nil

		case "a":
			if m.activePane == 0 {
				for _, name := range m.sortedNames {
					m.selected[name] = true
				}
				m.regeneratePreview()
				m.previewScroll = 0
			}
			return m, nil

		case "n":
			if m.activePane == 0 {
				for _, name := range m.sortedNames {
					m.selected[name] = false
				}
				m.regeneratePreview()
				m.previewScroll = 0
			}
			return m, nil

		case "up", "k":
			if m.activePane == 0 {
				if m.cursor > 0 {
					m.cursor--
				}
			} else {
				if m.previewScroll > 0 {
					m.previewScroll--
				}
			}
			return m, nil

		case "down", "j":
			if m.activePane == 0 {
				if m.cursor < len(m.sortedNames)-1 {
					m.cursor++
				}
			} else {
				m.previewScroll++
			}
			return m, nil

		case "c":
			err := clipboard.WriteAll(m.preview)
			if err != nil {
				m.flashMsg = "Clipboard unavailable"
			} else {
				m.flashMsg = "Copied to clipboard!"
			}
			m.flashTimer = 20
			return m, flashTickCmd()

		case "s":
			filename := fmt.Sprintf("mac-cleanup-report-%s.md", time.Now().Format("2006-01-02"))
			home, err := os.UserHomeDir()
			if err != nil {
				home = "."
			}
			savePath := filepath.Join(home, filename)
			err = os.WriteFile(savePath, []byte(m.preview), 0644)
			if err != nil {
				m.flashMsg = fmt.Sprintf("Error: %v", err)
			} else {
				m.flashMsg = fmt.Sprintf("Saved to ~/%s", filename)
			}
			m.flashTimer = 30
			return m, flashTickCmd()
		}
	}

	return m, nil
}

// View renders the export panel.
func (m exportModel) View() string {
	if m.width <= 0 {
		return "Loading..."
	}

	// Title
	titleBar := theme.TitleStyle.Render("Export Report")

	// Status bar
	statusBar := m.renderStatusBar()

	// Calculate available height for panes
	contentHeight := m.height - 4
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Left pane: checkbox list
	leftWidth := m.width / 3
	if leftWidth < 30 {
		leftWidth = 30
	}
	if leftWidth > 50 {
		leftWidth = 50
	}
	leftPane := m.renderCheckboxList(leftWidth, contentHeight)

	// Right pane: preview
	rightWidth := m.width - leftWidth - 3
	if rightWidth < 30 {
		rightWidth = 30
	}
	rightPane := m.renderPreview(rightWidth, contentHeight)

	// Join panes horizontally
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)

	// Flash message
	var flashLine string
	if m.flashMsg != "" {
		flashLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SuccessColor)).
			Bold(true).
			Render("  " + m.flashMsg)
	}

	parts := []string{titleBar, "", panes}
	if flashLine != "" {
		parts = append(parts, flashLine)
	}
	parts = append(parts, statusBar)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderCheckboxList renders the left pane with report checkboxes.
func (m exportModel) renderCheckboxList(width, height int) string {
	var rows []string

	title := theme.SubtitleStyle.Render("Select Reports")
	rows = append(rows, title, "")

	for i, name := range m.sortedNames {
		items := m.reports[name]
		var totalSize int64
		for _, item := range items {
			totalSize += item.Size
		}

		checkbox := "[ ]"
		if m.selected[name] {
			checkbox = "[x]"
		}

		desc := m.reportDescription(name)
		info := fmt.Sprintf("(%d items, %s)", len(items), theme.FormatSize(totalSize))

		lineStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextColor))
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.MutedColor))

		if i == m.cursor && m.activePane == 0 {
			lineStyle = theme.SelectedRowStyle.Copy()
			infoStyle = theme.SelectedRowStyle.Copy()
		}

		line := lineStyle.Render(fmt.Sprintf(" %s %s", checkbox, desc))
		infoLine := infoStyle.Render(fmt.Sprintf("     %s", info))

		rows = append(rows, line, infoLine)
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

// renderPreview renders the right pane with the markdown preview.
func (m exportModel) renderPreview(width, height int) string {
	title := theme.SubtitleStyle.Render("Preview")

	previewLines := strings.Split(m.preview, "\n")

	// Clamp scroll
	visibleLines := height - 6
	if visibleLines < 5 {
		visibleLines = 5
	}
	maxScroll := len(previewLines) - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	scroll := m.previewScroll
	if scroll > maxScroll {
		scroll = maxScroll
	}

	end := scroll + visibleLines
	if end > len(previewLines) {
		end = len(previewLines)
	}

	visible := previewLines[scroll:end]

	previewStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextColor))

	// Truncate lines to fit width
	maxLineWidth := width - 6
	if maxLineWidth < 20 {
		maxLineWidth = 20
	}
	for i, line := range visible {
		if len(line) > maxLineWidth {
			visible[i] = line[:maxLineWidth-1] + "\u2026"
		}
	}

	previewText := previewStyle.Render(strings.Join(visible, "\n"))

	// Token estimate
	tokenEst := export.EstimateTokens(m.preview)
	tokenLine := theme.MutedStyle.Render(fmt.Sprintf("~%d tokens", tokenEst))

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", previewText, "", tokenLine)

	panelStyle := theme.PanelStyle.Copy()
	if m.activePane == 1 {
		panelStyle = theme.ActivePanelStyle.Copy()
	}

	return panelStyle.
		Width(width).
		Height(height).
		Render(content)
}

// renderStatusBar renders the action key hints at the bottom.
func (m exportModel) renderStatusBar() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.OverlayColor))

	sep := sepStyle.Render(" | ")

	hints := []string{
		keyStyle.Render("c") + " " + descStyle.Render("Copy to Clipboard"),
		keyStyle.Render("s") + " " + descStyle.Render("Save to File"),
		keyStyle.Render("tab") + " " + descStyle.Render("Switch Pane"),
		keyStyle.Render("esc") + " " + descStyle.Render("Back"),
	}

	bar := theme.StatusBarStyle.Copy().
		Width(m.width).
		Align(lipgloss.Center).
		Render(strings.Join(hints, sep))

	return bar
}
