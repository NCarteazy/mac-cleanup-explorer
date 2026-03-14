package analyzer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

// LargeFilesReport finds files exceeding a size threshold.
type LargeFilesReport struct {
	Threshold int64 // in bytes, default 100MB
}

func (r *LargeFilesReport) Name() string        { return "large_files" }
func (r *LargeFilesReport) Description() string { return "Large Files — files exceeding size threshold" }
func (r *LargeFilesReport) AIContext() string {
	return "Lists individual files larger than 100MB. " +
		"These are candidates for archiving or removal. " +
		"High severity for files over 1GB, medium for over 500MB."
}

func (r *LargeFilesReport) Generate(root *scanner.FileNode) []ReportItem {
	if root == nil {
		return nil
	}

	var items []ReportItem
	walkTree(root, func(node *scanner.FileNode) {
		if node.IsDir {
			return
		}
		if node.Size > r.Threshold {
			ext := strings.ToLower(filepath.Ext(node.Name))
			if ext == "" {
				ext = "(no extension)"
			}
			severity := "low"
			if node.Size > 1024*1024*1024 { // >1GB
				severity = "high"
			} else if node.Size > 500*1024*1024 { // >500MB
				severity = "medium"
			}
			items = append(items, ReportItem{
				Path:        node.Path,
				Size:        node.Size,
				Category:    "Large File",
				Description: fmt.Sprintf("Large file (%s)", ext),
				LastMod:     formatTime(node.ModTime),
				LastAccess:  formatTime(node.AccessTime),
				Severity:    severity,
			})
		}
	})

	sort.Slice(items, func(i, j int) bool {
		return items[i].Size > items[j].Size
	})

	return items
}
