package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/nick/mac-cleanup-explorer/internal/scanner"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

type tickMsg time.Time

type scanModel struct {
	spinner  spinner.Model
	progress *scanner.ScanProgress
	start    time.Time
	done     bool
	width    int
	height   int
}

func newScanModel(progress *scanner.ScanProgress) scanModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.PrimaryColor))
	return scanModel{
		spinner:  s,
		progress: progress,
		start:    time.Now(),
	}
}

func (m scanModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m scanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		if !m.done {
			return m, tickCmd()
		}
	case scanCompleteMsg:
		m.done = true
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m scanModel) View() string {
	snap := m.progress.Snapshot()
	elapsed := time.Since(m.start).Round(time.Millisecond)

	logo := renderLogo()

	scanLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextColor)).
		Bold(true).
		Render("Scanning filesystem...")

	statsBlock := fmt.Sprintf(
		"  %s %s\n\n"+
			"  Files found:   %s\n"+
			"  Directories:   %s\n"+
			"  Data found:    %s\n"+
			"  Elapsed:       %s",
		m.spinner.View(),
		scanLabel,
		theme.TextStyle.Render(humanize.Comma(snap.FilesFound)),
		theme.TextStyle.Render(humanize.Comma(snap.DirsFound)),
		theme.SuccessStyle.Render(humanize.Bytes(uint64(snap.BytesFound))),
		theme.MutedStyle.Render(elapsed.String()),
	)

	currentPath := snap.CurrentPath
	maxPathWidth := 60
	if m.width > 0 && m.width-4 < maxPathWidth {
		maxPathWidth = m.width - 4
	}
	if maxPathWidth < 20 {
		maxPathWidth = 20
	}
	if len(currentPath) > maxPathWidth {
		currentPath = "..." + currentPath[len(currentPath)-(maxPathWidth-3):]
	}
	pathLine := theme.MutedStyle.Render("  " + currentPath)

	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor)).
		Render("  Press q to quit")

	content := lipgloss.JoinVertical(lipgloss.Left,
		logo,
		"",
		statsBlock,
		"",
		pathLine,
		"",
		hint,
	)

	// Center vertically and horizontally if we have dimensions
	if m.width > 0 && m.height > 0 {
		content = lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			content,
		)
	}

	return content
}

