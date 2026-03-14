package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirectory(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "file1.txt"), make([]byte, 1024), 0644)
	os.WriteFile(filepath.Join(tmp, "file2.txt"), make([]byte, 2048), 0644)
	sub := filepath.Join(tmp, "subdir")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "file3.txt"), make([]byte, 4096), 0644)

	progress := &ScanProgress{}
	result, err := Scan(tmp, progress)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.TotalFiles != 3 {
		t.Errorf("expected 3 files, got %d", result.TotalFiles)
	}
	if result.TotalDirs != 1 {
		t.Errorf("expected 1 subdir, got %d", result.TotalDirs)
	}
	expectedSize := int64(1024 + 2048 + 4096)
	if result.Root.Size != expectedSize {
		t.Errorf("expected root size %d, got %d", expectedSize, result.Root.Size)
	}
	if result.Root.FileCount != 3 {
		t.Errorf("expected root FileCount 3, got %d", result.Root.FileCount)
	}
}

func TestScanHandlesPermissionError(t *testing.T) {
	tmp := t.TempDir()
	restricted := filepath.Join(tmp, "noaccess")
	os.Mkdir(restricted, 0000)
	defer os.Chmod(restricted, 0755)

	progress := &ScanProgress{}
	result, err := Scan(tmp, progress)
	if err != nil {
		t.Fatalf("Scan should not fail on permission errors: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one error for restricted directory")
	}
}

func TestScanEmptyDirectory(t *testing.T) {
	tmp := t.TempDir()
	progress := &ScanProgress{}
	result, err := Scan(tmp, progress)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if result.TotalFiles != 0 {
		t.Errorf("expected 0 files, got %d", result.TotalFiles)
	}
	if result.Root.Size != 0 {
		t.Errorf("expected 0 size, got %d", result.Root.Size)
	}
}
