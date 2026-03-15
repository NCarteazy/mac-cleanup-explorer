# Bubbletea Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a reusable Go module that wraps Bubble Tea to eliminate TUI boilerplate — view routing, overlays, resize, breadcrumbs, status bar, flash messages, theming, and optional components (table, confirm, help).

**Architecture:** Single Go module with a core `App` orchestrator that manages a navigation stack of `View` interfaces. Views receive a `ViewContext` with dimensions, theme, and navigation helpers. Optional components (table, confirm, help) are independent packages under `components/`. A `theme/` package provides the `Theme` interface and Tokyo Night default.

**Tech Stack:** Go 1.22+, Bubble Tea (TUI framework), Lip Gloss (styling), go-humanize (size formatting)

**Design doc:** `docs/plans/2026-03-15-bubbletea-foundation-design.md` in the mac-cleanup-explorer repo.

---

### Task 0: Repository and Module Setup

**Files:**
- Create: `/Users/nick/bubbletea-foundation/go.mod`
- Create: `/Users/nick/bubbletea-foundation/go.sum`
- Create: `/Users/nick/bubbletea-foundation/CLAUDE.md`

**Step 1: Create the repo directory and initialize the Go module**

```bash
mkdir -p /Users/nick/bubbletea-foundation
cd /Users/nick/bubbletea-foundation
git init
go mod init github.com/NCarteazy/bubbletea-foundation
```

**Step 2: Add dependencies**

```bash
cd /Users/nick/bubbletea-foundation
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/dustin/go-humanize@latest
```

**Step 3: Create CLAUDE.md**

```markdown
# Bubbletea Foundation

## Build & Test
- `go build ./...` to build all packages
- `go test ./...` to run all tests
- `go test -v ./...` for verbose output
- `go test -race ./...` for race condition detection
- `go vet ./...` for linting

## Run Example
- `cd example && go run .`

## Project Structure
- `app.go` — App orchestrator (Init/Update/View, functional options)
- `view.go` — View interface, ViewContext, navigation messages
- `overlay.go` — Overlay interface and stack management
- `nav.go` — Navigation stack (push/pop/replace)
- `flash.go` — Flash message system with auto-decay
- `breadcrumb.go` — Auto breadcrumb bar from nav stack
- `statusbar.go` — Bottom status bar with key hints
- `theme/` — Theme interface and Tokyo Night default
- `components/` — Optional reusable components (table, confirm, help)
- `layout/` — Layout helpers (panes, centering)
- `example/` — Example app demonstrating all features

## Conventions
- All interfaces are minimal (3 methods or fewer)
- Components are independent — no cross-dependencies
- Theme is injected, never imported directly by components
- Views communicate through the App via tea.Msg, never directly
- Functional options pattern for App configuration
```

**Step 4: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add go.mod go.sum CLAUDE.md
git commit -m "chore: initialize module with dependencies"
```

---

### Task 1: Theme Interface and Tokyo Night Default

**Files:**
- Create: `theme/theme.go`
- Create: `theme/tokyo.go`
- Create: `theme/theme_test.go`

**Step 1: Write the failing test**

Create `theme/theme_test.go`:

```go
package theme_test

import (
	"testing"

	"github.com/NCarteazy/bubbletea-foundation/theme"
)

func TestTokyoNightImplementsTheme(t *testing.T) {
	var _ theme.Theme = theme.TokyoNight
}

func TestTokyoNightColors(t *testing.T) {
	tn := theme.TokyoNight
	if tn.Bg() == "" {
		t.Error("Bg() should not be empty")
	}
	if tn.Primary() == "" {
		t.Error("Primary() should not be empty")
	}
	if tn.Text() == "" {
		t.Error("Text() should not be empty")
	}
}

func TestFormatSize(t *testing.T) {
	tn := theme.TokyoNight
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1000, "1.0 kB"},
		{1000000, "1.0 MB"},
		{1000000000, "1.0 GB"},
	}
	for _, tt := range tests {
		got := tn.FormatSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestSizeBar(t *testing.T) {
	tn := theme.TokyoNight
	bar := tn.SizeBar(0.5, 20)
	if bar == "" {
		t.Error("SizeBar should not be empty")
	}
	if len(bar) == 0 {
		t.Error("SizeBar should produce visible output")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test ./theme/ -v`
Expected: FAIL — package/types don't exist yet

**Step 3: Write the Theme interface**

Create `theme/theme.go`:

```go
package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette and pre-built styles for a TUI application.
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
	Breadcrumb() lipgloss.Style
	TableHeader() lipgloss.Style
	SelectedRow() lipgloss.Style

	// Utilities
	FormatSize(bytes int64) string
	SizeBar(ratio float64, maxWidth int) string
}
```

**Step 4: Write the Tokyo Night implementation**

Create `theme/tokyo.go`:

```go
package theme

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/charmbracelet/lipgloss"
)

const (
	bgColor        = "#1a1b26"
	primaryColor   = "#7dcfff"
	secondaryColor = "#bb9af7"
	successColor   = "#9ece6a"
	warningColor   = "#e0af68"
	dangerColor    = "#f7768e"
	textColor      = "#c0caf5"
	mutedColor     = "#565f89"
	surfaceColor   = "#24283b"
	overlayColor   = "#414868"
)

// tokyoNight implements Theme with the Tokyo Night color palette.
type tokyoNight struct{}

// TokyoNight is the default theme instance.
var TokyoNight Theme = tokyoNight{}

func (t tokyoNight) Bg() string        { return bgColor }
func (t tokyoNight) Primary() string    { return primaryColor }
func (t tokyoNight) Secondary() string  { return secondaryColor }
func (t tokyoNight) Success() string    { return successColor }
func (t tokyoNight) Warning() string    { return warningColor }
func (t tokyoNight) Danger() string     { return dangerColor }
func (t tokyoNight) Text() string       { return textColor }
func (t tokyoNight) Muted() string      { return mutedColor }
func (t tokyoNight) Surface() string    { return surfaceColor }

func (t tokyoNight) Title() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(primaryColor)).
		Bold(true)
}

func (t tokyoNight) Panel() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(overlayColor)).
		Padding(1, 2)
}

func (t tokyoNight) ActivePanel() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(primaryColor)).
		Padding(1, 2)
}

func (t tokyoNight) StatusBar() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(surfaceColor)).
		Foreground(lipgloss.Color(textColor)).
		Padding(0, 1)
}

func (t tokyoNight) Breadcrumb() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(primaryColor))
}

func (t tokyoNight) TableHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(secondaryColor)).
		Bold(true)
}

func (t tokyoNight) SelectedRow() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(overlayColor)).
		Foreground(lipgloss.Color(textColor)).
		Bold(true)
}

func (t tokyoNight) FormatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	return humanize.Bytes(uint64(bytes))
}

func (t tokyoNight) SizeBar(ratio float64, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	filled := int(ratio * float64(maxWidth))
	if filled == 0 && ratio > 0 {
		filled = 1
	}

	color := successColor
	if ratio >= 0.7 {
		color = dangerColor
	} else if ratio >= 0.4 {
		color = warningColor
	}

	bar := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(strings.Repeat("█", filled))
	empty := lipgloss.NewStyle().
		Foreground(lipgloss.Color(mutedColor)).
		Render(strings.Repeat("░", maxWidth-filled))

	return fmt.Sprintf("%s%s", bar, empty)
}
```

**Step 5: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test ./theme/ -v`
Expected: PASS (all 4 tests)

**Step 6: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add theme/
git commit -m "feat: add Theme interface and Tokyo Night default palette"
```

---

### Task 2: View Interface, Navigation, and Messages

**Files:**
- Create: `view.go`
- Create: `nav.go`
- Create: `nav_test.go`

**Step 1: Write the failing test**

Create `nav_test.go`:

```go
package foundation

import (
	"testing"
)

func TestNavStackPush(t *testing.T) {
	s := &navStack{}
	s.push("dashboard", nil)
	s.push("detail", "some-data")

	if s.len() != 2 {
		t.Errorf("expected len 2, got %d", s.len())
	}
	if s.current().viewID != "detail" {
		t.Errorf("expected current 'detail', got %q", s.current().viewID)
	}
}

func TestNavStackPop(t *testing.T) {
	s := &navStack{}
	s.push("dashboard", nil)
	s.push("detail", nil)
	s.pop()

	if s.len() != 1 {
		t.Errorf("expected len 1, got %d", s.len())
	}
	if s.current().viewID != "dashboard" {
		t.Errorf("expected current 'dashboard', got %q", s.current().viewID)
	}
}

func TestNavStackPopEmpty(t *testing.T) {
	s := &navStack{}
	s.pop() // should not panic

	if s.len() != 0 {
		t.Errorf("expected len 0, got %d", s.len())
	}
}

func TestNavStackReplace(t *testing.T) {
	s := &navStack{}
	s.push("dashboard", nil)
	s.push("reports", nil)
	s.replace("settings", "data")

	if s.len() != 2 {
		t.Errorf("expected len 2, got %d", s.len())
	}
	if s.current().viewID != "settings" {
		t.Errorf("expected current 'settings', got %q", s.current().viewID)
	}
	if s.current().data != "data" {
		t.Errorf("expected data 'data', got %v", s.current().data)
	}
}

func TestNavStackBreadcrumbs(t *testing.T) {
	s := &navStack{}
	s.push("Dashboard", nil)
	s.push("Reports", nil)
	s.push("Detail", nil)

	crumbs := s.breadcrumbs()
	if len(crumbs) != 3 {
		t.Errorf("expected 3 breadcrumbs, got %d", len(crumbs))
	}
	if crumbs[0] != "Dashboard" || crumbs[1] != "Reports" || crumbs[2] != "Detail" {
		t.Errorf("unexpected breadcrumbs: %v", crumbs)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run TestNavStack`
Expected: FAIL — types don't exist

**Step 3: Write the View interface and navigation messages**

Create `view.go`:

```go
package foundation

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// View is the interface that application screens implement.
// Views are fully flexible — render whatever you want inside Render().
// The framework handles chrome (breadcrumbs, status bar), routing, and overlays.
type View interface {
	// ID returns a unique identifier for this view (used in nav stack, breadcrumbs).
	ID() string

	// Update handles messages. Returns the updated view and an optional command.
	Update(msg tea.Msg, ctx ViewContext) (View, tea.Cmd)

	// Render returns the view's content string.
	// Width/Height in ctx already account for framework chrome.
	Render(ctx ViewContext) string
}

// ViewContext provides views with dimensions, theme, and navigation helpers.
type ViewContext struct {
	Width  int
	Height int
	Theme  theme.Theme
	Data   any // data passed via Navigate() or Replace()
}

// Navigation messages — views return these as tea.Cmd to trigger navigation.

// NavigateMsg tells the app to push a new view onto the nav stack.
type NavigateMsg struct {
	ViewID string
	Data   any
}

// BackMsg tells the app to pop the current view off the nav stack.
type BackMsg struct{}

// ReplaceMsg tells the app to replace the current view on the stack.
type ReplaceMsg struct {
	ViewID string
	Data   any
}

// Navigate returns a tea.Cmd that pushes a view onto the nav stack.
func Navigate(viewID string, data any) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{ViewID: viewID, Data: data}
	}
}

// Back returns a tea.Cmd that pops back to the previous view.
func Back() tea.Cmd {
	return func() tea.Msg {
		return BackMsg{}
	}
}

// Replace returns a tea.Cmd that replaces the current view.
func Replace(viewID string, data any) tea.Cmd {
	return func() tea.Msg {
		return ReplaceMsg{ViewID: viewID, Data: data}
	}
}
```

**Step 4: Write the navigation stack**

Create `nav.go`:

```go
package foundation

// navEntry is a single entry in the navigation stack.
type navEntry struct {
	viewID string
	data   any
}

// navStack manages the view navigation history.
type navStack struct {
	entries []navEntry
}

func (s *navStack) push(viewID string, data any) {
	s.entries = append(s.entries, navEntry{viewID: viewID, data: data})
}

func (s *navStack) pop() {
	if len(s.entries) > 0 {
		s.entries = s.entries[:len(s.entries)-1]
	}
}

func (s *navStack) replace(viewID string, data any) {
	if len(s.entries) > 0 {
		s.entries[len(s.entries)-1] = navEntry{viewID: viewID, data: data}
	}
}

func (s *navStack) current() *navEntry {
	if len(s.entries) == 0 {
		return nil
	}
	return &s.entries[len(s.entries)-1]
}

func (s *navStack) len() int {
	return len(s.entries)
}

func (s *navStack) breadcrumbs() []string {
	crumbs := make([]string, len(s.entries))
	for i, e := range s.entries {
		crumbs[i] = e.viewID
	}
	return crumbs
}
```

**Step 5: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run TestNavStack`
Expected: PASS (all 5 tests)

**Step 6: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add view.go nav.go nav_test.go
git commit -m "feat: add View interface, ViewContext, and navigation stack"
```

---

### Task 3: Overlay Interface and Stack

**Files:**
- Create: `overlay.go`
- Create: `overlay_test.go`

**Step 1: Write the failing test**

Create `overlay_test.go`:

```go
package foundation

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type mockOverlay struct {
	done    bool
	content string
}

func (m mockOverlay) Update(msg tea.Msg) (Overlay, tea.Cmd) { return m, nil }
func (m mockOverlay) View() string                          { return m.content }
func (m mockOverlay) Done() bool                            { return m.done }

func TestOverlayStackPushPop(t *testing.T) {
	s := &overlayStack{}
	o1 := mockOverlay{content: "first"}
	o2 := mockOverlay{content: "second"}

	s.push(o1)
	s.push(o2)

	if s.len() != 2 {
		t.Errorf("expected len 2, got %d", s.len())
	}

	top := s.top()
	if top.View() != "second" {
		t.Errorf("expected top 'second', got %q", top.View())
	}

	s.pop()
	if s.len() != 1 {
		t.Errorf("expected len 1, got %d", s.len())
	}
	if s.top().View() != "first" {
		t.Errorf("expected top 'first', got %q", s.top().View())
	}
}

func TestOverlayStackEmpty(t *testing.T) {
	s := &overlayStack{}
	if s.active() {
		t.Error("empty stack should not be active")
	}
	if s.top() != nil {
		t.Error("top of empty stack should be nil")
	}
	s.pop() // should not panic
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run TestOverlay`
Expected: FAIL — types don't exist

**Step 3: Write the overlay interface and stack**

Create `overlay.go`:

```go
package foundation

import tea "github.com/charmbracelet/bubbletea"

// Overlay is a modal UI element rendered on top of the current view.
// Overlays form a stack — the topmost overlay receives input.
type Overlay interface {
	Update(msg tea.Msg) (Overlay, tea.Cmd)
	View() string
	Done() bool
}

// ShowOverlayMsg tells the app to push an overlay onto the stack.
type ShowOverlayMsg struct {
	Overlay Overlay
}

// ShowOverlay returns a tea.Cmd that pushes an overlay.
func ShowOverlay(o Overlay) tea.Cmd {
	return func() tea.Msg {
		return ShowOverlayMsg{Overlay: o}
	}
}

// overlayStack manages a stack of overlays.
type overlayStack struct {
	entries []Overlay
}

func (s *overlayStack) push(o Overlay) {
	s.entries = append(s.entries, o)
}

func (s *overlayStack) pop() {
	if len(s.entries) > 0 {
		s.entries = s.entries[:len(s.entries)-1]
	}
}

func (s *overlayStack) top() Overlay {
	if len(s.entries) == 0 {
		return nil
	}
	return s.entries[len(s.entries)-1]
}

func (s *overlayStack) active() bool {
	return len(s.entries) > 0
}

func (s *overlayStack) len() int {
	return len(s.entries)
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run TestOverlay`
Expected: PASS (all 2 tests)

**Step 5: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add overlay.go overlay_test.go
git commit -m "feat: add Overlay interface and stack management"
```

---

### Task 4: Flash Message System

**Files:**
- Create: `flash.go`
- Create: `flash_test.go`

**Step 1: Write the failing test**

Create `flash_test.go`:

```go
package foundation

import (
	"testing"
)

func TestFlashSetAndDecay(t *testing.T) {
	f := &flashState{}
	f.set("Hello!", 5)

	if f.message != "Hello!" {
		t.Errorf("expected message 'Hello!', got %q", f.message)
	}
	if !f.active() {
		t.Error("flash should be active")
	}

	// Tick down
	for i := 0; i < 5; i++ {
		f.tick()
	}

	if f.active() {
		t.Error("flash should be inactive after decay")
	}
	if f.message != "" {
		t.Errorf("expected empty message, got %q", f.message)
	}
}

func TestFlashInactiveByDefault(t *testing.T) {
	f := &flashState{}
	if f.active() {
		t.Error("flash should be inactive by default")
	}
	f.tick() // should not panic
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run TestFlash`
Expected: FAIL

**Step 3: Write flash implementation**

Create `flash.go`:

```go
package foundation

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// FlashMsg tells the app to display a flash message.
type FlashMsg struct {
	Message  string
	Duration int // number of ticks before auto-clear
}

// Flash returns a tea.Cmd that shows a flash message.
// Duration is in ticks (~100ms each), so 30 = ~3 seconds.
func Flash(msg string, duration int) tea.Cmd {
	return func() tea.Msg {
		return FlashMsg{Message: msg, Duration: duration}
	}
}

// flashTickMsg is the internal tick for flash decay.
type flashTickMsg struct{}

func flashTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return flashTickMsg{}
	})
}

// flashState manages a single flash message with auto-decay.
type flashState struct {
	message string
	timer   int
}

func (f *flashState) set(msg string, duration int) {
	f.message = msg
	f.timer = duration
}

func (f *flashState) tick() {
	if f.timer > 0 {
		f.timer--
		if f.timer <= 0 {
			f.message = ""
		}
	}
}

func (f *flashState) active() bool {
	return f.timer > 0 && f.message != ""
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run TestFlash`
Expected: PASS

**Step 5: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add flash.go flash_test.go
git commit -m "feat: add flash message system with auto-decay"
```

---

### Task 5: Breadcrumb and Status Bar Rendering

**Files:**
- Create: `breadcrumb.go`
- Create: `statusbar.go`
- Create: `chrome_test.go`

**Step 1: Write the failing test**

Create `chrome_test.go`:

```go
package foundation

import (
	"strings"
	"testing"

	"github.com/NCarteazy/bubbletea-foundation/theme"
)

func TestRenderBreadcrumbs(t *testing.T) {
	crumbs := []string{"Dashboard", "Reports", "Detail"}
	result := renderBreadcrumbs(crumbs, theme.TokyoNight)

	if result == "" {
		t.Error("breadcrumbs should not be empty")
	}
	// Should contain all crumb text
	for _, c := range crumbs {
		if !strings.Contains(result, c) {
			t.Errorf("breadcrumbs should contain %q", c)
		}
	}
}

func TestRenderBreadcrumbsEmpty(t *testing.T) {
	result := renderBreadcrumbs(nil, theme.TokyoNight)
	if result != "" {
		t.Error("empty breadcrumbs should return empty string")
	}
}

func TestRenderStatusBar(t *testing.T) {
	hints := []KeyHint{
		{Key: "j/k", Desc: "navigate"},
		{Key: "enter", Desc: "select"},
		{Key: "q", Desc: "quit"},
	}
	result := renderStatusBar(hints, nil, 80, theme.TokyoNight)

	if result == "" {
		t.Error("status bar should not be empty")
	}
	if !strings.Contains(result, "navigate") {
		t.Error("status bar should contain hint descriptions")
	}
}

func TestRenderStatusBarWithFlash(t *testing.T) {
	hints := []KeyHint{{Key: "q", Desc: "quit"}}
	flash := &flashState{message: "Copied!", timer: 10}
	result := renderStatusBar(hints, flash, 80, theme.TokyoNight)

	if !strings.Contains(result, "Copied!") {
		t.Error("status bar should show flash message when active")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run "TestRender(Breadcrumb|StatusBar)"`
Expected: FAIL

**Step 3: Write breadcrumb renderer**

Create `breadcrumb.go`:

```go
package foundation

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

func renderBreadcrumbs(crumbs []string, th theme.Theme) string {
	if len(crumbs) == 0 {
		return ""
	}

	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Muted())).
		Render(" > ")

	parts := make([]string, len(crumbs))
	for i, c := range crumbs {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted()))
		if i == len(crumbs)-1 {
			style = th.Breadcrumb()
		}
		parts[i] = style.Render(c)
	}

	return strings.Join(parts, sep)
}
```

**Step 4: Write status bar renderer**

Create `statusbar.go`:

```go
package foundation

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// KeyHint is a key-description pair shown in the status bar.
type KeyHint struct {
	Key  string
	Desc string
}

func renderStatusBar(hints []KeyHint, flash *flashState, width int, th theme.Theme) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Primary())).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Muted()))

	var content string

	if flash != nil && flash.active() {
		content = lipgloss.NewStyle().
			Foreground(lipgloss.Color(th.Success())).
			Bold(true).
			Render(flash.message)
	} else {
		parts := make([]string, len(hints))
		for i, h := range hints {
			parts[i] = keyStyle.Render(h.Key) + " " + descStyle.Render(h.Desc)
		}
		content = strings.Join(parts, " │ ")
	}

	return th.StatusBar().Copy().
		Width(width).
		Align(lipgloss.Center).
		Render(content)
}
```

**Step 5: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run "TestRender(Breadcrumb|StatusBar)"`
Expected: PASS

**Step 6: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add breadcrumb.go statusbar.go chrome_test.go
git commit -m "feat: add breadcrumb and status bar renderers"
```

---

### Task 6: Layout Helpers

**Files:**
- Create: `layout/panes.go`
- Create: `layout/center.go`
- Create: `layout/layout_test.go`

**Step 1: Write the failing test**

Create `layout/layout_test.go`:

```go
package layout_test

import (
	"testing"

	"github.com/NCarteazy/bubbletea-foundation/layout"
)

func TestTwoPaneBasic(t *testing.T) {
	result := layout.TwoPane("LEFT", "RIGHT", 80, 24)
	if result == "" {
		t.Error("TwoPane should produce output")
	}
}

func TestTwoPaneWithRatio(t *testing.T) {
	result := layout.TwoPaneWithRatio("LEFT", "RIGHT", 0.3, 80, 24)
	if result == "" {
		t.Error("TwoPaneWithRatio should produce output")
	}
}

func TestOverlayCenterBasic(t *testing.T) {
	result := layout.OverlayCenter(80, 24, "background", "overlay")
	if result == "" {
		t.Error("OverlayCenter should produce output")
	}
}

func TestOverlayCenterZeroSize(t *testing.T) {
	result := layout.OverlayCenter(0, 0, "bg", "overlay")
	if result != "overlay" {
		t.Errorf("expected fallback to overlay, got %q", result)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test ./layout/ -v`
Expected: FAIL

**Step 3: Write layout helpers**

Create `layout/panes.go`:

```go
package layout

import "github.com/charmbracelet/lipgloss"

// TwoPane renders two panes side by side, split evenly with a 1-char gap.
func TwoPane(left, right string, width, height int) string {
	return TwoPaneWithRatio(left, right, 0.5, width, height)
}

// TwoPaneWithRatio renders two panes with a custom left-side ratio (0.0 to 1.0).
func TwoPaneWithRatio(left, right string, ratio float64, width, height int) string {
	gap := 1
	available := width - gap
	leftW := int(float64(available) * ratio)
	rightW := available - leftW

	leftPane := lipgloss.NewStyle().
		Width(leftW).
		Height(height).
		Render(left)
	rightPane := lipgloss.NewStyle().
		Width(rightW).
		Height(height).
		Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)
}
```

Create `layout/center.go`:

```go
package layout

import "github.com/charmbracelet/lipgloss"

// OverlayCenter places an overlay panel centered on top of a background.
func OverlayCenter(width, height int, background, overlay string) string {
	if width <= 0 || height <= 0 {
		return overlay
	}
	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test ./layout/ -v`
Expected: PASS

**Step 5: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add layout/
git commit -m "feat: add layout helpers (two-pane split, overlay centering)"
```

---

### Task 7: App Orchestrator

**Files:**
- Create: `app.go`
- Create: `app_test.go`

**Step 1: Write the failing test**

Create `app_test.go`:

```go
package foundation

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// testView is a minimal View for testing.
type testView struct {
	id      string
	content string
}

func (v testView) ID() string                                        { return v.id }
func (v testView) Update(msg tea.Msg, ctx ViewContext) (View, tea.Cmd) { return v, nil }
func (v testView) Render(ctx ViewContext) string                      { return v.content }

func TestNewApp(t *testing.T) {
	app := New(
		WithTitle("Test App"),
		WithTheme(theme.TokyoNight),
		WithInitialView("home"),
		WithViews(
			testView{id: "home", content: "Home Page"},
			testView{id: "detail", content: "Detail Page"},
		),
		WithStatusHints([]KeyHint{
			{Key: "q", Desc: "quit"},
		}),
	)

	if app == nil {
		t.Fatal("New() returned nil")
	}
}

func TestAppView(t *testing.T) {
	app := New(
		WithTitle("Test App"),
		WithTheme(theme.TokyoNight),
		WithInitialView("home"),
		WithViews(
			testView{id: "home", content: "Home Content"},
		),
		WithStatusHints([]KeyHint{
			{Key: "q", Desc: "quit"},
		}),
	)

	// Simulate window size
	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	a := model.(*App)

	view := a.View()
	if !strings.Contains(view, "Home Content") {
		t.Error("View should contain the current view's content")
	}
}

func TestAppNavigation(t *testing.T) {
	app := New(
		WithTheme(theme.TokyoNight),
		WithInitialView("home"),
		WithViews(
			testView{id: "home", content: "Home"},
			testView{id: "detail", content: "Detail"},
		),
		WithStatusHints([]KeyHint{}),
	)

	// Simulate window size first
	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate to detail
	model, _ = model.Update(NavigateMsg{ViewID: "detail", Data: nil})
	a := model.(*App)

	if a.nav.current().viewID != "detail" {
		t.Errorf("expected current view 'detail', got %q", a.nav.current().viewID)
	}

	// Go back
	model, _ = model.Update(BackMsg{})
	a = model.(*App)

	if a.nav.current().viewID != "home" {
		t.Errorf("expected current view 'home', got %q", a.nav.current().viewID)
	}
}

func TestAppOverlay(t *testing.T) {
	app := New(
		WithTheme(theme.TokyoNight),
		WithInitialView("home"),
		WithViews(testView{id: "home", content: "Home"}),
		WithStatusHints([]KeyHint{}),
	)

	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	overlay := mockOverlay{content: "Confirm?"}
	model, _ = model.Update(ShowOverlayMsg{Overlay: overlay})
	a := model.(*App)

	if !a.overlays.active() {
		t.Error("overlay stack should be active")
	}

	view := a.View()
	if !strings.Contains(view, "Confirm?") {
		t.Error("View should contain overlay content")
	}
}

func TestAppFlash(t *testing.T) {
	app := New(
		WithTheme(theme.TokyoNight),
		WithInitialView("home"),
		WithViews(testView{id: "home", content: "Home"}),
		WithStatusHints([]KeyHint{}),
	)

	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model, _ = model.Update(FlashMsg{Message: "Saved!", Duration: 10})
	a := model.(*App)

	if !a.flash.active() {
		t.Error("flash should be active")
	}

	view := a.View()
	if !strings.Contains(view, "Saved!") {
		t.Error("View should contain flash message")
	}
}

func TestAppCtrlCQuits(t *testing.T) {
	app := New(
		WithTheme(theme.TokyoNight),
		WithInitialView("home"),
		WithViews(testView{id: "home", content: "Home"}),
		WithStatusHints([]KeyHint{}),
	)

	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("ctrl+c should return a quit command")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run TestApp`
Expected: FAIL — App doesn't exist

**Step 3: Write the App orchestrator**

Create `app.go`:

```go
package foundation

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NCarteazy/bubbletea-foundation/layout"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// App is the top-level Bubble Tea model that manages views, overlays,
// navigation, and chrome (breadcrumbs, status bar, flash messages).
type App struct {
	title       string
	th          theme.Theme
	views       map[string]View
	initialView string
	statusHints []KeyHint
	helpKeys    []KeySection

	nav      navStack
	overlays overlayStack
	flash    flashState

	width, height int
}

// KeySection groups key hints under a title for the help overlay.
type KeySection struct {
	Title string
	Keys  []KeyHint
}

// Option configures the App.
type Option func(*App)

// WithTitle sets the application title.
func WithTitle(title string) Option {
	return func(a *App) { a.title = title }
}

// WithTheme sets the color theme.
func WithTheme(th theme.Theme) Option {
	return func(a *App) { a.th = th }
}

// WithInitialView sets which view to show first.
func WithInitialView(id string) Option {
	return func(a *App) { a.initialView = id }
}

// WithViews registers views with the app.
func WithViews(views ...View) Option {
	return func(a *App) {
		for _, v := range views {
			a.views[v.ID()] = v
		}
	}
}

// WithStatusHints sets the default key hints in the status bar.
func WithStatusHints(hints []KeyHint) Option {
	return func(a *App) { a.statusHints = hints }
}

// WithHelp sets the help overlay content.
func WithHelp(sections []KeySection) Option {
	return func(a *App) { a.helpKeys = sections }
}

// New creates a new App with the given options.
func New(opts ...Option) *App {
	a := &App{
		views: make(map[string]View),
		th:    theme.TokyoNight,
	}
	for _, opt := range opts {
		opt(a)
	}
	if a.initialView != "" {
		a.nav.push(a.initialView, nil)
	}
	return a
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		// Global keys
		if msg.Type == tea.KeyCtrlC {
			return a, tea.Quit
		}

		// If overlay is active, route input there
		if a.overlays.active() {
			top := a.overlays.top()
			updated, cmd := top.Update(msg)
			if updated.Done() {
				a.overlays.pop()
			} else {
				a.overlays.entries[len(a.overlays.entries)-1] = updated
			}
			return a, cmd
		}

		// Help toggle
		if msg.String() == "?" && len(a.helpKeys) > 0 {
			a.overlays.push(newBuiltinHelp(a.helpKeys, a.width, a.height, a.th))
			return a, nil
		}

	case NavigateMsg:
		a.nav.push(msg.ViewID, msg.Data)
		return a, nil

	case BackMsg:
		if a.nav.len() > 1 {
			a.nav.pop()
		}
		return a, nil

	case ReplaceMsg:
		a.nav.replace(msg.ViewID, msg.Data)
		return a, nil

	case ShowOverlayMsg:
		a.overlays.push(msg.Overlay)
		return a, nil

	case FlashMsg:
		a.flash.set(msg.Message, msg.Duration)
		return a, flashTickCmd()

	case flashTickMsg:
		a.flash.tick()
		if a.flash.active() {
			return a, flashTickCmd()
		}
		return a, nil
	}

	// Route to current view
	current := a.nav.current()
	if current == nil {
		return a, nil
	}
	view, ok := a.views[current.viewID]
	if !ok {
		return a, nil
	}

	ctx := a.viewContext()
	updated, cmd := view.Update(msg, ctx)
	a.views[current.viewID] = updated
	return a, cmd
}

func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Initializing..."
	}

	// Render chrome
	crumbs := renderBreadcrumbs(a.nav.breadcrumbs(), a.th)
	bar := renderStatusBar(a.statusHints, &a.flash, a.width, a.th)

	// Calculate content area
	chromeHeight := 2 // breadcrumbs + status bar
	if crumbs == "" {
		chromeHeight = 1
	}
	contentHeight := a.height - chromeHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Render current view
	content := ""
	current := a.nav.current()
	if current != nil {
		if view, ok := a.views[current.viewID]; ok {
			ctx := ViewContext{
				Width:  a.width,
				Height: contentHeight,
				Theme:  a.th,
				Data:   current.data,
			}
			content = view.Render(ctx)
		}
	}

	// Compose: breadcrumbs + content + status bar
	var parts []string
	if crumbs != "" {
		parts = append(parts, crumbs)
	}
	parts = append(parts, content, bar)
	base := strings.Join(parts, "\n")

	// Render overlays on top
	if a.overlays.active() {
		overlay := a.overlays.top().View()
		base = layout.OverlayCenter(a.width, a.height, base, overlay)
	}

	return base
}

func (a *App) viewContext() ViewContext {
	chromeHeight := 2
	if len(a.nav.breadcrumbs()) == 0 {
		chromeHeight = 1
	}
	contentHeight := a.height - chromeHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	var data any
	if c := a.nav.current(); c != nil {
		data = c.data
	}

	return ViewContext{
		Width:  a.width,
		Height: contentHeight,
		Theme:  a.th,
		Data:   data,
	}
}

// builtinHelp is the built-in help overlay.
type builtinHelp struct {
	sections []KeySection
	width    int
	height   int
	th       theme.Theme
	done     bool
}

func newBuiltinHelp(sections []KeySection, width, height int, th theme.Theme) *builtinHelp {
	return &builtinHelp{sections: sections, width: width, height: height, th: th}
}

func (h *builtinHelp) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "?", "esc", "q":
			h.done = true
		}
	}
	return h, nil
}

func (h *builtinHelp) View() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Primary())).
		Bold(true).
		Width(16)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Text()))
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Secondary())).
		Bold(true).
		MarginTop(1)

	var lines []string
	title := h.th.Title().Render("Keyboard Shortcuts")
	lines = append(lines, title, "")

	for _, section := range h.sections {
		lines = append(lines, sectionStyle.Render(section.Title))
		for _, k := range section.Keys {
			lines = append(lines, keyStyle.Render(k.Key)+descStyle.Render(k.Desc))
		}
	}

	lines = append(lines, "", lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Muted())).
		Render("Press ? or esc to close"))

	content := strings.Join(lines, "\n")

	maxW := h.width - 10
	if maxW > 60 {
		maxW = 60
	}
	return h.th.Panel().Copy().
		Width(maxW).
		Render(content)
}

func (h *builtinHelp) Done() bool {
	return h.done
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test . -v -run TestApp`
Expected: PASS (all 6 tests)

**Step 5: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add app.go app_test.go
git commit -m "feat: add App orchestrator with view routing, overlays, flash, and help"
```

---

### Task 8: Table Component

**Files:**
- Create: `components/table.go`
- Create: `components/table_test.go`

**Step 1: Write the failing test**

Create `components/table_test.go`:

```go
package components_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NCarteazy/bubbletea-foundation/components"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

func TestTableRender(t *testing.T) {
	cols := []components.Column{
		{Name: "Name", Width: 20, Align: lipgloss.Left},
		{Name: "Size", Width: 10, Align: lipgloss.Right},
	}
	rows := [][]string{
		{"file1.txt", "1.2 MB"},
		{"file2.txt", "3.4 GB"},
	}

	tbl := components.NewTable(cols, rows, theme.TokyoNight)
	tbl.SetSize(40, 10)
	view := tbl.View()

	if view == "" {
		t.Error("table should render content")
	}
}

func TestTableNavigation(t *testing.T) {
	cols := []components.Column{
		{Name: "Name", Width: 20, Align: lipgloss.Left},
	}
	rows := [][]string{{"a"}, {"b"}, {"c"}}

	tbl := components.NewTable(cols, rows, theme.TokyoNight)
	tbl.SetSize(30, 10)

	if tbl.SelectedRow() != 0 {
		t.Errorf("initial selection should be 0, got %d", tbl.SelectedRow())
	}

	// Move down
	tbl, _ = tbl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tbl.SelectedRow() != 1 {
		t.Errorf("after j, selection should be 1, got %d", tbl.SelectedRow())
	}

	// Move up
	tbl, _ = tbl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tbl.SelectedRow() != 0 {
		t.Errorf("after k, selection should be 0, got %d", tbl.SelectedRow())
	}

	// Don't go above 0
	tbl, _ = tbl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tbl.SelectedRow() != 0 {
		t.Errorf("should not go below 0, got %d", tbl.SelectedRow())
	}
}

func TestTableSort(t *testing.T) {
	cols := []components.Column{
		{Name: "Name", Width: 20, Align: lipgloss.Left},
		{Name: "Count", Width: 10, Align: lipgloss.Right},
	}
	rows := [][]string{{"a", "10"}, {"b", "5"}, {"c", "20"}}

	tbl := components.NewTable(cols, rows, theme.TokyoNight)
	tbl.SetSize(40, 10)

	// Press s to sort
	tbl, _ = tbl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	// Table should still have 3 rows
	if tbl.RowCount() != 3 {
		t.Errorf("expected 3 rows, got %d", tbl.RowCount())
	}
}

func TestTableEmpty(t *testing.T) {
	cols := []components.Column{
		{Name: "Name", Width: 20, Align: lipgloss.Left},
	}
	tbl := components.NewTable(cols, nil, theme.TokyoNight)
	tbl.SetSize(30, 10)
	view := tbl.View()
	if view == "" {
		t.Error("empty table should still render header")
	}
}

func TestTableSetRows(t *testing.T) {
	cols := []components.Column{{Name: "Name", Width: 20, Align: lipgloss.Left}}
	tbl := components.NewTable(cols, [][]string{{"a"}, {"b"}}, theme.TokyoNight)

	tbl.SetRows([][]string{{"x"}, {"y"}, {"z"}})
	if tbl.RowCount() != 3 {
		t.Errorf("expected 3 rows after SetRows, got %d", tbl.RowCount())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test ./components/ -v`
Expected: FAIL

**Step 3: Write the Table component**

Create `components/table.go`:

```go
package components

import (
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// Column defines a single table column.
type Column struct {
	Name  string
	Width int
	Align lipgloss.Position
}

// Table is a sortable, scrollable table component.
type Table struct {
	columns  []Column
	rows     [][]string
	cursor   int
	offset   int
	height   int
	width    int
	sortCol  int
	sortDesc bool
	th       theme.Theme
}

// NewTable creates a new Table with columns, rows, and a theme.
func NewTable(columns []Column, rows [][]string, th theme.Theme) Table {
	if rows == nil {
		rows = [][]string{}
	}
	return Table{
		columns:  columns,
		rows:     rows,
		sortCol:  -1,
		sortDesc: true,
		th:       th,
	}
}

// Update handles keyboard navigation and sorting.
func (t Table) Update(msg tea.Msg) (Table, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
			}
			if t.cursor < t.offset {
				t.offset = t.cursor
			}
		case "down", "j":
			if t.cursor < len(t.rows)-1 {
				t.cursor++
			}
			visible := t.visibleRowCount()
			if t.cursor >= t.offset+visible {
				t.offset = t.cursor - visible + 1
			}
		case "s":
			t.sortCol++
			if t.sortCol >= len(t.columns) {
				t.sortCol = 0
				t.sortDesc = !t.sortDesc
			}
			t.sortRows()
		case "home", "g":
			t.cursor = 0
			t.offset = 0
		case "end", "G":
			if len(t.rows) > 0 {
				t.cursor = len(t.rows) - 1
				visible := t.visibleRowCount()
				if t.cursor >= visible {
					t.offset = t.cursor - visible + 1
				}
			}
		}
	}
	return t, nil
}

// View renders the table.
func (t Table) View() string {
	if len(t.columns) == 0 {
		return ""
	}

	const colGap = 2
	gap := strings.Repeat(" ", colGap)

	var b strings.Builder

	// Header
	headerParts := make([]string, len(t.columns))
	for i, col := range t.columns {
		name := col.Name
		if i == t.sortCol {
			if t.sortDesc {
				name += " v"
			} else {
				name += " ^"
			}
		}
		headerParts[i] = t.th.TableHeader().Copy().
			Width(col.Width).
			Align(col.Align).
			Render(truncate(name, col.Width))
	}
	b.WriteString(strings.Join(headerParts, gap))
	b.WriteString("\n")

	// Rows
	visible := t.visibleRowCount()
	end := t.offset + visible
	if end > len(t.rows) {
		end = len(t.rows)
	}

	altRowStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.th.Surface()))

	for i := t.offset; i < end; i++ {
		row := t.rows[i]
		cellParts := make([]string, len(t.columns))
		for j, col := range t.columns {
			val := ""
			if j < len(row) {
				val = row[j]
			}
			cellStyle := lipgloss.NewStyle().
				Width(col.Width).
				Align(col.Align)

			if i == t.cursor {
				cellStyle = t.th.SelectedRow().Copy().
					Width(col.Width).
					Align(col.Align)
			} else if (i-t.offset)%2 == 1 {
				cellStyle = altRowStyle.Copy().
					Width(col.Width).
					Align(col.Align)
			}

			cellParts[j] = cellStyle.Render(truncate(val, col.Width))
		}
		b.WriteString(strings.Join(cellParts, gap))
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// SelectedRow returns the current cursor position.
func (t Table) SelectedRow() int {
	return t.cursor
}

// RowCount returns the number of rows.
func (t Table) RowCount() int {
	return len(t.rows)
}

// SetRows replaces row data and resets cursor if needed.
func (t *Table) SetRows(rows [][]string) {
	t.rows = rows
	if t.cursor >= len(rows) {
		t.cursor = 0
	}
	t.offset = 0
}

// SetSize updates the table dimensions.
func (t *Table) SetSize(width, height int) {
	t.width = width
	t.height = height
}

func (t Table) visibleRowCount() int {
	if t.height <= 1 {
		return 10
	}
	return t.height
}

func (t *Table) sortRows() {
	if t.sortCol < 0 || t.sortCol >= len(t.columns) {
		return
	}
	col := t.sortCol
	desc := t.sortDesc
	sort.SliceStable(t.rows, func(i, j int) bool {
		a := ""
		b := ""
		if col < len(t.rows[i]) {
			a = t.rows[i][col]
		}
		if col < len(t.rows[j]) {
			b = t.rows[j][col]
		}
		aNum, aErr := parseNumeric(a)
		bNum, bErr := parseNumeric(b)
		if aErr == nil && bErr == nil {
			if desc {
				return aNum > bNum
			}
			return aNum < bNum
		}
		if desc {
			return a > b
		}
		return a < b
	})
}

func parseNumeric(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	s = strings.ReplaceAll(s, ",", "")

	multipliers := map[string]float64{
		"B":  1,
		"kB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}
	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, " "+suffix) {
			numStr := strings.TrimSuffix(s, " "+suffix)
			numStr = strings.TrimSpace(numStr)
			f, err := strconv.ParseFloat(numStr, 64)
			if err == nil {
				return f * mult, nil
			}
		}
	}

	return strconv.ParseFloat(s, 64)
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(s) <= width {
		return s
	}
	if width <= 1 {
		return "\u2026"
	}
	return s[:width-1] + "\u2026"
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test ./components/ -v`
Expected: PASS (all 5 tests)

**Step 5: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add components/
git commit -m "feat: add Table component with sorting, scrolling, and navigation"
```

---

### Task 9: Confirm Component

**Files:**
- Create: `components/confirm.go`
- Create: `components/confirm_test.go`

**Step 1: Write the failing test**

Create `components/confirm_test.go`:

```go
package components_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/NCarteazy/bubbletea-foundation/components"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

func TestConfirmRender(t *testing.T) {
	c := components.NewConfirm("Delete file.txt?", theme.TokyoNight)
	view := c.View()

	if !strings.Contains(view, "Delete file.txt?") {
		t.Error("confirm should contain the prompt")
	}
}

func TestConfirmYes(t *testing.T) {
	called := false
	c := components.NewConfirm("Delete?", theme.TokyoNight)
	c.OnConfirm = func() tea.Cmd {
		called = true
		return nil
	}

	updated, _ := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	confirm := updated.(*components.Confirm)

	if !confirm.Done() {
		t.Error("confirm should be done after 'y'")
	}
	if !called {
		t.Error("OnConfirm should have been called")
	}
}

func TestConfirmNo(t *testing.T) {
	c := components.NewConfirm("Delete?", theme.TokyoNight)

	updated, _ := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	confirm := updated.(*components.Confirm)

	if !confirm.Done() {
		t.Error("confirm should be done after 'n'")
	}
}

func TestConfirmEsc(t *testing.T) {
	c := components.NewConfirm("Delete?", theme.TokyoNight)

	updated, _ := c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	confirm := updated.(*components.Confirm)

	if !confirm.Done() {
		t.Error("confirm should be done after esc")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test ./components/ -v -run TestConfirm`
Expected: FAIL

**Step 3: Write the Confirm component**

Create `components/confirm.go`:

```go
package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// Confirm is a generic confirmation dialog overlay.
type Confirm struct {
	Prompt    string
	OnConfirm func() tea.Cmd
	th        theme.Theme
	done      bool
	confirmed bool
}

// NewConfirm creates a new confirmation dialog.
func NewConfirm(prompt string, th theme.Theme) *Confirm {
	return &Confirm{
		Prompt: prompt,
		th:     th,
	}
}

// Update handles y/n/esc input.
func (c *Confirm) Update(msg tea.Msg) (interface{ Update(tea.Msg) (interface{ Update(tea.Msg) (interface{}, tea.Cmd); View() string; Done() bool }, tea.Cmd); View() string; Done() bool }, tea.Cmd) {
	// This signature is unwieldy — we implement the Overlay interface from the parent package.
	// But since components can't import foundation (circular), we use a simpler approach.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			c.done = true
			c.confirmed = true
			if c.OnConfirm != nil {
				return c, c.OnConfirm()
			}
			return c, nil
		case "n", "N", "esc":
			c.done = true
			return c, nil
		}
	}
	return c, nil
}

// View renders the confirmation dialog.
func (c *Confirm) View() string {
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.th.Warning())).
		Bold(true)
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.th.Muted()))

	content := fmt.Sprintf(
		"%s\n\n%s",
		promptStyle.Render(c.Prompt),
		hintStyle.Render("[y] confirm  [n/esc] cancel"),
	)

	return c.th.Panel().Copy().
		Width(50).
		Render(content)
}

// Done returns whether the dialog has been dismissed.
func (c *Confirm) Done() bool {
	return c.done
}

// Confirmed returns whether the user confirmed.
func (c *Confirm) Confirmed() bool {
	return c.confirmed
}
```

Wait — that `Update` signature is wrong because of circular imports. Let me fix the approach. The Confirm component should implement a simple interface that the App can wrap.

Replace `components/confirm.go` with:

```go
package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// Confirm is a generic confirmation dialog.
// It implements a standalone Update/View/Done pattern.
// To use as an Overlay, wrap it with foundation.OverlayFromConfirm().
type Confirm struct {
	Prompt    string
	OnConfirm func() tea.Cmd
	th        theme.Theme
	done      bool
	confirmed bool
}

// NewConfirm creates a new confirmation dialog.
func NewConfirm(prompt string, th theme.Theme) *Confirm {
	return &Confirm{
		Prompt: prompt,
		th:     th,
	}
}

// Update handles y/n/esc input.
func (c *Confirm) Update(msg tea.Msg) (*Confirm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			c.done = true
			c.confirmed = true
			if c.OnConfirm != nil {
				return c, c.OnConfirm()
			}
			return c, nil
		case "n", "N", "esc":
			c.done = true
			return c, nil
		}
	}
	return c, nil
}

// View renders the confirmation dialog.
func (c *Confirm) View() string {
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.th.Warning())).
		Bold(true)
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.th.Muted()))

	content := fmt.Sprintf(
		"%s\n\n%s",
		promptStyle.Render(c.Prompt),
		hintStyle.Render("[y] confirm  [n/esc] cancel"),
	)

	return c.th.Panel().Copy().
		Width(50).
		Render(content)
}

// Done returns whether the dialog has been dismissed.
func (c *Confirm) Done() bool {
	return c.done
}

// Confirmed returns whether the user confirmed.
func (c *Confirm) Confirmed() bool {
	return c.confirmed
}
```

And update `components/confirm_test.go` to match:

```go
package components_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/NCarteazy/bubbletea-foundation/components"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

func TestConfirmRender(t *testing.T) {
	c := components.NewConfirm("Delete file.txt?", theme.TokyoNight)
	view := c.View()

	if !strings.Contains(view, "Delete file.txt?") {
		t.Error("confirm should contain the prompt")
	}
}

func TestConfirmYes(t *testing.T) {
	called := false
	c := components.NewConfirm("Delete?", theme.TokyoNight)
	c.OnConfirm = func() tea.Cmd {
		called = true
		return nil
	}

	updated, _ := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if !updated.Done() {
		t.Error("confirm should be done after 'y'")
	}
	if !called {
		t.Error("OnConfirm should have been called")
	}
}

func TestConfirmNo(t *testing.T) {
	c := components.NewConfirm("Delete?", theme.TokyoNight)

	updated, _ := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if !updated.Done() {
		t.Error("confirm should be done after 'n'")
	}
}

func TestConfirmEsc(t *testing.T) {
	c := components.NewConfirm("Delete?", theme.TokyoNight)

	updated, _ := c.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !updated.Done() {
		t.Error("confirm should be done after esc")
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test ./components/ -v -run TestConfirm`
Expected: PASS (all 4 tests)

**Step 5: Add an Overlay adapter in the root package**

Add to `app.go` (or create `confirm_overlay.go`):

```go
// confirmOverlay wraps a components.Confirm as an Overlay.
type confirmOverlay struct {
	confirm *components.Confirm
}

// ConfirmOverlay creates an Overlay from a Confirm component.
func ConfirmOverlay(c *components.Confirm) Overlay {
	return &confirmOverlay{confirm: c}
}

func (o *confirmOverlay) Update(msg tea.Msg) (Overlay, tea.Cmd) {
	updated, cmd := o.confirm.Update(msg)
	o.confirm = updated
	return o, cmd
}

func (o *confirmOverlay) View() string {
	return o.confirm.View()
}

func (o *confirmOverlay) Done() bool {
	return o.confirm.Done()
}
```

**Step 6: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add components/confirm.go components/confirm_test.go app.go
git commit -m "feat: add Confirm component with Overlay adapter"
```

---

### Task 10: Help Component

**Files:**
- Create: `components/help.go`
- Create: `components/help_test.go`

**Step 1: Write the failing test**

Create `components/help_test.go`:

```go
package components_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/NCarteazy/bubbletea-foundation/components"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

func TestHelpRender(t *testing.T) {
	sections := []components.HelpSection{
		{
			Title: "Navigation",
			Keys: []components.HelpKey{
				{Key: "j/k", Desc: "move up/down"},
				{Key: "enter", Desc: "select"},
			},
		},
	}

	h := components.NewHelp(sections, 80, 24, theme.TokyoNight)
	view := h.View()

	if !strings.Contains(view, "Navigation") {
		t.Error("help should contain section title")
	}
	if !strings.Contains(view, "move up/down") {
		t.Error("help should contain key descriptions")
	}
}

func TestHelpDismiss(t *testing.T) {
	h := components.NewHelp(nil, 80, 24, theme.TokyoNight)

	updated, _ := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !updated.Done() {
		t.Error("help should be done after esc")
	}
}

func TestHelpDismissQuestion(t *testing.T) {
	h := components.NewHelp(nil, 80, 24, theme.TokyoNight)

	updated, _ := h.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !updated.Done() {
		t.Error("help should be done after ?")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/nick/bubbletea-foundation && go test ./components/ -v -run TestHelp`
Expected: FAIL

**Step 3: Write the Help component**

Create `components/help.go`:

```go
package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// HelpKey is a single key binding entry.
type HelpKey struct {
	Key  string
	Desc string
}

// HelpSection groups key bindings under a title.
type HelpSection struct {
	Title string
	Keys  []HelpKey
}

// Help is a modal help overlay showing key bindings.
type Help struct {
	Sections []HelpSection
	width    int
	height   int
	th       theme.Theme
	done     bool
}

// NewHelp creates a new help overlay.
func NewHelp(sections []HelpSection, width, height int, th theme.Theme) *Help {
	return &Help{
		Sections: sections,
		width:    width,
		height:   height,
		th:       th,
	}
}

// Update handles dismiss keys.
func (h *Help) Update(msg tea.Msg) (*Help, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "?", "esc", "q":
			h.done = true
		}
	}
	return h, nil
}

// View renders the help panel.
func (h *Help) View() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Primary())).
		Bold(true).
		Width(16)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Text()))
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Secondary())).
		Bold(true).
		MarginTop(1)

	var lines []string
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Primary())).
		Bold(true).
		Render("Keyboard Shortcuts")
	lines = append(lines, title, "")

	for _, section := range h.Sections {
		lines = append(lines, sectionStyle.Render(section.Title))
		for _, k := range section.Keys {
			lines = append(lines, keyStyle.Render(k.Key)+descStyle.Render(k.Desc))
		}
	}

	lines = append(lines, "", lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.th.Muted())).
		Render("Press ? or esc to close"))

	content := strings.Join(lines, "\n")

	maxW := h.width - 10
	if maxW > 60 {
		maxW = 60
	}
	return h.th.Panel().Copy().
		Width(maxW).
		Render(content)
}

// Done returns whether the overlay has been dismissed.
func (h *Help) Done() bool {
	return h.done
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/nick/bubbletea-foundation && go test ./components/ -v -run TestHelp`
Expected: PASS (all 3 tests)

**Step 5: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add components/help.go components/help_test.go
git commit -m "feat: add Help component with grouped key binding sections"
```

---

### Task 11: Example App

**Files:**
- Create: `example/go.mod`
- Create: `example/main.go`

**Step 1: Create the example module**

```bash
mkdir -p /Users/nick/bubbletea-foundation/example
cd /Users/nick/bubbletea-foundation/example
go mod init github.com/NCarteazy/bubbletea-foundation/example
```

Then edit `example/go.mod` to add a replace directive for local development:

```
module github.com/NCarteazy/bubbletea-foundation/example

go 1.22

require github.com/NCarteazy/bubbletea-foundation v0.0.0

replace github.com/NCarteazy/bubbletea-foundation => ../
```

**Step 2: Write the example app**

Create `example/main.go`:

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	foundation "github.com/NCarteazy/bubbletea-foundation"
	"github.com/NCarteazy/bubbletea-foundation/components"
	"github.com/NCarteazy/bubbletea-foundation/theme"
)

// --- Home View ---

type homeView struct{}

func (v homeView) ID() string { return "Home" }

func (v homeView) Update(msg tea.Msg, ctx foundation.ViewContext) (foundation.View, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter":
			return v, foundation.Navigate("Files", nil)
		case "q":
			return v, tea.Quit
		}
	}
	return v, nil
}

func (v homeView) Render(ctx foundation.ViewContext) string {
	title := ctx.Theme.Title().Render("Welcome to the Example App")
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ctx.Theme.Muted())).
		Render("This demonstrates bubbletea-foundation")

	bar := ctx.Theme.SizeBar(0.65, 40)
	barLabel := fmt.Sprintf("Disk Usage: 65%% %s", bar)

	return fmt.Sprintf("%s\n%s\n\n%s\n\nPress enter to browse files, q to quit.",
		title, subtitle, barLabel)
}

// --- Files View ---

type filesView struct {
	table components.Table
	built bool
}

func (v filesView) ID() string { return "Files" }

func (v filesView) Update(msg tea.Msg, ctx foundation.ViewContext) (foundation.View, tea.Cmd) {
	if !v.built {
		v.table = buildFileTable(ctx.Theme)
		v.table.SetSize(ctx.Width, ctx.Height-2)
		v.built = true
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "esc":
			return v, foundation.Back()
		case "enter":
			return v, foundation.Navigate("Detail", v.table.SelectedRow())
		case "d":
			c := components.NewConfirm("Delete this file?", ctx.Theme)
			c.OnConfirm = func() tea.Cmd {
				return foundation.Flash("File deleted!", 30)
			}
			return v, foundation.ShowOverlay(foundation.ConfirmOverlay(c))
		}
	}

	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

func (v filesView) Render(ctx foundation.ViewContext) string {
	if !v.built {
		return "Loading..."
	}
	return v.table.View()
}

func buildFileTable(th theme.Theme) components.Table {
	cols := []components.Column{
		{Name: "Name", Width: 30, Align: lipgloss.Left},
		{Name: "Size", Width: 12, Align: lipgloss.Right},
		{Name: "Modified", Width: 20, Align: lipgloss.Left},
	}
	rows := [][]string{
		{"Documents", "2.4 GB", "2026-03-14"},
		{"Downloads", "8.1 GB", "2026-03-15"},
		{"node_modules", "1.2 GB", "2026-02-20"},
		{"Library/Caches", "4.7 GB", "2026-03-15"},
		{".docker", "12.3 GB", "2026-03-10"},
		{"Applications", "15.6 GB", "2026-03-01"},
		{"Pictures", "3.2 GB", "2025-12-25"},
		{"Music", "900 MB", "2025-06-15"},
	}
	return components.NewTable(cols, rows, th)
}

// --- Detail View ---

type detailView struct{}

func (v detailView) ID() string { return "Detail" }

func (v detailView) Update(msg tea.Msg, ctx foundation.ViewContext) (foundation.View, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "esc" {
			return v, foundation.Back()
		}
	}
	return v, nil
}

func (v detailView) Render(ctx foundation.ViewContext) string {
	idx, _ := ctx.Data.(int)
	return fmt.Sprintf("Detail view for item %d\n\nPress esc to go back.", idx)
}

// --- Main ---

func main() {
	app := foundation.New(
		foundation.WithTitle("Example App"),
		foundation.WithTheme(theme.TokyoNight),
		foundation.WithInitialView("Home"),
		foundation.WithViews(
			homeView{},
			filesView{},
			detailView{},
		),
		foundation.WithHelp([]foundation.KeySection{
			{Title: "Navigation", Keys: []foundation.KeyHint{
				{Key: "j/k", Desc: "move up/down"},
				{Key: "enter", Desc: "select / drill in"},
				{Key: "esc", Desc: "go back"},
			}},
			{Title: "Actions", Keys: []foundation.KeyHint{
				{Key: "d", Desc: "delete selected"},
				{Key: "s", Desc: "cycle sort"},
			}},
		}),
		foundation.WithStatusHints([]foundation.KeyHint{
			{Key: "j/k", Desc: "navigate"},
			{Key: "enter", Desc: "select"},
			{Key: "esc", Desc: "back"},
			{Key: "?", Desc: "help"},
			{Key: "q", Desc: "quit"},
		}),
	)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Install dependencies and build**

```bash
cd /Users/nick/bubbletea-foundation/example
go mod tidy
go build .
```

**Step 4: Test it runs**

```bash
cd /Users/nick/bubbletea-foundation/example
go run .
```

Expected: Opens a TUI with the Home view. Press enter → Files table. Press ? → Help overlay. Press d → Confirm dialog. Press esc → Back. Press q → Quit.

**Step 5: Commit**

```bash
cd /Users/nick/bubbletea-foundation
git add example/
git commit -m "feat: add example app demonstrating all framework features"
```

---

### Task 12: Run All Tests and Push

**Step 1: Run full test suite**

```bash
cd /Users/nick/bubbletea-foundation
go test ./... -v
go vet ./...
```

Expected: All tests pass, no vet issues.

**Step 2: Run race detection**

```bash
cd /Users/nick/bubbletea-foundation
go test -race ./...
```

Expected: No race conditions.

**Step 3: Create GitHub repo and push**

```bash
cd /Users/nick/bubbletea-foundation
gh repo create bubbletea-foundation --public --source=. --push
```

**Step 4: Tag initial release**

```bash
cd /Users/nick/bubbletea-foundation
git tag v0.1.0
git push origin v0.1.0
```

---

### Task 13: Migrate mac-cleanup-explorer to Use Foundation

This is an optional follow-up task. After the foundation module is published, update mac-cleanup-explorer to import and use it instead of its own internal table, theme, and overlay code. This validates the framework with a real app.

**Not detailed here** — save for a separate plan after the foundation is stable.
