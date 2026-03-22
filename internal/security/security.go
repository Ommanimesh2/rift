// Package security provides security-relevant analysis of a diff.DiffResult.
// It detects changes that could indicate privilege escalation or reduced
// filesystem security, such as new SUID/SGID binaries, new executables, and
// world-writable files.
//
// The Analyze function is a pure function: no I/O, no global state.
package security

import (
	"os"

	"github.com/Ommanimesh2/rift/internal/diff"
)

// Permission bit constants for clarity and avoiding magic numbers.
const (
	ModeSetUID os.FileMode = 0o4000 // SUID bit
	ModeSetGID os.FileMode = 0o2000 // SGID bit
	ModeExec   os.FileMode = 0o111  // any execute bit (owner, group, or other)
	ModeWorldW os.FileMode = 0o002  // world-writable bit
)

// SecurityEventKind classifies the type of security-relevant change detected.
type SecurityEventKind string

const (
	// KindNewSUID is emitted when an Added file has the SUID bit set.
	KindNewSUID SecurityEventKind = "new_suid"

	// KindNewSGID is emitted when an Added file has the SGID bit set.
	KindNewSGID SecurityEventKind = "new_sgid"

	// KindSUIDAdded is emitted when a Modified file gains the SUID bit.
	KindSUIDAdded SecurityEventKind = "suid_added"

	// KindSGIDAdded is emitted when a Modified file gains the SGID bit.
	KindSGIDAdded SecurityEventKind = "sgid_added"

	// KindNewExecutable is emitted when an Added non-directory file has any execute bit set.
	KindNewExecutable SecurityEventKind = "new_executable"

	// KindWorldWritable is emitted when an Added or Modified non-directory file is world-writable.
	KindWorldWritable SecurityEventKind = "world_writable"

	// KindPermEscalation is emitted when a Modified file's permission bits are strictly more permissive.
	KindPermEscalation SecurityEventKind = "perm_escalation"
)

// SecurityEvent describes a single security-relevant change detected by Analyze.
type SecurityEvent struct {
	// Kind classifies the type of security event.
	Kind SecurityEventKind

	// Path is the filesystem path of the affected entry.
	Path string

	// Before holds the file mode of the entry before the change.
	// This is 0 for Added entries (no Before node).
	Before os.FileMode

	// After holds the file mode of the entry after the change.
	After os.FileMode
}

// Analyze inspects a *diff.DiffResult and returns all security-relevant events
// found within it. Events are returned in the same order as result.Entries.
// The returned slice is always non-nil (may be empty).
//
// Detection rules:
//   - Added files: SUID, SGID, executable, world-writable bits are flagged.
//   - Modified files: newly gained SUID/SGID, world-writable, and permission escalation.
//   - Removed files: no events (removing files is not a security concern).
//   - Directories are excluded from new_executable and world_writable checks.
func Analyze(result *diff.DiffResult) []SecurityEvent {
	events := make([]SecurityEvent, 0)

	for _, entry := range result.Entries {
		switch entry.Type {
		case diff.Added:
			after := entry.After
			mode := after.Mode

			if mode&ModeSetUID != 0 {
				events = append(events, SecurityEvent{
					Kind:  KindNewSUID,
					Path:  entry.Path,
					After: mode,
				})
			}
			if mode&ModeSetGID != 0 {
				events = append(events, SecurityEvent{
					Kind:  KindNewSGID,
					Path:  entry.Path,
					After: mode,
				})
			}
			if mode&ModeExec != 0 && !after.IsDir {
				events = append(events, SecurityEvent{
					Kind:  KindNewExecutable,
					Path:  entry.Path,
					After: mode,
				})
			}
			if mode&ModeWorldW != 0 && !after.IsDir {
				events = append(events, SecurityEvent{
					Kind:  KindWorldWritable,
					Path:  entry.Path,
					After: mode,
				})
			}

		case diff.Modified:
			before := entry.Before
			after := entry.After
			beforeMode := before.Mode
			afterMode := after.Mode

			if beforeMode&ModeSetUID == 0 && afterMode&ModeSetUID != 0 {
				events = append(events, SecurityEvent{
					Kind:   KindSUIDAdded,
					Path:   entry.Path,
					Before: beforeMode,
					After:  afterMode,
				})
			}
			if beforeMode&ModeSetGID == 0 && afterMode&ModeSetGID != 0 {
				events = append(events, SecurityEvent{
					Kind:   KindSGIDAdded,
					Path:   entry.Path,
					Before: beforeMode,
					After:  afterMode,
				})
			}
			if afterMode&ModeWorldW != 0 && !after.IsDir {
				events = append(events, SecurityEvent{
					Kind:   KindWorldWritable,
					Path:   entry.Path,
					Before: beforeMode,
					After:  afterMode,
				})
			}
			if afterMode&0o7777 > beforeMode&0o7777 {
				events = append(events, SecurityEvent{
					Kind:   KindPermEscalation,
					Path:   entry.Path,
					Before: beforeMode,
					After:  afterMode,
				})
			}

		case diff.Removed:
			// Removing files is not a security risk — no events emitted.
		}
	}

	return events
}
