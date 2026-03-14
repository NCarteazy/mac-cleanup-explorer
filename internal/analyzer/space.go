package analyzer

import (
	"fmt"
	"strings"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

// SpaceReport generates a hierarchical directory breakdown showing
// top-level directories by size, percentage, and file count.
type SpaceReport struct{}

func (r *SpaceReport) Name() string        { return "space" }
func (r *SpaceReport) Description() string { return "Space Treemap — top-level directory breakdown" }
func (r *SpaceReport) AIContext() string {
	return "Hierarchical disk usage breakdown by top-level directory. " +
		"Shows where disk space is concentrated. " +
		"High severity means the directory uses more than 20% of total space."
}

func (r *SpaceReport) Generate(root *scanner.FileNode) []ReportItem {
	if root == nil || len(root.Children) == 0 {
		return nil
	}

	totalSize := root.Size
	if totalSize == 0 {
		totalSize = 1 // avoid division by zero
	}

	var items []ReportItem
	for _, child := range root.Children {
		pct := float64(child.Size) / float64(totalSize) * 100
		severity := "low"
		if pct > 20 {
			severity = "high"
		} else if pct > 10 {
			severity = "medium"
		}

		items = append(items, ReportItem{
			Path:        child.Path,
			Size:        child.Size,
			Category:    categorizeDir(child.Name),
			Description: fmt.Sprintf("%.1f%% of total space", pct),
			FileCount:   child.FileCount,
			Severity:    severity,
		})
	}
	return items
}

func categorizeDir(name string) string {
	lower := strings.ToLower(name)
	switch {
	case lower == "library":
		return "System & App Data"
	case lower == "applications":
		return "Applications"
	case lower == "documents":
		return "Documents"
	case lower == "downloads":
		return "Downloads"
	case lower == "pictures" || lower == "movies" || lower == "music":
		return "Media"
	case lower == "developer" || lower == "projects" || lower == "code":
		return "Developer"
	case lower == ".trash":
		return "Trash"
	default:
		return "Other"
	}
}
