package output_test

import (
	"os"
	"strings"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/output"
	"github.com/Ommanimesh2/rift/internal/tree"
)

// makeTree builds a *tree.FileTree from a map of path → *tree.FileNode.
func makeTree(entries map[string]*tree.FileNode) *tree.FileTree {
	return &tree.FileTree{Entries: entries}
}

// makeNode builds a *tree.FileNode with the given attributes.
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

// --- FormatBytes tests ---

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{name: "zero", input: 0, want: "0 bytes"},
		{name: "500 bytes", input: 500, want: "500 bytes"},
		{name: "1 KB exactly", input: 1024, want: "1.0 KB"},
		{name: "1.5 KB", input: 1536, want: "1.5 KB"},
		{name: "1 MB exactly", input: 1048576, want: "1.0 MB"},
		{name: "1 GB exactly", input: 1073741824, want: "1.0 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := output.FormatBytes(tc.input)
			if got != tc.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// --- FormatSizeDelta tests ---

func TestFormatSizeDelta(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{name: "positive 1 KB", input: 1024, want: "+1.0 KB"},
		{name: "negative 2 KB", input: -2048, want: "-2.0 KB"},
		{name: "zero", input: 0, want: "0 bytes"},
		{name: "positive 512 bytes", input: 512, want: "+512 bytes"},
		{name: "negative 256 bytes", input: -256, want: "-256 bytes"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := output.FormatSizeDelta(tc.input)
			if got != tc.want {
				t.Errorf("FormatSizeDelta(%d) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// --- FormatEntry tests ---

func TestFormatEntry_Added(t *testing.T) {
	entry := &diff.DiffEntry{
		Path:      "usr/bin/foo",
		Type:      diff.Added,
		After:     makeNode("usr/bin/foo", "abc123", 1024, 0755, 0, 0, false, ""),
		SizeDelta: 1024,
	}

	got := output.FormatEntry(entry)
	want := "+ usr/bin/foo  (+1.0 KB)"

	if got != want {
		t.Errorf("FormatEntry(added) = %q, want %q", got, want)
	}
}

func TestFormatEntry_Removed(t *testing.T) {
	entry := &diff.DiffEntry{
		Path:      "usr/lib/old.so",
		Type:      diff.Removed,
		Before:    makeNode("usr/lib/old.so", "def456", 2048, 0644, 0, 0, false, ""),
		SizeDelta: -2048,
	}

	got := output.FormatEntry(entry)
	want := "- usr/lib/old.so  (-2.0 KB)"

	if got != want {
		t.Errorf("FormatEntry(removed) = %q, want %q", got, want)
	}
}

func TestFormatEntry_Modified_ContentAndMode(t *testing.T) {
	entry := &diff.DiffEntry{
		Path:           "etc/config",
		Type:           diff.Modified,
		Before:         makeNode("etc/config", "aaa", 100, 0644, 0, 0, false, ""),
		After:          makeNode("etc/config", "bbb", 100, 0755, 0, 0, false, ""),
		SizeDelta:      0,
		ContentChanged: true,
		ModeChanged:    true,
	}

	got := output.FormatEntry(entry)
	want := "~ etc/config  [content, mode]  (0 bytes)"

	if got != want {
		t.Errorf("FormatEntry(modified content+mode) = %q, want %q", got, want)
	}
}

func TestFormatEntry_Modified_ContentOnly_PositiveDelta(t *testing.T) {
	entry := &diff.DiffEntry{
		Path:           "etc/app.conf",
		Type:           diff.Modified,
		Before:         makeNode("etc/app.conf", "aaa", 100, 0644, 0, 0, false, ""),
		After:          makeNode("etc/app.conf", "bbb", 612, 0644, 0, 0, false, ""),
		SizeDelta:      512,
		ContentChanged: true,
	}

	got := output.FormatEntry(entry)
	want := "~ etc/app.conf  [content]  (+512 bytes)"

	if got != want {
		t.Errorf("FormatEntry(modified content only) = %q, want %q", got, want)
	}
}

func TestFormatEntry_Modified_AllFlags(t *testing.T) {
	entry := &diff.DiffEntry{
		Path:              "bin/tool",
		Type:              diff.Modified,
		SizeDelta:         0,
		ContentChanged:    true,
		ModeChanged:       true,
		UIDChanged:        true,
		GIDChanged:        true,
		LinkTargetChanged: true,
		TypeChanged:       true,
	}

	got := output.FormatEntry(entry)

	// All six flags should appear in order
	wantFlags := "[content, mode, uid, gid, link, type]"
	if !strings.Contains(got, wantFlags) {
		t.Errorf("FormatEntry(all flags) = %q, want it to contain %q", got, wantFlags)
	}
	if !strings.HasPrefix(got, "~ bin/tool") {
		t.Errorf("FormatEntry(all flags) = %q, want prefix %q", got, "~ bin/tool")
	}
}

// --- FormatSummary tests ---

func TestFormatSummary_AllZero(t *testing.T) {
	result := diff.Diff(
		makeTree(map[string]*tree.FileNode{}),
		makeTree(map[string]*tree.FileNode{}),
	)

	got := output.FormatSummary(result)
	want := "0 added, 0 removed, 0 modified"

	if got != want {
		t.Errorf("FormatSummary(empty) = %q, want %q", got, want)
	}
}

func TestFormatSummary_WithBytes(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"removed.txt": makeNode("removed.txt", "aaa", 256*1024, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"added1.txt": makeNode("added1.txt", "bbb", 5*1024*1024, 0644, 0, 0, false, ""),
		"added2.txt": makeNode("added2.txt", "ccc", 4*1024*1024+900*1024, 0644, 0, 0, false, ""),
		// roughly 12.3 MB combined? Let's use precise values
	})

	result := diff.Diff(a, b)
	got := output.FormatSummary(result)

	// Should contain counts
	if !strings.Contains(got, "2 added") {
		t.Errorf("FormatSummary = %q, want it to contain %q", got, "2 added")
	}
	if !strings.Contains(got, "1 removed") {
		t.Errorf("FormatSummary = %q, want it to contain %q", got, "1 removed")
	}
	if !strings.Contains(got, "0 modified") {
		t.Errorf("FormatSummary = %q, want it to contain %q", got, "0 modified")
	}
	// Added bytes should appear as parenthetical
	if !strings.Contains(got, "(+") {
		t.Errorf("FormatSummary = %q, want it to contain added bytes parenthetical", got)
	}
	// Removed bytes should appear as parenthetical
	if !strings.Contains(got, "(-") {
		t.Errorf("FormatSummary = %q, want it to contain removed bytes parenthetical", got)
	}
}

func TestFormatSummary_NoByteParenthetical_WhenZeroBytes(t *testing.T) {
	// A modified entry contributes to neither added nor removed bytes.
	a := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "aaa", 100, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "bbb", 100, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)
	got := output.FormatSummary(result)

	// Modified count appears but no byte parentheticals
	if !strings.Contains(got, "1 modified") {
		t.Errorf("FormatSummary = %q, want it to contain %q", got, "1 modified")
	}
	if strings.Contains(got, "(") {
		t.Errorf("FormatSummary = %q, want no parentheticals when bytes are zero", got)
	}
}

func TestFormatSummary_ModifiedDoesNotContributeToByteTotals(t *testing.T) {
	// Only a modified entry with a size delta — the modified count appears
	// but the delta should NOT appear in parentheticals (only added/removed get that).
	a := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "aaa", 100, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "bbb", 200, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)
	got := output.FormatSummary(result)

	if !strings.Contains(got, "0 added") {
		t.Errorf("FormatSummary = %q, want %q", got, "0 added")
	}
	if !strings.Contains(got, "0 removed") {
		t.Errorf("FormatSummary = %q, want %q", got, "0 removed")
	}
	if !strings.Contains(got, "1 modified") {
		t.Errorf("FormatSummary = %q, want %q", got, "1 modified")
	}
	// No byte parentheticals — modified doesn't contribute
	if strings.Contains(got, "(") {
		t.Errorf("FormatSummary = %q, want no byte parentheticals for modified-only changes", got)
	}
}

// --- Render tests ---

func TestRender_Empty(t *testing.T) {
	result := diff.Diff(
		makeTree(map[string]*tree.FileNode{}),
		makeTree(map[string]*tree.FileNode{}),
	)

	got := output.Render(result)

	// Empty diff: just summary, no entries
	if !strings.Contains(got, "0 added, 0 removed, 0 modified") {
		t.Errorf("Render(empty) = %q, want summary line", got)
	}
}

func TestRender_ContainsAllEntries(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"removed.txt": makeNode("removed.txt", "aaa", 200, 0644, 0, 0, false, ""),
		"modified.txt": makeNode("modified.txt", "bbb", 300, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"added.txt":    makeNode("added.txt", "ccc", 400, 0644, 0, 0, false, ""),
		"modified.txt": makeNode("modified.txt", "ddd", 300, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)
	got := output.Render(result)

	// Each entry path appears in the output
	if !strings.Contains(got, "added.txt") {
		t.Errorf("Render: expected %q in output, got:\n%s", "added.txt", got)
	}
	if !strings.Contains(got, "removed.txt") {
		t.Errorf("Render: expected %q in output, got:\n%s", "removed.txt", got)
	}
	if !strings.Contains(got, "modified.txt") {
		t.Errorf("Render: expected %q in output, got:\n%s", "modified.txt", got)
	}

	// Summary appears at the end
	if !strings.Contains(got, "1 added") {
		t.Errorf("Render: expected summary containing %q, got:\n%s", "1 added", got)
	}
	if !strings.Contains(got, "1 removed") {
		t.Errorf("Render: expected summary containing %q, got:\n%s", "1 removed", got)
	}
	if !strings.Contains(got, "1 modified") {
		t.Errorf("Render: expected summary containing %q, got:\n%s", "1 modified", got)
	}
}

func TestRender_SummaryIsLastLine(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "abc", 100, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)
	got := output.Render(result)

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	lastLine := lines[len(lines)-1]

	if !strings.Contains(lastLine, "1 added") {
		t.Errorf("Render: last line should be summary, got: %q", lastLine)
	}
}

func TestRender_EntriesPrecedeSummary(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"aaa.txt": makeNode("aaa.txt", "aaa", 100, 0644, 0, 0, false, ""),
		"zzz.txt": makeNode("zzz.txt", "zzz", 200, 0644, 0, 0, false, ""),
	})

	result := diff.Diff(a, b)
	got := output.Render(result)

	aaaIdx := strings.Index(got, "aaa.txt")
	zzzIdx := strings.Index(got, "zzz.txt")
	summaryIdx := strings.Index(got, "2 added")

	if aaaIdx < 0 || zzzIdx < 0 || summaryIdx < 0 {
		t.Fatalf("Render: missing expected content in:\n%s", got)
	}
	if aaaIdx > summaryIdx {
		t.Errorf("Render: entry aaa.txt should appear before summary")
	}
	if zzzIdx > summaryIdx {
		t.Errorf("Render: entry zzz.txt should appear before summary")
	}
	if aaaIdx > zzzIdx {
		t.Errorf("Render: aaa.txt should appear before zzz.txt (alphabetical order)")
	}
}
