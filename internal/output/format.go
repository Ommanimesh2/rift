// Package output provides text formatting and terminal rendering functions
// for DiffResult output. FormatEntry/FormatSummary/Render produce plain-text
// structured output; terminal.go wraps these with lipgloss color styles.
package output

import (
	"fmt"
	"strings"

	"github.com/ommmishra/imgdiff/internal/diff"
)

// FormatBytes converts a byte count to a human-readable string using binary
// SI suffixes (KB, MB, GB). Values below 1 KB are rendered as "N bytes".
func FormatBytes(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d bytes", b)
	}
}

// FormatSizeDelta formats a signed byte delta: positive values get a "+"
// prefix, negative values keep their "-" sign, and zero is "0 bytes".
func FormatSizeDelta(delta int64) string {
	if delta == 0 {
		return "0 bytes"
	}
	if delta > 0 {
		return "+" + FormatBytes(delta)
	}
	// delta < 0: FormatBytes takes the absolute value, so re-add sign.
	return "-" + FormatBytes(-delta)
}

// FormatEntry renders a single DiffEntry as a one-line string.
//
// Format by change type:
//   - Added:    "+ {path}  ({delta})"
//   - Removed:  "- {path}  ({delta})"
//   - Modified: "~ {path}  [{flags}]  ({delta})"
//
// Flags for Modified entries are listed in canonical order:
// content, mode, uid, gid, link, type.
func FormatEntry(entry *diff.DiffEntry) string {
	switch entry.Type {
	case diff.Added:
		return fmt.Sprintf("+ %s  (%s)", entry.Path, FormatSizeDelta(entry.SizeDelta))
	case diff.Removed:
		return fmt.Sprintf("- %s  (%s)", entry.Path, FormatSizeDelta(entry.SizeDelta))
	case diff.Modified:
		flags := collectFlags(entry)
		return fmt.Sprintf("~ %s  [%s]  (%s)", entry.Path, flags, FormatSizeDelta(entry.SizeDelta))
	default:
		return fmt.Sprintf("? %s", entry.Path)
	}
}

// collectFlags returns a comma-separated list of active change flag labels
// for a Modified DiffEntry in canonical order: content, mode, uid, gid, link, type.
func collectFlags(entry *diff.DiffEntry) string {
	var parts []string
	if entry.ContentChanged {
		parts = append(parts, "content")
	}
	if entry.ModeChanged {
		parts = append(parts, "mode")
	}
	if entry.UIDChanged {
		parts = append(parts, "uid")
	}
	if entry.GIDChanged {
		parts = append(parts, "gid")
	}
	if entry.LinkTargetChanged {
		parts = append(parts, "link")
	}
	if entry.TypeChanged {
		parts = append(parts, "type")
	}
	return strings.Join(parts, ", ")
}

// FormatSummary renders a one-line summary of a DiffResult.
//
// Format: "{N} added (+{bytes}), {N} removed (-{bytes}), {N} modified"
// Byte parentheticals are omitted when the byte count for that category is 0.
// Modified entries do not contribute to added or removed byte totals.
func FormatSummary(result *diff.DiffResult) string {
	var addedPart, removedPart string

	if result.AddedBytes > 0 {
		addedPart = fmt.Sprintf(" (+%s)", FormatBytes(result.AddedBytes))
	}
	if result.RemovedBytes > 0 {
		removedPart = fmt.Sprintf(" (-%s)", FormatBytes(result.RemovedBytes))
	}

	return fmt.Sprintf("%d added%s, %d removed%s, %d modified",
		result.Added, addedPart,
		result.Removed, removedPart,
		result.Modified)
}

// Render produces a complete plain-text diff output from a DiffResult.
//
// Output format:
//   - One FormatEntry line per changed entry, newline separated
//   - Blank line
//   - FormatSummary line
//
// For an empty DiffResult (no entries), only the summary line is returned.
func Render(result *diff.DiffResult) string {
	var sb strings.Builder

	for _, entry := range result.Entries {
		sb.WriteString(FormatEntry(entry))
		sb.WriteByte('\n')
	}

	if len(result.Entries) > 0 {
		sb.WriteByte('\n')
	}

	sb.WriteString(FormatSummary(result))
	sb.WriteByte('\n')

	return sb.String()
}
