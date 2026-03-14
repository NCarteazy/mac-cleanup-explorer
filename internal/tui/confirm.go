package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

// deleteItemMsg requests the app to show a delete confirmation.
type deleteItemMsg struct {
	path string
	size int64
}

// trashItemMsg requests the app to show a trash confirmation.
type trashItemMsg struct {
	path string
	size int64
}

// bulkDeleteMsg requests confirmation for bulk delete of multiple items.
type bulkDeleteMsg struct {
	paths []string
	sizes []int64
	total int64
}

// itemDeletedMsg is sent after a delete or trash operation completes.
type itemDeletedMsg struct {
	path    string
	freed   int64
	success bool
	err     string
	action  string // "delete" or "trash"
}

// bulkDeletedMsg is sent after a bulk delete completes.
type bulkDeletedMsg struct {
	deleted int
	failed  int
	freed   int64
}

// confirmModel is a modal confirmation dialog.
type confirmModel struct {
	title     string
	message   string
	path      string
	paths     []string // for bulk operations
	size      int64
	action    string // "delete" or "trash"
	confirmed bool
	cancelled bool
	width     int
}

// newDeleteConfirm creates a confirmation dialog for a delete action.
func newDeleteConfirm(path string, size int64, width int) confirmModel {
	return confirmModel{
		title:   "Delete Item",
		message: "Are you sure? This cannot be undone.",
		path:    path,
		size:    size,
		action:  "delete",
		width:   width,
	}
}

// newTrashConfirm creates a confirmation dialog for a trash action.
func newTrashConfirm(path string, size int64, width int) confirmModel {
	return confirmModel{
		title:   "Move to Trash",
		message: "Move to Trash? You can restore from Trash later.",
		path:    path,
		size:    size,
		action:  "trash",
		width:   width,
	}
}

// newBulkDeleteConfirm creates a confirmation dialog for bulk delete.
func newBulkDeleteConfirm(paths []string, sizes []int64, total int64, width int) confirmModel {
	return confirmModel{
		title:   "Bulk Delete",
		message: "Are you sure? This cannot be undone.",
		paths:   paths,
		size:    total,
		action:  "delete",
		width:   width,
	}
}

// Update handles input for the confirmation dialog.
func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.confirmed = true
			return m, m.execute()
		case "n", "N", "esc":
			m.cancelled = true
			return m, nil
		}
	}
	return m, nil
}

// execute performs the confirmed action and returns the result as a message.
func (m confirmModel) execute() tea.Cmd {
	if len(m.paths) > 0 {
		// Bulk delete
		paths := m.paths
		return func() tea.Msg {
			var deleted, failed int
			var freed int64
			for _, p := range paths {
				info, err := os.Stat(p)
				var sz int64
				if err == nil {
					sz = info.Size()
				}
				if err := os.RemoveAll(p); err != nil {
					failed++
				} else {
					deleted++
					freed += sz
				}
			}
			return bulkDeletedMsg{
				deleted: deleted,
				failed:  failed,
				freed:   freed,
			}
		}
	}

	path := m.path
	size := m.size
	action := m.action

	return func() tea.Msg {
		if action == "trash" {
			home, err := os.UserHomeDir()
			if err != nil {
				return itemDeletedMsg{
					path:    path,
					freed:   0,
					success: false,
					err:     fmt.Sprintf("cannot find home dir: %v", err),
					action:  action,
				}
			}
			trashDir := filepath.Join(home, ".Trash")
			baseName := filepath.Base(path)
			trashPath := filepath.Join(trashDir, baseName)

			// If a file with the same name already exists in Trash, add a suffix
			if _, err := os.Stat(trashPath); err == nil {
				ext := filepath.Ext(baseName)
				name := strings.TrimSuffix(baseName, ext)
				for i := 1; ; i++ {
					trashPath = filepath.Join(trashDir, fmt.Sprintf("%s_%d%s", name, i, ext))
					if _, err := os.Stat(trashPath); os.IsNotExist(err) {
						break
					}
				}
			}

			err = os.Rename(path, trashPath)
			if err != nil {
				return itemDeletedMsg{
					path:    path,
					freed:   0,
					success: false,
					err:     err.Error(),
					action:  action,
				}
			}
			return itemDeletedMsg{
				path:    path,
				freed:   size,
				success: true,
				action:  action,
			}
		}

		// Delete
		err := os.RemoveAll(path)
		if err != nil {
			return itemDeletedMsg{
				path:    path,
				freed:   0,
				success: false,
				err:     err.Error(),
				action:  action,
			}
		}
		return itemDeletedMsg{
			path:    path,
			freed:   size,
			success: true,
			action:  action,
		}
	}
}

// View renders the confirmation dialog as a centered overlay.
func (m confirmModel) View() string {
	panelWidth := 50
	if m.width > 0 && m.width < 60 {
		panelWidth = m.width - 10
	}
	if panelWidth < 30 {
		panelWidth = 30
	}

	// Choose border color based on action
	var borderColor string
	var titleStyle lipgloss.Style
	if m.action == "delete" {
		borderColor = theme.DangerColor
		titleStyle = theme.DangerStyle.Copy().Bold(true)
	} else {
		borderColor = theme.WarningColor
		titleStyle = theme.WarningStyle.Copy().Bold(true)
	}

	title := titleStyle.Render(m.title)

	var pathLine string
	if len(m.paths) > 0 {
		pathLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextColor)).
			Render(fmt.Sprintf("%d items selected", len(m.paths)))
	} else {
		displayPath := m.path
		maxPathLen := panelWidth - 6
		if maxPathLen < 20 {
			maxPathLen = 20
		}
		if len(displayPath) > maxPathLen {
			displayPath = "..." + displayPath[len(displayPath)-(maxPathLen-3):]
		}
		pathLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.SecondaryColor)).
			Render(displayPath)
	}

	sizeLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextColor)).
		Render("Size: " + theme.FormatSize(m.size))

	messageLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedColor)).
		Render(m.message)

	// Extra warning for large deletions (>1GB)
	var extraWarning string
	if m.size > 1024*1024*1024 {
		extraWarning = theme.DangerStyle.Copy().Bold(true).
			Render("WARNING: This is a large deletion!")
	}

	promptLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true).
		Render("Press y to confirm, n or esc to cancel")

	var parts []string
	parts = append(parts, title, "", pathLine, sizeLine, "", messageLine)
	if extraWarning != "" {
		parts = append(parts, extraWarning)
	}
	parts = append(parts, "", promptLine)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1, 2).
		Width(panelWidth).
		Render(content)

	return panel
}
