package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
	"github.com/nick/mac-cleanup-explorer/internal/scanner"
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
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q":
			if a.currentView == viewScanning {
				return a, tea.Quit
			}
			if a.currentView == viewDashboard {
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
	switch a.currentView {
	case viewScanning:
		return a.scanModel.View()
	case viewDashboard:
		return a.dashboardModel.View()
	case viewReports:
		return a.reportsModel.View()
	case viewDetail:
		return a.detailModel.View()
	case viewExport:
		return a.exportModel.View()
	case viewExecutor:
		return a.executorModel.View()
	default:
		return "Loading..."
	}
}
