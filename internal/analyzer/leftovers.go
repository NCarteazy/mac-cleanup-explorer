package analyzer

import (
	"fmt"
	"strings"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

// LeftoversReport finds application support data for apps that are no longer installed.
type LeftoversReport struct{}

func (r *LeftoversReport) Name() string { return "leftovers" }
func (r *LeftoversReport) Description() string {
	return "Application Leftovers — support data for uninstalled apps"
}
func (r *LeftoversReport) AIContext() string {
	return "Identifies application support data, preferences, containers, and caches " +
		"that remain after an application has been uninstalled. These are typically safe " +
		"to remove if you don't plan to reinstall the application."
}

// supportPathPatterns are the well-known paths where app leftovers live.
var supportPathPatterns = []string{
	"Library/Application Support",
	"Library/Containers",
	"Library/Preferences",
	"Library/Caches",
}

func (r *LeftoversReport) Generate(root *scanner.FileNode) []ReportItem {
	if root == nil {
		return nil
	}

	// First, collect all app names from /Applications/ directories
	appNames := collectAppNames(root)
	if len(appNames) == 0 {
		// If no applications directory found, we can't determine leftovers
		return nil
	}

	var items []ReportItem

	walkTree(root, func(node *scanner.FileNode) {
		if !node.IsDir {
			return
		}
		// Check if this node is a direct child of a support path
		if node.Parent == nil {
			return
		}

		parentPath := node.Parent.Path
		isSupportDir := false
		for _, pattern := range supportPathPatterns {
			if strings.HasSuffix(parentPath, pattern) {
				isSupportDir = true
				break
			}
		}
		if !isSupportDir {
			return
		}

		// Check if any installed app matches this support directory name
		if matchesAnyApp(node.Name, appNames) {
			return
		}

		severity := "low"
		if node.Size > 100*1024*1024 { // >100MB
			severity = "high"
		} else if node.Size > 10*1024*1024 { // >10MB
			severity = "medium"
		}

		items = append(items, ReportItem{
			Path:        node.Path,
			Size:        node.Size,
			Category:    "Application Leftover",
			Description: fmt.Sprintf("No matching app found for %q", node.Name),
			FileCount:   node.FileCount,
			Severity:    severity,
		})
	})

	return items
}

// collectAppNames walks the tree looking for nodes under paths containing "/Applications/"
// and collects normalized app names.
func collectAppNames(root *scanner.FileNode) map[string]bool {
	names := make(map[string]bool)
	walkTree(root, func(node *scanner.FileNode) {
		if !node.IsDir {
			return
		}
		// Look for .app directories under an Applications parent
		if !strings.HasSuffix(node.Name, ".app") {
			return
		}
		if !strings.Contains(node.Path, "/Applications/") {
			// Also handle the Applications node itself being a parent
			if node.Parent == nil || !strings.HasSuffix(node.Parent.Path, "/Applications") {
				return
			}
		}
		names[normalizeAppName(node.Name)] = true
	})
	return names
}

// normalizeAppName strips ".app" and common suffixes, lowercases.
func normalizeAppName(name string) string {
	name = strings.TrimSuffix(name, ".app")
	name = strings.ToLower(name)
	// Strip common suffixes for fuzzy matching
	for _, suffix := range []string{" helper", " agent", " updater", " daemon"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}

// normalizeSupportName normalizes a support directory name for matching.
func normalizeSupportName(name string) string {
	name = strings.ToLower(name)
	// Strip common prefixes/suffixes
	for _, s := range []string{" helper", " agent", " updater", " daemon"} {
		name = strings.TrimSuffix(name, s)
	}
	// Remove "com." and "org." bundle ID prefixes for common cases
	// e.g., "com.apple.Safari" -> "safari"
	parts := strings.Split(name, ".")
	if len(parts) >= 3 && (parts[0] == "com" || parts[0] == "org" || parts[0] == "io" || parts[0] == "net") {
		name = parts[len(parts)-1]
	}
	return name
}

// matchesAnyApp checks if a support directory name matches any installed app.
func matchesAnyApp(supportDirName string, appNames map[string]bool) bool {
	normalized := normalizeSupportName(supportDirName)
	if appNames[normalized] {
		return true
	}
	// Also try exact lowercase match
	if appNames[strings.ToLower(supportDirName)] {
		return true
	}
	// Check if any app name is contained in the support dir name or vice versa
	for appName := range appNames {
		if strings.Contains(normalized, appName) || strings.Contains(appName, normalized) {
			return true
		}
	}
	return false
}
