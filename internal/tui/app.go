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
	scanPath      string
	width, height int

	// Scan state
	scanProgress *scanner.ScanProgress
	scanResult   *scanner.ScanResult

	// Analysis state
	reports    map[string][]analyzer.ReportItem
	reportList []analyzer.Report

	// Sub-models
	scanModel scanModel
	// (dashboard, reports, etc. added in later tasks)
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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q":
			if a.currentView == viewScanning {
				return a, tea.Quit
			}
			// Other views: q goes back or quits depending on view
			if a.currentView == viewDashboard {
				return a, tea.Quit
			}
		}

	case scanCompleteMsg:
		a.scanResult = msg.result
		a.currentView = viewDashboard
		return a, func() tea.Msg {
			results := analyzer.GenerateAll(a.scanResult)
			return reportsReadyMsg{reports: results}
		}

	case reportsReadyMsg:
		a.reports = msg.reports
	}

	// Route to current view's update
	switch a.currentView {
	case viewScanning:
		model, cmd := a.scanModel.Update(msg)
		a.scanModel = model.(scanModel)
		return a, cmd
	}

	return a, nil
}

func (a *App) View() string {
	switch a.currentView {
	case viewScanning:
		return a.scanModel.View()
	case viewDashboard:
		return "Dashboard (coming in Task 7)...\n\nPress q to quit"
	default:
		return "Loading..."
	}
}
