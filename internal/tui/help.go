package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// helpModel is the help overlay that shows all key bindings.
type helpModel struct {
	width  int
	height int
}

// newHelpModel creates a new help overlay model.
func newHelpModel(width, height int) helpModel {
	return helpModel{
		width:  width,
		height: height,
	}
}

// Update handles input for the help overlay.
func (m helpModel) Update(msg tea.Msg) (helpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "?", "esc":
			return m, func() tea.Msg { return toggleHelpMsg{} }
		}
	}
	return m, nil
}

// toggleHelpMsg requests the app to toggle the help overlay.
type toggleHelpMsg struct{}

// View renders the help overlay panel.
func (m helpModel) View() string {
	panelWidth := 45
	if m.width > 0 && m.width < 55 {
		panelWidth = m.width - 10
	}
	if panelWidth < 30 {
		panelWidth = 30
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true).
		Width(12)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextColor))
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SecondaryColor)).
		Bold(true)

	title := theme.TitleStyle.Copy().Render("MAC CLEANUP EXPLORER")
	subtitle := theme.MutedStyle.Copy().Render("Keyboard Help")

	renderBinding := func(key, desc string) string {
		return keyStyle.Render(key) + descStyle.Render(desc)
	}

	nav := []string{
		sectionStyle.Render("NAVIGATION"),
		renderBinding("up/k", "Move up"),
		renderBinding("down/j", "Move down"),
		renderBinding("enter", "Select / drill down"),
		renderBinding("esc", "Go back"),
		renderBinding("tab", "Switch pane"),
		renderBinding("q", "Quit"),
	}

	actions := []string{
		sectionStyle.Render("ACTIONS"),
		renderBinding("d", "Delete selected item"),
		renderBinding("m", "Move to Trash"),
		renderBinding("D", "Bulk delete selected"),
		renderBinding("space", "Toggle selection"),
		renderBinding("c", "Copy path to clipboard"),
	}

	views := []string{
		sectionStyle.Render("VIEWS"),
		renderBinding("e", "Export reports"),
		renderBinding("x", "Command executor"),
		renderBinding("?", "Toggle this help"),
	}

	exportKeys := []string{
		sectionStyle.Render("EXPORT"),
		renderBinding("c", "Copy to clipboard"),
		renderBinding("s", "Save to file"),
		renderBinding("a", "Select all reports"),
		renderBinding("n", "Deselect all reports"),
	}

	closeHint := theme.MutedStyle.Copy().Render("Press ? or esc to close")

	var lines []string
	lines = append(lines, title, subtitle, "")
	lines = append(lines, nav...)
	lines = append(lines, "")
	lines = append(lines, actions...)
	lines = append(lines, "")
	lines = append(lines, views...)
	lines = append(lines, "")
	lines = append(lines, exportKeys...)
	lines = append(lines, "", closeHint)

	content := strings.Join(lines, "\n")

	panel := theme.PanelStyle.Copy().
		Width(panelWidth).
		Render(content)

	return panel
}
