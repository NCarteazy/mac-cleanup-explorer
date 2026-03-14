package analyzer

import (
	"fmt"
	"time"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

// StaleReport finds files that haven't been accessed in a long time.
type StaleReport struct {
	MaxAge int // months since last access
}

func (r *StaleReport) Name() string        { return "stale" }
func (r *StaleReport) Description() string { return "Stale Files — not accessed in 6+ months" }
func (r *StaleReport) AIContext() string {
	return "Files not accessed in 6 or more months. " +
		"These may be forgotten or no longer needed. " +
		"Severity is based on both file size and how long since last access."
}

func (r *StaleReport) Generate(root *scanner.FileNode) []ReportItem {
	if root == nil {
		return nil
	}

	cutoff := time.Now().AddDate(0, -r.MaxAge, 0)
	var items []ReportItem

	walkTree(root, func(node *scanner.FileNode) {
		if node.IsDir {
			return
		}
		// Skip items with zero AccessTime (metadata not available)
		if node.AccessTime.IsZero() {
			return
		}
		if node.AccessTime.Before(cutoff) {
			monthsStale := int(time.Since(node.AccessTime).Hours() / (24 * 30))
			severity := staleSeverity(node.Size, monthsStale)
			items = append(items, ReportItem{
				Path:        node.Path,
				Size:        node.Size,
				Category:    categorizeByPath(node.Path),
				Description: fmt.Sprintf("Not accessed in %d months", monthsStale),
				LastAccess:  formatTime(node.AccessTime),
				LastMod:     formatTime(node.ModTime),
				Severity:    severity,
			})
		}
	})

	return items
}

func staleSeverity(size int64, monthsStale int) string {
	// Combine size and staleness for severity
	score := float64(size)/(1024*1024) * float64(monthsStale)
	if score > 500 { // e.g., 50MB * 10 months
		return "high"
	}
	if score > 50 { // e.g., 5MB * 10 months
		return "medium"
	}
	return "low"
}
