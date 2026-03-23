package output

import (
	"strings"
	"testing"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/tree"
)

func TestGroupByLayer_Empty(t *testing.T) {
	groups := GroupByLayer(nil)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestGroupByLayer_SingleLayer(t *testing.T) {
	entries := []*diff.DiffEntry{
		{Path: "a", Type: diff.Added, After: &tree.FileNode{LayerIndex: 0, Size: 100}, SizeDelta: 100},
		{Path: "b", Type: diff.Added, After: &tree.FileNode{LayerIndex: 0, Size: 200}, SizeDelta: 200},
	}
	groups := GroupByLayer(entries)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Index != 0 {
		t.Errorf("expected index 0, got %d", groups[0].Index)
	}
	if len(groups[0].Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(groups[0].Entries))
	}
	if groups[0].TotalDelta != 300 {
		t.Errorf("expected delta 300, got %d", groups[0].TotalDelta)
	}
}

func TestGroupByLayer_MultipleLayers(t *testing.T) {
	entries := []*diff.DiffEntry{
		{Path: "a", Type: diff.Added, After: &tree.FileNode{LayerIndex: 0}, SizeDelta: 100},
		{Path: "b", Type: diff.Added, After: &tree.FileNode{LayerIndex: 2}, SizeDelta: 200},
		{Path: "c", Type: diff.Modified, After: &tree.FileNode{LayerIndex: 2}, Before: &tree.FileNode{LayerIndex: 0}, SizeDelta: 50},
		{Path: "d", Type: diff.Removed, Before: &tree.FileNode{LayerIndex: 1}, SizeDelta: -80},
	}
	groups := GroupByLayer(entries)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	// Groups should be sorted by index: 0, 1, 2
	if groups[0].Index != 0 || groups[1].Index != 1 || groups[2].Index != 2 {
		t.Errorf("unexpected group ordering: %d, %d, %d", groups[0].Index, groups[1].Index, groups[2].Index)
	}
}

func TestFormatLayerAttribution(t *testing.T) {
	groups := []LayerGroup{
		{
			Index:   0,
			Command: "/bin/sh -c apt-get install -y curl",
			Entries: []*diff.DiffEntry{
				{Path: "usr/bin/curl", Type: diff.Added, After: &tree.FileNode{Size: 2100000}, SizeDelta: 2100000},
			},
			TotalDelta: 2100000,
		},
		{
			Index:   1,
			Command: "COPY . /app",
			Entries: []*diff.DiffEntry{
				{Path: "app/main.go", Type: diff.Modified, SizeDelta: 12, Before: &tree.FileNode{}, After: &tree.FileNode{}},
			},
			TotalDelta: 12,
		},
	}

	out := FormatLayerAttribution(groups, "myapp:v1", "myapp:v2")

	if !strings.Contains(out, "myapp:v1 → myapp:v2") {
		t.Error("expected image names in header")
	}
	if !strings.Contains(out, "Layer 0 (/bin/sh -c apt-get install -y curl)") {
		t.Error("expected layer 0 header with command")
	}
	if !strings.Contains(out, "+ usr/bin/curl") {
		t.Error("expected added file in layer 0")
	}
	if !strings.Contains(out, "Layer 1 (COPY . /app)") {
		t.Error("expected layer 1 header")
	}
	if !strings.Contains(out, "~ app/main.go") {
		t.Error("expected modified file in layer 1")
	}
}

func TestFormatLayerAttribution_TruncatesLongCommand(t *testing.T) {
	longCmd := strings.Repeat("x", 100)
	groups := []LayerGroup{
		{Index: 0, Command: longCmd, TotalDelta: 0},
	}
	out := FormatLayerAttribution(groups, "a", "b")
	if strings.Contains(out, longCmd) {
		t.Error("expected long command to be truncated")
	}
	if !strings.Contains(out, "...") {
		t.Error("expected truncation ellipsis")
	}
}

func TestFormatLayerAttribution_UnknownCommand(t *testing.T) {
	groups := []LayerGroup{
		{Index: 0, Command: "", TotalDelta: 0},
	}
	out := FormatLayerAttribution(groups, "a", "b")
	if !strings.Contains(out, "(unknown)") {
		t.Error("expected (unknown) for empty command")
	}
}
