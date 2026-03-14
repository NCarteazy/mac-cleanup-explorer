package export

import (
	"strings"
	"testing"

	"github.com/nick/mac-cleanup-explorer/internal/analyzer"
)

func TestFormatReport(t *testing.T) {
	items := []analyzer.ReportItem{
		{Path: "/Users/test/big.dmg", Size: 5368709120, Category: "Downloads", Severity: "high"},
		{Path: "/Users/test/old.zip", Size: 1073741824, Category: "Downloads", Severity: "medium"},
	}
	sysInfo := SystemInfo{
		OSVersion: "macOS 15.0",
		DiskSize:  "500 GB",
		FreeSpace: "50 GB",
		Machine:   "MacBook Pro (Apple M1)",
		ScanScope: "/Users/test",
		ScanTime:  "2.3s",
	}

	md := FormatMarkdown("Large Files", "Files over 100MB sorted by size", items, sysInfo)

	if !strings.Contains(md, "# Mac Cleanup Explorer Report") {
		t.Error("missing report header")
	}
	if !strings.Contains(md, "Large Files") {
		t.Error("missing report name")
	}
	if !strings.Contains(md, "big.dmg") {
		t.Error("missing file entry")
	}
	if !strings.Contains(md, "macOS 15.0") {
		t.Error("missing system info")
	}
}

func TestFormatMultipleReports(t *testing.T) {
	reports := map[string][]analyzer.ReportItem{
		"Large Files": {
			{Path: "/big.iso", Size: 4000000000, Category: "Downloads", Severity: "high"},
		},
		"Caches": {
			{Path: "/Library/Caches/old", Size: 2000000000, Category: "Cache", Severity: "medium"},
		},
	}
	sysInfo := SystemInfo{OSVersion: "macOS 15.0", DiskSize: "500 GB", FreeSpace: "50 GB"}

	md := FormatMultipleReports(reports, sysInfo)
	if !strings.Contains(md, "Large Files") {
		t.Error("missing Large Files section")
	}
	if !strings.Contains(md, "Caches") {
		t.Error("missing Caches section")
	}
}

func TestTokenEstimate(t *testing.T) {
	text := strings.Repeat("word ", 100)
	estimate := EstimateTokens(text)
	if estimate < 100 || estimate > 200 {
		t.Errorf("token estimate %d seems off for 100 words", estimate)
	}
}

func TestTruncateItems(t *testing.T) {
	items := make([]analyzer.ReportItem, 100)
	for i := range items {
		items[i] = analyzer.ReportItem{Path: "/test", Size: int64(100 - i)}
	}
	truncated, summary := TruncateItems(items, 20)
	if len(truncated) != 20 {
		t.Errorf("expected 20 items, got %d", len(truncated))
	}
	if summary == "" {
		t.Error("expected truncation summary")
	}
}

func TestTruncateItemsNoTruncation(t *testing.T) {
	items := make([]analyzer.ReportItem, 5)
	for i := range items {
		items[i] = analyzer.ReportItem{Path: "/test", Size: int64(i)}
	}
	truncated, summary := TruncateItems(items, 20)
	if len(truncated) != 5 {
		t.Errorf("expected 5 items, got %d", len(truncated))
	}
	if summary != "" {
		t.Error("expected no truncation summary")
	}
}

func TestFormatJSON(t *testing.T) {
	items := []analyzer.ReportItem{
		{Path: "/test.txt", Size: 1024, Category: "Test"},
	}
	sysInfo := SystemInfo{OSVersion: "macOS 15.0"}

	jsonStr := FormatJSON("Test Report", items, sysInfo)
	if !strings.Contains(jsonStr, "test.txt") {
		t.Error("missing file in JSON output")
	}
	if !strings.Contains(jsonStr, "macOS 15.0") {
		t.Error("missing system info in JSON")
	}
}

func TestSuggestedPrompts(t *testing.T) {
	prompts := SuggestedPrompts("Large Files")
	if len(prompts) == 0 {
		t.Error("expected at least one suggested prompt for Large Files")
	}
	prompts = SuggestedPrompts("Unknown Report")
	if len(prompts) == 0 {
		t.Error("expected a default prompt even for unknown reports")
	}
}
