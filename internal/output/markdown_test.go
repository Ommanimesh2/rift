package output_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ommmishra/imgdiff/internal/diff"
	"github.com/ommmishra/imgdiff/internal/output"
	"github.com/ommmishra/imgdiff/internal/security"
	"github.com/ommmishra/imgdiff/internal/tree"
)

// --- FormatMarkdown tests ---

// TestFormatMarkdown_EmptyResult verifies that an empty DiffResult with no
// security events produces the header, summary with zeros, "(No changes)"
// in the changes section, and no security findings section.
func TestFormatMarkdown_EmptyResult(t *testing.T) {
	result := diff.Diff(
		makeTree(map[string]*tree.FileNode{}),
		makeTree(map[string]*tree.FileNode{}),
	)

	got := output.FormatMarkdown(result, "img1:latest", "img2:latest", nil)

	// Header
	if !strings.Contains(got, "## Image Diff: `img1:latest` → `img2:latest`") {
		t.Errorf("missing header, got:\n%s", got)
	}

	// Summary section header
	if !strings.Contains(got, "### Summary") {
		t.Errorf("missing summary section header, got:\n%s", got)
	}

	// Summary table rows with zeros
	if !strings.Contains(got, "| Added") {
		t.Errorf("missing Added row, got:\n%s", got)
	}
	if !strings.Contains(got, "| Removed") {
		t.Errorf("missing Removed row, got:\n%s", got)
	}
	if !strings.Contains(got, "| Modified") {
		t.Errorf("missing Modified row, got:\n%s", got)
	}

	// Changes section header
	if !strings.Contains(got, "### Changes") {
		t.Errorf("missing changes section header, got:\n%s", got)
	}

	// (No changes) instead of the table
	if !strings.Contains(got, "(No changes)") {
		t.Errorf("expected '(No changes)', got:\n%s", got)
	}

	// No security findings section
	if strings.Contains(got, "### Security Findings") {
		t.Errorf("unexpected security findings section for empty events, got:\n%s", got)
	}

	// Ends with single trailing newline
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("output should end with newline, got:\n%q", got)
	}
}

// TestFormatMarkdown_OneAdded verifies that a single added entry produces the
// correct summary row and changes table row.
func TestFormatMarkdown_OneAdded(t *testing.T) {
	// 123456 bytes = 120.6 KB
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/wget": makeNode("usr/bin/wget", "abc123", 123456, 0755, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	got := output.FormatMarkdown(result, "base:1.0", "new:2.0", nil)

	// Summary: Added row should show count 1 and positive size
	if !strings.Contains(got, "| Added") {
		t.Errorf("missing Added row, got:\n%s", got)
	}
	// The summary Added row includes "+120.6 KB"
	if !strings.Contains(got, "+120.6 KB") {
		t.Errorf("expected '+120.6 KB' in summary Added row, got:\n%s", got)
	}

	// Changes table: status `+`, path backtick-wrapped, delta
	if !strings.Contains(got, "| `+`") {
		t.Errorf("expected '| `+`' status cell, got:\n%s", got)
	}
	if !strings.Contains(got, "`usr/bin/wget`") {
		t.Errorf("expected backtick-wrapped path, got:\n%s", got)
	}
}

// TestFormatMarkdown_OneRemoved verifies that a single removed entry produces
// the correct summary and changes rows.
func TestFormatMarkdown_OneRemoved(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"usr/lib/old.so": makeNode("usr/lib/old.so", "def456", 2048, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{})
	result := diff.Diff(a, b)

	got := output.FormatMarkdown(result, "base:1.0", "new:2.0", nil)

	// Removed row in summary shows negative size
	if !strings.Contains(got, "-2.0 KB") {
		t.Errorf("expected '-2.0 KB' in summary Removed row, got:\n%s", got)
	}

	// Changes table status cell
	if !strings.Contains(got, "| `-`") {
		t.Errorf("expected '| `-`' status cell, got:\n%s", got)
	}
	if !strings.Contains(got, "`usr/lib/old.so`") {
		t.Errorf("expected backtick-wrapped path, got:\n%s", got)
	}
}

// TestFormatMarkdown_OneModified verifies that a modified entry with
// ContentChanged and ModeChanged shows "content, mode" in the Details column.
func TestFormatMarkdown_OneModified(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"etc/config": makeNode("etc/config", "aaa", 100, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"etc/config": makeNode("etc/config", "bbb", 100, 0755, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	got := output.FormatMarkdown(result, "a:1", "b:2", nil)

	// Changes table status cell for modified
	if !strings.Contains(got, "| `~`") {
		t.Errorf("expected '| `~`' status cell, got:\n%s", got)
	}
	if !strings.Contains(got, "`etc/config`") {
		t.Errorf("expected backtick-wrapped path, got:\n%s", got)
	}
	// Details column shows "content, mode"
	if !strings.Contains(got, "content, mode") {
		t.Errorf("expected 'content, mode' in details column, got:\n%s", got)
	}
}

// TestFormatMarkdown_SecurityEvent_AddedNoMode verifies that a security event
// for an Added file (Before == 0) produces an empty mode column.
func TestFormatMarkdown_SecurityEvent_AddedNoMode(t *testing.T) {
	events := []security.SecurityEvent{
		{
			Kind:   security.KindNewSUID,
			Path:   "usr/bin/sudo",
			Before: 0,
			After:  os.FileMode(0o4755),
		},
	}

	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/sudo": makeNode("usr/bin/sudo", "abc", 12345, 0o4755, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	got := output.FormatMarkdown(result, "a:1", "b:2", events)

	// Security Findings section present
	if !strings.Contains(got, "### Security Findings") {
		t.Errorf("expected security findings section, got:\n%s", got)
	}
	// Kind label
	if !strings.Contains(got, "SUID") {
		t.Errorf("expected 'SUID' label, got:\n%s", got)
	}
	// Path in backticks
	if !strings.Contains(got, "`usr/bin/sudo`") {
		t.Errorf("expected backtick-wrapped path, got:\n%s", got)
	}
	// Mode column is empty (no backtick-formatted mode string)
	// Before == 0 → mode column should be empty (no "→" arrow)
	if strings.Contains(got, "→") {
		t.Errorf("expected empty mode column for Added event (Before==0), but found '→' in:\n%s", got)
	}
}

// TestFormatMarkdown_SecurityEvent_ModifiedWithMode verifies that a security
// event for a Modified file produces the octal mode transition in the Mode column.
func TestFormatMarkdown_SecurityEvent_ModifiedWithMode(t *testing.T) {
	events := []security.SecurityEvent{
		{
			Kind:   security.KindSUIDAdded,
			Path:   "usr/bin/tool",
			Before: os.FileMode(0o0755),
			After:  os.FileMode(0o4755),
		},
	}

	a := makeTree(map[string]*tree.FileNode{
		"usr/bin/tool": makeNode("usr/bin/tool", "aaa", 5000, 0o0755, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/tool": makeNode("usr/bin/tool", "bbb", 5000, 0o4755, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	got := output.FormatMarkdown(result, "a:1", "b:2", events)

	// Security Findings section present
	if !strings.Contains(got, "### Security Findings") {
		t.Errorf("expected security findings section, got:\n%s", got)
	}
	// Mode column uses backtick-formatted octal: `00755 → 04755`
	if !strings.Contains(got, "`00755 → 04755`") {
		t.Errorf("expected '`00755 → 04755`' in mode column, got:\n%s", got)
	}
}

// TestFormatMarkdown_NoSecuritySection verifies that passing nil or empty
// events omits the Security Findings section entirely.
func TestFormatMarkdown_NoSecuritySection(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "abc", 100, 0644, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	// nil events
	got := output.FormatMarkdown(result, "a:1", "b:2", nil)
	if strings.Contains(got, "### Security Findings") {
		t.Errorf("nil events: unexpected security findings section, got:\n%s", got)
	}

	// empty events slice
	got2 := output.FormatMarkdown(result, "a:1", "b:2", []security.SecurityEvent{})
	if strings.Contains(got2, "### Security Findings") {
		t.Errorf("empty events: unexpected security findings section, got:\n%s", got2)
	}
}

// TestFormatMarkdown_FullReport verifies that a mixed diff (added, removed,
// modified) with security events produces all sections with correct structure.
func TestFormatMarkdown_FullReport(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"usr/lib/old.so":  makeNode("usr/lib/old.so", "aaa", 4096, 0644, 0, 0, false, ""),
		"etc/config.conf": makeNode("etc/config.conf", "bbb", 200, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/wget":    makeNode("usr/bin/wget", "ccc", 123456, 0755, 0, 0, false, ""),
		"etc/config.conf": makeNode("etc/config.conf", "ddd", 250, 0644, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	events := []security.SecurityEvent{
		{
			Kind:   security.KindNewExecutable,
			Path:   "usr/bin/wget",
			Before: 0,
			After:  os.FileMode(0o0755),
		},
	}

	got := output.FormatMarkdown(result, "base:1.0", "new:2.0", events)

	// Header present
	if !strings.Contains(got, "## Image Diff:") {
		t.Errorf("missing header, got:\n%s", got)
	}

	// All three summary rows
	if !strings.Contains(got, "| Added") {
		t.Errorf("missing Added summary row, got:\n%s", got)
	}
	if !strings.Contains(got, "| Removed") {
		t.Errorf("missing Removed summary row, got:\n%s", got)
	}
	if !strings.Contains(got, "| Modified") {
		t.Errorf("missing Modified summary row, got:\n%s", got)
	}

	// Changes table has entries for all three types
	if !strings.Contains(got, "`usr/bin/wget`") {
		t.Errorf("missing added entry in changes table, got:\n%s", got)
	}
	if !strings.Contains(got, "`usr/lib/old.so`") {
		t.Errorf("missing removed entry in changes table, got:\n%s", got)
	}
	if !strings.Contains(got, "`etc/config.conf`") {
		t.Errorf("missing modified entry in changes table, got:\n%s", got)
	}

	// Security findings section present
	if !strings.Contains(got, "### Security Findings") {
		t.Errorf("missing security findings section, got:\n%s", got)
	}
	if !strings.Contains(got, "NEW EXEC") {
		t.Errorf("expected 'NEW EXEC' label, got:\n%s", got)
	}

	// Sections appear in order: Summary before Changes before Security Findings
	summaryIdx := strings.Index(got, "### Summary")
	changesIdx := strings.Index(got, "### Changes")
	securityIdx := strings.Index(got, "### Security Findings")
	if summaryIdx < 0 || changesIdx < 0 || securityIdx < 0 {
		t.Fatalf("missing required sections in:\n%s", got)
	}
	if summaryIdx > changesIdx {
		t.Errorf("Summary should appear before Changes")
	}
	if changesIdx > securityIdx {
		t.Errorf("Changes should appear before Security Findings")
	}
}

// TestFormatMarkdown_SummaryDashWhenZeroBytes verifies that the summary size
// column shows "—" when the byte count for that category is zero.
func TestFormatMarkdown_SummaryDashWhenZeroBytes(t *testing.T) {
	// Only a modified entry — AddedBytes and RemovedBytes are both 0
	a := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "aaa", 100, 0644, 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "bbb", 100, 0644, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	got := output.FormatMarkdown(result, "a:1", "b:2", nil)

	// Both Added and Removed size columns should be "—"
	// Count the number of "—" occurrences (em dash)
	dashCount := strings.Count(got, " — ")
	if dashCount < 2 {
		t.Errorf("expected at least 2 '—' dashes in summary for zero-byte categories, got %d in:\n%s", dashCount, got)
	}
}

// TestFormatMarkdown_ChangesTableHeader verifies the exact table header format.
func TestFormatMarkdown_ChangesTableHeader(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"x": makeNode("x", "abc", 10, 0644, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	got := output.FormatMarkdown(result, "a", "b", nil)

	// Changes table header row
	if !strings.Contains(got, "| Status | Path | Size Delta | Details |") {
		t.Errorf("missing changes table header, got:\n%s", got)
	}
	// Changes table separator row
	if !strings.Contains(got, "|--------|------|------------|---------|") {
		t.Errorf("missing changes table separator, got:\n%s", got)
	}
}

// TestFormatMarkdown_TrailingNewline verifies the output ends with exactly one
// trailing newline.
func TestFormatMarkdown_TrailingNewline(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{})
	result := diff.Diff(a, b)

	got := output.FormatMarkdown(result, "a", "b", nil)

	if !strings.HasSuffix(got, "\n") {
		t.Errorf("output should end with newline")
	}
	// Should not end with double newline
	if strings.HasSuffix(got, "\n\n") {
		t.Errorf("output should not end with double newline, got:\n%q", got)
	}
}
