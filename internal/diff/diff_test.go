package diff_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/tree"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// makeTree builds a *tree.FileTree from a map of path → *tree.FileNode.
func makeTree(entries map[string]*tree.FileNode) *tree.FileTree {
	return &tree.FileTree{Entries: entries}
}

// makeNode builds a *tree.FileNode with the given attributes.
// digestHex is the sha256 hex string; pass "" for dirs and symlinks.
func makeNode(path, digestHex string, size int64, mode os.FileMode, uid, gid int, isDir bool, linkTarget string) *tree.FileNode {
	n := &tree.FileNode{
		Path:       path,
		Size:       size,
		Mode:       mode,
		UID:        uid,
		GID:        gid,
		IsDir:      isDir,
		LinkTarget: linkTarget,
	}
	if digestHex != "" {
		n.Digest = v1.Hash{Algorithm: "sha256", Hex: digestHex}
	}
	return n
}

// --- Test cases ---

func TestDiff_IdenticalSingleFileTrees(t *testing.T) {
	node := makeNode("etc/passwd", "abc123", 100, 0644, 0, 0, false, "")
	a := makeTree(map[string]*tree.FileNode{"etc/passwd": node})
	b := makeTree(map[string]*tree.FileNode{"etc/passwd": node})

	result := diff.Diff(a, b)

	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries for identical trees, got %d", len(result.Entries))
	}
	if result.Added != 0 || result.Removed != 0 || result.Modified != 0 {
		t.Errorf("expected all counts 0, got added=%d removed=%d modified=%d",
			result.Added, result.Removed, result.Modified)
	}
}

func TestDiff_EmptyATreeBHasTwoFiles(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/cat": makeNode("usr/bin/cat", "deadbeef", 1024, 0755, 0, 0, false, ""),
		"etc/hosts":   makeNode("etc/hosts", "cafebabe", 256, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)

	if result.Added != 2 {
		t.Errorf("expected 2 added, got %d", result.Added)
	}
	if result.Removed != 0 {
		t.Errorf("expected 0 removed, got %d", result.Removed)
	}
	if result.Modified != 0 {
		t.Errorf("expected 0 modified, got %d", result.Modified)
	}
	wantBytes := int64(1024 + 256)
	if result.AddedBytes != wantBytes {
		t.Errorf("expected AddedBytes=%d, got %d", wantBytes, result.AddedBytes)
	}
	for _, e := range result.Entries {
		if e.Type != diff.Added {
			t.Errorf("entry %q: expected type Added, got %v", e.Path, e.Type)
		}
		if e.Before != nil {
			t.Errorf("entry %q: expected Before=nil for Added entry", e.Path)
		}
		if e.After == nil {
			t.Errorf("entry %q: expected After to be set for Added entry", e.Path)
		}
	}
}

func TestDiff_AHasTwoFilesBEmpty(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"usr/bin/cat": makeNode("usr/bin/cat", "deadbeef", 1024, 0755, 0, 0, false, ""),
		"etc/hosts":   makeNode("etc/hosts", "cafebabe", 256, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{})

	result := diff.Diff(a, b)

	if result.Removed != 2 {
		t.Errorf("expected 2 removed, got %d", result.Removed)
	}
	if result.Added != 0 {
		t.Errorf("expected 0 added, got %d", result.Added)
	}
	if result.Modified != 0 {
		t.Errorf("expected 0 modified, got %d", result.Modified)
	}
	wantBytes := int64(1024 + 256)
	if result.RemovedBytes != wantBytes {
		t.Errorf("expected RemovedBytes=%d, got %d", wantBytes, result.RemovedBytes)
	}
	for _, e := range result.Entries {
		if e.Type != diff.Removed {
			t.Errorf("entry %q: expected type Removed, got %v", e.Path, e.Type)
		}
		if e.After != nil {
			t.Errorf("entry %q: expected After=nil for Removed entry", e.Path)
		}
		if e.Before == nil {
			t.Errorf("entry %q: expected Before to be set for Removed entry", e.Path)
		}
	}
}

func TestDiff_ContentChanged(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"app/main.go": makeNode("app/main.go", "aaaaaa", 500, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"app/main.go": makeNode("app/main.go", "bbbbbb", 500, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)

	if result.Modified != 1 {
		t.Fatalf("expected 1 modified, got %d", result.Modified)
	}
	e := result.Entries[0]
	if e.Type != diff.Modified {
		t.Errorf("expected type Modified, got %v", e.Type)
	}
	if !e.ContentChanged {
		t.Error("expected ContentChanged=true when digest hex differs")
	}
	if e.ModeChanged || e.UIDChanged || e.GIDChanged || e.TypeChanged || e.LinkTargetChanged {
		t.Error("no other change flags should be set")
	}
}

func TestDiff_ModeChanged(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"bin/tool": makeNode("bin/tool", "aaaaaa", 200, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"bin/tool": makeNode("bin/tool", "aaaaaa", 200, 0755, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)

	if result.Modified != 1 {
		t.Fatalf("expected 1 modified, got %d", result.Modified)
	}
	e := result.Entries[0]
	if !e.ModeChanged {
		t.Error("expected ModeChanged=true when permission bits differ")
	}
	if e.ContentChanged || e.UIDChanged || e.GIDChanged || e.TypeChanged || e.LinkTargetChanged {
		t.Error("no other change flags should be set")
	}
}

func TestDiff_UIDChanged(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"etc/file": makeNode("etc/file", "aaaaaa", 100, 0644, 1000, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"etc/file": makeNode("etc/file", "aaaaaa", 100, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)

	if result.Modified != 1 {
		t.Fatalf("expected 1 modified, got %d", result.Modified)
	}
	e := result.Entries[0]
	if !e.UIDChanged {
		t.Error("expected UIDChanged=true when UID differs")
	}
	if e.ContentChanged || e.ModeChanged || e.GIDChanged || e.TypeChanged || e.LinkTargetChanged {
		t.Error("no other change flags should be set")
	}
}

func TestDiff_GIDChanged(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"etc/file": makeNode("etc/file", "aaaaaa", 100, 0644, 0, 1000, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"etc/file": makeNode("etc/file", "aaaaaa", 100, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)

	if result.Modified != 1 {
		t.Fatalf("expected 1 modified, got %d", result.Modified)
	}
	e := result.Entries[0]
	if !e.GIDChanged {
		t.Error("expected GIDChanged=true when GID differs")
	}
	if e.ContentChanged || e.ModeChanged || e.UIDChanged || e.TypeChanged || e.LinkTargetChanged {
		t.Error("no other change flags should be set")
	}
}

func TestDiff_FileBecomesDir(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"app": makeNode("app", "aaaaaa", 100, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"app": makeNode("app", "", 0, 0755, 0, 0, true, ""),
	})

	result := diff.Diff(a, b)

	if result.Modified != 1 {
		t.Fatalf("expected 1 modified, got %d", result.Modified)
	}
	e := result.Entries[0]
	if !e.TypeChanged {
		t.Error("expected TypeChanged=true when IsDir flips")
	}
}

func TestDiff_SymlinkTargetChanged(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"usr/bin/sh": makeNode("usr/bin/sh", "", 0, 0777, 0, 0, false, "/bin/bash"),
	})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/sh": makeNode("usr/bin/sh", "", 0, 0777, 0, 0, false, "/bin/dash"),
	})

	result := diff.Diff(a, b)

	if result.Modified != 1 {
		t.Fatalf("expected 1 modified, got %d", result.Modified)
	}
	e := result.Entries[0]
	if !e.LinkTargetChanged {
		t.Error("expected LinkTargetChanged=true when symlink target differs")
	}
	if e.ContentChanged || e.ModeChanged || e.UIDChanged || e.GIDChanged || e.TypeChanged {
		t.Error("no other change flags should be set")
	}
}

func TestDiff_MixedChanges(t *testing.T) {
	// a: has 3 files. b: removes one, modifies one, adds one new.
	a := makeTree(map[string]*tree.FileNode{
		"kept.txt":    makeNode("kept.txt", "aaaaaa", 100, 0644, 0, 0, false, ""),
		"removed.txt": makeNode("removed.txt", "bbbbbb", 200, 0644, 0, 0, false, ""),
		"modified.txt": makeNode("modified.txt", "cccccc", 300, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"kept.txt":    makeNode("kept.txt", "aaaaaa", 100, 0644, 0, 0, false, ""),
		"added.txt":   makeNode("added.txt", "dddddd", 400, 0644, 0, 0, false, ""),
		"modified.txt": makeNode("modified.txt", "eeeeee", 350, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)

	if result.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Added)
	}
	if result.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Removed)
	}
	if result.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", result.Modified)
	}
	if result.AddedBytes != 400 {
		t.Errorf("expected AddedBytes=400, got %d", result.AddedBytes)
	}
	if result.RemovedBytes != 200 {
		t.Errorf("expected RemovedBytes=200, got %d", result.RemovedBytes)
	}
	if len(result.Entries) != 3 {
		t.Errorf("expected 3 total entries, got %d", len(result.Entries))
	}
}

func TestDiff_BothTreesEmpty(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{})

	result := diff.Diff(a, b)

	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries for both-empty diff, got %d", len(result.Entries))
	}
	if result.Added != 0 || result.Removed != 0 || result.Modified != 0 {
		t.Error("expected all counts 0 for both-empty trees")
	}
}

func TestDiff_EntriesSortedAlphabetically(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"zzz.txt": makeNode("zzz.txt", "111111", 10, 0644, 0, 0, false, ""),
		"aaa.txt": makeNode("aaa.txt", "222222", 20, 0644, 0, 0, false, ""),
		"mmm.txt": makeNode("mmm.txt", "333333", 30, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)

	if len(result.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Entries))
	}
	paths := []string{result.Entries[0].Path, result.Entries[1].Path, result.Entries[2].Path}
	wantPaths := []string{"aaa.txt", "mmm.txt", "zzz.txt"}
	for i, p := range paths {
		if p != wantPaths[i] {
			t.Errorf("entry[%d]: expected path %q, got %q", i, wantPaths[i], p)
		}
	}
}

func TestDiff_SizeDelta(t *testing.T) {
	tests := []struct {
		name          string
		beforeSize    int64
		afterSize     int64
		wantSizeDelta int64
	}{
		{name: "grew", beforeSize: 100, afterSize: 300, wantSizeDelta: 200},
		{name: "shrank", beforeSize: 300, afterSize: 100, wantSizeDelta: -200},
		{name: "same size but content changed", beforeSize: 200, afterSize: 200, wantSizeDelta: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := makeTree(map[string]*tree.FileNode{
				"file.bin": makeNode("file.bin", "aaaaaa", tc.beforeSize, 0644, 0, 0, false, ""),
			})
			b := makeTree(map[string]*tree.FileNode{
				"file.bin": makeNode("file.bin", "bbbbbb", tc.afterSize, 0644, 0, 0, false, ""),
			})

			result := diff.Diff(a, b)

			if len(result.Entries) == 0 {
				t.Fatalf("expected at least 1 entry")
			}
			e := result.Entries[0]
			if e.SizeDelta != tc.wantSizeDelta {
				t.Errorf("SizeDelta: want %d, got %d", tc.wantSizeDelta, e.SizeDelta)
			}
		})
	}
}

func TestDiff_SizeDeltaForAdded(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"new.txt": makeNode("new.txt", "aaaaaa", 512, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)

	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	e := result.Entries[0]
	if e.SizeDelta != 512 {
		t.Errorf("SizeDelta for Added: want 512, got %d", e.SizeDelta)
	}
}

func TestDiff_SizeDeltaForRemoved(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"old.txt": makeNode("old.txt", "aaaaaa", 512, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{})

	result := diff.Diff(a, b)

	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	e := result.Entries[0]
	if e.SizeDelta != -512 {
		t.Errorf("SizeDelta for Removed: want -512, got %d", e.SizeDelta)
	}
}

func TestDiffResult_String(t *testing.T) {
	tests := []struct {
		name        string
		a, b        *tree.FileTree
		wantContain []string
	}{
		{
			name: "added files summary",
			a:    makeTree(map[string]*tree.FileNode{}),
			b: makeTree(map[string]*tree.FileNode{
				"f1": makeNode("f1", "aaa", 1024*1024, 0644, 0, 0, false, ""),
				"f2": makeNode("f2", "bbb", 200*1024, 0644, 0, 0, false, ""),
			}),
			wantContain: []string{"2 added", "0 removed", "0 modified"},
		},
		{
			name: "mixed diff summary",
			a: makeTree(map[string]*tree.FileNode{
				"rem.txt": makeNode("rem.txt", "aaa", 100, 0644, 0, 0, false, ""),
				"mod.txt": makeNode("mod.txt", "bbb", 200, 0644, 0, 0, false, ""),
			}),
			b: makeTree(map[string]*tree.FileNode{
				"add.txt": makeNode("add.txt", "ccc", 300, 0644, 0, 0, false, ""),
				"mod.txt": makeNode("mod.txt", "ddd", 250, 0644, 0, 0, false, ""),
			}),
			wantContain: []string{"1 added", "1 removed", "1 modified"},
		},
		{
			name: "empty diff summary",
			a:    makeTree(map[string]*tree.FileNode{}),
			b:    makeTree(map[string]*tree.FileNode{}),
			wantContain: []string{"0 added", "0 removed", "0 modified"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := diff.Diff(tc.a, tc.b)
			s := result.String()
			for _, want := range tc.wantContain {
				if !strings.Contains(s, want) {
					t.Errorf("String() = %q, want it to contain %q", s, want)
				}
			}
		})
	}
}
