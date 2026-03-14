package analyzer

import (
	"fmt"
	"strings"
	"time"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

// walkTree recursively visits every node in the tree.
func walkTree(node *scanner.FileNode, fn func(*scanner.FileNode)) {
	if node == nil {
		return
	}
	fn(node)
	for _, child := range node.Children {
		walkTree(child, fn)
	}
}

// formatTime returns a human-readable time string, or empty if zero.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

// formatBytes returns a human-readable byte count.
func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// categorizeByPath returns a category based on the file path.
func categorizeByPath(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "/library/"):
		return "System & App Data"
	case strings.Contains(lower, "/applications/"):
		return "Applications"
	case strings.Contains(lower, "/documents/"):
		return "Documents"
	case strings.Contains(lower, "/downloads/"):
		return "Downloads"
	case strings.Contains(lower, "/pictures/") || strings.Contains(lower, "/movies/") || strings.Contains(lower, "/music/"):
		return "Media"
	case strings.Contains(lower, "/developer/") || strings.Contains(lower, "/projects/") || strings.Contains(lower, "/code/"):
		return "Developer"
	case strings.Contains(lower, "/.trash/"):
		return "Trash"
	default:
		return "Other"
	}
}
