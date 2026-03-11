package security_test

import (
	"os"
	"testing"

	"github.com/ommmishra/imgdiff/internal/diff"
	"github.com/ommmishra/imgdiff/internal/security"
	"github.com/ommmishra/imgdiff/internal/tree"
)

// makeAdded builds a DiffEntry for an Added file with the given path, mode, and IsDir.
func makeAdded(path string, mode os.FileMode, isDir bool) *diff.DiffEntry {
	return &diff.DiffEntry{
		Path:  path,
		Type:  diff.Added,
		After: &tree.FileNode{Path: path, Mode: mode, IsDir: isDir},
	}
}

// makeModified builds a DiffEntry for a Modified file with the given path, before/after modes.
func makeModified(path string, beforeMode, afterMode os.FileMode, isDir bool) *diff.DiffEntry {
	return &diff.DiffEntry{
		Path:   path,
		Type:   diff.Modified,
		Before: &tree.FileNode{Path: path, Mode: beforeMode, IsDir: isDir},
		After:  &tree.FileNode{Path: path, Mode: afterMode, IsDir: isDir},
	}
}

// makeRemoved builds a DiffEntry for a Removed file.
func makeRemoved(path string, mode os.FileMode) *diff.DiffEntry {
	return &diff.DiffEntry{
		Path:   path,
		Type:   diff.Removed,
		Before: &tree.FileNode{Path: path, Mode: mode},
	}
}

// kindSlice extracts just the Kind strings from a []SecurityEvent.
func kindSlice(events []security.SecurityEvent) []string {
	out := make([]string, len(events))
	for i, e := range events {
		out[i] = string(e.Kind)
	}
	return out
}

// TestAnalyzeEmptyResult verifies that an empty DiffResult produces an empty (non-nil) slice.
func TestAnalyzeEmptyResult(t *testing.T) {
	result := &diff.DiffResult{}
	events := security.Analyze(result)
	if events == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

// TestAnalyzeAddedRegularFileNoSpecialBits verifies that a plain added file emits no events.
func TestAnalyzeAddedRegularFileNoSpecialBits(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeAdded("/usr/bin/foo", 0o644, false)},
	}
	events := security.Analyze(result)
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %v", kindSlice(events))
	}
}

// TestAnalyzeAddedSUIDFile verifies "new_suid" + "new_executable" for SUID file.
func TestAnalyzeAddedSUIDFile(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeAdded("/usr/bin/sudo", 0o4755, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	want := []string{"new_suid", "new_executable"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
	// Verify path and mode are correctly populated.
	for _, e := range events {
		if e.Path != "/usr/bin/sudo" {
			t.Errorf("expected path /usr/bin/sudo, got %s", e.Path)
		}
		if e.Before != 0 {
			t.Errorf("expected Before=0 for Added entry, got %v", e.Before)
		}
		if e.After != 0o4755 {
			t.Errorf("expected After=0o4755, got %v", e.After)
		}
	}
}

// TestAnalyzeAddedSGIDFile verifies "new_sgid" + "new_executable" for SGID file.
func TestAnalyzeAddedSGIDFile(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeAdded("/usr/bin/wall", 0o2755, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	want := []string{"new_sgid", "new_executable"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
}

// TestAnalyzeAddedExecOnlyFile verifies "new_executable" only for a plain exec file.
func TestAnalyzeAddedExecOnlyFile(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeAdded("/usr/bin/cat", 0o755, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	want := []string{"new_executable"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
}

// TestAnalyzeAddedDirectoryWithExecBit verifies no events for an added directory with exec bits.
func TestAnalyzeAddedDirectoryWithExecBit(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeAdded("/usr/bin", 0o755, true)},
	}
	events := security.Analyze(result)
	if len(events) != 0 {
		t.Fatalf("expected 0 events for directory, got %v", kindSlice(events))
	}
}

// TestAnalyzeAddedWorldWritableExecFile verifies "new_executable" + "world_writable" for mode 0o777.
func TestAnalyzeAddedWorldWritableExecFile(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeAdded("/tmp/script.sh", 0o777, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	want := []string{"new_executable", "world_writable"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
}

// TestAnalyzeAddedWorldWritableNonExecFile verifies "world_writable" only for mode 0o666.
func TestAnalyzeAddedWorldWritableNonExecFile(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeAdded("/etc/writable.conf", 0o666, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	want := []string{"world_writable"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
}

// TestAnalyzeModifiedSUIDAdded verifies "suid_added" + "perm_escalation" when SUID is added.
func TestAnalyzeModifiedSUIDAdded(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeModified("/usr/bin/prog", 0o755, 0o4755, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	want := []string{"suid_added", "perm_escalation"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
	for _, e := range events {
		if e.Before != 0o755 {
			t.Errorf("expected Before=0o755, got %v", e.Before)
		}
		if e.After != 0o4755 {
			t.Errorf("expected After=0o4755, got %v", e.After)
		}
	}
}

// TestAnalyzeModifiedSGIDAdded verifies "sgid_added" + "perm_escalation" when SGID is added.
func TestAnalyzeModifiedSGIDAdded(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeModified("/usr/bin/prog", 0o755, 0o2755, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	want := []string{"sgid_added", "perm_escalation"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
}

// TestAnalyzeModifiedPermissionsLoosenedNoNewExec verifies that "new_executable" is NOT emitted
// for Modified entries, only "perm_escalation".
func TestAnalyzeModifiedPermissionsLoosenedNoNewExec(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeModified("/etc/script", 0o644, 0o755, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	// Only perm_escalation; new_executable is only for Added entries.
	want := []string{"perm_escalation"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
}

// TestAnalyzeModifiedWorldWritableAdded verifies "world_writable" + "perm_escalation".
func TestAnalyzeModifiedWorldWritableAdded(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeModified("/etc/config", 0o644, 0o646, false)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	want := []string{"world_writable", "perm_escalation"}
	if !equalStringSlice(kinds, want) {
		t.Fatalf("expected %v, got %v", want, kinds)
	}
}

// TestAnalyzeModifiedPermissionsTightened verifies no events when permissions are reduced.
func TestAnalyzeModifiedPermissionsTightened(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeModified("/usr/bin/prog", 0o755, 0o644, false)},
	}
	events := security.Analyze(result)
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %v", kindSlice(events))
	}
}

// TestAnalyzeRemovedEntry verifies that a removed file emits no events.
func TestAnalyzeRemovedEntry(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeRemoved("/usr/bin/oldtool", 0o4755)},
	}
	events := security.Analyze(result)
	if len(events) != 0 {
		t.Fatalf("expected 0 events for removed entry, got %v", kindSlice(events))
	}
}

// TestAnalyzeMultipleEventsPerEntry verifies that a single entry can produce multiple events.
func TestAnalyzeMultipleEventsPerEntry(t *testing.T) {
	// SUID + exec = 2 events for one entry.
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeAdded("/usr/bin/newsetuid", 0o4755, false)},
	}
	events := security.Analyze(result)
	if len(events) != 2 {
		t.Fatalf("expected 2 events from one entry, got %d: %v", len(events), kindSlice(events))
	}
}

// TestAnalyzeEventOrder verifies events are returned in entry iteration order.
func TestAnalyzeEventOrder(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{
			makeAdded("/a", 0o755, false), // new_executable
			makeAdded("/b", 0o666, false), // world_writable
		},
	}
	events := security.Analyze(result)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Kind != security.KindNewExecutable {
		t.Errorf("expected first event to be new_executable, got %s", events[0].Kind)
	}
	if events[1].Kind != security.KindWorldWritable {
		t.Errorf("expected second event to be world_writable, got %s", events[1].Kind)
	}
}

// TestAnalyzeModifiedDirectoryWorldWritable verifies no event when a directory is world-writable
// (directories are excluded from world_writable detection).
func TestAnalyzeModifiedDirectoryWorldWritable(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{makeModified("/tmp", 0o755, 0o777, true)},
	}
	events := security.Analyze(result)
	kinds := kindSlice(events)
	// perm_escalation yes, world_writable no (it's a directory).
	for _, k := range kinds {
		if k == "world_writable" {
			t.Errorf("unexpected world_writable event for directory, events: %v", kinds)
		}
	}
}

// equalStringSlice returns true if two string slices are equal element by element.
func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
