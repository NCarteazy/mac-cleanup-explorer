package analyzer

import (
	"testing"
	"time"

	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

func makeTestTree() *scanner.FileNode {
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 100000,
		Children: []*scanner.FileNode{
			{Name: "Library", Path: "/Library", IsDir: true, Size: 50000, FileCount: 100},
			{Name: "Applications", Path: "/Applications", IsDir: true, Size: 30000, FileCount: 50},
			{Name: "Documents", Path: "/Documents", IsDir: true, Size: 20000, FileCount: 200},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}
	return root
}

func TestSpaceReport(t *testing.T) {
	root := makeTestTree()
	r := &SpaceReport{}
	items := r.Generate(root)

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Library is 50% of 100000 => high
	found := false
	for _, item := range items {
		if item.Path == "/Library" {
			found = true
			if item.Category != "System & App Data" {
				t.Errorf("Library category = %q, want %q", item.Category, "System & App Data")
			}
			if item.Severity != "high" {
				t.Errorf("Library severity = %q, want %q (50%% of total)", item.Severity, "high")
			}
			if item.Size != 50000 {
				t.Errorf("Library size = %d, want 50000", item.Size)
			}
		}
		if item.Path == "/Applications" {
			if item.Category != "Applications" {
				t.Errorf("Applications category = %q, want %q", item.Category, "Applications")
			}
			if item.Severity != "high" {
				t.Errorf("Applications severity = %q, want %q (30%% of total)", item.Severity, "high")
			}
		}
		if item.Path == "/Documents" {
			if item.Category != "Documents" {
				t.Errorf("Documents category = %q, want %q", item.Category, "Documents")
			}
			// 20000/100000 = 20%, which is exactly 20% (not >20%), so severity is medium
			if item.Severity != "medium" {
				t.Errorf("Documents severity = %q, want %q (exactly 20%% of total)", item.Severity, "medium")
			}
		}
	}
	if !found {
		t.Error("Library item not found in results")
	}
}

func TestSpaceReportEmptyRoot(t *testing.T) {
	r := &SpaceReport{}

	// nil root
	items := r.Generate(nil)
	if items != nil {
		t.Errorf("expected nil for nil root, got %d items", len(items))
	}

	// root with no children
	root := &scanner.FileNode{Name: "/", Path: "/", IsDir: true}
	items = r.Generate(root)
	if items != nil {
		t.Errorf("expected nil for empty root, got %d items", len(items))
	}
}

func TestSpaceReportCategories(t *testing.T) {
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 1000,
		Children: []*scanner.FileNode{
			{Name: "Downloads", Path: "/Downloads", IsDir: true, Size: 200},
			{Name: "Pictures", Path: "/Pictures", IsDir: true, Size: 150},
			{Name: "Movies", Path: "/Movies", IsDir: true, Size: 100},
			{Name: "Music", Path: "/Music", IsDir: true, Size: 100},
			{Name: "Developer", Path: "/Developer", IsDir: true, Size: 200},
			{Name: ".Trash", Path: "/.Trash", IsDir: true, Size: 50},
			{Name: "randomdir", Path: "/randomdir", IsDir: true, Size: 200},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}

	r := &SpaceReport{}
	items := r.Generate(root)

	expected := map[string]string{
		"/Downloads": "Downloads",
		"/Pictures":  "Media",
		"/Movies":    "Media",
		"/Music":     "Media",
		"/Developer": "Developer",
		"/.Trash":    "Trash",
		"/randomdir": "Other",
	}

	for _, item := range items {
		want, ok := expected[item.Path]
		if !ok {
			t.Errorf("unexpected item path: %s", item.Path)
			continue
		}
		if item.Category != want {
			t.Errorf("path %s: category = %q, want %q", item.Path, item.Category, want)
		}
	}
}

func TestLargeFilesReport(t *testing.T) {
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 3000000000,
		Children: []*scanner.FileNode{
			{
				Name: "Documents", Path: "/Documents", IsDir: true, Size: 3000000000,
				Children: []*scanner.FileNode{
					{Name: "huge.dmg", Path: "/Documents/huge.dmg", Size: 2000000000},   // 2GB
					{Name: "big.iso", Path: "/Documents/big.iso", Size: 600000000},        // 600MB
					{Name: "medium.zip", Path: "/Documents/medium.zip", Size: 200000000},  // 200MB
					{Name: "small.txt", Path: "/Documents/small.txt", Size: 1000},          // 1KB
				},
			},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
		for _, gc := range c.Children {
			gc.Parent = c
		}
	}

	r := &LargeFilesReport{Threshold: 100 * 1024 * 1024} // 100MB
	items := r.Generate(root)

	if len(items) != 3 {
		t.Fatalf("expected 3 large files, got %d", len(items))
	}

	// Should be sorted by size descending
	if items[0].Path != "/Documents/huge.dmg" {
		t.Errorf("first item = %s, want /Documents/huge.dmg", items[0].Path)
	}
	if items[0].Severity != "high" {
		t.Errorf("2GB file severity = %q, want high", items[0].Severity)
	}
	if items[1].Severity != "medium" {
		t.Errorf("600MB file severity = %q, want medium", items[1].Severity)
	}
	if items[2].Severity != "low" {
		t.Errorf("200MB file severity = %q, want low", items[2].Severity)
	}
}

func TestLargeFilesReportNoneLarge(t *testing.T) {
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 5000,
		Children: []*scanner.FileNode{
			{Name: "a.txt", Path: "/a.txt", Size: 1000},
			{Name: "b.txt", Path: "/b.txt", Size: 2000},
			{Name: "c.txt", Path: "/c.txt", Size: 2000},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}

	r := &LargeFilesReport{Threshold: 100 * 1024 * 1024}
	items := r.Generate(root)

	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestStaleReport(t *testing.T) {
	now := time.Now()
	oldTime := now.AddDate(0, -12, 0) // 12 months ago
	recentTime := now.AddDate(0, -1, 0) // 1 month ago

	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 100000,
		Children: []*scanner.FileNode{
			{
				Name: "Documents", Path: "/Documents", IsDir: true, Size: 100000,
				Children: []*scanner.FileNode{
					{Name: "old.txt", Path: "/Documents/old.txt", Size: 5000000, AccessTime: oldTime},
					{Name: "recent.txt", Path: "/Documents/recent.txt", Size: 3000, AccessTime: recentTime},
					{Name: "no_atime.txt", Path: "/Documents/no_atime.txt", Size: 1000}, // zero AccessTime
				},
			},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
		for _, gc := range c.Children {
			gc.Parent = c
		}
	}

	r := &StaleReport{MaxAge: 6}
	items := r.Generate(root)

	if len(items) != 1 {
		t.Fatalf("expected 1 stale file, got %d", len(items))
	}
	if items[0].Path != "/Documents/old.txt" {
		t.Errorf("stale item = %s, want /Documents/old.txt", items[0].Path)
	}
	if items[0].Category != "Documents" {
		t.Errorf("stale item category = %q, want Documents", items[0].Category)
	}
}

func TestCachesReport(t *testing.T) {
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 500000,
		Children: []*scanner.FileNode{
			{
				Name: "Library", Path: "/Library", IsDir: true, Size: 400000,
				Children: []*scanner.FileNode{
					{
						Name: "Caches", Path: "/Library/Caches", IsDir: true, Size: 200000, FileCount: 50,
						Children: []*scanner.FileNode{
							{Name: "com.apple.Safari", Path: "/Library/Caches/com.apple.Safari", IsDir: true, Size: 100000, FileCount: 20},
						},
					},
					{Name: "Logs", Path: "/Library/Logs", IsDir: true, Size: 100000, FileCount: 30},
				},
			},
			{
				Name: ".cache", Path: "/.cache", IsDir: true, Size: 50000, FileCount: 10,
			},
			{
				Name: "normaldir", Path: "/normaldir", IsDir: true, Size: 50000,
			},
		},
	}
	// Set up parent pointers
	for _, c := range root.Children {
		c.Parent = root
		for _, gc := range c.Children {
			gc.Parent = c
			for _, ggc := range gc.Children {
				ggc.Parent = gc
			}
		}
	}

	r := &CachesReport{}
	items := r.Generate(root)

	if len(items) < 2 {
		t.Fatalf("expected at least 2 cache items, got %d", len(items))
	}

	paths := make(map[string]bool)
	for _, item := range items {
		paths[item.Path] = true
	}

	// /Library/Caches should match (contains /Library/Caches/)
	// But note: the pattern "/Library/Caches/" needs the path to contain it.
	// /Library/Caches itself: path + "/" = "/Library/Caches/" which contains "/Library/Caches/"
	if !paths["/Library/Caches"] {
		t.Error("expected /Library/Caches in results")
	}
	// /Library/Logs should match
	if !paths["/Library/Logs"] {
		t.Error("expected /Library/Logs in results")
	}
	// normaldir should not match
	if paths["/normaldir"] {
		t.Error("/normaldir should not be in cache results")
	}
	// com.apple.Safari should NOT be reported separately since its parent was already reported
	if paths["/Library/Caches/com.apple.Safari"] {
		t.Error("/Library/Caches/com.apple.Safari should not be separately reported (parent already matched)")
	}
}

func TestDevBloatReport(t *testing.T) {
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 1000000,
		Children: []*scanner.FileNode{
			{
				Name: "projects", Path: "/projects", IsDir: true, Size: 800000,
				Children: []*scanner.FileNode{
					{Name: "node_modules", Path: "/projects/node_modules", IsDir: true, Size: 300000, FileCount: 5000},
					{Name: ".git", Path: "/projects/.git", IsDir: true, Size: 60 * 1024 * 1024, FileCount: 100}, // 60MB, above 50MB threshold
					{Name: "build", Path: "/projects/build", IsDir: true, Size: 50000, FileCount: 20},
					{Name: "venv", Path: "/projects/venv", IsDir: true, Size: 100000, FileCount: 300},
					{Name: "__pycache__", Path: "/projects/__pycache__", IsDir: true, Size: 5000, FileCount: 10},
					{Name: "src", Path: "/projects/src", IsDir: true, Size: 10000, FileCount: 50}, // not bloat
				},
			},
			{
				Name: "small-repo", Path: "/small-repo", IsDir: true, Size: 200000,
				Children: []*scanner.FileNode{
					{Name: ".git", Path: "/small-repo/.git", IsDir: true, Size: 10 * 1024 * 1024, FileCount: 50}, // 10MB, below 50MB threshold
				},
			},
		},
	}
	// Set up parents
	for _, c := range root.Children {
		c.Parent = root
		for _, gc := range c.Children {
			gc.Parent = c
		}
	}

	r := &DevBloatReport{}
	items := r.Generate(root)

	paths := make(map[string]bool)
	for _, item := range items {
		paths[item.Path] = true
	}

	if !paths["/projects/node_modules"] {
		t.Error("expected node_modules in results")
	}
	if !paths["/projects/.git"] {
		t.Error("expected large .git in results")
	}
	if !paths["/projects/build"] {
		t.Error("expected build in results")
	}
	if !paths["/projects/venv"] {
		t.Error("expected venv in results")
	}
	if !paths["/projects/__pycache__"] {
		t.Error("expected __pycache__ in results")
	}
	if paths["/projects/src"] {
		t.Error("src should not be in devbloat results")
	}
	// Small .git should NOT be reported (below 50MB threshold)
	if paths["/small-repo/.git"] {
		t.Error("small .git should not be reported (below 50MB threshold)")
	}
}

func TestLeftoversReport(t *testing.T) {
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 500000,
		Children: []*scanner.FileNode{
			{
				Name: "Applications", Path: "/Applications", IsDir: true, Size: 100000,
				Children: []*scanner.FileNode{
					{Name: "Safari.app", Path: "/Applications/Safari.app", IsDir: true, Size: 50000},
					{Name: "Slack.app", Path: "/Applications/Slack.app", IsDir: true, Size: 50000},
				},
			},
			{
				Name: "Library", Path: "/Library", IsDir: true, Size: 400000,
				Children: []*scanner.FileNode{
					{
						Name: "Application Support", Path: "/Library/Application Support", IsDir: true, Size: 300000,
						Children: []*scanner.FileNode{
							{Name: "Safari", Path: "/Library/Application Support/Safari", IsDir: true, Size: 100000, FileCount: 20},
							{Name: "Slack", Path: "/Library/Application Support/Slack", IsDir: true, Size: 100000, FileCount: 15},
							{Name: "OldApp", Path: "/Library/Application Support/OldApp", IsDir: true, Size: 100000, FileCount: 30},
						},
					},
				},
			},
		},
	}
	// Set up parents
	for _, c := range root.Children {
		c.Parent = root
		for _, gc := range c.Children {
			gc.Parent = c
			for _, ggc := range gc.Children {
				ggc.Parent = gc
			}
		}
	}

	r := &LeftoversReport{}
	items := r.Generate(root)

	if len(items) != 1 {
		t.Fatalf("expected 1 leftover, got %d", len(items))
	}
	if items[0].Path != "/Library/Application Support/OldApp" {
		t.Errorf("leftover = %s, want /Library/Application Support/OldApp", items[0].Path)
	}
	if items[0].Category != "Application Leftover" {
		t.Errorf("category = %q, want Application Leftover", items[0].Category)
	}
}

func TestDuplicatesReport(t *testing.T) {
	// Test the size-grouping phase with in-memory tree.
	// Full hashing requires real files on disk and is tested via integration.
	root := &scanner.FileNode{
		Name: "/", Path: "/", IsDir: true, Size: 10000000,
		Children: []*scanner.FileNode{
			{Name: "a.bin", Path: "/a.bin", Size: 2000000},  // 2MB
			{Name: "b.bin", Path: "/b.bin", Size: 2000000},  // 2MB (same size as a)
			{Name: "c.bin", Path: "/c.bin", Size: 3000000},  // 3MB (unique size)
			{Name: "d.bin", Path: "/d.bin", Size: 3000000},  // 3MB (same size as c)
			{Name: "tiny.txt", Path: "/tiny.txt", Size: 500}, // below MinSize
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}

	r := &DuplicatesReport{MinSize: 1024 * 1024} // 1MB
	groups := r.GroupBySize(root)

	// Should have 2 groups: 2MB and 3MB
	if len(groups) != 2 {
		t.Fatalf("expected 2 size groups, got %d", len(groups))
	}

	for size, nodes := range groups {
		if len(nodes) != 2 {
			t.Errorf("size group %d: expected 2 files, got %d", size, len(nodes))
		}
	}

	// tiny.txt should not be in any group
	for _, nodes := range groups {
		for _, n := range nodes {
			if n.Name == "tiny.txt" {
				t.Error("tiny.txt should not be in size groups (below MinSize)")
			}
		}
	}
}

func TestAllReports(t *testing.T) {
	reports := AllReports()

	if len(reports) != 7 {
		t.Fatalf("expected 7 reports, got %d", len(reports))
	}

	names := make(map[string]bool)
	for _, r := range reports {
		name := r.Name()
		if name == "" {
			t.Error("report has empty name")
		}
		if names[name] {
			t.Errorf("duplicate report name: %s", name)
		}
		names[name] = true

		if r.Description() == "" {
			t.Errorf("report %s has empty description", name)
		}
		if r.AIContext() == "" {
			t.Errorf("report %s has empty AIContext", name)
		}
	}

	expectedNames := []string{"space", "large_files", "stale", "caches", "devbloat", "leftovers", "duplicates"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("expected report %q not registered", name)
		}
	}
}

func TestGenerateAll(t *testing.T) {
	root := makeTestTree()
	result := &scanner.ScanResult{
		Root:       root,
		TotalSize:  100000,
		TotalFiles: 350,
		TotalDirs:  3,
	}

	out := GenerateAll(result)

	if len(out) != 7 {
		t.Fatalf("expected 7 report outputs, got %d", len(out))
	}

	expectedNames := []string{"space", "large_files", "stale", "caches", "devbloat", "leftovers", "duplicates"}
	for _, name := range expectedNames {
		if _, ok := out[name]; !ok {
			t.Errorf("missing report output for %q", name)
		}
	}

	// Space report should have items for the test tree
	if len(out["space"]) != 3 {
		t.Errorf("space report: expected 3 items, got %d", len(out["space"]))
	}
}
