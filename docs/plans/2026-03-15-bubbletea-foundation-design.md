# Bubbletea Foundation — Design Document

**Date:** 2026-03-15
**Status:** Approved

## Overview

A Go module that wraps Bubble Tea to eliminate TUI boilerplate. You define views (screens) as a simple interface; the framework handles view routing, overlay stack, resize propagation, breadcrumbs, status bar, flash messages, and theming. Ships with optional reusable components (table, confirm dialog, help overlay).

**Module:** `github.com/NCarteazy/bubbletea-foundation`
**Dependencies:** Bubble Tea, Lip Gloss

## Goals

- Stand up new TUI apps in minutes, not hours
- Kill the ~300 lines of routing/overlay/resize boilerplate every app needs
- Views are fully flexible — render whatever you want inside them
- Optional components (table, confirm, help) are independent and composable
- Tokyo Night default theme, swappable via interface

## Architecture

Three layers:

### 1. App (Orchestrator)
- Top-level Bubble Tea model
- Owns view routing via navigation stack (push/pop/replace)
- Manages overlay stack (help, confirm, custom modals)
- Propagates resize to all views
- Renders chrome: breadcrumbs (top), status bar (bottom), flash messages
- Handles global keybindings (ctrl+c quit, ? help)
- Created with functional options pattern

### 2. Views (Your Screens)
- Implement a 3-method `View` interface: `ID()`, `Update()`, `Render()`
- Receive `ViewContext` with dimensions, theme, navigation, and overlay helpers
- Width/height in context already accounts for chrome (breadcrumbs + status bar)
- Render anything — tables, forms, plain text, art — framework doesn't care
- No direct coupling between views; communicate through the app via messages

### 3. Components (Optional Building Blocks)
- Reusable widgets that work inside any view
- Each component is independent — use all, some, or none
- v1 ships: Table, Confirm dialog, Help overlay
- Extensible — add more later without breaking anything

## Module Structure

```
bubbletea-foundation/
├── app.go          # App struct, Init/Update/View, functional options
├── view.go         # View interface, ViewContext
├── overlay.go      # Overlay stack (help, confirm, custom modals)
├── nav.go          # Navigation stack (push/pop/replace)
├── flash.go        # Flash message system with auto-decay timer
├── breadcrumb.go   # Auto breadcrumb bar from navigation stack
├── statusbar.go    # Bottom status bar with configurable key hints
├── theme/
│   ├── theme.go    # Theme interface + utility functions
│   └── tokyo.go    # Tokyo Night palette (default)
├── components/
│   ├── table.go    # Sortable, scrollable table with column config
│   ├── confirm.go  # Generic confirmation dialog
│   └── help.go     # Help overlay with grouped key binding sections
└── layout/
    ├── panes.go    # Two-pane split layout helper
    └── center.go   # Overlay centering utility
```

## Core Interfaces

### View

```go
type View interface {
    ID() string
    Update(msg tea.Msg, ctx ViewContext) (View, tea.Cmd)
    Render(ctx ViewContext) string
}
```

### ViewContext

```go
type ViewContext struct {
    Width  int
    Height int
    Theme  theme.Theme

    // Navigation
    Navigate(viewID string, data any)
    Back()
    Replace(viewID string, data any)

    // Overlays
    ShowOverlay(overlay Overlay)
    Flash(msg string, duration int)
    Confirm(prompt string, onYes func() tea.Cmd)
}
```

### Theme

```go
type Theme interface {
    // Colors
    Bg() string
    Primary() string
    Secondary() string
    Success() string
    Warning() string
    Danger() string
    Text() string
    Muted() string
    Surface() string

    // Pre-built styles
    Title() lipgloss.Style
    Panel() lipgloss.Style
    ActivePanel() lipgloss.Style
    StatusBar() lipgloss.Style
    TableHeader() lipgloss.Style
    SelectedRow() lipgloss.Style

    // Utilities
    FormatSize(bytes int64) string
    SizeBar(ratio float64, maxWidth int) string
}
```

### Overlay

```go
type Overlay interface {
    Update(msg tea.Msg) (Overlay, tea.Cmd)
    View() string
    Done() bool
}
```

## App Construction API

```go
func main() {
    app := foundation.New(
        foundation.WithTitle("My Tool"),
        foundation.WithTheme(tokyo.Night),
        foundation.WithInitialView("dashboard"),
        foundation.WithViews(
            myDashboardView{},
            myDetailView{},
        ),
        foundation.WithHelp([]foundation.KeySection{
            {Title: "Navigation", Keys: []foundation.KeyHint{
                {Key: "j/k", Desc: "move up/down"},
                {Key: "enter", Desc: "select"},
            }},
        }),
        foundation.WithStatusHints("enter", "select", "q", "quit", "?", "help"),
    )
    tea.NewProgram(app, tea.WithAltScreen()).Run()
}
```

## Components (v1)

### Table
Sortable, scrollable table with configurable columns. Supports keyboard navigation (j/k, g/G, s to cycle sort), alternating row colors, selection highlighting, dynamic column widths.

### Confirm
Generic confirmation dialog. Rendered as centered overlay. Supports custom prompt text, y/n/esc handling, and a callback that returns a `tea.Cmd` on confirm.

### Help
Modal overlay toggled with `?`. Displays key bindings organized into titled sections. Two-column layout (key + description). Auto-generated from `KeySection` definitions passed at app creation.

## What the Framework Handles (So You Don't)

| Concern | How |
|---------|-----|
| View routing | Navigation stack with push/pop/replace |
| Resize | Auto-propagates to all views via ViewContext |
| Overlays | Stack-based, blocks input to views underneath |
| Breadcrumbs | Auto-generated from navigation stack view IDs |
| Status bar | Rendered from configurable key hints |
| Flash messages | Timed auto-decay, rendered above status bar |
| Help overlay | Built-in `?` toggle, content from KeySection config |
| Quit | ctrl+c always works |
| Theme | Injected via ViewContext, swappable at creation |

## Future Components (Post-v1)

- Text input / form fields
- Progress bar / spinner
- List (non-tabular)
- Tabs
- Tree view
- Toast notifications

These would be added as independent components under `components/` without breaking existing code.
