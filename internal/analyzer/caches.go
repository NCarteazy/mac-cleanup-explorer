package analyzer

import (
	"fmt"
	"strings"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

// CachesReport detects cache and temporary data directories.
type CachesReport struct{}

func (r *CachesReport) Name() string        { return "caches" }
func (r *CachesReport) Description() string { return "Cache & Temp Data — clearable caches and temp files" }
func (r *CachesReport) AIContext() string {
	return "Identifies cache directories, temp files, and other data that can typically " +
		"be safely cleared. Includes system caches, package manager caches, build caches, " +
		"and log files. These are usually regenerated on demand."
}

// cachePatterns maps path substrings to human-friendly category labels.
var cachePatterns = []struct {
	pattern  string
	category string
}{
	{"/Library/Caches/", "System Cache"},
	{"/Library/Logs/", "Log Files"},
	{"/.cache/", "User Cache"},
	{"/DerivedData/", "Xcode Derived Data"},
	{"/.npm/", "npm Cache"},
	{"/.yarn/", "Yarn Cache"},
	{"/.pnpm-store/", "pnpm Cache"},
	{"/pip/cache/", "pip Cache"},
	{"/.cargo/registry/", "Cargo Cache"},
	{"/.gradle/caches/", "Gradle Cache"},
	{"/var/folders/", "System Temp"},
	{"/.Trash/", "Trash"},
}

func (r *CachesReport) Generate(root *scanner.FileNode) []ReportItem {
	if root == nil {
		return nil
	}

	// Track which directories we've already reported to avoid duplicates
	reported := make(map[string]bool)
	var items []ReportItem

	walkTree(root, func(node *scanner.FileNode) {
		if !node.IsDir {
			return
		}
		for _, cp := range cachePatterns {
			// Check if this directory's path contains the cache pattern
			if !strings.Contains(node.Path+"/", cp.pattern) {
				continue
			}
			// Only report the highest-level match to avoid duplication
			if reported[node.Path] {
				continue
			}
			// Skip if any ancestor was already reported for the same pattern
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
				Category:    cp.category,
				Description: fmt.Sprintf("Cache/temp directory (%s)", cp.category),
				FileCount:   node.FileCount,
				Severity:    severity,
			})
			break // only match the first pattern per directory
		}
	})

	return items
}

func ancestorReported(node *scanner.FileNode, reported map[string]bool) bool {
	parent := node.Parent
	for parent != nil {
		if reported[parent.Path] {
			return true
		}
		parent = parent.Parent
	}
	return false
}
