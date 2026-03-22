package diff

import (
	"testing"

	"github.com/Ommanimesh2/rift/internal/tree"
)

func TestFilterEntries_NoFilters(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			&DiffEntry{Path: "a.txt", Type: Added, After: &tree.FileNode{Size: 100}},
			&DiffEntry{Path: "b.txt", Type: Removed, Before: &tree.FileNode{Size: 200}},
		},
		Added: 1, Removed: 1, AddedBytes: 100, RemovedBytes: 200,
	}

	filtered := FilterEntries(result, nil, nil)
	if filtered != result {
		t.Error("expected same result when no filters")
	}
}

func TestFilterEntries_IncludeOnly(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			&DiffEntry{Path: "etc/nginx/nginx.conf", Type: Modified, Before: &tree.FileNode{Size: 100}, After: &tree.FileNode{Size: 120}},
			&DiffEntry{Path: "usr/bin/curl", Type: Added, After: &tree.FileNode{Size: 500}},
			&DiffEntry{Path: "etc/apt/sources.list", Type: Modified, Before: &tree.FileNode{Size: 50}, After: &tree.FileNode{Size: 55}},
		},
	}

	filtered := FilterEntries(result, []string{"etc/**"}, nil)
	if len(filtered.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(filtered.Entries))
	}
	if filtered.Entries[0].Path != "etc/nginx/nginx.conf" {
		t.Errorf("expected etc/nginx/nginx.conf, got %s", filtered.Entries[0].Path)
	}
	if filtered.Entries[1].Path != "etc/apt/sources.list" {
		t.Errorf("expected etc/apt/sources.list, got %s", filtered.Entries[1].Path)
	}
	if filtered.Modified != 2 {
		t.Errorf("expected Modified=2, got %d", filtered.Modified)
	}
	if filtered.Added != 0 {
		t.Errorf("expected Added=0, got %d", filtered.Added)
	}
}

func TestFilterEntries_ExcludeOnly(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			&DiffEntry{Path: "usr/bin/curl", Type: Added, After: &tree.FileNode{Size: 500}},
			&DiffEntry{Path: "var/cache/apt/pkgcache.bin", Type: Added, After: &tree.FileNode{Size: 1000}},
			&DiffEntry{Path: "var/cache/apt/srcpkgcache.bin", Type: Added, After: &tree.FileNode{Size: 800}},
			&DiffEntry{Path: "etc/hostname", Type: Modified, Before: &tree.FileNode{Size: 10}, After: &tree.FileNode{Size: 12}},
		},
	}

	filtered := FilterEntries(result, nil, []string{"var/cache/**"})
	if len(filtered.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(filtered.Entries))
	}
	if filtered.Added != 1 {
		t.Errorf("expected Added=1, got %d", filtered.Added)
	}
	if filtered.AddedBytes != 500 {
		t.Errorf("expected AddedBytes=500, got %d", filtered.AddedBytes)
	}
}

func TestFilterEntries_IncludeAndExclude(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			&DiffEntry{Path: "usr/bin/curl", Type: Added, After: &tree.FileNode{Size: 500}},
			&DiffEntry{Path: "usr/bin/wget", Type: Added, After: &tree.FileNode{Size: 300}},
			&DiffEntry{Path: "usr/share/doc/curl/README", Type: Added, After: &tree.FileNode{Size: 50}},
			&DiffEntry{Path: "etc/hosts", Type: Modified, Before: &tree.FileNode{Size: 10}, After: &tree.FileNode{Size: 12}},
		},
	}

	filtered := FilterEntries(result, []string{"usr/**"}, []string{"usr/share/doc/**"})
	if len(filtered.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(filtered.Entries))
	}
	if filtered.Entries[0].Path != "usr/bin/curl" {
		t.Errorf("expected usr/bin/curl, got %s", filtered.Entries[0].Path)
	}
	if filtered.Entries[1].Path != "usr/bin/wget" {
		t.Errorf("expected usr/bin/wget, got %s", filtered.Entries[1].Path)
	}
}

func TestFilterEntries_ExtensionPattern(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			&DiffEntry{Path: "app/main.py", Type: Modified, Before: &tree.FileNode{Size: 100}, After: &tree.FileNode{Size: 110}},
			&DiffEntry{Path: "app/__pycache__/main.cpython-311.pyc", Type: Added, After: &tree.FileNode{Size: 2000}},
			&DiffEntry{Path: "app/utils.py", Type: Added, After: &tree.FileNode{Size: 50}},
		},
	}

	filtered := FilterEntries(result, nil, []string{"**/*.pyc"})
	if len(filtered.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(filtered.Entries))
	}
	if filtered.Added != 1 {
		t.Errorf("expected Added=1, got %d", filtered.Added)
	}
}

func TestFilterEntries_RecomputesSummary(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			&DiffEntry{Path: "a.txt", Type: Added, After: &tree.FileNode{Size: 100}},
			&DiffEntry{Path: "b.txt", Type: Removed, Before: &tree.FileNode{Size: 200}},
			&DiffEntry{Path: "c.txt", Type: Modified, Before: &tree.FileNode{Size: 50}, After: &tree.FileNode{Size: 60}},
			&DiffEntry{Path: "d.log", Type: Added, After: &tree.FileNode{Size: 999}},
		},
		Added: 2, Removed: 1, Modified: 1,
		AddedBytes: 1099, RemovedBytes: 200,
	}

	filtered := FilterEntries(result, nil, []string{"*.log"})
	if filtered.Added != 1 {
		t.Errorf("expected Added=1, got %d", filtered.Added)
	}
	if filtered.AddedBytes != 100 {
		t.Errorf("expected AddedBytes=100, got %d", filtered.AddedBytes)
	}
	if filtered.Removed != 1 {
		t.Errorf("expected Removed=1, got %d", filtered.Removed)
	}
	if filtered.RemovedBytes != 200 {
		t.Errorf("expected RemovedBytes=200, got %d", filtered.RemovedBytes)
	}
	if filtered.Modified != 1 {
		t.Errorf("expected Modified=1, got %d", filtered.Modified)
	}
}

func TestFilterEntries_EmptyResult(t *testing.T) {
	result := &DiffResult{
		Entries: []*DiffEntry{
			&DiffEntry{Path: "a.txt", Type: Added, After: &tree.FileNode{Size: 100}},
		},
		Added: 1, AddedBytes: 100,
	}

	filtered := FilterEntries(result, []string{"*.go"}, nil)
	if len(filtered.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(filtered.Entries))
	}
	if filtered.Added != 0 {
		t.Errorf("expected Added=0, got %d", filtered.Added)
	}
}

func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		include []string
		exclude []string
		want    bool
	}{
		{"no filters", "a.txt", nil, nil, true},
		{"include match", "etc/nginx.conf", []string{"etc/**"}, nil, true},
		{"include no match", "usr/bin/ls", []string{"etc/**"}, nil, false},
		{"exclude match", "var/cache/foo", nil, []string{"var/cache/**"}, false},
		{"exclude no match", "usr/bin/ls", nil, []string{"var/cache/**"}, true},
		{"include+exclude pass", "etc/nginx.conf", []string{"etc/**"}, []string{"etc/default/**"}, true},
		{"include+exclude blocked", "etc/default/grub", []string{"etc/**"}, []string{"etc/default/**"}, false},
		{"doublestar extension", "deep/path/to/file.pyc", nil, []string{"**/*.pyc"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesFilter(tt.path, tt.include, tt.exclude)
			if got != tt.want {
				t.Errorf("matchesFilter(%q, %v, %v) = %v, want %v", tt.path, tt.include, tt.exclude, got, tt.want)
			}
		})
	}
}
