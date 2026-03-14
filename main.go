package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nick/mac-cleanup-explorer/internal/tui"
)

func main() {
	scanPath := flag.String("path", "/", "Root path to scan")
	flag.Parse()

	app := tui.NewApp(*scanPath)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
