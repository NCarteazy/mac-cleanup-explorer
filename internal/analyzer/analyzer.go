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
		&LargeFilesReport{Threshold: 100 * 1024 * 1024}, // 100MB
		&StaleReport{MaxAge: 6},                           // 6 months
		&CachesReport{},
		&DevBloatReport{},
		&LeftoversReport{},
		&DuplicatesReport{MinSize: 1024 * 1024}, // 1MB
	}
}
