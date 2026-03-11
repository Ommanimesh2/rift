// Package cmd_test contains integration tests for the cmd package.
// These tests validate that the output formatters are importable and produce
// correctly structured output — they do not invoke rootCmd with real images.
package cmd_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ommmishra/imgdiff/internal/diff"
	"github.com/ommmishra/imgdiff/internal/output"
	"github.com/ommmishra/imgdiff/internal/security"
	"github.com/ommmishra/imgdiff/internal/tree"
)

// buildTestResult creates a minimal DiffResult with one added and one modified entry
// suitable for integration smoke tests.
func buildTestResult() *diff.DiffResult {
	ft1 := &tree.FileTree{Entries: map[string]*tree.FileNode{
		"usr/bin/old": {Path: "usr/bin/old", Size: 1024, Mode: 0o755},
	}}
	ft2 := &tree.FileTree{Entries: map[string]*tree.FileNode{
		// Mode differs from ft1 → classified as Modified.
		"usr/bin/old": {Path: "usr/bin/old", Size: 2048, Mode: 0o644},
		"usr/bin/new": {Path: "usr/bin/new", Size: 512},
	}}
	return diff.Diff(ft1, ft2)
}

// TestFormatJSON_ValidJSON verifies that FormatJSON returns valid JSON and that
// the top-level keys image1, image2, summary, changes, and security_events are present.
func TestFormatJSON_ValidJSON(t *testing.T) {
	result := buildTestResult()
	events := security.Analyze(result)

	data, err := output.FormatJSON(result, "app:v1", "app:v2", events)
	if err != nil {
		t.Fatalf("FormatJSON returned unexpected error: %v", err)
	}

	// Must be valid JSON.
	var report output.DiffReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("FormatJSON output is not valid JSON: %v\noutput:\n%s", err, data)
	}

	// Check image names are preserved.
	if report.Image1 != "app:v1" {
		t.Errorf("Image1 = %q, want %q", report.Image1, "app:v1")
	}
	if report.Image2 != "app:v2" {
		t.Errorf("Image2 = %q, want %q", report.Image2, "app:v2")
	}

	// The test result has 1 added entry (usr/bin/new) and 1 modified entry (usr/bin/old).
	if report.Summary.Added != 1 {
		t.Errorf("Summary.Added = %d, want 1", report.Summary.Added)
	}
	if report.Summary.Modified != 1 {
		t.Errorf("Summary.Modified = %d, want 1", report.Summary.Modified)
	}

	// Changes slice must be non-nil and contain 2 entries.
	if report.Changes == nil {
		t.Fatal("Changes field must not be nil")
	}
	if len(report.Changes) != 2 {
		t.Errorf("len(Changes) = %d, want 2", len(report.Changes))
	}

	// SecurityEvents must be non-nil (may be empty).
	if report.SecurityEvents == nil {
		t.Fatal("SecurityEvents must not be nil")
	}
}

// TestFormatJSON_SecurityEventsPresent verifies that security events are serialized
// with the expected kind and path fields.
func TestFormatJSON_SecurityEventsPresent(t *testing.T) {
	result := &diff.DiffResult{
		Modified: 1,
		Entries: []*diff.DiffEntry{
			{
				Path:        "usr/bin/sudo",
				Type:        diff.Modified,
				ModeChanged: true,
				Before:      &tree.FileNode{Path: "usr/bin/sudo", Mode: 0o755},
				After:       &tree.FileNode{Path: "usr/bin/sudo", Mode: 0o4755},
			},
		},
	}
	events := security.Analyze(result)

	data, err := output.FormatJSON(result, "base:1", "base:2", events)
	if err != nil {
		t.Fatalf("FormatJSON returned unexpected error: %v", err)
	}

	var report output.DiffReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("FormatJSON output is not valid JSON: %v", err)
	}

	if len(report.SecurityEvents) == 0 {
		t.Fatal("expected at least one SecurityEvent, got none")
	}

	ev := report.SecurityEvents[0]
	if ev.Path != "usr/bin/sudo" {
		t.Errorf("SecurityEvent.Path = %q, want %q", ev.Path, "usr/bin/sudo")
	}
	if ev.Kind == "" {
		t.Error("SecurityEvent.Kind must not be empty")
	}
}

// TestFormatMarkdown_ContainsRequiredSections verifies that FormatMarkdown produces
// output containing the "## Image Diff", "### Summary", and "### Changes" section headers.
func TestFormatMarkdown_ContainsRequiredSections(t *testing.T) {
	result := buildTestResult()
	events := security.Analyze(result)

	md := output.FormatMarkdown(result, "app:v1", "app:v2", events)

	requiredSections := []string{
		"## Image Diff",
		"### Summary",
		"### Changes",
	}
	for _, section := range requiredSections {
		if !strings.Contains(md, section) {
			t.Errorf("FormatMarkdown output missing section %q\noutput:\n%s", section, md)
		}
	}
}

// TestFormatMarkdown_ImageNamesInHeader verifies that the image names appear in
// the Markdown header line.
func TestFormatMarkdown_ImageNamesInHeader(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{},
	}

	md := output.FormatMarkdown(result, "nginx:1.24", "nginx:1.25", nil)

	if !strings.Contains(md, "nginx:1.24") {
		t.Errorf("FormatMarkdown header missing image1 name\noutput:\n%s", md)
	}
	if !strings.Contains(md, "nginx:1.25") {
		t.Errorf("FormatMarkdown header missing image2 name\noutput:\n%s", md)
	}

	// Empty result should show "(No changes)" in the changes section.
	if !strings.Contains(md, "(No changes)") {
		t.Errorf("FormatMarkdown empty result should contain \"(No changes)\"\noutput:\n%s", md)
	}
}
