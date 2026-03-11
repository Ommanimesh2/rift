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

// --- FormatSecurityEvent tests ---

func TestFormatSecurityEvent_AddedSUID(t *testing.T) {
	ev := security.SecurityEvent{
		Kind:  security.KindNewSUID,
		Path:  "usr/bin/newbinary",
		After: os.FileMode(0o4755),
	}

	got := output.FormatSecurityEvent(ev)

	// Added events (Before == 0) should not show mode arrows.
	if !strings.Contains(got, "[SUID]") {
		t.Errorf("FormatSecurityEvent(new_suid) = %q, want it to contain [SUID]", got)
	}
	if !strings.Contains(got, "usr/bin/newbinary") {
		t.Errorf("FormatSecurityEvent(new_suid) = %q, want it to contain the path", got)
	}
	if strings.Contains(got, "→") {
		t.Errorf("FormatSecurityEvent(new_suid added) = %q, want no mode arrows for Added event", got)
	}
}

func TestFormatSecurityEvent_AddedSGID(t *testing.T) {
	ev := security.SecurityEvent{
		Kind:  security.KindNewSGID,
		Path:  "usr/bin/sgidbin",
		After: os.FileMode(0o2755),
	}

	got := output.FormatSecurityEvent(ev)

	if !strings.Contains(got, "[SGID]") {
		t.Errorf("FormatSecurityEvent(new_sgid) = %q, want [SGID]", got)
	}
	if strings.Contains(got, "→") {
		t.Errorf("FormatSecurityEvent(new_sgid) = %q, want no arrows", got)
	}
}

func TestFormatSecurityEvent_ModifiedPermEscalation(t *testing.T) {
	ev := security.SecurityEvent{
		Kind:   security.KindPermEscalation,
		Path:   "etc/shadow",
		Before: os.FileMode(0o640),
		After:  os.FileMode(0o646),
	}

	got := output.FormatSecurityEvent(ev)

	if !strings.Contains(got, "[PERM ESCALATION]") {
		t.Errorf("FormatSecurityEvent(perm_escalation) = %q, want [PERM ESCALATION]", got)
	}
	if !strings.Contains(got, "etc/shadow") {
		t.Errorf("FormatSecurityEvent(perm_escalation) = %q, want path etc/shadow", got)
	}
	if !strings.Contains(got, "→") {
		t.Errorf("FormatSecurityEvent(perm_escalation) = %q, want mode arrows", got)
	}
	if !strings.Contains(got, "0640") {
		t.Errorf("FormatSecurityEvent(perm_escalation) = %q, want before mode 0640", got)
	}
	if !strings.Contains(got, "0646") {
		t.Errorf("FormatSecurityEvent(perm_escalation) = %q, want after mode 0646", got)
	}
}

func TestFormatSecurityEvent_WorldWritable(t *testing.T) {
	ev := security.SecurityEvent{
		Kind:   security.KindWorldWritable,
		Path:   "tmp/data",
		Before: os.FileMode(0o644),
		After:  os.FileMode(0o646),
	}

	got := output.FormatSecurityEvent(ev)

	if !strings.Contains(got, "[WORLD-WRITABLE]") {
		t.Errorf("FormatSecurityEvent(world_writable) = %q, want [WORLD-WRITABLE]", got)
	}
}

func TestFormatSecurityEvent_SUIDAdded(t *testing.T) {
	ev := security.SecurityEvent{
		Kind:   security.KindSUIDAdded,
		Path:   "usr/bin/modified",
		Before: os.FileMode(0o755),
		After:  os.FileMode(0o4755),
	}

	got := output.FormatSecurityEvent(ev)

	if !strings.Contains(got, "[SUID ADDED]") {
		t.Errorf("FormatSecurityEvent(suid_added) = %q, want [SUID ADDED]", got)
	}
	if !strings.Contains(got, "0755") {
		t.Errorf("FormatSecurityEvent(suid_added) = %q, want before mode 0755", got)
	}
	if !strings.Contains(got, "4755") {
		t.Errorf("FormatSecurityEvent(suid_added) = %q, want after mode 4755", got)
	}
}

func TestFormatSecurityEvent_NewExecutable(t *testing.T) {
	ev := security.SecurityEvent{
		Kind:  security.KindNewExecutable,
		Path:  "usr/local/bin/tool",
		After: os.FileMode(0o755),
	}

	got := output.FormatSecurityEvent(ev)

	if !strings.Contains(got, "[NEW EXEC]") {
		t.Errorf("FormatSecurityEvent(new_executable) = %q, want [NEW EXEC]", got)
	}
	if strings.Contains(got, "→") {
		t.Errorf("FormatSecurityEvent(new_executable) = %q, want no arrows for Added event", got)
	}
}

// --- RenderTerminalWithSecurity tests ---

func TestRenderTerminalWithSecurity_EmptyEvents_NilSlice(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "abc", 100, 0644, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	withSecurity := output.RenderTerminalWithSecurity(result, "img1", "img2", nil, nil)
	withLayers := output.RenderTerminalWithLayers(result, "img1", "img2", nil)

	if withSecurity != withLayers {
		t.Errorf("RenderTerminalWithSecurity(nil events) should match RenderTerminalWithLayers\ngot:  %q\nwant: %q", withSecurity, withLayers)
	}
}

func TestRenderTerminalWithSecurity_EmptyEvents_EmptySlice(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"file.txt": makeNode("file.txt", "abc", 100, 0644, 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	withSecurity := output.RenderTerminalWithSecurity(result, "img1", "img2", nil, []security.SecurityEvent{})
	withLayers := output.RenderTerminalWithLayers(result, "img1", "img2", nil)

	if withSecurity != withLayers {
		t.Errorf("RenderTerminalWithSecurity(empty events) should match RenderTerminalWithLayers\ngot:  %q\nwant: %q", withSecurity, withLayers)
	}
}

func TestRenderTerminalWithSecurity_SingleEvent_SectionAppears(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/newbinary": makeNode("usr/bin/newbinary", "abc", 1024, os.FileMode(0o4755), 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	events := []security.SecurityEvent{
		{
			Kind:  security.KindNewSUID,
			Path:  "usr/bin/newbinary",
			After: os.FileMode(0o4755),
		},
	}

	got := output.RenderTerminalWithSecurity(result, "img1", "img2", nil, events)

	if !strings.Contains(got, "Security Findings") {
		t.Errorf("RenderTerminalWithSecurity: expected security section header, got:\n%s", got)
	}
	if !strings.Contains(got, "SUID") {
		t.Errorf("RenderTerminalWithSecurity: expected SUID label, got:\n%s", got)
	}
	if !strings.Contains(got, "usr/bin/newbinary") {
		t.Errorf("RenderTerminalWithSecurity: expected path in security section, got:\n%s", got)
	}
}

func TestRenderTerminalWithSecurity_MultipleEvents_AllListed(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{
		"etc/shadow": makeNode("etc/shadow", "aaa", 100, os.FileMode(0o640), 0, 0, false, ""),
	})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/tool": makeNode("usr/bin/tool", "bbb", 1024, os.FileMode(0o755), 0, 0, false, ""),
		"etc/shadow":   makeNode("etc/shadow", "ccc", 100, os.FileMode(0o646), 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	events := []security.SecurityEvent{
		{
			Kind:  security.KindNewExecutable,
			Path:  "usr/bin/tool",
			After: os.FileMode(0o755),
		},
		{
			Kind:   security.KindWorldWritable,
			Path:   "etc/shadow",
			Before: os.FileMode(0o640),
			After:  os.FileMode(0o646),
		},
	}

	got := output.RenderTerminalWithSecurity(result, "img1", "img2", nil, events)

	if !strings.Contains(got, "Security Findings (2)") {
		t.Errorf("RenderTerminalWithSecurity(2 events): expected 'Security Findings (2)', got:\n%s", got)
	}
	if !strings.Contains(got, "usr/bin/tool") {
		t.Errorf("RenderTerminalWithSecurity: expected usr/bin/tool in output, got:\n%s", got)
	}
	if !strings.Contains(got, "etc/shadow") {
		t.Errorf("RenderTerminalWithSecurity: expected etc/shadow in output, got:\n%s", got)
	}
	if !strings.Contains(got, "NEW EXEC") {
		t.Errorf("RenderTerminalWithSecurity: expected NEW EXEC label, got:\n%s", got)
	}
	if !strings.Contains(got, "WORLD-WRITABLE") {
		t.Errorf("RenderTerminalWithSecurity: expected WORLD-WRITABLE label, got:\n%s", got)
	}
}

func TestRenderTerminalWithSecurity_SecuritySectionBeforeSummary(t *testing.T) {
	a := makeTree(map[string]*tree.FileNode{})
	b := makeTree(map[string]*tree.FileNode{
		"usr/bin/suid": makeNode("usr/bin/suid", "abc", 1024, os.FileMode(0o4755), 0, 0, false, ""),
	})
	result := diff.Diff(a, b)

	events := []security.SecurityEvent{
		{
			Kind:  security.KindNewSUID,
			Path:  "usr/bin/suid",
			After: os.FileMode(0o4755),
		},
	}

	got := output.RenderTerminalWithSecurity(result, "img1", "img2", nil, events)

	secIdx := strings.Index(got, "Security Findings")
	summaryIdx := strings.Index(got, "Summary:")

	if secIdx < 0 {
		t.Fatalf("RenderTerminalWithSecurity: no security section found in:\n%s", got)
	}
	if summaryIdx < 0 {
		t.Fatalf("RenderTerminalWithSecurity: no summary found in:\n%s", got)
	}
	if secIdx > summaryIdx {
		t.Errorf("RenderTerminalWithSecurity: security section should appear before summary")
	}
}
