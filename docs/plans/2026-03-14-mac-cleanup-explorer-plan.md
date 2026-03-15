# Mac Cleanup Explorer — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a beautiful TUI app that scans a Mac filesystem, generates structured reports about disk usage, exports AI-friendly summaries, and executes cleanup commands safely.

**Architecture:** Single Go binary with three layers — Scanner (concurrent filesystem walker), Analyzer (7 report generators), and TUI (Bubble Tea + Lip Gloss). In-memory tree, no database.

**Tech Stack:** Go 1.22+, Bubble Tea (TUI framework), Lip Gloss (styling), Bubbles (components), clipboard lib, xxhash (fast hashing for duplicates)

---

## Task 0: Project Setup

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `.gitignore`
- Create: `CLAUDE.md`

**Step 1: Install Go via Homebrew**

```bash
brew install go
go version
```

Expected: `go version go1.22+`

**Step 2: Initialize git repo**

```bash
cd /Users/nick/mac-cleanup-explorer
git init
```

**Step 3: Create .gitignore**

```gitignore
# Binaries
mac-cleanup-explorer
*.exe
*.dylib

# Go
vendor/

# IDE
.idea/
.vscode/
*.swp

# macOS
.DS_Store

# App data
*.log
```

**Step 4: Initialize Go module**

```bash
go mod init github.com/nick/mac-cleanup-explorer
```

**Step 5: Install dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/atotto/clipboard
go get github.com/cespare/xxhash/v2
go get github.com/dustin/go-humanize
```

**Step 6: Create minimal main.go**

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m model) View() string {
	return "Mac Cleanup Explorer — press q to quit\n"
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 7: Create CLAUDE.md**

```markdown
# Mac Cleanup Explorer

## Build & Run
- `go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer`
- `go run .` for quick dev runs

## Test
- `go test ./...` to run all tests
- `go test ./internal/scanner/ -v` for verbose scanner tests
- `go test -race ./...` for race condition detection

## Lint
- `go vet ./...`

## Project Structure
- `main.go` — entrypoint
- `internal/scanner/` — filesystem scanner
- `internal/analyzer/` — report generators
- `internal/tui/` — Bubble Tea TUI views and components
- `internal/theme/` — color palette and styling constants
- `internal/export/` — AI export formatting
- `internal/executor/` — command parsing and execution

## Conventions
- Use `internal/` for all packages (not importable by external code)
- Each TUI view is its own file in `internal/tui/`
- Reports implement the `analyzer.Report` interface
- Tokyo Night color palette defined in `internal/theme/`
- All file sizes displayed via go-humanize
- Test files live next to their source files
```

**Step 8: Build and run to verify**

```bash
go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer
```

Expected: Alt screen with message, press `q` to exit cleanly.

**Step 9: Commit**

```bash
git add .gitignore go.mod go.sum main.go CLAUDE.md
git commit -m "feat: initial project setup with Go module and minimal Bubble Tea app"
```

---

## Task 1: Theme & Styling Constants

**Files:**
- Create: `internal/theme/theme.go`
- Create: `internal/theme/theme_test.go`

**Step 1: Write test**

```go
package theme

import "testing"

func TestColorsAreDefined(t *testing.T) {
	// Verify all theme colors render non-empty strings
	colors := []struct {
		name  string
		style string
	}{
		{"Background", BgColor},
		{"Primary", PrimaryColor},
		{"Secondary", SecondaryColor},
		{"Success", SuccessColor},
		{"Warning", WarningColor},
		{"Danger", DangerColor},
		{"Muted", MutedColor},
		{"Text", TextColor},
	}
	for _, c := range colors {
		if c.style == "" {
			t.Errorf("color %s is empty", c.name)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1024, "1.0 kB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tt := range tests {
		got := FormatSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestSizeBarColor(t *testing.T) {
	// Small percentage should be success color
	c := SizeBarColor(0.1)
	if c != SuccessColor {
		t.Errorf("expected success color for 10%%, got %s", c)
	}
	// Large percentage should be danger color
	c = SizeBarColor(0.9)
	if c != DangerColor {
		t.Errorf("expected danger color for 90%%, got %s", c)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/theme/ -v
```

Expected: FAIL — package doesn't exist yet.

**Step 3: Implement theme.go**

```go
package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

// Tokyo Night color palette
const (
	BgColor        = "#1a1b26"
	PrimaryColor   = "#7dcfff"
	SecondaryColor = "#bb9af7"
	SuccessColor   = "#9ece6a"
	WarningColor   = "#e0af68"
	DangerColor    = "#f7768e"
	MutedColor     = "#565f89"
	TextColor      = "#c0caf5"
	SurfaceColor   = "#24283b"
	OverlayColor   = "#414868"
)

// Reusable styles
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PrimaryColor)).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SecondaryColor))

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(MutedColor))

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SuccessColor))

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(WarningColor))

	DangerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(DangerColor))

	TextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor))

	// Panel with rounded borders
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(PrimaryColor)).
			Padding(1, 2)

	// Active/selected panel
	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(SecondaryColor)).
				Padding(1, 2)

	// Status bar at bottom
	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(SurfaceColor)).
			Foreground(lipgloss.Color(TextColor)).
			Padding(0, 1)

	// Breadcrumb bar at top
	BreadcrumbStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(SurfaceColor)).
			Foreground(lipgloss.Color(PrimaryColor)).
			Padding(0, 1)

	// Selected row
	SelectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(OverlayColor)).
				Foreground(lipgloss.Color(TextColor))

	// Table header
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(PrimaryColor)).
				Bold(true).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(OverlayColor))
)

// FormatSize formats bytes into human-readable sizes.
func FormatSize(bytes int64) string {
	return humanize.Bytes(uint64(bytes))
}

// SizeBarColor returns the appropriate color for a size bar given a ratio (0.0 to 1.0).
func SizeBarColor(ratio float64) string {
	switch {
	case ratio >= 0.7:
		return DangerColor
	case ratio >= 0.4:
		return WarningColor
	default:
		return SuccessColor
	}
}

// SizeBar renders a horizontal bar chart segment.
func SizeBar(width int, ratio float64, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 30
	}
	filled := int(ratio * float64(maxWidth))
	if filled < 1 && ratio > 0 {
		filled = 1
	}
	color := SizeBarColor(ratio)
	bar := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(repeatChar('█', filled))
	empty := lipgloss.NewStyle().
		Foreground(lipgloss.Color(MutedColor)).
		Render(repeatChar('░', maxWidth-filled))
	return bar + empty
}

func repeatChar(ch rune, count int) string {
	if count <= 0 {
		return ""
	}
	s := make([]rune, count)
	for i := range s {
		s[i] = ch
	}
	return string(s)
}
```

**Step 4: Run tests**

```bash
go test ./internal/theme/ -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/theme/
git commit -m "feat: add Tokyo Night theme with styling constants and size formatting"
```

---

## Task 2: Filesystem Scanner

**Files:**
- Create: `internal/scanner/scanner.go`
- Create: `internal/scanner/tree.go`
- Create: `internal/scanner/scanner_test.go`

**Step 1: Write tree.go — data structures**

```go
package scanner

import (
	"sync"
	"time"
)

// FileNode represents a file or directory in the scanned tree.
type FileNode struct {
	Name       string
	Path       string
	Size       int64      // For files: file size. For dirs: total recursive size.
	IsDir      bool
	ModTime    time.Time
	AccessTime time.Time
	Children   []*FileNode
	Parent     *FileNode
	FileCount  int64 // Number of files (recursive for dirs)
	DirCount   int64 // Number of subdirs (recursive for dirs)
	Error      string // Non-empty if there was an error accessing this node
}

// ScanResult holds the complete scan output.
type ScanResult struct {
	Root       *FileNode
	TotalSize  int64
	TotalFiles int64
	TotalDirs  int64
	Errors     []string
	Duration   time.Duration
}

// ScanProgress is sent during scanning to report status.
type ScanProgress struct {
	CurrentPath string
	FilesFound  int64
	DirsFound   int64
	BytesFound  int64
	mu          sync.Mutex
}

func (p *ScanProgress) Update(path string, isDir bool, size int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CurrentPath = path
	p.BytesFound += size
	if isDir {
		p.DirsFound++
	} else {
		p.FilesFound++
	}
}

func (p *ScanProgress) Snapshot() ScanProgress {
	p.mu.Lock()
	defer p.mu.Unlock()
	return ScanProgress{
		CurrentPath: p.CurrentPath,
		FilesFound:  p.FilesFound,
		DirsFound:   p.DirsFound,
		BytesFound:  p.BytesFound,
	}
}
```

**Step 2: Write scanner_test.go**

```go
package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirectory(t *testing.T) {
	// Create a temp directory structure
	tmp := t.TempDir()
	// Create files
	os.WriteFile(filepath.Join(tmp, "file1.txt"), make([]byte, 1024), 0644)
	os.WriteFile(filepath.Join(tmp, "file2.txt"), make([]byte, 2048), 0644)
	// Create subdirectory with a file
	sub := filepath.Join(tmp, "subdir")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "file3.txt"), make([]byte, 4096), 0644)

	progress := &ScanProgress{}
	result, err := Scan(tmp, progress)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.TotalFiles != 3 {
		t.Errorf("expected 3 files, got %d", result.TotalFiles)
	}
	if result.TotalDirs != 1 {
		t.Errorf("expected 1 subdir, got %d", result.TotalDirs)
	}
	// Root size should be sum of all files
	expectedSize := int64(1024 + 2048 + 4096)
	if result.Root.Size != expectedSize {
		t.Errorf("expected root size %d, got %d", expectedSize, result.Root.Size)
	}
	if result.Root.FileCount != 3 {
		t.Errorf("expected root FileCount 3, got %d", result.Root.FileCount)
	}
}

func TestScanHandlesPermissionError(t *testing.T) {
	tmp := t.TempDir()
	restricted := filepath.Join(tmp, "noaccess")
	os.Mkdir(restricted, 0000)
	defer os.Chmod(restricted, 0755) // cleanup

	progress := &ScanProgress{}
	result, err := Scan(tmp, progress)
	if err != nil {
		t.Fatalf("Scan should not fail on permission errors: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one error for restricted directory")
	}
}

func TestScanEmptyDirectory(t *testing.T) {
	tmp := t.TempDir()
	progress := &ScanProgress{}
	result, err := Scan(tmp, progress)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if result.TotalFiles != 0 {
		t.Errorf("expected 0 files, got %d", result.TotalFiles)
	}
	if result.Root.Size != 0 {
		t.Errorf("expected 0 size, got %d", result.Root.Size)
	}
}
```

**Step 3: Run tests to verify they fail**

```bash
go test ./internal/scanner/ -v
```

Expected: FAIL — `Scan` function doesn't exist.

**Step 4: Implement scanner.go**

```go
package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

// Scan walks the filesystem starting from root and returns a ScanResult.
func Scan(root string, progress *ScanProgress) (*ScanResult, error) {
	start := time.Now()

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving root path: %w", err)
	}

	rootInfo, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat root: %w", err)
	}

	rootNode := &FileNode{
		Name:  rootInfo.Name(),
		Path:  absRoot,
		IsDir: true,
	}

	result := &ScanResult{Root: rootNode}
	nodeMap := map[string]*FileNode{absRoot: rootNode}

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", path, err))
			return nil // continue scanning
		}

		if path == absRoot {
			return nil // skip root itself, already created
		}

		info, err := d.Info()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", path, err))
			return nil
		}

		node := &FileNode{
			Name:    d.Name(),
			Path:    path,
			IsDir:   d.IsDir(),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		}

		// Get access time from syscall
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			node.AccessTime = time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
		}

		if !d.IsDir() {
			node.Size = info.Size()
			result.TotalFiles++
			result.TotalSize += node.Size
		} else {
			result.TotalDirs++
		}

		// Find parent and attach
		parentPath := filepath.Dir(path)
		if parent, ok := nodeMap[parentPath]; ok {
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		}

		if d.IsDir() {
			nodeMap[path] = node
		}

		if progress != nil {
			progress.Update(path, d.IsDir(), node.Size)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking filesystem: %w", err)
	}

	// Calculate recursive sizes for directories (bottom-up)
	calculateDirSizes(rootNode)

	// Sort children by size descending at each level
	sortChildren(rootNode)

	result.Duration = time.Since(start)
	return result, nil
}

// calculateDirSizes recursively computes directory sizes bottom-up.
func calculateDirSizes(node *FileNode) {
	if !node.IsDir {
		node.FileCount = 1
		return
	}

	var totalSize int64
	var fileCount, dirCount int64

	for _, child := range node.Children {
		calculateDirSizes(child)
		totalSize += child.Size
		fileCount += child.FileCount
		if child.IsDir {
			dirCount += 1 + child.DirCount
		} else {
			// fileCount already counted recursively
		}
	}

	node.Size = totalSize
	node.FileCount = fileCount
	node.DirCount = dirCount
}

// sortChildren sorts all children at every level by size descending.
func sortChildren(node *FileNode) {
	if !node.IsDir {
		return
	}
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})
	for _, child := range node.Children {
		sortChildren(child)
	}
}

// SkipPaths returns paths that should be skipped during scanning.
func SkipPaths() []string {
	return []string{
		"/System/Volumes/Data", // avoid double-counting firmlinked content
	}
}

// ShouldSkip checks if a path should be skipped.
func ShouldSkip(path string) bool {
	for _, skip := range SkipPaths() {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}
```

**Step 5: Run tests**

```bash
go test ./internal/scanner/ -v
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/scanner/
git commit -m "feat: add filesystem scanner with concurrent walking and tree building"
```

---

## Task 3: Analyzer Framework + Space Treemap Report

**Files:**
- Create: `internal/analyzer/analyzer.go`
- Create: `internal/analyzer/space.go`
- Create: `internal/analyzer/analyzer_test.go`

**Step 1: Write analyzer.go — report interface**

```go
package analyzer

import "github.com/nick/mac-cleanup-explorer/internal/scanner"

// ReportItem represents a single item in a report.
type ReportItem struct {
	Path        string
	Size        int64
	Category    string
	Description string
	LastAccess  string
	LastMod     string
	FileCount   int64
	Severity    string // "low", "medium", "high"
}

// Report is the interface all report generators implement.
type Report interface {
	Name() string
	Description() string
	Generate(root *scanner.FileNode) []ReportItem
	AIContext() string // Explanation for AI export
}

// GenerateAll runs all reports against a scan result.
func GenerateAll(result *scanner.ScanResult) map[string][]ReportItem {
	reports := AllReports()
	out := make(map[string][]ReportItem, len(reports))
	for _, r := range reports {
		out[r.Name()] = r.Generate(result.Root)
	}
	return out
}

// AllReports returns all available report generators.
func AllReports() []Report {
	return []Report{
		&SpaceReport{},
		// More reports added in subsequent tasks
	}
}
```

**Step 2: Write space.go — space treemap report**

```go
package analyzer

import (
	"fmt"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
	"github.com/dustin/go-humanize"
)

// SpaceReport generates a hierarchical space breakdown.
type SpaceReport struct{}

func (r *SpaceReport) Name() string        { return "Space Treemap" }
func (r *SpaceReport) Description() string  { return "Hierarchical breakdown of disk space by directory" }
func (r *SpaceReport) AIContext() string {
	return "This report shows disk space consumption by directory. " +
		"Identify the largest directories and suggest which can be safely reduced or removed. " +
		"Consider whether large directories contain caches, temporary files, or outdated data."
}

func (r *SpaceReport) Generate(root *scanner.FileNode) []ReportItem {
	var items []ReportItem
	// Top-level children only (user drills down in TUI)
	for _, child := range root.Children {
		pct := float64(0)
		if root.Size > 0 {
			pct = float64(child.Size) / float64(root.Size) * 100
		}
		severity := "low"
		if pct > 20 {
			severity = "high"
		} else if pct > 10 {
			severity = "medium"
		}
		items = append(items, ReportItem{
			Path:        child.Path,
			Size:        child.Size,
			Category:    categoryFor(child),
			Description: fmt.Sprintf("%.1f%% of total (%s, %d files)", pct, humanize.Bytes(uint64(child.Size)), child.FileCount),
			FileCount:   child.FileCount,
			Severity:    severity,
		})
	}
	return items
}

func categoryFor(node *scanner.FileNode) string {
	// Simple categorization based on path
	switch node.Name {
	case "Library":
		return "System & App Data"
	case "Applications":
		return "Applications"
	case "Documents":
		return "Documents"
	case "Downloads":
		return "Downloads"
	case "Desktop":
		return "Desktop"
	case "Pictures", "Photos":
		return "Media"
	case "Movies", "Music":
		return "Media"
	case "Developer", "Projects", "Code", "src", "repos":
		return "Developer"
	case ".Trash":
		return "Trash"
	default:
		return "Other"
	}
}
```

**Step 3: Write analyzer_test.go**

```go
package analyzer

import (
	"testing"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

func makeTestTree() *scanner.FileNode {
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 100000,
		Children: []*scanner.FileNode{
			{Name: "Library", Path: "/Library", IsDir: true, Size: 50000, FileCount: 100},
			{Name: "Applications", Path: "/Applications", IsDir: true, Size: 30000, FileCount: 50},
			{Name: "Documents", Path: "/Documents", IsDir: true, Size: 20000, FileCount: 200},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}
	return root
}

func TestSpaceReport(t *testing.T) {
	root := makeTestTree()
	report := &SpaceReport{}

	items := report.Generate(root)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// First item should be Library (largest)
	if items[0].Path != "/Library" {
		t.Errorf("expected first item to be /Library, got %s", items[0].Path)
	}
	if items[0].Severity != "high" {
		t.Errorf("expected high severity for 50%%, got %s", items[0].Severity)
	}
	if items[0].Category != "System & App Data" {
		t.Errorf("expected category 'System & App Data', got %s", items[0].Category)
	}
}

func TestSpaceReportEmptyRoot(t *testing.T) {
	root := &scanner.FileNode{Name: "/", Path: "/", IsDir: true}
	report := &SpaceReport{}
	items := report.Generate(root)
	if len(items) != 0 {
		t.Errorf("expected 0 items for empty root, got %d", len(items))
	}
}

func TestAllReports(t *testing.T) {
	reports := AllReports()
	if len(reports) == 0 {
		t.Fatal("expected at least one report")
	}
	for _, r := range reports {
		if r.Name() == "" {
			t.Error("report has empty name")
		}
		if r.AIContext() == "" {
			t.Error("report has empty AI context")
		}
	}
}
```

**Step 4: Run tests**

```bash
go test ./internal/analyzer/ -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/analyzer/
git commit -m "feat: add analyzer framework with space treemap report"
```

---

## Task 4: Remaining Reports (Large Files, Stale, Caches, Dev Bloat, App Leftovers, Duplicates)

**Files:**
- Create: `internal/analyzer/large_files.go`
- Create: `internal/analyzer/stale.go`
- Create: `internal/analyzer/caches.go`
- Create: `internal/analyzer/devbloat.go`
- Create: `internal/analyzer/leftovers.go`
- Create: `internal/analyzer/duplicates.go`
- Modify: `internal/analyzer/analyzer.go` — register all reports in `AllReports()`
- Create: `internal/analyzer/reports_test.go`

Each report implements the `Report` interface. The plan provides the core logic for each:

**Step 1: Write tests for all reports**

Create `internal/analyzer/reports_test.go` with tests for each report type using the test tree helper. Tests should verify:
- Correct number of items returned for known test data
- Correct categorization and severity
- Edge case: empty tree returns no items
- Each report's Name(), Description(), and AIContext() are non-empty

**Step 2: Implement large_files.go**

Walk the tree recursively, collect all files over a threshold (default 100MB). Sort by size descending. Return as ReportItems with severity based on size (>1GB = high, >500MB = medium, else low).

**Step 3: Implement stale.go**

Walk tree, collect files/dirs with AccessTime older than 6 months. Group by category based on parent path. Severity based on size * staleness.

**Step 4: Implement caches.go**

Match known cache paths:
- `*/Library/Caches/*`
- `*/Library/Logs/*`
- `*/.cache/*`
- `*/DerivedData/*`
- `*/.npm/*`, `*/.yarn/*`, `*/.pnpm-store/*`
- `*/pip/cache/*`, `*/.cargo/registry/*`
- `*/.gradle/caches/*`
- `*/Docker/` (in Library)

Return matched directories with their sizes.

**Step 5: Implement devbloat.go**

Find:
- `node_modules` directories
- `.git` directories over 50MB
- `build/`, `dist/`, `target/`, `out/` directories
- `venv/`, `.venv/`, `env/` directories
- `__pycache__` directories

Severity: high for items over 500MB, medium over 100MB.

**Step 6: Implement leftovers.go**

Scan `~/Library/Application Support/`, `~/Library/Preferences/`, `~/Library/Containers/`. Cross-reference with installed apps in `/Applications/` and `~/Applications/`. Flag support files where no matching `.app` exists.

**Step 7: Implement duplicates.go**

Phase 1: Group files by size (only files > 1MB).
Phase 2: For groups with 2+ files, compute xxhash of first 4KB.
Phase 3: For still-matching files, compute full hash.
Return groups of duplicate files.

**Step 8: Register all reports in AllReports()**

Update `analyzer.go` to include all 7 reports.

**Step 9: Run all tests**

```bash
go test ./internal/analyzer/ -v
```

Expected: PASS

**Step 10: Commit**

```bash
git add internal/analyzer/
git commit -m "feat: add all 7 report generators (large files, stale, caches, dev bloat, leftovers, duplicates)"
```

---

## Task 5: AI Export Formatter

**Files:**
- Create: `internal/export/export.go`
- Create: `internal/export/export_test.go`

**Step 1: Write tests**

```go
package export

import (
	"strings"
	"testing"

	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
)

func TestFormatReport(t *testing.T) {
	items := []analyzer.ReportItem{
		{Path: "/Users/test/big.dmg", Size: 5368709120, Category: "Downloads", Severity: "high"},
		{Path: "/Users/test/old.zip", Size: 1073741824, Category: "Downloads", Severity: "medium"},
	}
	sysInfo := SystemInfo{
		OSVersion:  "macOS 15.0",
		DiskSize:   "500 GB",
		FreeSpace:  "50 GB",
		Machine:    "MacBook Pro (Apple M1)",
		ScanScope:  "/Users/test",
		ScanTime:   "2.3s",
	}

	md := FormatMarkdown("Large Files", "Files over 100MB", items, sysInfo)

	if !strings.Contains(md, "# Mac Cleanup Explorer Report") {
		t.Error("missing report header")
	}
	if !strings.Contains(md, "Large Files") {
		t.Error("missing report name")
	}
	if !strings.Contains(md, "big.dmg") {
		t.Error("missing file entry")
	}
	if !strings.Contains(md, "macOS 15.0") {
		t.Error("missing system info")
	}
}

func TestTokenEstimate(t *testing.T) {
	text := strings.Repeat("word ", 100) // ~100 words ≈ ~133 tokens
	estimate := EstimateTokens(text)
	if estimate < 100 || estimate > 200 {
		t.Errorf("token estimate %d seems off for 100 words", estimate)
	}
}

func TestTruncateItems(t *testing.T) {
	items := make([]analyzer.ReportItem, 100)
	for i := range items {
		items[i] = analyzer.ReportItem{Path: "/test", Size: int64(100 - i)}
	}
	truncated, summary := TruncateItems(items, 20)
	if len(truncated) != 20 {
		t.Errorf("expected 20 items, got %d", len(truncated))
	}
	if summary == "" {
		t.Error("expected truncation summary")
	}
}
```

**Step 2: Implement export.go**

The formatter takes report items and system info, produces structured markdown with:
- System info header
- Report context section
- Summary stats
- Data table (markdown table format)
- Token estimate
- Suggested AI prompts per report type
- Truncation with summary when items exceed limit

Also implement JSON export as alternative format.

**Step 3: Run tests**

```bash
go test ./internal/export/ -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/export/
git commit -m "feat: add AI export formatter with markdown/JSON output and token estimation"
```

---

## Task 6: TUI — Scan Screen

**Files:**
- Create: `internal/tui/app.go` — main Bubble Tea model, view routing
- Create: `internal/tui/scan.go` — scanning view with progress
- Create: `internal/tui/keys.go` — key bindings
- Modify: `main.go` — wire up the real TUI

**Step 1: Implement app.go — main model with view routing**

The main model holds:
- Current view state (scanning, dashboard, reports, detail, export, executor)
- Scan result
- Window dimensions
- Active report data

It routes Update/View to the active sub-view.

**Step 2: Implement keys.go — global key bindings**

Define key map: q/ctrl+c quit, esc back, tab switch pane, ? help, j/k/up/down navigate, enter select.

**Step 3: Implement scan.go — scanning progress view**

Shows:
- App title/logo at top (ASCII art, styled with theme colors)
- Animated spinner (charm bubbles spinner)
- Progress stats: files scanned, dirs scanned, data found
- Current path being scanned (truncated to fit)
- Elapsed time

Uses a `tea.Tick` command to poll the `ScanProgress` struct every 100ms.

Scanning runs in a goroutine, sends a `scanCompleteMsg` when done.

**Step 4: Update main.go to use the real TUI**

Replace the placeholder model with the real `tui.App` model.

**Step 5: Build and manually test**

```bash
go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer
```

Expected: See scan screen with animated progress, then it transitions (to a placeholder for now).

**Step 6: Commit**

```bash
git add internal/tui/ main.go
git commit -m "feat: add TUI framework with animated scan screen"
```

---

## Task 7: TUI — Dashboard

**Files:**
- Create: `internal/tui/dashboard.go`

**Step 1: Implement dashboard view**

After scanning completes, show:
- Summary card: total disk size, used, free, scan duration
- Category breakdown with size bars (using theme.SizeBar)
- Quick stats highlights ("5 items over 1GB", "12GB in caches", etc.)
- Navigation hint: "Press enter to browse reports, e to export"

The dashboard computes categories by grouping top-level children of the scan tree.

**Step 2: Build and test visually**

```bash
go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer
```

Expected: Scan completes → dashboard shows with colored bars and stats.

**Step 3: Commit**

```bash
git add internal/tui/dashboard.go
git commit -m "feat: add dashboard view with category breakdown and size bars"
```

---

## Task 8: TUI — Report Browser

**Files:**
- Create: `internal/tui/reports.go`
- Create: `internal/tui/table.go` — reusable table component

**Step 1: Implement table.go — reusable styled table**

A generic table component with:
- Column headers (styled with TableHeaderStyle)
- Sortable columns (click header or key to cycle sort)
- Scrollable rows with selected row highlighting
- Alternating row background
- Column width auto-calculation
- Truncation with `…` for long values

**Step 2: Implement reports.go — report browser view**

Left sidebar: list of report names with item counts and total sizes.
Right pane: selected report rendered as a table.

Navigation:
- `tab` to switch between sidebar and table
- `j/k` to navigate within active pane
- `enter` on sidebar selects report
- `enter` on table row opens detail view
- Breadcrumbs update: `Dashboard > Reports > [Report Name]`

For the Space Treemap report, `enter` on a directory drills into its children (pushes to a navigation stack for back/forward).

**Step 3: Build and test**

```bash
go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer
```

Expected: Can browse all 7 reports, sort tables, drill into space treemap.

**Step 4: Commit**

```bash
git add internal/tui/reports.go internal/tui/table.go
git commit -m "feat: add report browser with sortable tables and drill-down navigation"
```

---

## Task 9: TUI — Detail View & Export Panel

**Files:**
- Create: `internal/tui/detail.go`
- Create: `internal/tui/export.go`

**Step 1: Implement detail.go**

When a report item is selected, show:
- Full path (copyable)
- Size (formatted)
- Last accessed, last modified
- Category, severity (color-coded)
- Why it was flagged
- Action keys: `d` delete, `m` move to trash, `y` copy path

**Step 2: Implement export.go — export panel overlay**

Modal overlay showing:
- Checkbox list of available reports
- Preview pane showing formatted markdown
- Token estimate for selected reports
- Keys: `space` toggle report, `c` copy to clipboard, `s` save to file, `esc` close

Uses `atotto/clipboard` for clipboard copy. Saves to `~/mac-cleanup-report-YYYY-MM-DD.md`.

**Step 3: Build and test**

```bash
go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer
```

Expected: Can view details, export reports to clipboard and file.

**Step 4: Commit**

```bash
git add internal/tui/detail.go internal/tui/export.go
git commit -m "feat: add detail view and export panel with clipboard support"
```

---

## Task 10: TUI — Command Executor

**Files:**
- Create: `internal/executor/executor.go`
- Create: `internal/executor/executor_test.go`
- Create: `internal/tui/executor.go`

**Step 1: Write executor tests**

```go
package executor

import "testing"

func TestParseCommands(t *testing.T) {
	input := "rm -rf ~/Library/Caches/old\nrm ~/Downloads/big.dmg"
	cmds := ParseCommands(input)
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
}

func TestBlocksDangerousPaths(t *testing.T) {
	cmd := Command{Raw: "rm -rf /System/Library"}
	err := ValidateCommand(cmd)
	if err == nil {
		t.Error("expected error for dangerous path")
	}
}

func TestBlocksSudo(t *testing.T) {
	cmd := Command{Raw: "sudo rm -rf /tmp/old"}
	err := ValidateCommand(cmd)
	if err == nil {
		t.Error("expected error for sudo command")
	}
}
```

**Step 2: Implement executor.go**

```go
package executor

// Command represents a parsed cleanup command.
type Command struct {
	Raw         string
	Validated   bool
	Error       string
	Output      string
	ExitCode    int
	Executed    bool
}

// ParseCommands splits a block of text into individual commands.
// Filters empty lines and comments.

// ValidateCommand checks a command against safety rules.
// Blocks: system paths, sudo (unless allowed), excessively large deletes.

// ExecuteCommand runs a single command and captures output.
// Returns error if command fails.

// DangerousPaths lists paths that should never be targeted.
var DangerousPaths = []string{
	"/System", "/usr", "/bin", "/sbin", "/etc",
	"/Library/Apple", "/private/var/db",
}
```

**Step 3: Implement TUI executor view**

Text input area for pasting commands. Parsed commands shown as a list with:
- Command text (syntax-highlighted — keywords in cyan, paths in purple, flags in amber)
- Validation status (green check or red X)
- Execute button per command or batch execute
- Dry-run toggle (`n` key)
- Output area below each command after execution

**Step 4: Run tests**

```bash
go test ./internal/executor/ -v
```

Expected: PASS

**Step 5: Build and test end-to-end**

```bash
go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer
```

Expected: Full flow works — scan, browse, export, paste commands, execute.

**Step 6: Commit**

```bash
git add internal/executor/ internal/tui/executor.go
git commit -m "feat: add command executor with safety validation and dry-run mode"
```

---

## Task 11: Direct Delete Actions

**Files:**
- Modify: `internal/tui/reports.go` — add delete/trash key handlers
- Modify: `internal/tui/detail.go` — add delete/trash key handlers
- Create: `internal/tui/confirm.go` — confirmation dialog component

**Step 1: Implement confirm.go**

Reusable confirmation dialog:
- Styled modal overlay
- Shows action description, item path, size
- `y` to confirm, `n`/`esc` to cancel
- Red-highlighted warning for large deletions

**Step 2: Add delete actions to reports and detail views**

- `d` key triggers confirmation dialog, then `os.RemoveAll` on confirm
- `m` key moves to `~/.Trash/` instead
- `space` to select multiple items, `D` for bulk delete with summary confirmation
- After action: remove item from current view, update size totals

**Step 3: Build and test**

```bash
go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer
```

Expected: Can delete/trash items from report browser and detail view.

**Step 4: Commit**

```bash
git add internal/tui/confirm.go internal/tui/reports.go internal/tui/detail.go
git commit -m "feat: add direct delete/trash actions with confirmation dialogs"
```

---

## Task 12: Help Overlay & Polish

**Files:**
- Create: `internal/tui/help.go`
- Modify: `internal/tui/app.go` — wire help toggle
- Create: `internal/tui/logo.go` — ASCII art logo

**Step 1: Implement logo.go**

ASCII art "Mac Cleanup Explorer" styled with gradient theme colors (cyan → purple). Shown on scan screen and help overlay.

**Step 2: Implement help.go**

`?` toggles a help overlay showing all key bindings grouped by context:
- Global keys
- Navigation keys
- Report browser keys
- Action keys
- Export keys

Styled as a centered panel with the theme's panel border.

**Step 3: Polish pass**

- Ensure consistent padding/margins across all views
- Verify breadcrumbs update correctly for all navigation
- Status bar shows context-appropriate key hints on every view
- Window resize handling (respond to `tea.WindowSizeMsg`)
- Truncate long paths cleanly with `…`
- Ensure all colors use theme constants (no hardcoded colors)

**Step 4: Build and full manual test**

```bash
go build -o mac-cleanup-explorer . && ./mac-cleanup-explorer
```

Test: scan → dashboard → each report → drill down → detail → export → paste commands → execute → help.

**Step 5: Commit**

```bash
git add internal/tui/
git commit -m "feat: add help overlay, ASCII logo, and visual polish pass"
```

---

## Task 13: Action History Log

**Files:**
- Create: `internal/executor/history.go`
- Create: `internal/executor/history_test.go`

**Step 1: Write test**

Test that actions are logged to `~/.mac-cleanup-explorer/history.log` with timestamp, command, result, and size freed.

**Step 2: Implement history.go**

- Creates `~/.mac-cleanup-explorer/` directory if needed
- Appends to `history.log` with format: `[2026-03-14 15:04:05] COMMAND: rm ... | RESULT: success | FREED: 1.5 GB`
- Log both direct TUI actions and executor commands

**Step 3: Run tests**

```bash
go test ./internal/executor/ -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add internal/executor/history.go internal/executor/history_test.go
git commit -m "feat: add action history logging to ~/.mac-cleanup-explorer/history.log"
```

---

## Task 14: Final Integration & README

**Files:**
- Modify: `main.go` — final wiring, flag parsing for scan scope
- Create: `Makefile`

**Step 1: Add CLI flags to main.go**

```go
flag.StringVar(&scanPath, "path", "/", "Root path to scan")
flag.BoolVar(&skipSystem, "skip-system", false, "Skip /System and other OS paths")
```

**Step 2: Create Makefile**

```makefile
.PHONY: build run test clean

build:
	go build -o mac-cleanup-explorer .

run: build
	./mac-cleanup-explorer

test:
	go test ./... -v

clean:
	rm -f mac-cleanup-explorer
```

**Step 3: Run full test suite**

```bash
go test ./... -v -race
```

Expected: All tests PASS, no race conditions.

**Step 4: Build and final end-to-end test**

```bash
make build && ./mac-cleanup-explorer
```

**Step 5: Commit**

```bash
git add main.go Makefile
git commit -m "feat: add CLI flags and Makefile for final integration"
```

---

## Summary

| Task | Description | Key Files |
|------|-------------|-----------|
| 0 | Project setup | `go.mod`, `main.go`, `.gitignore`, `CLAUDE.md` |
| 1 | Theme & styling | `internal/theme/` |
| 2 | Filesystem scanner | `internal/scanner/` |
| 3 | Analyzer framework + space report | `internal/analyzer/` |
| 4 | All 7 reports | `internal/analyzer/*.go` |
| 5 | AI export formatter | `internal/export/` |
| 6 | TUI scan screen | `internal/tui/scan.go`, `app.go` |
| 7 | TUI dashboard | `internal/tui/dashboard.go` |
| 8 | TUI report browser | `internal/tui/reports.go`, `table.go` |
| 9 | TUI detail & export | `internal/tui/detail.go`, `export.go` |
| 10 | Command executor | `internal/executor/`, `internal/tui/executor.go` |
| 11 | Direct delete actions | `internal/tui/confirm.go` |
| 12 | Help overlay & polish | `internal/tui/help.go`, `logo.go` |
| 13 | Action history log | `internal/executor/history.go` |
| 14 | Final integration | `main.go`, `Makefile` |
