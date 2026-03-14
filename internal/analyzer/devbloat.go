package analyzer

import (
	"fmt"
	"strings"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

// DevBloatReport finds developer-related bloat directories.
type DevBloatReport struct{}

func (r *DevBloatReport) Name() string        { return "devbloat" }
func (r *DevBloatReport) Description() string { return "Developer Bloat — build artifacts, deps, caches" }
func (r *DevBloatReport) AIContext() string {
	return "Identifies developer-related bloat: node_modules, build output directories, " +
		"Python virtual environments, large .git directories, and Java build caches. " +
		"These can often be cleaned and regenerated with build tools."
}

// devBloatTargets maps directory names to their categories.
// Some have a minimum size threshold.
type devBloatTarget struct {
	name     string
	category string
	minSize  int64 // 0 means always report
}

var devBloatTargets = []devBloatTarget{
	{"node_modules", "Node.js Dependencies", 0},
	{".git", "Git Repository Data", 50 * 1024 * 1024}, // only if >50MB
	{"build", "Build Output", 0},
	{"dist", "Build Output", 0},
	{"target", "Build Output", 0},
	{"out", "Build Output", 0},
	{"venv", "Python Virtual Environment", 0},
	{".venv", "Python Virtual Environment", 0},
	{"env", "Python Virtual Environment", 0},
	{"__pycache__", "Python Cache", 0},
	{".gradle", "Gradle Cache", 0},
	{".m2", "Maven Cache", 0},
}

func (r *DevBloatReport) Generate(root *scanner.FileNode) []ReportItem {
	if root == nil {
		return nil
	}

	reported := make(map[string]bool)
	var items []ReportItem

	walkTree(root, func(node *scanner.FileNode) {
		if !node.IsDir {
			return
		}
		for _, target := range devBloatTargets {
			if !strings.EqualFold(node.Name, target.name) {
				continue
			}
			if target.minSize > 0 && node.Size < target.minSize {
				continue
			}
			// Skip if an ancestor is already reported (e.g., node_modules inside node_modules)
			if ancestorReported(node, reported) {
				continue
			}
			reported[node.Path] = true

			severity := "low"
			if node.Size > 500*1024*1024 { // >500MB
				severity = "high"
			} else if node.Size > 100*1024*1024 { // >100MB
				severity = "medium"
			}

			items = append(items, ReportItem{
				Path:        node.Path,
				Size:        node.Size,
				Category:    target.category,
				Description: fmt.Sprintf("%s directory", target.category),
				FileCount:   node.FileCount,
				Severity:    severity,
			})
			break
		}
	})

	return items
}
