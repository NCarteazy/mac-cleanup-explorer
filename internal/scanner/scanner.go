package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

// Scan walks the filesystem starting from root and returns a ScanResult.
func Scan(root string, progress *ScanProgress) (*ScanResult, error) {
	start := time.Now()

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving root path: %w", err)
	}

	rootInfo, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat root: %w", err)
	}

	rootNode := &FileNode{
		Name:  rootInfo.Name(),
		Path:  absRoot,
		IsDir: true,
	}

	result := &ScanResult{Root: rootNode}
	nodeMap := map[string]*FileNode{absRoot: rootNode}

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", path, err))
			return nil
		}

		if path == absRoot {
			return nil
		}

		// Skip known problematic paths (Docker VM images, system firmlinks, etc.)
		if ShouldSkip(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", path, err))
			return nil
		}

		node := &FileNode{
			Name:    d.Name(),
			Path:    path,
			IsDir:   d.IsDir(),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		}

		// Get access time from syscall
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			node.AccessTime = time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
		}

		if !d.IsDir() {
			result.TotalFiles++
			result.TotalSize += node.Size
		} else {
			result.TotalDirs++
		}

		parentPath := filepath.Dir(path)
		if parent, ok := nodeMap[parentPath]; ok {
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		}

		if d.IsDir() {
			nodeMap[path] = node
		}

		if progress != nil {
			progress.Update(path, d.IsDir(), node.Size)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking filesystem: %w", err)
	}

	calculateDirSizes(rootNode)
	sortChildren(rootNode)

	result.Duration = time.Since(start)
	return result, nil
}

func calculateDirSizes(node *FileNode) {
	if !node.IsDir {
		node.FileCount = 1
		return
	}

	var totalSize int64
	var fileCount, dirCount int64

	for _, child := range node.Children {
		calculateDirSizes(child)
		totalSize += child.Size
		fileCount += child.FileCount
		if child.IsDir {
			dirCount += 1 + child.DirCount
		}
	}

	node.Size = totalSize
	node.FileCount = fileCount
	node.DirCount = dirCount
}

func sortChildren(node *FileNode) {
	if !node.IsDir {
		return
	}
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})
	for _, child := range node.Children {
		sortChildren(child)
	}
}

// SkipPaths returns paths that should be skipped during scanning.
func SkipPaths() []string {
	home, _ := os.UserHomeDir()
	paths := []string{
		"/System/Volumes/Data",
	}
	if home != "" {
		// Docker Desktop VM disk image — sparse file that reports 1TB+ apparent size
		paths = append(paths, filepath.Join(home, "Library/Containers/com.docker.docker/Data/vms"))
	}
	return paths
}

// ShouldSkip checks if a path should be skipped.
func ShouldSkip(path string) bool {
	for _, skip := range SkipPaths() {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}
