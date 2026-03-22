package content

import (
	"strings"
	"testing"
)

func TestUnifiedDiff_Identical(t *testing.T) {
	result := UnifiedDiff([]byte("hello\n"), []byte("hello\n"), "a", "b")
	if result != "" {
		t.Errorf("expected empty diff for identical content, got %q", result)
	}
}

func TestUnifiedDiff_SimpleChange(t *testing.T) {
	old := []byte("line1\nline2\nline3\n")
	new := []byte("line1\nmodified\nline3\n")

	result := UnifiedDiff(old, new, "a/file.txt", "b/file.txt")

	if !strings.Contains(result, "--- a/file.txt") {
		t.Error("expected --- header")
	}
	if !strings.Contains(result, "+++ b/file.txt") {
		t.Error("expected +++ header")
	}
	if !strings.Contains(result, "-line2") {
		t.Error("expected deleted line")
	}
	if !strings.Contains(result, "+modified") {
		t.Error("expected added line")
	}
}

func TestUnifiedDiff_Addition(t *testing.T) {
	old := []byte("line1\n")
	new := []byte("line1\nline2\n")

	result := UnifiedDiff(old, new, "a", "b")
	if !strings.Contains(result, "+line2") {
		t.Error("expected added line2")
	}
}

func TestUnifiedDiff_Deletion(t *testing.T) {
	old := []byte("line1\nline2\n")
	new := []byte("line1\n")

	result := UnifiedDiff(old, new, "a", "b")
	if !strings.Contains(result, "-line2") {
		t.Error("expected deleted line2")
	}
}

func TestUnifiedDiff_EmptyOld(t *testing.T) {
	result := UnifiedDiff([]byte{}, []byte("new content\n"), "a", "b")
	if !strings.Contains(result, "+new content") {
		t.Error("expected added content")
	}
}

func TestUnifiedDiff_EmptyNew(t *testing.T) {
	result := UnifiedDiff([]byte("old content\n"), []byte{}, "a", "b")
	if !strings.Contains(result, "-old content") {
		t.Error("expected deleted content")
	}
}
