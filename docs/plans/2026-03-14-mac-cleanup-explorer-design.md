# Mac Cleanup Explorer — Design Document

**Date:** 2026-03-14
**Status:** Approved

## Overview

A TUI application for macOS that scans the entire filesystem, generates structured reports about disk usage, and makes it easy to export findings to an AI for cleanup recommendations — then execute those recommendations safely.

**Stack:** Go + Bubble Tea (TUI) + Lip Gloss (styling)
**Architecture:** Single binary, in-memory scanner + analyzer + TUI

## Core Architecture

Three layers in a single Go binary:

### 1. Scanner
- Walks filesystem using `filepath.WalkDir` with concurrent worker pool
- Builds in-memory tree of directories and files with metadata (size, mod time, access time)
- Handles permission errors gracefully (skips and notes them)
- Scans full filesystem by default, with option to scope to specific paths
- Live progress reporting back to TUI during scan

### 2. Analyzer
- Runs report generators against the scanned tree
- Each report is a Go function implementing a common interface
- Reports are generated after scan completes (fast — just traversing in-memory data)
- 7 report types (see Reports section)

### 3. TUI
- Bubble Tea for application framework
- Lip Gloss for styling
- Keyboard-driven navigation
- 6 main views (see TUI Layout section)

**Flow:** Scan → Analyze → Browse → Export → Act

## TUI Layout & Navigation

### Views

1. **Scan screen** — animated progress bar with live stats (files scanned, dirs visited, current path, elapsed time). Starts automatically on launch.

2. **Dashboard** — landing after scan. Summary card with total disk usage, top-level breakdown by category (apps, caches, dev tools, media, system, other), quick stats ("5 items over 1GB", "12GB in stale caches").

3. **Report browser** — left sidebar lists available reports, right pane shows selected report data. Each report is a sorted, filterable table or tree. Drill into directories, expand subtrees, toggle sorting by size/count/age.

4. **Detail view** — full path, size, last accessed, last modified, associated app/tool, and why it was flagged.

5. **Export panel** — overlay/modal to select reports, preview structured output, copy to clipboard or save to file. Markdown format optimized for AI consumption.

6. **Command executor** — paste AI-generated bash commands, syntax-highlighted preview, approve individually or batch, live output.

### Navigation

- Vim-style: `j/k` move, `enter` drill in, `esc`/`q` back, `tab` switch panes, `?` help
- Arrow keys also supported
- Breadcrumbs in top bar: `Dashboard > Caches > ~/Library/Caches`
- Status bar at bottom with key hints, scan stats, current filter

## Visual Design System

### Color Palette (Tokyo Night inspired)

| Role | Color | Hex |
|------|-------|-----|
| Background | Deep charcoal | `#1a1b26` |
| Primary accent | Electric cyan | `#7dcfff` |
| Secondary accent | Soft purple | `#bb9af7` |
| Success/safe | Mint green | `#9ece6a` |
| Warning | Warm amber | `#e0af68` |
| Danger/large | Coral red | `#f7768e` |
| Muted text | Slate gray | `#565f89` |
| Bright text | Soft white | `#c0caf5` |

### Components

- **Size bars** — horizontal gradient fills (green → amber → red) showing proportional size
- **Cards/panels** — rounded box-drawing chars (`╭╮╰╯`) with accent-colored borders, consistent padding
- **Tables** — alternating subtle row shading, aligned columns, truncation with `…`
- **Breadcrumbs** — persistent top bar showing current location
- **Status bar** — bottom bar with key hints and context
- **Spinners/progress** — charm animated spinners, smooth progress bar with ETA
- **Icons** — unicode symbols (`📁 📄 🗑️ ⚠️ ✓`) with ASCII/nerd font fallback

### Micro-interactions

- Smooth list scrolling via Bubble Tea viewport
- Highlighted row with subtle background shift
- Flash confirmation on clipboard copy
- Auto-formatted sizes (bytes → KB → MB → GB)

## Reports & Analysis

### 1. Space Treemap
Hierarchical directory breakdown. Size, percentage of total, child count per directory. Drill into any level. Primary exploration view.

### 2. Large Files
Files over 100MB (configurable). Sorted by size descending. Shows path, size, last modified, last accessed, file type.

### 3. Stale Files & Directories
Items not accessed in 6+ months (configurable). Grouped by category (downloads, documents, app data). Highlights forgotten items.

### 4. Cache & Temp Data
Known cache locations:
- `~/Library/Caches/*`, `~/Library/Logs/*`
- Homebrew — old formula downloads, outdated versions
- npm/yarn/pnpm cache, pip cache, Go module cache
- Xcode derived data, iOS simulators
- Docker images/volumes
- Browser caches (Chrome, Safari, Firefox)
- System temp (`/tmp`, `/var/folders`)

### 5. Developer Bloat
Finds: `node_modules`, large `.git` directories, build output (`dist/`, `build/`, `target/`), virtual environments (`venv`, `.venv`, `conda`), inactive project directories.

### 6. Application Leftovers
Cross-references installed apps (`~/Applications`, `/Applications`) with support files (`~/Library/Application Support`, `~/Library/Preferences`, `~/Library/Containers`). Flags orphaned support files.

### 7. Duplicates
Size-based candidates, hash-confirmed. Grouped by content, all locations shown. Focuses on files over 1MB.

## AI Export Format

```markdown
# Mac Cleanup Explorer Report
## System Info
- macOS version, disk size, free space, machine model
- Scan timestamp, scan scope, scan duration

## Report: [Report Name]
### Context
Brief explanation of what this report shows and what kind
of cleanup advice is useful.

### Summary
- Total items: X
- Total size: X.XX GB
- Top category: ...

### Data
| Path | Size | Last Accessed | Category | Notes |
|------|------|---------------|----------|-------|
(sorted, truncated to top N items)

### Suggested Prompts
Pre-written prompts for the AI, e.g.:
"Review these cache directories and suggest which are safe
to delete. Provide exact rm commands."
```

**Features:**
- Token estimate per report for context window awareness
- Intelligent truncation — top N by size, summary of omitted items
- Markdown (for pasting) or JSON (for programmatic use)
- Clipboard copy defaults to markdown
- Multiple reports combinable into single export

## Command Executor

### Direct Actions (in-TUI)
- `d` — delete selected item (with confirmation)
- `D` — bulk delete (select with `space`, then delete)
- `m` — move to Trash (safer default)
- Confirmation dialog shows exactly what will happen
- Undo via Trash for moved items

### Paste-back Executor
- Import commands view (`ctrl+v` or dedicated view)
- Parses and displays each command individually, syntax-highlighted
- Shows: command, target paths, estimated space freed
- Execute one-by-one (`enter`) or batch (`ctrl+a` + `enter`)
- Live output streaming per command
- Color-coded results (green ✓ / red ✗ with error)
- **Dry-run mode** (`n` toggle) — preview before executing

### Safety Guardrails
- Blocks commands targeting system-critical paths (`/System`, `/usr`, `/bin`)
- Warns on deletions exceeding configurable threshold (default 10GB)
- No `sudo` unless explicitly toggled on
- Full action log at `~/.mac-cleanup-explorer/history.log`
