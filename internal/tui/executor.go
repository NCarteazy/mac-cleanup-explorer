package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/executor"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// commandExecutedMsg is sent when a command finishes executing.
type commandExecutedMsg struct {
	index int
}

// executorModel is the TUI view for pasting and executing cleanup commands.
type executorModel struct {
	textarea   textarea.Model
	commands   []executor.Command
	cursor     int
	editing    bool // true = editing textarea, false = browsing commands
	dryRun     bool
	width      int
	height     int
	flashMsg   string
	flashTimer int
}

// newExecutorModel creates a new executor model.
func newExecutorModel(width, height int) executorModel {
	ta := textarea.New()
	ta.Placeholder = "Paste cleanup commands here (one per line)..."
	ta.ShowLineNumbers = true
	ta.Focus()
	ta.SetWidth(width - 6)
	ta.SetHeight(8)
	ta.CharLimit = 4096

	return executorModel{
		textarea: ta,
		editing:  true,
		width:    width,
		height:   height,
	}
}

// Update handles messages for the executor model.
func (m executorModel) Update(msg tea.Msg) (executorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(m.width - 6)
		return m, nil

	case flashTickMsg:
		m.flashTimer--
		if m.flashTimer <= 0 {
			m.flashMsg = ""
			return m, nil
		}
		return m, flashTickCmd()

	case commandExecutedMsg:
		if msg.index >= 0 && msg.index < len(m.commands) {
			m.commands[msg.index].Running = false
		}
		return m, nil

	case tea.KeyMsg:
		if m.editing {
			return m.handleEditingKeys(msg)
		}
		return m.handleCommandListKeys(msg)
	}

	// Forward to textarea if editing
	if m.editing {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleEditingKeys handles key events while in textarea editing mode.
func (m executorModel) handleEditingKeys(msg tea.KeyMsg) (executorModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if len(m.commands) > 0 {
			// If commands already parsed, go back to app
			return m, func() tea.Msg { return navigateBackMsg{} }
		}
		return m, func() tea.Msg { return navigateBackMsg{} }

	case "ctrl+d":
		// Parse commands from textarea
		input := m.textarea.Value()
		m.commands = executor.ParseCommands(input)
		// Validate all commands
		for i := range m.commands {
			executor.ValidateCommand(&m.commands[i])
		}
		if len(m.commands) > 0 {
			m.editing = false
			m.cursor = 0
		} else {
			m.flashMsg = "No commands found"
			m.flashTimer = 20
			return m, flashTickCmd()
		}
		return m, nil

	case "tab":
		if len(m.commands) > 0 {
			m.editing = false
			m.cursor = 0
			return m, nil
		}
		// Forward tab to textarea for indentation
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	// Forward all other keys to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// handleCommandListKeys handles key events while browsing the command list.
func (m executorModel) handleCommandListKeys(msg tea.KeyMsg) (executorModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return navigateBackMsg{} }

	case "tab":
		m.editing = true
		m.textarea.Focus()
		return m, nil

	case "j", "down":
		if m.cursor < len(m.commands)-1 {
			m.cursor++
		}
		return m, nil

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "enter":
		return m.executeSelected()

	case "ctrl+a":
		return m.executeAll()

	case "n":
		m.dryRun = !m.dryRun
		if m.dryRun {
			m.flashMsg = "Dry-run mode ON"
		} else {
			m.flashMsg = "Dry-run mode OFF"
		}
		m.flashTimer = 15
		return m, flashTickCmd()
	}

	return m, nil
}

// executeSelected executes the currently selected command.
func (m executorModel) executeSelected() (executorModel, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.commands) {
		return m, nil
	}

	cmd := &m.commands[m.cursor]
	if cmd.Error != "" && !cmd.Validated {
		m.flashMsg = "Cannot execute: " + cmd.Error
		m.flashTimer = 20
		return m, flashTickCmd()
	}

	if m.dryRun {
		cmd.Output = "[DRY RUN] Would execute: " + cmd.Raw
		cmd.Executed = true
		cmd.ExitCode = 0
		m.flashMsg = "Dry run complete"
		m.flashTimer = 15
		return m, flashTickCmd()
	}

	// Execute the command
	idx := m.cursor
	executor.ExecuteCommand(cmd)

	return m, func() tea.Msg {
		return commandExecutedMsg{index: idx}
	}
}

// executeAll executes all validated commands.
func (m executorModel) executeAll() (executorModel, tea.Cmd) {
	executed := 0
	for i := range m.commands {
		cmd := &m.commands[i]
		if !cmd.Validated || cmd.Error != "" {
			continue
		}

		if m.dryRun {
			cmd.Output = "[DRY RUN] Would execute: " + cmd.Raw
			cmd.Executed = true
			cmd.ExitCode = 0
		} else {
			executor.ExecuteCommand(cmd)
		}
		executed++
	}

	if m.dryRun {
		m.flashMsg = fmt.Sprintf("Dry run: %d command(s) simulated", executed)
	} else {
		m.flashMsg = fmt.Sprintf("Executed %d command(s)", executed)
	}
	m.flashTimer = 20
	return m, flashTickCmd()
}

// View renders the executor view.
func (m executorModel) View() string {
	if m.width <= 0 {
		return "Loading..."
	}

	// Title
	titleText := "Command Executor"
	if m.dryRun {
		dryRunBadge := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.WarningColor)).
			Foreground(lipgloss.Color(theme.BgColor)).
			Bold(true).
			Padding(0, 1).
			Render("DRY RUN")
		titleText = titleText + "  " + dryRunBadge
	}
	titleBar := theme.TitleStyle.Render(titleText)

	// Textarea section
	taLabel := theme.SubtitleStyle.Render("Commands Input")
	var taSection string
	if m.editing {
		taSection = theme.ActivePanelStyle.Copy().
			Width(m.width - 4).
			Render(lipgloss.JoinVertical(lipgloss.Left, taLabel, "", m.textarea.View()))
	} else {
		taSection = theme.PanelStyle.Copy().
			Width(m.width - 4).
			Render(lipgloss.JoinVertical(lipgloss.Left, taLabel, "", m.textarea.View()))
	}

	// Command list section
	cmdList := m.renderCommandList()

	// Flash message
	var flashLine string
	if m.flashMsg != "" {
		flashLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SuccessColor)).
			Bold(true).
			Render("  " + m.flashMsg)
	}

	// Status bar
	statusBar := m.renderStatusBar()

	parts := []string{titleBar, "", taSection, "", cmdList}
	if flashLine != "" {
		parts = append(parts, flashLine)
	}
	parts = append(parts, statusBar)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderCommandList renders the parsed command list with syntax highlighting.
func (m executorModel) renderCommandList() string {
	if len(m.commands) == 0 {
		hint := theme.MutedStyle.Render("  No commands parsed yet. Paste commands above and press Ctrl+D.")
		return hint
	}

	listLabel := theme.SubtitleStyle.Render("Parsed Commands")

	// Calculate available height for the command list
	listHeight := m.height - 20
	if listHeight < 5 {
		listHeight = 5
	}

	var rows []string
	for i, cmd := range m.commands {
		rows = append(rows, m.renderCommandRow(i, cmd)...)

		// Show output if executed
		if cmd.Executed && cmd.Output != "" {
			outputStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.MutedColor)).
				PaddingLeft(4)
			// Truncate output to 3 lines max
			outLines := strings.Split(strings.TrimSpace(cmd.Output), "\n")
			if len(outLines) > 3 {
				outLines = append(outLines[:3], fmt.Sprintf("... (%d more lines)", len(outLines)-3))
			}
			for _, ol := range outLines {
				rows = append(rows, outputStyle.Render(ol))
			}
		}
	}

	content := strings.Join(rows, "\n")

	// Wrap in panel
	panelStyle := theme.PanelStyle.Copy()
	if !m.editing {
		panelStyle = theme.ActivePanelStyle.Copy()
	}

	return panelStyle.
		Width(m.width - 4).
		Height(listHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left, listLabel, "", content))
}

// renderCommandRow renders a single command entry with syntax highlighting.
func (m executorModel) renderCommandRow(index int, cmd executor.Command) []string {
	// Cursor indicator
	cursor := "  "
	if !m.editing && index == m.cursor {
		cursor = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.PrimaryColor)).
			Bold(true).
			Render("> ")
	}

	// Status icon
	var statusIcon string
	if cmd.Executed {
		if cmd.ExitCode == 0 {
			statusIcon = lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.SuccessColor)).
				Render(" [OK]")
		} else {
			statusIcon = lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DangerColor)).
				Render(fmt.Sprintf(" [FAIL:%d]", cmd.ExitCode))
		}
	} else if cmd.Error != "" && !cmd.Validated {
		statusIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.DangerColor)).
			Render(" [BLOCKED]")
	} else if cmd.Validated {
		statusIcon = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SuccessColor)).
			Render(" [VALID]")
	}

	// Syntax-highlighted command text
	highlighted := highlightCommand(cmd.Raw)

	// Row background for selected item
	var line string
	if !m.editing && index == m.cursor {
		line = theme.SelectedRowStyle.Copy().
			Width(m.width - 12).
			Render(cursor + highlighted + statusIcon)
	} else {
		line = cursor + highlighted + statusIcon
	}

	rows := []string{line}

	// Show error message below if validation failed
	if cmd.Error != "" && !cmd.Validated {
		errMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.DangerColor)).
			PaddingLeft(4).
			Render("Error: " + cmd.Error)
		rows = append(rows, errMsg)
	}

	// Show warning if present
	if cmd.Warning != "" {
		warnMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.WarningColor)).
			PaddingLeft(4).
			Render("Warning: " + cmd.Warning)
		rows = append(rows, warnMsg)
	}

	return rows
}

// highlightCommand applies basic syntax highlighting to a command string.
func highlightCommand(raw string) string {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return raw
	}

	rmStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.DangerColor)).
		Bold(true)
	flagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.WarningColor))
	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SecondaryColor))
	defaultStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextColor))

	var result []string
	for i, part := range parts {
		switch {
		case i == 0 && (part == "rm" || part == "rmdir" || part == "trash"):
			result = append(result, rmStyle.Render(part))
		case strings.HasPrefix(part, "-"):
			result = append(result, flagStyle.Render(part))
		case strings.HasPrefix(part, "/") || strings.HasPrefix(part, "~") || strings.HasPrefix(part, "."):
			result = append(result, pathStyle.Render(part))
		default:
			result = append(result, defaultStyle.Render(part))
		}
	}

	return strings.Join(result, " ")
}

// renderStatusBar renders the status bar with key hints.
func (m executorModel) renderStatusBar() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.OverlayColor))

	sep := sepStyle.Render(" | ")

	var hints []string
	if m.editing {
		hints = []string{
			keyStyle.Render("ctrl+d") + " " + descStyle.Render("Parse Commands"),
			keyStyle.Render("tab") + " " + descStyle.Render("Switch to List"),
			keyStyle.Render("esc") + " " + descStyle.Render("Back"),
		}
	} else {
		hints = []string{
			keyStyle.Render("enter") + " " + descStyle.Render("Execute"),
			keyStyle.Render("ctrl+a") + " " + descStyle.Render("Execute All"),
			keyStyle.Render("n") + " " + descStyle.Render("Dry-Run"),
			keyStyle.Render("tab") + " " + descStyle.Render("Edit"),
			keyStyle.Render("j/k") + " " + descStyle.Render("Navigate"),
			keyStyle.Render("esc") + " " + descStyle.Render("Back"),
		}
	}

	bar := theme.StatusBarStyle.Copy().
		Width(m.width).
		Align(lipgloss.Center).
		Render(strings.Join(hints, sep))

	return bar
}
