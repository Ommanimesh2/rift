package content

import (
	"fmt"
	"strings"

	difflib "github.com/sergi/go-diff/diffmatchpatch"
)

// UnifiedDiff generates a unified diff between two text contents.
// Returns an empty string if the contents are identical.
func UnifiedDiff(oldContent, newContent []byte, oldName, newName string) string {
	oldText := string(oldContent)
	newText := string(newContent)

	if oldText == newText {
		return ""
	}

	dmp := difflib.New()
	a, b, c := dmp.DiffLinesToChars(oldText, newText)
	diffs := dmp.DiffMain(a, b, false)
	diffs = dmp.DiffCharsToLines(diffs, c)
	diffs = dmp.DiffCleanupSemantic(diffs)

	return formatUnifiedDiff(diffs, oldName, newName)
}

// formatUnifiedDiff formats diffs in a unified diff style.
func formatUnifiedDiff(diffs []difflib.Diff, oldName, newName string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "--- %s\n", oldName)
	fmt.Fprintf(&sb, "+++ %s\n", newName)

	for _, d := range diffs {
		lines := strings.Split(d.Text, "\n")
		// Remove trailing empty string from split
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		for _, line := range lines {
			switch d.Type {
			case difflib.DiffEqual:
				fmt.Fprintf(&sb, " %s\n", line)
			case difflib.DiffDelete:
				fmt.Fprintf(&sb, "-%s\n", line)
			case difflib.DiffInsert:
				fmt.Fprintf(&sb, "+%s\n", line)
			}
		}
	}

	return sb.String()
}
