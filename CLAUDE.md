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
