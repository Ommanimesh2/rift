// Package output — markdown.go provides FormatMarkdown, which renders a
// DiffResult (plus image names and security events) into a GitHub-Flavored
// Markdown report suitable for PR comments and documentation.
package output

import (
	"fmt"
	"strings"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/security"
)

// FormatMarkdown renders a DiffResult as a GitHub-Flavored Markdown report.
//
// Output structure:
//
//	## Image Diff: `{image1}` → `{image2}`
//
//	### Summary
//	| Category | Count | Size |
//	...
//
//	### Changes
//	| Status | Path | Size Delta | Details |
//	...  (or "(No changes)" when empty)
//
//	### Security Findings ({N})   ← omitted when events is nil or empty
//	| Kind | Path | Mode |
//	...
//
// The output ends with a single trailing newline.
func FormatMarkdown(result *diff.DiffResult, image1, image2 string, events []security.SecurityEvent) string {
	var sb strings.Builder

	// --- Header ---
	fmt.Fprintf(&sb, "## Image Diff: `%s` → `%s`\n", image1, image2)
	sb.WriteByte('\n')

	// --- Summary section ---
	sb.WriteString("### Summary\n")
	sb.WriteByte('\n')
	sb.WriteString("| Category | Count | Size |\n")
	sb.WriteString("|----------|-------|------|\n")

	// Added row: show "+X.Y KB" when AddedBytes > 0, else "—"
	var addedSize string
	if result.AddedBytes > 0 {
		addedSize = "+" + FormatBytes(result.AddedBytes)
	} else {
		addedSize = "—"
	}
	fmt.Fprintf(&sb, "| Added    | %d     | %s |\n", result.Added, addedSize)

	// Removed row: show "-X.Y KB" when RemovedBytes > 0, else "—"
	var removedSize string
	if result.RemovedBytes > 0 {
		removedSize = "-" + FormatBytes(result.RemovedBytes)
	} else {
		removedSize = "—"
	}
	fmt.Fprintf(&sb, "| Removed  | %d     | %s |\n", result.Removed, removedSize)

	fmt.Fprintf(&sb, "| Modified | %d     | — |\n", result.Modified)
	sb.WriteByte('\n')

	// --- Changes section ---
	sb.WriteString("### Changes\n")
	sb.WriteByte('\n')

	if len(result.Entries) == 0 {
		sb.WriteString("(No changes)\n")
	} else {
		sb.WriteString("| Status | Path | Size Delta | Details |\n")
		sb.WriteString("|--------|------|------------|---------|")

		for _, entry := range result.Entries {
			var status, details string
			switch entry.Type {
			case diff.Added:
				status = "`+`"
				details = ""
			case diff.Removed:
				status = "`-`"
				details = ""
			case diff.Modified:
				status = "`~`"
				details = collectFlags(entry)
			}
			delta := FormatSizeDelta(entry.SizeDelta)
			fmt.Fprintf(&sb, "\n| %s | `%s` | %s | %s |", status, entry.Path, delta, details)
		}
		sb.WriteByte('\n')
	}

	// --- Security Findings section (omitted when no events) ---
	if len(events) > 0 {
		sb.WriteByte('\n')
		fmt.Fprintf(&sb, "### Security Findings (%d)\n", len(events))
		sb.WriteByte('\n')
		sb.WriteString("| Kind | Path | Mode |\n")
		sb.WriteString("|------|------|------|\n")

		for _, event := range events {
			label := securityKindLabel(event.Kind)
			var modeCol string
			if event.Before != 0 {
				modeCol = fmt.Sprintf("`0%04o → 0%04o`", event.Before, event.After)
			}
			fmt.Fprintf(&sb, "| %s | `%s` | %s |\n", label, event.Path, modeCol)
		}
	}

	// Ensure the output ends with exactly one trailing newline.
	out := sb.String()
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	// Remove any double trailing newlines introduced during construction.
	for strings.HasSuffix(out, "\n\n") {
		out = out[:len(out)-1]
	}
	return out
}
