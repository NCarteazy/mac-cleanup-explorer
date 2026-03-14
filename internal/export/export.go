package export

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
)

// SystemInfo holds machine and scan metadata for report context.
type SystemInfo struct {
	OSVersion string
	DiskSize  string
	FreeSpace string
	Machine   string
	ScanScope string
	ScanTime  string
}

// FormatMarkdown produces an AI-friendly markdown report for a single report.
func FormatMarkdown(reportName, aiContext string, items []analyzer.ReportItem, sysInfo SystemInfo) string {
	var b strings.Builder

	b.WriteString("# Mac Cleanup Explorer Report\n\n")

	writeSystemInfo(&b, sysInfo)

	b.WriteString(fmt.Sprintf("## Report: %s\n\n", reportName))

	if aiContext != "" {
		b.WriteString("### Context\n")
		b.WriteString(aiContext + "\n\n")
	}

	writeSummary(&b, items)
	writeDataTable(&b, items)
	writeSuggestedPrompts(&b, reportName)

	tokenEst := EstimateTokens(b.String())
	b.WriteString(fmt.Sprintf("Token estimate: ~%d tokens\n", tokenEst))

	return b.String()
}

// FormatMultipleReports combines multiple reports with a single system info header.
func FormatMultipleReports(reports map[string][]analyzer.ReportItem, sysInfo SystemInfo) string {
	var b strings.Builder

	b.WriteString("# Mac Cleanup Explorer Report\n\n")

	writeSystemInfo(&b, sysInfo)

	// Sort report names for deterministic output.
	names := make([]string, 0, len(reports))
	for name := range reports {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		items := reports[name]
		b.WriteString(fmt.Sprintf("## Report: %s\n\n", name))
		writeSummary(&b, items)
		writeDataTable(&b, items)
		writeSuggestedPrompts(&b, name)
	}

	tokenEst := EstimateTokens(b.String())
	b.WriteString(fmt.Sprintf("Token estimate: ~%d tokens\n", tokenEst))

	return b.String()
}

// jsonReport is the structure used for JSON export.
type jsonReport struct {
	ReportName string              `json:"report_name"`
	SystemInfo jsonSystemInfo      `json:"system_info"`
	Items      []jsonReportItem    `json:"items"`
	Summary    jsonSummary         `json:"summary"`
	Prompts    []string            `json:"suggested_prompts"`
}

type jsonSystemInfo struct {
	OSVersion string `json:"os_version,omitempty"`
	DiskSize  string `json:"disk_size,omitempty"`
	FreeSpace string `json:"free_space,omitempty"`
	Machine   string `json:"machine,omitempty"`
	ScanScope string `json:"scan_scope,omitempty"`
	ScanTime  string `json:"scan_time,omitempty"`
}

type jsonReportItem struct {
	Path       string `json:"path"`
	Size       int64  `json:"size_bytes"`
	SizeHuman  string `json:"size_human"`
	Category   string `json:"category,omitempty"`
	Severity   string `json:"severity,omitempty"`
}

type jsonSummary struct {
	TotalItems int    `json:"total_items"`
	TotalSize  string `json:"total_size"`
}

// FormatJSON produces a JSON representation of a report.
func FormatJSON(reportName string, items []analyzer.ReportItem, sysInfo SystemInfo) string {
	jItems := make([]jsonReportItem, len(items))
	var totalSize int64
	for i, item := range items {
		totalSize += item.Size
		jItems[i] = jsonReportItem{
			Path:      item.Path,
			Size:      item.Size,
			SizeHuman: humanize.IBytes(uint64(item.Size)),
			Category:  item.Category,
			Severity:  item.Severity,
		}
	}

	report := jsonReport{
		ReportName: reportName,
		SystemInfo: jsonSystemInfo{
			OSVersion: sysInfo.OSVersion,
			DiskSize:  sysInfo.DiskSize,
			FreeSpace: sysInfo.FreeSpace,
			Machine:   sysInfo.Machine,
			ScanScope: sysInfo.ScanScope,
			ScanTime:  sysInfo.ScanTime,
		},
		Items: jItems,
		Summary: jsonSummary{
			TotalItems: len(items),
			TotalSize:  humanize.IBytes(uint64(totalSize)),
		},
		Prompts: SuggestedPrompts(reportName),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(data)
}

// EstimateTokens gives a rough token count. Average English token is ~4 characters.
func EstimateTokens(text string) int {
	return len(text) / 4
}

// TruncateItems returns up to maxItems items and a summary of omitted items.
// If no truncation is needed, summary is empty.
func TruncateItems(items []analyzer.ReportItem, maxItems int) ([]analyzer.ReportItem, string) {
	if len(items) <= maxItems {
		return items, ""
	}
	truncated := items[:maxItems]
	remaining := items[maxItems:]
	var totalSize int64
	for _, item := range remaining {
		totalSize += item.Size
	}
	summary := fmt.Sprintf("...and %d more items totaling %s",
		len(remaining), humanize.IBytes(uint64(totalSize)))
	return truncated, summary
}

// SuggestedPrompts returns AI prompts tailored to the given report type.
func SuggestedPrompts(reportName string) []string {
	switch reportName {
	case "Large Files":
		return []string{
			"Review these large files and suggest which are safe to delete based on file type and location.",
			"Identify any large files that appear to be temporary or redundant.",
		}
	case "Caches":
		return []string{
			"Review these cache directories and tell me which are safe to clear without affecting running applications.",
			"Estimate how much space I can safely reclaim from these caches.",
		}
	case "Dev Bloat":
		return []string{
			"Analyze these developer directories and suggest cleanup commands for each tool.",
			"Which of these developer artifacts can be safely removed and regenerated when needed?",
		}
	case "Stale Files":
		return []string{
			"Review these unused files and suggest which ones are safe to archive or delete.",
			"Identify any stale files that may still be important despite not being accessed recently.",
		}
	case "Leftovers":
		return []string{
			"Identify which of these leftover files belong to applications that are no longer installed.",
			"Suggest a safe cleanup plan for these application leftovers.",
		}
	case "Duplicates":
		return []string{
			"Review these duplicate files and suggest which copies to keep and which to remove.",
			"Are any of these duplicates intentional (e.g., backups) that should be preserved?",
		}
	case "Disk Space":
		return []string{
			"Analyze this disk space breakdown and suggest the best areas to focus cleanup efforts.",
			"What directories are consuming the most space relative to their expected size?",
		}
	default:
		return []string{
			"Analyze this report and suggest which items are safe to clean up.",
		}
	}
}

// writeSystemInfo writes the system info section.
func writeSystemInfo(b *strings.Builder, sysInfo SystemInfo) {
	b.WriteString("## System Info\n")
	if sysInfo.OSVersion != "" {
		b.WriteString(fmt.Sprintf("- OS: %s\n", sysInfo.OSVersion))
	}
	if sysInfo.DiskSize != "" || sysInfo.FreeSpace != "" {
		b.WriteString(fmt.Sprintf("- Disk: %s total, %s free\n", sysInfo.DiskSize, sysInfo.FreeSpace))
	}
	if sysInfo.Machine != "" {
		b.WriteString(fmt.Sprintf("- Machine: %s\n", sysInfo.Machine))
	}
	if sysInfo.ScanScope != "" {
		scanLine := fmt.Sprintf("- Scanned: %s", sysInfo.ScanScope)
		if sysInfo.ScanTime != "" {
			scanLine += fmt.Sprintf(" (%s)", sysInfo.ScanTime)
		}
		b.WriteString(scanLine + "\n")
	}
	b.WriteString("\n")
}

// writeSummary writes the summary section for a list of items.
func writeSummary(b *strings.Builder, items []analyzer.ReportItem) {
	var totalSize int64
	for _, item := range items {
		totalSize += item.Size
	}
	b.WriteString("### Summary\n")
	b.WriteString(fmt.Sprintf("- Total items: %d\n", len(items)))
	b.WriteString(fmt.Sprintf("- Total size: %s\n\n", humanize.IBytes(uint64(totalSize))))
}

// writeDataTable writes the markdown table of report items.
func writeDataTable(b *strings.Builder, items []analyzer.ReportItem) {
	b.WriteString("### Data\n")
	b.WriteString("| Path | Size | Category | Severity |\n")
	b.WriteString("|------|------|----------|----------|\n")
	for _, item := range items {
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			item.Path,
			humanize.IBytes(uint64(item.Size)),
			item.Category,
			item.Severity,
		))
	}
	b.WriteString("\n")
}

// writeSuggestedPrompts writes the suggested prompts section.
func writeSuggestedPrompts(b *strings.Builder, reportName string) {
	prompts := SuggestedPrompts(reportName)
	b.WriteString("### Suggested Prompts\n")
	for _, p := range prompts {
		b.WriteString(fmt.Sprintf("- %q\n", p))
	}
	b.WriteString("\n")
}
