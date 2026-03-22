// Package diff compares two FileTree instances and produces a structured
// DiffResult describing every file-level change: additions, removals, and
// modifications including content hash, permission, ownership, and type deltas.
package diff

import (
	"fmt"
	"sort"

	"github.com/Ommanimesh2/rift/internal/tree"
)

// ChangeType classifies a single diff entry.
type ChangeType int

const (
	// Added means the path exists in B but not in A.
	Added ChangeType = iota
	// Removed means the path exists in A but not in B.
	Removed
	// Modified means the path exists in both trees but at least one attribute differs.
	Modified
)

// String returns a human-readable label for the ChangeType.
func (c ChangeType) String() string {
	switch c {
	case Added:
		return "added"
	case Removed:
		return "removed"
	case Modified:
		return "modified"
	default:
		return "unknown"
	}
}

// DiffEntry describes a single file-system path that changed between two trees.
type DiffEntry struct {
	// Path is the normalized path of the filesystem entry.
	Path string

	// Type is the kind of change (Added, Removed, Modified).
	Type ChangeType

	// Before is the FileNode from tree A. Nil for Added entries.
	Before *tree.FileNode

	// After is the FileNode from tree B. Nil for Removed entries.
	After *tree.FileNode

	// SizeDelta is the change in bytes:
	//   Added    → After.Size
	//   Removed  → -Before.Size
	//   Modified → After.Size - Before.Size
	SizeDelta int64

	// Change flags — set only for Modified entries.

	// ContentChanged is true when the sha256 digest hex strings differ.
	ContentChanged bool

	// ModeChanged is true when the permission bits differ.
	ModeChanged bool

	// UIDChanged is true when the numeric owner UID differs.
	UIDChanged bool

	// GIDChanged is true when the numeric owner GID differs.
	GIDChanged bool

	// LinkTargetChanged is true when the symlink target path differs.
	LinkTargetChanged bool

	// TypeChanged is true when IsDir flipped between Before and After.
	TypeChanged bool
}

// DiffResult holds the complete set of changes between two FileTrees.
type DiffResult struct {
	// Entries is the ordered list of changed paths (sorted alphabetically).
	Entries []*DiffEntry

	// Summary counts.
	Added    int
	Removed  int
	Modified int

	// Byte totals.
	AddedBytes   int64
	RemovedBytes int64
}

// String returns a human-readable one-line summary of the diff.
// Example: "5 added, 2 removed, 3 modified (+1.2 MB / -200 bytes)"
func (r *DiffResult) String() string {
	var addedTotal, removedTotal int64
	for _, e := range r.Entries {
		if e.SizeDelta > 0 {
			addedTotal += e.SizeDelta
		} else if e.SizeDelta < 0 {
			removedTotal += -e.SizeDelta
		}
	}

	if addedTotal == 0 && removedTotal == 0 {
		return fmt.Sprintf("%d added, %d removed, %d modified",
			r.Added, r.Removed, r.Modified)
	}
	return fmt.Sprintf("%d added, %d removed, %d modified (+%s / -%s)",
		r.Added, r.Removed, r.Modified,
		formatBytes(addedTotal), formatBytes(removedTotal))
}

// Diff compares two FileTrees and returns a *DiffResult with all changes.
//
// Algorithm:
//  1. Scan b.Entries: path not in a → Added; path in both → compare attributes.
//  2. Scan a.Entries: path not in b → Removed.
//  3. Sort result entries alphabetically by Path.
//  4. Compute summary counters and byte totals.
func Diff(a, b *tree.FileTree) *DiffResult {
	result := &DiffResult{}

	// Pass 1: iterate over b — find Added and Modified entries.
	for path, bNode := range b.Entries {
		aNode, inA := a.Entries[path]
		if !inA {
			// File exists only in b → Added.
			result.Entries = append(result.Entries, &DiffEntry{
				Path:      path,
				Type:      Added,
				Before:    nil,
				After:     bNode,
				SizeDelta: bNode.Size,
			})
			continue
		}
		// File exists in both — compare attributes.
		if entry := compareNodes(path, aNode, bNode); entry != nil {
			result.Entries = append(result.Entries, entry)
		}
	}

	// Pass 2: iterate over a — find Removed entries.
	for path, aNode := range a.Entries {
		if _, inB := b.Entries[path]; !inB {
			result.Entries = append(result.Entries, &DiffEntry{
				Path:      path,
				Type:      Removed,
				Before:    aNode,
				After:     nil,
				SizeDelta: -aNode.Size,
			})
		}
	}

	// Sort entries alphabetically by path for deterministic output.
	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].Path < result.Entries[j].Path
	})

	// Compute summary counters and byte totals.
	for _, e := range result.Entries {
		switch e.Type {
		case Added:
			result.Added++
			result.AddedBytes += e.After.Size
		case Removed:
			result.Removed++
			result.RemovedBytes += e.Before.Size
		case Modified:
			result.Modified++
		}
	}

	return result
}

// compareNodes compares two FileNodes at the same path and returns a Modified
// DiffEntry if any attribute differs, or nil if they are identical.
func compareNodes(path string, a, b *tree.FileNode) *DiffEntry {
	entry := &DiffEntry{
		Path:   path,
		Type:   Modified,
		Before: a,
		After:  b,
	}

	changed := false

	if a.Digest.Hex != b.Digest.Hex {
		entry.ContentChanged = true
		changed = true
	}
	if a.Mode != b.Mode {
		entry.ModeChanged = true
		changed = true
	}
	if a.UID != b.UID {
		entry.UIDChanged = true
		changed = true
	}
	if a.GID != b.GID {
		entry.GIDChanged = true
		changed = true
	}
	if a.IsDir != b.IsDir {
		entry.TypeChanged = true
		changed = true
	}
	if a.LinkTarget != b.LinkTarget {
		entry.LinkTargetChanged = true
		changed = true
	}

	if !changed {
		return nil
	}

	entry.SizeDelta = b.Size - a.Size
	return entry
}

// formatBytes converts a byte count to a human-readable string using SI
// suffixes (KB, MB, GB). Values below 1 KB are rendered as "N bytes".
func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d bytes", b)
	}
}
