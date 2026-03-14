package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogAction(t *testing.T) {
	tmp := t.TempDir()
	logFile := filepath.Join(tmp, "history.log")

	err := LogAction(logFile, "rm -rf ~/Library/Caches/old", "success", 1073741824)
	if err != nil {
		t.Fatalf("LogAction failed: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "rm -rf ~/Library/Caches/old") {
		t.Error("log missing command")
	}
	if !strings.Contains(content, "success") {
		t.Error("log missing result")
	}
	if !strings.Contains(content, "1.1 GB") {
		t.Error("log missing freed size")
	}
}

func TestLogActionCreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	logFile := filepath.Join(tmp, "subdir", "history.log")

	err := LogAction(logFile, "rm test", "success", 0)
	if err != nil {
		t.Fatalf("LogAction should create parent dirs: %v", err)
	}
}

func TestLogActionAppendsMultiple(t *testing.T) {
	tmp := t.TempDir()
	logFile := filepath.Join(tmp, "history.log")

	LogAction(logFile, "cmd1", "success", 100)
	LogAction(logFile, "cmd2", "failed", 0)

	data, _ := os.ReadFile(logFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 log lines, got %d", len(lines))
	}
}

func TestDefaultLogPath(t *testing.T) {
	path := DefaultLogPath()
	if !strings.Contains(path, ".mac-cleanup-explorer") {
		t.Errorf("unexpected default path: %s", path)
	}
	if !strings.HasSuffix(path, "history.log") {
		t.Errorf("expected history.log suffix: %s", path)
	}
}
