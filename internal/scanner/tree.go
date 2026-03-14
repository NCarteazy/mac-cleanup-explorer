package scanner

import (
	"sync"
	"time"
)

// FileNode represents a file or directory in the scanned tree.
type FileNode struct {
	Name       string
	Path       string
	Size       int64      // For files: file size. For dirs: total recursive size.
	IsDir      bool
	ModTime    time.Time
	AccessTime time.Time
	Children   []*FileNode
	Parent     *FileNode
	FileCount  int64 // Number of files (recursive for dirs)
	DirCount   int64 // Number of subdirs (recursive for dirs)
	Error      string // Non-empty if there was an error accessing this node
}

// ScanResult holds the complete scan output.
type ScanResult struct {
	Root       *FileNode
	TotalSize  int64
	TotalFiles int64
	TotalDirs  int64
	Errors     []string
	Duration   time.Duration
}

// ScanProgress is sent during scanning to report status.
type ScanProgress struct {
	CurrentPath string
	FilesFound  int64
	DirsFound   int64
	BytesFound  int64
	mu          sync.Mutex
}

func (p *ScanProgress) Update(path string, isDir bool, size int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CurrentPath = path
	p.BytesFound += size
	if isDir {
		p.DirsFound++
	} else {
		p.FilesFound++
	}
}

func (p *ScanProgress) Snapshot() ScanProgress {
	p.mu.Lock()
	defer p.mu.Unlock()
	return ScanProgress{
		CurrentPath: p.CurrentPath,
		FilesFound:  p.FilesFound,
		DirsFound:   p.DirsFound,
		BytesFound:  p.BytesFound,
	}
}
