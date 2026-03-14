package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
)

// DefaultLogPath returns the default history log file path.
func DefaultLogPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mac-cleanup-explorer", "history.log")
}

// LogAction appends an action to the history log.
// Format: [2026-03-14 15:04:05] COMMAND: rm ... | RESULT: success | FREED: 1.5 GB
func LogAction(logPath, command, result string, freedBytes int64) error {
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating log directory: %w", err)
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer f.Close()

	freed := "0 B"
	if freedBytes > 0 {
		freed = humanize.Bytes(uint64(freedBytes))
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] COMMAND: %s | RESULT: %s | FREED: %s\n", timestamp, command, result, freed)

	_, err = f.WriteString(line)
	return err
}
