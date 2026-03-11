// Package output provides text formatting and terminal rendering functions
// for DiffResult output. This file implements the JSON formatter for
// machine-readable CI/CD pipeline consumption.
package output

import (
	"encoding/json"

	"github.com/ommmishra/imgdiff/internal/diff"
	"github.com/ommmishra/imgdiff/internal/security"
)

// DiffReport is the top-level JSON structure produced by FormatJSON.
type DiffReport struct {
	Image1         string          `json:"image1"`
	Image2         string          `json:"image2"`
	Summary        ReportSummary   `json:"summary"`
	Changes        []ChangeEntry   `json:"changes"`
	SecurityEvents []SecurityEntry `json:"security_events"`
}

// ReportSummary holds aggregate counts and byte totals from a DiffResult.
type ReportSummary struct {
	Added        int   `json:"added"`
	Removed      int   `json:"removed"`
	Modified     int   `json:"modified"`
	AddedBytes   int64 `json:"added_bytes"`
	RemovedBytes int64 `json:"removed_bytes"`
}

// ChangeEntry represents a single file-system path change in the JSON output.
// The Changes field is only populated for Modified entries and is omitted when empty.
type ChangeEntry struct {
	Path      string   `json:"path"`
	Type      string   `json:"type"`
	SizeDelta int64    `json:"size_delta"`
	Changes   []string `json:"changes,omitempty"`
}

// SecurityEntry represents a single security-relevant change in the JSON output.
// BeforeMode and AfterMode are stored as uint32 decimal integers.
type SecurityEntry struct {
	Kind       string `json:"kind"`
	Path       string `json:"path"`
	BeforeMode uint32 `json:"before_mode"`
	AfterMode  uint32 `json:"after_mode"`
}

// FormatJSON serializes a DiffResult along with image names and security events
// into indented JSON bytes suitable for CI/CD pipeline consumption.
//
// The output uses json.MarshalIndent with "" prefix and "  " (2-space) indent.
// No trailing newline is appended.
//
// A nil events slice is treated as an empty slice: "security_events" renders
// as [] rather than null.
func FormatJSON(result *diff.DiffResult, image1, image2 string, events []security.SecurityEvent) ([]byte, error) {
	// Build changes slice.
	changes := make([]ChangeEntry, 0, len(result.Entries))
	for _, e := range result.Entries {
		ce := ChangeEntry{
			Path:      e.Path,
			Type:      e.Type.String(),
			SizeDelta: e.SizeDelta,
		}
		// Only Modified entries carry a changes flag array.
		if e.Type == diff.Modified {
			flags := collectFlagSlice(e)
			if len(flags) > 0 {
				ce.Changes = flags
			}
		}
		changes = append(changes, ce)
	}

	// Build security events slice. Always non-nil so JSON renders [] not null.
	secEntries := make([]SecurityEntry, 0, len(events))
	for _, ev := range events {
		secEntries = append(secEntries, SecurityEntry{
			Kind:       string(ev.Kind),
			Path:       ev.Path,
			BeforeMode: uint32(ev.Before),
			AfterMode:  uint32(ev.After),
		})
	}

	report := DiffReport{
		Image1: image1,
		Image2: image2,
		Summary: ReportSummary{
			Added:        result.Added,
			Removed:      result.Removed,
			Modified:     result.Modified,
			AddedBytes:   result.AddedBytes,
			RemovedBytes: result.RemovedBytes,
		},
		Changes:        changes,
		SecurityEvents: secEntries,
	}

	return json.MarshalIndent(report, "", "  ")
}

// collectFlagSlice returns a slice of active change flag labels for a Modified
// DiffEntry in canonical order: content, mode, uid, gid, link, type.
func collectFlagSlice(entry *diff.DiffEntry) []string {
	var flags []string
	if entry.ContentChanged {
		flags = append(flags, "content")
	}
	if entry.ModeChanged {
		flags = append(flags, "mode")
	}
	if entry.UIDChanged {
		flags = append(flags, "uid")
	}
	if entry.GIDChanged {
		flags = append(flags, "gid")
	}
	if entry.LinkTargetChanged {
		flags = append(flags, "link")
	}
	if entry.TypeChanged {
		flags = append(flags, "type")
	}
	return flags
}
