// Package exitcode_test provides tests for the exitcode package.
// Test strategy: TDD — tests written first (RED), then implementation (GREEN).
package exitcode_test

import (
	"testing"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/exitcode"
	"github.com/Ommanimesh2/rift/internal/security"
	"github.com/Ommanimesh2/rift/internal/tree"
)

// --- ParseSizeThreshold tests ---

// TestParseSizeThreshold_Empty verifies empty string returns 0 (disabled).
func TestParseSizeThreshold_Empty(t *testing.T) {
	got, err := exitcode.ParseSizeThreshold("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("ParseSizeThreshold(\"\") = %d, want 0", got)
	}
}

// TestParseSizeThreshold_Zero verifies "0" returns 0 (disabled).
func TestParseSizeThreshold_Zero(t *testing.T) {
	got, err := exitcode.ParseSizeThreshold("0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("ParseSizeThreshold(\"0\") = %d, want 0", got)
	}
}

// TestParseSizeThreshold_BareNumber verifies a bare number is treated as bytes.
func TestParseSizeThreshold_BareNumber(t *testing.T) {
	got, err := exitcode.ParseSizeThreshold("100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 100 {
		t.Errorf("ParseSizeThreshold(\"100\") = %d, want 100", got)
	}
}

// TestParseSizeThreshold_ByteSuffix verifies "B" suffix is treated as bytes.
func TestParseSizeThreshold_ByteSuffix(t *testing.T) {
	got, err := exitcode.ParseSizeThreshold("500B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 500 {
		t.Errorf("ParseSizeThreshold(\"500B\") = %d, want 500", got)
	}
}

// TestParseSizeThreshold_KilobyteSuffix verifies "KB" and "K" suffixes.
func TestParseSizeThreshold_KilobyteSuffix(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{"10KB", 10240},
		{"10K", 10240},
	}
	for _, tc := range cases {
		got, err := exitcode.ParseSizeThreshold(tc.input)
		if err != nil {
			t.Fatalf("ParseSizeThreshold(%q) unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParseSizeThreshold(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// TestParseSizeThreshold_MegabyteSuffix verifies "MB" and "M" suffixes.
func TestParseSizeThreshold_MegabyteSuffix(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{"5MB", 5242880},
		{"5M", 5242880},
	}
	for _, tc := range cases {
		got, err := exitcode.ParseSizeThreshold(tc.input)
		if err != nil {
			t.Fatalf("ParseSizeThreshold(%q) unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParseSizeThreshold(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// TestParseSizeThreshold_GigabyteSuffix verifies "GB" and "G" suffixes.
func TestParseSizeThreshold_GigabyteSuffix(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{"1GB", 1073741824},
		{"1G", 1073741824},
	}
	for _, tc := range cases {
		got, err := exitcode.ParseSizeThreshold(tc.input)
		if err != nil {
			t.Fatalf("ParseSizeThreshold(%q) unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParseSizeThreshold(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// TestParseSizeThreshold_DecimalValue verifies decimal values like "1.5MB".
func TestParseSizeThreshold_DecimalValue(t *testing.T) {
	got, err := exitcode.ParseSizeThreshold("1.5MB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1572864 {
		t.Errorf("ParseSizeThreshold(\"1.5MB\") = %d, want 1572864", got)
	}
}

// TestParseSizeThreshold_CaseInsensitive verifies suffix matching is case-insensitive.
func TestParseSizeThreshold_CaseInsensitive(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{"10kb", 10240},
		{"5mb", 5242880},
		{"1gb", 1073741824},
	}
	for _, tc := range cases {
		got, err := exitcode.ParseSizeThreshold(tc.input)
		if err != nil {
			t.Fatalf("ParseSizeThreshold(%q) unexpected error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParseSizeThreshold(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// TestParseSizeThreshold_Negative verifies negative values return an error.
func TestParseSizeThreshold_Negative(t *testing.T) {
	_, err := exitcode.ParseSizeThreshold("-5MB")
	if err == nil {
		t.Error("ParseSizeThreshold(\"-5MB\") expected error, got nil")
	}
}

// TestParseSizeThreshold_UnknownSuffix verifies an unknown suffix returns an error.
func TestParseSizeThreshold_UnknownSuffix(t *testing.T) {
	_, err := exitcode.ParseSizeThreshold("10TB")
	if err == nil {
		t.Error("ParseSizeThreshold(\"10TB\") expected error, got nil")
	}
}

// TestParseSizeThreshold_NonNumeric verifies a non-numeric string returns an error.
func TestParseSizeThreshold_NonNumeric(t *testing.T) {
	_, err := exitcode.ParseSizeThreshold("notanumber")
	if err == nil {
		t.Error("ParseSizeThreshold(\"notanumber\") expected error, got nil")
	}
}

// --- Evaluate tests ---

// makeResult is a helper to build a DiffResult with given added/removed/modified counts
// and corresponding byte totals.
func makeResult(added, removed, modified int, addedBytes, removedBytes int64) *diff.DiffResult {
	r := &diff.DiffResult{
		Added:        added,
		Removed:      removed,
		Modified:     modified,
		AddedBytes:   addedBytes,
		RemovedBytes: removedBytes,
	}
	// Build synthetic Entries so counts and bytes are consistent.
	for i := 0; i < added; i++ {
		size := addedBytes / int64(added)
		r.Entries = append(r.Entries, &diff.DiffEntry{
			Type:  diff.Added,
			Path:  "added/file",
			After: &tree.FileNode{Size: size},
		})
	}
	for i := 0; i < removed; i++ {
		size := removedBytes / int64(removed)
		r.Entries = append(r.Entries, &diff.DiffEntry{
			Type:   diff.Removed,
			Path:   "removed/file",
			Before: &tree.FileNode{Size: size},
		})
	}
	for i := 0; i < modified; i++ {
		r.Entries = append(r.Entries, &diff.DiffEntry{
			Type:   diff.Modified,
			Path:   "modified/file",
			Before: &tree.FileNode{},
			After:  &tree.FileNode{},
		})
	}
	return r
}

// makeSecurity returns a slice of security events for testing.
func makeSecurity(n int) []security.SecurityEvent {
	events := make([]security.SecurityEvent, n)
	for i := range events {
		events[i] = security.SecurityEvent{Kind: "new_suid", Path: "usr/bin/something"}
	}
	return events
}

// TestEvaluate_NoConditions verifies 0 is returned when nothing is triggered.
func TestEvaluate_NoConditions(t *testing.T) {
	result := makeResult(0, 0, 0, 0, 0)
	opts := exitcode.Options{ExitOnChange: false, ExitOnSecurity: false, SizeThreshold: 0}
	if code := exitcode.Evaluate(result, nil, opts); code != 0 {
		t.Errorf("Evaluate() = %d, want 0", code)
	}
}

// TestEvaluate_ExitOnChange_WithChanges verifies exit 2 when ExitOnChange=true and changes exist.
func TestEvaluate_ExitOnChange_WithChanges(t *testing.T) {
	result := makeResult(1, 0, 0, 100, 0)
	opts := exitcode.Options{ExitOnChange: true}
	if code := exitcode.Evaluate(result, nil, opts); code != 2 {
		t.Errorf("Evaluate() = %d, want 2", code)
	}
}

// TestEvaluate_ExitOnChange_NoChanges verifies exit 0 when ExitOnChange=true but no changes.
func TestEvaluate_ExitOnChange_NoChanges(t *testing.T) {
	result := makeResult(0, 0, 0, 0, 0)
	opts := exitcode.Options{ExitOnChange: true}
	if code := exitcode.Evaluate(result, nil, opts); code != 0 {
		t.Errorf("Evaluate() = %d, want 0", code)
	}
}

// TestEvaluate_ExitOnChange_False_WithChanges verifies exit 0 when ExitOnChange=false even with changes.
func TestEvaluate_ExitOnChange_False_WithChanges(t *testing.T) {
	result := makeResult(3, 2, 1, 1000, 500)
	opts := exitcode.Options{ExitOnChange: false}
	if code := exitcode.Evaluate(result, nil, opts); code != 0 {
		t.Errorf("Evaluate() = %d, want 0", code)
	}
}

// TestEvaluate_ExitOnSecurity_WithEvents verifies exit 2 when ExitOnSecurity=true and events exist.
func TestEvaluate_ExitOnSecurity_WithEvents(t *testing.T) {
	result := makeResult(0, 0, 0, 0, 0)
	events := makeSecurity(1)
	opts := exitcode.Options{ExitOnSecurity: true}
	if code := exitcode.Evaluate(result, events, opts); code != 2 {
		t.Errorf("Evaluate() = %d, want 2", code)
	}
}

// TestEvaluate_ExitOnSecurity_False verifies exit 0 when ExitOnSecurity=false even with events.
func TestEvaluate_ExitOnSecurity_False(t *testing.T) {
	result := makeResult(0, 0, 0, 0, 0)
	events := makeSecurity(3)
	opts := exitcode.Options{ExitOnSecurity: false}
	if code := exitcode.Evaluate(result, events, opts); code != 0 {
		t.Errorf("Evaluate() = %d, want 0", code)
	}
}

// TestEvaluate_SizeThreshold_Exceeded verifies exit 2 when net delta exceeds threshold.
// net delta = 5MB added - 0 removed = 5MB > 1MB threshold → 2
func TestEvaluate_SizeThreshold_Exceeded(t *testing.T) {
	const fiveMB = int64(5 * 1024 * 1024)
	const oneMB = int64(1 * 1024 * 1024)
	result := makeResult(1, 0, 0, fiveMB, 0)
	opts := exitcode.Options{SizeThreshold: oneMB}
	if code := exitcode.Evaluate(result, nil, opts); code != 2 {
		t.Errorf("Evaluate() = %d, want 2", code)
	}
}

// TestEvaluate_SizeThreshold_UnderThreshold verifies exit 0 when net delta is under threshold.
// net delta = 500KB < 1MB threshold → 0
func TestEvaluate_SizeThreshold_UnderThreshold(t *testing.T) {
	const fiveHundredKB = int64(500 * 1024)
	const oneMB = int64(1 * 1024 * 1024)
	result := makeResult(1, 0, 0, fiveHundredKB, 0)
	opts := exitcode.Options{SizeThreshold: oneMB}
	if code := exitcode.Evaluate(result, nil, opts); code != 0 {
		t.Errorf("Evaluate() = %d, want 0", code)
	}
}

// TestEvaluate_SizeThreshold_Disabled verifies exit 0 when SizeThreshold=0 (disabled).
func TestEvaluate_SizeThreshold_Disabled(t *testing.T) {
	const tenMB = int64(10 * 1024 * 1024)
	result := makeResult(1, 0, 0, tenMB, 0)
	opts := exitcode.Options{SizeThreshold: 0}
	if code := exitcode.Evaluate(result, nil, opts); code != 0 {
		t.Errorf("Evaluate() = %d, want 0", code)
	}
}

// TestEvaluate_AllOptionsTrue_NoChanges verifies exit 0 when all options true but no triggers.
func TestEvaluate_AllOptionsTrue_NoChanges(t *testing.T) {
	result := makeResult(0, 0, 0, 0, 0)
	opts := exitcode.Options{
		ExitOnChange:   true,
		ExitOnSecurity: true,
		SizeThreshold:  1024,
	}
	if code := exitcode.Evaluate(result, nil, opts); code != 0 {
		t.Errorf("Evaluate() = %d, want 0", code)
	}
}

// TestEvaluate_SizeThreshold_NegativeDelta verifies exit 0 when image shrank (negative net delta).
// Shrinking image should not trigger size threshold — only growth triggers.
func TestEvaluate_SizeThreshold_NegativeDelta(t *testing.T) {
	const oneMB = int64(1 * 1024 * 1024)
	const fiveMB = int64(5 * 1024 * 1024)
	// addedBytes=0, removedBytes=5MB → net delta = -5MB (image shrank)
	result := makeResult(0, 1, 0, 0, fiveMB)
	opts := exitcode.Options{SizeThreshold: oneMB}
	if code := exitcode.Evaluate(result, nil, opts); code != 0 {
		t.Errorf("Evaluate() with shrinking image = %d, want 0", code)
	}
}
