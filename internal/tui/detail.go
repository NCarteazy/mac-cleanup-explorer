package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// detailModel displays detailed information about a single report item.
type detailModel struct {
	item       analyzer.ReportItem
	reportName string
	width      int
	height     int

	flashMsg   string
	flashTimer int
}

// newDetailModel creates a detail view for the given item.
func newDetailModel(item analyzer.ReportItem, reportName string, width, height int) detailModel {
	return detailModel{
		item:       item,
		reportName: reportName,
		width:      width,
		height:     height,
	}
}

// Update handles messages for the detail model.
func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
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
		case "c":
			err := clipboard.WriteAll(m.item.Path)
			if err != nil {
				m.flashMsg = "Clipboard unavailable"
			} else {
				m.flashMsg = "Copied!"
			}
			m.flashTimer = 15
			return m, flashTickCmd()
		}
	}

	return m, nil
}

// View renders the detail view.
func (m detailModel) View() string {
	if m.width <= 0 {
		return "Loading..."
	}

	// Breadcrumbs
	breadcrumbs := m.renderBreadcrumbs()

	// Status bar
	statusBar := m.renderStatusBar()

	// Main content panel
	contentHeight := m.height - 4
	if contentHeight < 10 {
		contentHeight = 10
	}
	contentWidth := m.width - 4
	if contentWidth > 80 {
		contentWidth = 80
	}
	if contentWidth < 40 {
		contentWidth = 40
	}

	panel := m.renderPanel(contentWidth, contentHeight)

	// Center the panel
	centeredPanel := lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Center, panel)

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumbs,
		centeredPanel,
		statusBar,
	)
}

// renderBreadcrumbs renders the breadcrumb bar at the top.
func (m detailModel) renderBreadcrumbs() string {
	parts := []string{"Dashboard", "Reports", m.reportName, "Detail"}

	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor)).
		Render(" > ")

	styledParts := make([]string, len(parts))
	for i, p := range parts {
		styledParts[i] = theme.BreadcrumbStyle.Render(p)
	}

	return strings.Join(styledParts, sep)
}

// renderPanel renders the main detail panel.
func (m detailModel) renderPanel(width, height int) string {
	item := m.item

	title := theme.TitleStyle.Render("Item Detail")

	labelStyle := theme.MutedStyle.Copy()
	valueStyle := theme.TextStyle.Copy()

	var rows []string

	// Path
	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true)
	rows = append(rows, fmt.Sprintf("  %s  %s",
		labelStyle.Render("Path:"),
		pathStyle.Render(item.Path)))

	// Size with severity-based coloring
	sizeStr := theme.FormatSize(item.Size)
	sizeStyle := sizeColorStyle(item.Severity)
	rows = append(rows, fmt.Sprintf("  %s  %s",
		labelStyle.Render("Size:"),
		sizeStyle.Render(sizeStr)))

	// Last Accessed
	if item.LastAccess != "" {
		rows = append(rows, fmt.Sprintf("  %s  %s",
			labelStyle.Render("Last Accessed:"),
			valueStyle.Render(item.LastAccess)))
	}

	// Last Modified
	if item.LastMod != "" {
		rows = append(rows, fmt.Sprintf("  %s  %s",
			labelStyle.Render("Last Modified:"),
			valueStyle.Render(item.LastMod)))
	}

	// Category
	if item.Category != "" {
		rows = append(rows, fmt.Sprintf("  %s  %s",
			labelStyle.Render("Category:"),
			valueStyle.Render(item.Category)))
	}

	// Severity badge
	if item.Severity != "" {
		badge := severityBadge(item.Severity)
		rows = append(rows, fmt.Sprintf("  %s  %s",
			labelStyle.Render("Severity:"),
			badge))
	}

	// Description
	if item.Description != "" {
		rows = append(rows, fmt.Sprintf("  %s  %s",
			labelStyle.Render("Description:"),
			valueStyle.Render(item.Description)))
	}

	// File Count (for directories)
	if item.FileCount > 0 {
		rows = append(rows, fmt.Sprintf("  %s  %s",
			labelStyle.Render("File Count:"),
			valueStyle.Render(fmt.Sprintf("%d", item.FileCount))))
	}

	body := strings.Join(rows, "\n\n")

	// Flash message
	var flashLine string
	if m.flashMsg != "" {
		flashLine = "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SuccessColor)).
			Bold(true).
			Render("  "+m.flashMsg)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, "", body+flashLine)

	return theme.PanelStyle.Copy().
		Width(width).
		Height(height).
		Render(content)
}

// renderStatusBar renders the action key hints at the bottom.
func (m detailModel) renderStatusBar() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.OverlayColor))

	sep := sepStyle.Render(" | ")

	hints := []string{
		keyStyle.Render("d") + " " + descStyle.Render("Delete"),
		keyStyle.Render("m") + " " + descStyle.Render("Move to Trash"),
		keyStyle.Render("c") + " " + descStyle.Render("Copy Path"),
		keyStyle.Render("esc") + " " + descStyle.Render("Back"),
	}

	bar := theme.StatusBarStyle.Copy().
		Width(m.width).
		Align(lipgloss.Center).
		Render(strings.Join(hints, sep))

	return bar
}

// severityBadge returns a color-coded severity label.
func severityBadge(severity string) string {
	var style lipgloss.Style
	switch strings.ToLower(severity) {
	case "high":
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.DangerColor)).
			Bold(true)
	case "medium":
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.WarningColor)).
			Bold(true)
	default:
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SuccessColor)).
			Bold(true)
	}
	return style.Render(strings.ToUpper(severity))
}

// sizeColorStyle returns a style color-coded by severity.
func sizeColorStyle(severity string) lipgloss.Style {
	switch strings.ToLower(severity) {
	case "high":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DangerColor)).Bold(true)
	case "medium":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.WarningColor)).Bold(true)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SuccessColor)).Bold(true)
	}
}

// flashTickMsg is sent periodically to decrement the flash timer.
type flashTickMsg struct{}

func flashTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return flashTickMsg{}
	})
}
