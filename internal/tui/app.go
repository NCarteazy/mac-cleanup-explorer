package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
	"github.com/nick/mac-cleanup-explorer/internal/executor"
	"github.com/nick/mac-cleanup-explorer/internal/scanner"
	"github.com/nick/mac-cleanup-explorer/internal/theme"
)

type viewState int

const (
	viewScanning viewState = iota
	viewDashboard
	viewReports
	viewDetail
	viewExport
	viewExecutor
)

// scanCompleteMsg is sent when the filesystem scan finishes.
type scanCompleteMsg struct {
	result *scanner.ScanResult
}

// reportsReadyMsg is sent when analysis is complete.
type reportsReadyMsg struct {
	reports map[string][]analyzer.ReportItem
}

// App is the top-level Bubble Tea model that routes to sub-views.
type App struct {
	currentView   viewState
	previousView  viewState // for returning from overlay views
	scanPath      string
	width, height int

	// Scan state
	scanProgress *scanner.ScanProgress
	scanResult   *scanner.ScanResult

	// Analysis state
	reports    map[string][]analyzer.ReportItem
	reportList []analyzer.Report

	// Sub-models
	scanModel      scanModel
	dashboardModel dashboardModel
	reportsModel   reportsModel
	detailModel    detailModel
	exportModel    exportModel
	executorModel  executorModel

	// Overlays
	showHelp     bool
	helpModel    helpModel
	showConfirm  bool
	confirmModel confirmModel
}

// NewApp creates a new App model configured to scan the given path.
func NewApp(scanPath string) *App {
	progress := &scanner.ScanProgress{}
	return &App{
		currentView:  viewScanning,
		scanPath:     scanPath,
		scanProgress: progress,
		reportList:   analyzer.AllReports(),
		scanModel:    newScanModel(progress),
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.scanModel.Init(),
		a.startScan(),
	)
}

func (a *App) startScan() tea.Cmd {
	return func() tea.Msg {
		result, err := scanner.Scan(a.scanPath, a.scanProgress)
		if err != nil {
			return scanCompleteMsg{result: &scanner.ScanResult{}}
		}
		return scanCompleteMsg{result: result}
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.helpModel.width = msg.Width
		a.helpModel.height = msg.Height
		a.confirmModel.width = msg.Width
		// Forward resize to active sub-model
		switch a.currentView {
		case viewDashboard:
			a.dashboardModel.width = msg.Width
			a.dashboardModel.height = msg.Height
		case viewReports:
			var cmd tea.Cmd
			a.reportsModel, cmd = a.reportsModel.Update(msg)
			return a, cmd
		case viewDetail:
			var cmd tea.Cmd
			a.detailModel, cmd = a.detailModel.Update(msg)
			return a, cmd
		case viewExport:
			var cmd tea.Cmd
			a.exportModel, cmd = a.exportModel.Update(msg)
			return a, cmd
		case viewExecutor:
			var cmd tea.Cmd
			a.executorModel, cmd = a.executorModel.Update(msg)
			return a, cmd
		}

	case tea.KeyMsg:
		// If confirmation dialog is active, route to it
		if a.showConfirm {
			var cmd tea.Cmd
			a.confirmModel, cmd = a.confirmModel.Update(msg)
			if a.confirmModel.cancelled {
				a.showConfirm = false
				return a, nil
			}
			if a.confirmModel.confirmed {
				a.showConfirm = false
				return a, cmd
			}
			return a, cmd
		}

		// If help overlay is active, only ? and esc are active
		if a.showHelp {
			var cmd tea.Cmd
			a.helpModel, cmd = a.helpModel.Update(msg)
			return a, cmd
		}

		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "?":
			a.showHelp = true
			a.helpModel = newHelpModel(a.width, a.height)
			return a, nil
		case "q":
			if a.currentView == viewScanning || a.currentView == viewDashboard {
				return a, tea.Quit
			}
		case "enter":
			if a.currentView == viewDashboard {
				a.currentView = viewReports
				a.reportsModel = newReportsModel(a.reports, a.reportList, a.scanResult, a.width, a.height)
				return a, nil
			}
		case "e":
			if a.currentView == viewDashboard || a.currentView == viewReports {
				a.previousView = a.currentView
				a.currentView = viewExport
				a.exportModel = newExportModel(
					a.reports, a.reportList, a.scanResult,
					a.scanPath, a.width, a.height, a.previousView,
				)
				return a, nil
			}
		case "x":
			if a.currentView == viewDashboard || a.currentView == viewReports {
				a.previousView = a.currentView
				a.currentView = viewExecutor
				a.executorModel = newExecutorModel(a.width, a.height)
				return a, nil
			}
		}

	case toggleHelpMsg:
		a.showHelp = !a.showHelp
		if a.showHelp {
			a.helpModel = newHelpModel(a.width, a.height)
		}
		return a, nil

	case deleteItemMsg:
		a.showConfirm = true
		a.confirmModel = newDeleteConfirm(msg.path, msg.size, a.width)
		return a, nil

	case trashItemMsg:
		a.showConfirm = true
		a.confirmModel = newTrashConfirm(msg.path, msg.size, a.width)
		return a, nil

	case bulkDeleteMsg:
		a.showConfirm = true
		a.confirmModel = newBulkDeleteConfirm(msg.paths, msg.sizes, msg.total, a.width)
		return a, nil

	case itemDeletedMsg:
		if msg.success {
			actionStr := "Deleted"
			cmdStr := fmt.Sprintf("rm -rf %s", msg.path)
			if msg.action == "trash" {
				actionStr = "Moved to Trash"
				cmdStr = fmt.Sprintf("trash %s", msg.path)
			}
			flash := fmt.Sprintf("%s: %s (freed %s)", actionStr, msg.path, theme.FormatSize(msg.freed))

			// Log the action to history (non-blocking)
			logCmd := logActionCmd(cmdStr, "success", msg.freed)

			// Remove from reports model
			a.reportsModel.removeReportItem(msg.path)
			// Also update the app-level reports reference
			a.reports = a.reportsModel.reports

			// If in detail view, navigate back to reports
			if a.currentView == viewDetail {
				a.currentView = viewReports
				a.reportsModel.flashMsg = flash
				a.reportsModel.flashTimer = 30
				return a, tea.Batch(flashTickCmd(), logCmd)
			}
			a.reportsModel.flashMsg = flash
			a.reportsModel.flashTimer = 30
			return a, tea.Batch(flashTickCmd(), logCmd)
		}
		// Error case
		errFlash := fmt.Sprintf("Error: %s", msg.err)
		a.reportsModel.flashMsg = errFlash
		a.reportsModel.flashTimer = 30
		return a, flashTickCmd()

	case bulkDeletedMsg:
		flash := fmt.Sprintf("Bulk delete: %d deleted, %d failed (freed %s)",
			msg.deleted, msg.failed, theme.FormatSize(msg.freed))
		// Remove deleted items from the reports model
		// The confirm model already deleted them from disk
		a.reportsModel.flashMsg = flash
		a.reportsModel.flashTimer = 30
		// Re-sync the reports
		a.reports = a.reportsModel.reports

		// Log bulk delete action to history (non-blocking)
		logCmd := logActionCmd(
			fmt.Sprintf("bulk delete (%d items)", msg.deleted+msg.failed),
			fmt.Sprintf("%d deleted, %d failed", msg.deleted, msg.failed),
			msg.freed,
		)
		return a, tea.Batch(flashTickCmd(), logCmd)

	case scanCompleteMsg:
		a.scanResult = msg.result
		a.currentView = viewDashboard
		a.dashboardModel = newDashboardModel(a.scanResult, a.width, a.height)
		return a, func() tea.Msg {
			results := analyzer.GenerateAll(a.scanResult)
			return reportsReadyMsg{reports: results}
		}

	case reportsReadyMsg:
		a.reports = msg.reports
		a.dashboardModel.reports = msg.reports

	case navigateBackMsg:
		switch a.currentView {
		case viewReports:
			a.currentView = viewDashboard
			return a, nil
		case viewDetail:
			a.currentView = viewReports
			return a, nil
		case viewExport:
			a.currentView = a.exportModel.previousView
			return a, nil
		case viewExecutor:
			a.currentView = a.previousView
			return a, nil
		}

	case navigateToDetailMsg:
		// Determine the report name from the currently selected sidebar item
		reportName := ""
		if a.reportsModel.sidebarCursor < len(a.reportsModel.reportList) {
			reportName = a.reportsModel.reportList[a.reportsModel.sidebarCursor].Description()
		}
		a.currentView = viewDetail
		a.detailModel = newDetailModel(msg.item, reportName, a.width, a.height)
		return a, nil
	}

	// Route to current view's update
	switch a.currentView {
	case viewScanning:
		model, cmd := a.scanModel.Update(msg)
		a.scanModel = model.(scanModel)
		return a, cmd
	case viewDashboard:
		var cmd tea.Cmd
		a.dashboardModel, cmd = a.dashboardModel.Update(msg)
		return a, cmd
	case viewReports:
		var cmd tea.Cmd
		a.reportsModel, cmd = a.reportsModel.Update(msg)
		return a, cmd
	case viewDetail:
		var cmd tea.Cmd
		a.detailModel, cmd = a.detailModel.Update(msg)
		return a, cmd
	case viewExport:
		var cmd tea.Cmd
		a.exportModel, cmd = a.exportModel.Update(msg)
		return a, cmd
	case viewExecutor:
		var cmd tea.Cmd
		a.executorModel, cmd = a.executorModel.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *App) View() string {
	var base string
	switch a.currentView {
	case viewScanning:
		base = a.scanModel.View()
	case viewDashboard:
		base = a.dashboardModel.View()
	case viewReports:
		base = a.reportsModel.View()
	case viewDetail:
		base = a.detailModel.View()
	case viewExport:
		base = a.exportModel.View()
	case viewExecutor:
		base = a.executorModel.View()
	default:
		base = "Loading..."
	}

	// Render confirmation dialog as centered overlay
	if a.showConfirm {
		overlay := a.confirmModel.View()
		base = overlayCenter(a.width, a.height, base, overlay)
	}

	// Render help overlay on top of everything
	if a.showHelp {
		overlay := a.helpModel.View()
		base = overlayCenter(a.width, a.height, base, overlay)
	}

	return base
}

// overlayCenter places an overlay panel centered on top of a background view.
func overlayCenter(width, height int, background, overlay string) string {
	if width <= 0 || height <= 0 {
		return overlay
	}
	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color(theme.BgColor)),
	)
}

// logActionCmd returns a tea.Cmd that logs an action to the history file
// without blocking the UI.
func logActionCmd(command, result string, freedBytes int64) tea.Cmd {
	return func() tea.Msg {
		_ = executor.LogAction(executor.DefaultLogPath(), command, result, freedBytes)
		return nil
	}
}
