package analyzer

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/cespare/xxhash/v2"
	"github.com/nick/mac-cleanup-explorer/internal/scanner"
)

// DuplicatesReport finds duplicate files based on size and content hashing.
type DuplicatesReport struct {
	MinSize int64 // minimum file size to consider, default 1MB
}

func (r *DuplicatesReport) Name() string        { return "duplicates" }
func (r *DuplicatesReport) Description() string { return "Duplicate Files — files with identical content" }
func (r *DuplicatesReport) AIContext() string {
	return "Identifies duplicate files by comparing sizes and content hashes (xxhash). " +
		"Only files above 1MB are checked. Uses a multi-phase approach: first groups by " +
		"file size, then compares partial hashes (first 4KB), then full file hashes. " +
		"Each duplicate group represents wasted disk space."
}

func (r *DuplicatesReport) Generate(root *scanner.FileNode) []ReportItem {
	if root == nil {
		return nil
	}

	// Phase 1: Group files by size
	sizeGroups := make(map[int64][]*scanner.FileNode)
	walkTree(root, func(node *scanner.FileNode) {
		if node.IsDir {
			return
		}
		if node.Size < r.MinSize {
			return
		}
		sizeGroups[node.Size] = append(sizeGroups[node.Size], node)
	})

	// Phase 2: For groups with 2+ files, compute partial hash (first 4KB)
	type hashGroup struct {
		hash  uint64
		nodes []*scanner.FileNode
	}

	var candidates []hashGroup
	for _, nodes := range sizeGroups {
		if len(nodes) < 2 {
			continue
		}
		partialGroups := make(map[uint64][]*scanner.FileNode)
		for _, node := range nodes {
			h, err := hashFilePartial(node.Path, 4096)
			if err != nil {
				continue
			}
			partialGroups[h] = append(partialGroups[h], node)
		}
		for h, pNodes := range partialGroups {
			if len(pNodes) >= 2 {
				candidates = append(candidates, hashGroup{hash: h, nodes: pNodes})
			}
		}
	}

	// Phase 3: Full file hash for remaining candidates
	var items []ReportItem
	for _, cand := range candidates {
		fullGroups := make(map[uint64][]*scanner.FileNode)
		for _, node := range cand.nodes {
			h, err := hashFileFull(node.Path)
			if err != nil {
				continue
			}
			fullGroups[h] = append(fullGroups[h], node)
		}
		for _, dupNodes := range fullGroups {
			if len(dupNodes) < 2 {
				continue
			}
			wastedSpace := int64(len(dupNodes)-1) * dupNodes[0].Size

			// Sort by path for deterministic output
			sort.Slice(dupNodes, func(i, j int) bool {
				return dupNodes[i].Path < dupNodes[j].Path
			})

			for _, node := range dupNodes {
				severity := "low"
				if wastedSpace > 500*1024*1024 { // >500MB wasted
					severity = "high"
				} else if wastedSpace > 50*1024*1024 { // >50MB wasted
					severity = "medium"
				}
				items = append(items, ReportItem{
					Path:        node.Path,
					Size:        node.Size,
					Category:    "Duplicate File",
					Description: fmt.Sprintf("%d duplicates, %s wasted", len(dupNodes)-1, formatBytes(wastedSpace)),
					LastMod:     formatTime(node.ModTime),
					Severity:    severity,
				})
			}
		}
	}

	return items
}

// GroupBySize is exported for testing the size-grouping phase independently.
func (r *DuplicatesReport) GroupBySize(root *scanner.FileNode) map[int64][]*scanner.FileNode {
	sizeGroups := make(map[int64][]*scanner.FileNode)
	walkTree(root, func(node *scanner.FileNode) {
		if node.IsDir {
			return
		}
		if node.Size < r.MinSize {
			return
		}
		sizeGroups[node.Size] = append(sizeGroups[node.Size], node)
	})
	return sizeGroups
}

func hashFilePartial(path string, maxBytes int64) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	h := xxhash.New()
	if _, err := io.CopyN(h, f, maxBytes); err != nil && err != io.EOF {
		return 0, err
	}
	return h.Sum64(), nil
}

func hashFileFull(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	h := xxhash.New()
	if _, err := io.Copy(h, f); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}
