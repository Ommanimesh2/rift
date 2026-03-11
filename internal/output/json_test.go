package output_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ommmishra/imgdiff/internal/diff"
	"github.com/ommmishra/imgdiff/internal/output"
	"github.com/ommmishra/imgdiff/internal/security"
	"github.com/ommmishra/imgdiff/internal/tree"
)

// TestFormatJSON_EmptyResult verifies that an empty DiffResult produces a
// valid JSON object with all-zero summary counts, an empty changes array,
// and an empty security_events array (not null).
func TestFormatJSON_EmptyResult(t *testing.T) {
	result := diff.Diff(
		&tree.FileTree{Entries: map[string]*tree.FileNode{}},
		&tree.FileTree{Entries: map[string]*tree.FileNode{}},
	)

	got, err := output.FormatJSON(result, "img1:latest", "img2:latest", nil)
	if err != nil {
		t.Fatalf("FormatJSON returned error: %v", err)
	}

	var report output.DiffReport
	if err := json.Unmarshal(got, &report); err != nil {
		t.Fatalf("FormatJSON output is not valid JSON: %v\noutput: %s", err, got)
	}

	if report.Image1 != "img1:latest" {
		t.Errorf("Image1 = %q, want %q", report.Image1, "img1:latest")
	}
	if report.Image2 != "img2:latest" {
		t.Errorf("Image2 = %q, want %q", report.Image2, "img2:latest")
	}

	s := report.Summary
	if s.Added != 0 || s.Removed != 0 || s.Modified != 0 {
		t.Errorf("Summary counts non-zero for empty diff: %+v", s)
	}
	if s.AddedBytes != 0 || s.RemovedBytes != 0 {
		t.Errorf("Summary bytes non-zero for empty diff: %+v", s)
	}

	if report.Changes == nil || len(report.Changes) != 0 {
		t.Errorf("Changes should be empty slice, got: %v", report.Changes)
	}
	if report.SecurityEvents == nil || len(report.SecurityEvents) != 0 {
		t.Errorf("SecurityEvents should be empty slice (not nil), got: %v", report.SecurityEvents)
	}
}

// TestFormatJSON_NilEventsBecomesEmptyArray checks that passing nil as the
// events slice still renders "security_events": [] in JSON (not null).
func TestFormatJSON_NilEventsBecomesEmptyArray(t *testing.T) {
	result := diff.Diff(
		&tree.FileTree{Entries: map[string]*tree.FileNode{}},
		&tree.FileTree{Entries: map[string]*tree.FileNode{}},
	)

	got, err := output.FormatJSON(result, "a", "b", nil)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	// Verify the raw JSON contains [] not null for security_events
	raw := string(got)
	if !containsSubstr(raw, `"security_events": []`) {
		t.Errorf("expected security_events to be [], got JSON: %s", raw)
	}
}

// TestFormatJSON_AddedEntry verifies that an Added DiffEntry maps to a
// ChangeEntry with type "added", the correct path, and positive size_delta.
func TestFormatJSON_AddedEntry(t *testing.T) {
	after := &tree.FileNode{Path: "usr/bin/wget", Size: 123456, Mode: 0755}
	entry := &diff.DiffEntry{
		Path:      "usr/bin/wget",
		Type:      diff.Added,
		After:     after,
		SizeDelta: 123456,
	}
	result := &diff.DiffResult{
		Entries:    []*diff.DiffEntry{entry},
		Added:      1,
		AddedBytes: 123456,
	}

	got, err := output.FormatJSON(result, "img1", "img2", nil)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	var report output.DiffReport
	if err := json.Unmarshal(got, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(report.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(report.Changes))
	}

	c := report.Changes[0]
	if c.Path != "usr/bin/wget" {
		t.Errorf("Path = %q, want %q", c.Path, "usr/bin/wget")
	}
	if c.Type != "added" {
		t.Errorf("Type = %q, want %q", c.Type, "added")
	}
	if c.SizeDelta != 123456 {
		t.Errorf("SizeDelta = %d, want %d", c.SizeDelta, 123456)
	}
	// Added entries must NOT include a "changes" array
	if c.Changes != nil {
		t.Errorf("Changes should be nil for Added entry, got: %v", c.Changes)
	}

	if report.Summary.Added != 1 {
		t.Errorf("Summary.Added = %d, want 1", report.Summary.Added)
	}
	if report.Summary.AddedBytes != 123456 {
		t.Errorf("Summary.AddedBytes = %d, want 123456", report.Summary.AddedBytes)
	}
}

// TestFormatJSON_RemovedEntry verifies that a Removed DiffEntry maps to a
// ChangeEntry with type "removed" and negative size_delta.
func TestFormatJSON_RemovedEntry(t *testing.T) {
	before := &tree.FileNode{Path: "usr/lib/old.so", Size: 89012, Mode: 0644}
	entry := &diff.DiffEntry{
		Path:      "usr/lib/old.so",
		Type:      diff.Removed,
		Before:    before,
		SizeDelta: -89012,
	}
	result := &diff.DiffResult{
		Entries:      []*diff.DiffEntry{entry},
		Removed:      1,
		RemovedBytes: 89012,
	}

	got, err := output.FormatJSON(result, "img1", "img2", nil)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	var report output.DiffReport
	if err := json.Unmarshal(got, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(report.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(report.Changes))
	}

	c := report.Changes[0]
	if c.Path != "usr/lib/old.so" {
		t.Errorf("Path = %q, want %q", c.Path, "usr/lib/old.so")
	}
	if c.Type != "removed" {
		t.Errorf("Type = %q, want %q", c.Type, "removed")
	}
	if c.SizeDelta != -89012 {
		t.Errorf("SizeDelta = %d, want %d", c.SizeDelta, -89012)
	}
	if c.Changes != nil {
		t.Errorf("Changes should be nil for Removed entry, got: %v", c.Changes)
	}

	if report.Summary.Removed != 1 {
		t.Errorf("Summary.Removed = %d, want 1", report.Summary.Removed)
	}
	if report.Summary.RemovedBytes != 89012 {
		t.Errorf("Summary.RemovedBytes = %d, want 89012", report.Summary.RemovedBytes)
	}
}

// TestFormatJSON_ModifiedWithFlags verifies that a Modified DiffEntry with
// content and mode changes produces a ChangeEntry with "changes": ["content","mode"].
func TestFormatJSON_ModifiedWithFlags(t *testing.T) {
	beforeNode := &tree.FileNode{Path: "etc/nginx.conf", Size: 500, Mode: 0644}
	afterNode := &tree.FileNode{Path: "etc/nginx.conf", Size: 545, Mode: 0755}
	entry := &diff.DiffEntry{
		Path:           "etc/nginx.conf",
		Type:           diff.Modified,
		Before:         beforeNode,
		After:          afterNode,
		SizeDelta:      45,
		ContentChanged: true,
		ModeChanged:    true,
	}
	result := &diff.DiffResult{
		Entries:  []*diff.DiffEntry{entry},
		Modified: 1,
	}

	got, err := output.FormatJSON(result, "img1", "img2", nil)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	var report output.DiffReport
	if err := json.Unmarshal(got, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(report.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(report.Changes))
	}

	c := report.Changes[0]
	if c.Path != "etc/nginx.conf" {
		t.Errorf("Path = %q, want %q", c.Path, "etc/nginx.conf")
	}
	if c.Type != "modified" {
		t.Errorf("Type = %q, want %q", c.Type, "modified")
	}
	if c.SizeDelta != 45 {
		t.Errorf("SizeDelta = %d, want 45", c.SizeDelta)
	}
	if len(c.Changes) != 2 {
		t.Fatalf("expected 2 change flags, got %v", c.Changes)
	}
	if c.Changes[0] != "content" {
		t.Errorf("Changes[0] = %q, want %q", c.Changes[0], "content")
	}
	if c.Changes[1] != "mode" {
		t.Errorf("Changes[1] = %q, want %q", c.Changes[1], "mode")
	}

	if report.Summary.Modified != 1 {
		t.Errorf("Summary.Modified = %d, want 1", report.Summary.Modified)
	}
}

// TestFormatJSON_ModifiedNoFlags verifies that a Modified entry with no flags
// set omits the "changes" key entirely from the JSON output (omitempty).
func TestFormatJSON_ModifiedNoFlags(t *testing.T) {
	beforeNode := &tree.FileNode{Path: "etc/hosts", Size: 100, Mode: 0644}
	afterNode := &tree.FileNode{Path: "etc/hosts", Size: 100, Mode: 0644}
	entry := &diff.DiffEntry{
		Path:      "etc/hosts",
		Type:      diff.Modified,
		Before:    beforeNode,
		After:     afterNode,
		SizeDelta: 0,
		// all change flags are false
	}
	result := &diff.DiffResult{
		Entries:  []*diff.DiffEntry{entry},
		Modified: 1,
	}

	got, err := output.FormatJSON(result, "img1", "img2", nil)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	// Verify "changes" key is absent in the raw JSON
	raw := string(got)
	if containsSubstr(raw, `"changes"`) {
		t.Errorf("expected no 'changes' key when flags are empty, got: %s", raw)
	}
}

// TestFormatJSON_SecurityEvents_NewSUID checks that a KindNewSUID event is
// rendered correctly: before_mode == 0, after_mode as uint32 decimal.
func TestFormatJSON_SecurityEvents_NewSUID(t *testing.T) {
	result := &diff.DiffResult{Entries: []*diff.DiffEntry{}}
	events := []security.SecurityEvent{
		{
			Kind:  security.KindNewSUID,
			Path:  "/usr/bin/ping",
			Before: 0,
			After:  os.FileMode(0o4755), // 2541 decimal
		},
	}

	got, err := output.FormatJSON(result, "img1", "img2", events)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	var report output.DiffReport
	if err := json.Unmarshal(got, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(report.SecurityEvents) != 1 {
		t.Fatalf("expected 1 security event, got %d", len(report.SecurityEvents))
	}

	ev := report.SecurityEvents[0]
	if ev.Kind != "new_suid" {
		t.Errorf("Kind = %q, want %q", ev.Kind, "new_suid")
	}
	if ev.Path != "/usr/bin/ping" {
		t.Errorf("Path = %q, want %q", ev.Path, "/usr/bin/ping")
	}
	if ev.BeforeMode != 0 {
		t.Errorf("BeforeMode = %d, want 0", ev.BeforeMode)
	}
	// 0o4755 = 2541
	if ev.AfterMode != 2541 {
		t.Errorf("AfterMode = %d, want 2541 (0o4755)", ev.AfterMode)
	}
}

// TestFormatJSON_SecurityEvents_PermEscalation checks that a KindPermEscalation
// event carries both before_mode and after_mode.
func TestFormatJSON_SecurityEvents_PermEscalation(t *testing.T) {
	result := &diff.DiffResult{Entries: []*diff.DiffEntry{}}
	events := []security.SecurityEvent{
		{
			Kind:  security.KindPermEscalation,
			Path:  "/bin/sh",
			Before: os.FileMode(0o755), // 493 decimal
			After:  os.FileMode(0o4755), // 2541 decimal
		},
	}

	got, err := output.FormatJSON(result, "img1", "img2", events)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	var report output.DiffReport
	if err := json.Unmarshal(got, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(report.SecurityEvents) != 1 {
		t.Fatalf("expected 1 security event, got %d", len(report.SecurityEvents))
	}

	ev := report.SecurityEvents[0]
	if ev.Kind != "perm_escalation" {
		t.Errorf("Kind = %q, want %q", ev.Kind, "perm_escalation")
	}
	if ev.Path != "/bin/sh" {
		t.Errorf("Path = %q, want %q", ev.Path, "/bin/sh")
	}
	// 0o755 = 493
	if ev.BeforeMode != 493 {
		t.Errorf("BeforeMode = %d, want 493 (0o755)", ev.BeforeMode)
	}
	// 0o4755 = 2541
	if ev.AfterMode != 2541 {
		t.Errorf("AfterMode = %d, want 2541 (0o4755)", ev.AfterMode)
	}
}

// TestFormatJSON_ModifiedAllFlags verifies canonical flag order for all six flags.
func TestFormatJSON_ModifiedAllFlags(t *testing.T) {
	beforeNode := &tree.FileNode{Path: "bin/tool", Size: 100, Mode: 0644}
	afterNode := &tree.FileNode{Path: "bin/tool", Size: 100, Mode: 0755}
	entry := &diff.DiffEntry{
		Path:              "bin/tool",
		Type:              diff.Modified,
		Before:            beforeNode,
		After:             afterNode,
		SizeDelta:         0,
		ContentChanged:    true,
		ModeChanged:       true,
		UIDChanged:        true,
		GIDChanged:        true,
		LinkTargetChanged: true,
		TypeChanged:       true,
	}
	result := &diff.DiffResult{
		Entries:  []*diff.DiffEntry{entry},
		Modified: 1,
	}

	got, err := output.FormatJSON(result, "img1", "img2", nil)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	var report output.DiffReport
	if err := json.Unmarshal(got, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(report.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(report.Changes))
	}

	c := report.Changes[0]
	wantFlags := []string{"content", "mode", "uid", "gid", "link", "type"}
	if len(c.Changes) != len(wantFlags) {
		t.Fatalf("Changes = %v, want %v", c.Changes, wantFlags)
	}
	for i, want := range wantFlags {
		if c.Changes[i] != want {
			t.Errorf("Changes[%d] = %q, want %q", i, c.Changes[i], want)
		}
	}
}

// TestFormatJSON_OutputIsIndented verifies the output uses 2-space indentation
// (json.MarshalIndent with "" prefix and "  " indent).
func TestFormatJSON_OutputIsIndented(t *testing.T) {
	result := &diff.DiffResult{Entries: []*diff.DiffEntry{}}

	got, err := output.FormatJSON(result, "a", "b", nil)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	raw := string(got)
	// Indented JSON starts with "{\n"
	if !containsSubstr(raw, "{\n") {
		t.Errorf("expected indented JSON (containing {\\n), got: %s", raw)
	}
	// Two-space indent means we should see "  \"" (2 spaces before a key)
	if !containsSubstr(raw, "  \"") {
		t.Errorf("expected 2-space indentation, got: %s", raw)
	}
	// No trailing newline after the final "}"
	if len(raw) > 0 && raw[len(raw)-1] == '\n' {
		t.Errorf("FormatJSON must not append a trailing newline")
	}
}

// containsSubstr is a helper to check string containment without importing strings.
func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstr(s, sub))
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
