package output_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ommmishra/imgdiff/internal/output"
)

// testLayerSummary returns a non-nil LayerSummary for use in FormatQuick tests.
func testLayerSummary() *output.LayerSummary {
	return &output.LayerSummary{
		TotalA:       5,
		TotalB:       6,
		SharedCount:  4,
		SharedBytes:  524288000,
		OnlyInACount: 1,
		OnlyInABytes: 10240,
		OnlyInBCount: 2,
		OnlyInBBytes: 20480,
	}
}

// --- terminal format ---

func TestFormatQuick_Terminal_ContainsImageNames(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "terminal")
	if !strings.Contains(got, "img1") {
		t.Errorf("terminal output missing image1: %q", got)
	}
	if !strings.Contains(got, "img2") {
		t.Errorf("terminal output missing image2: %q", got)
	}
}

func TestFormatQuick_Terminal_ContainsSharedCount(t *testing.T) {
	summary := testLayerSummary()
	got := output.FormatQuick(summary, "img1", "img2", "terminal")
	// FormatLayerSummary includes "Shared:    4 layers"
	if !strings.Contains(got, "4") {
		t.Errorf("terminal output missing shared count: %q", got)
	}
}

func TestFormatQuick_Terminal_ContainsNote(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "terminal")
	if !strings.Contains(got, "--quick") {
		t.Errorf("terminal output missing --quick note: %q", got)
	}
}

// --- empty format treated as terminal ---

func TestFormatQuick_EmptyFormat_TreatedAsTerminal(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "")
	if !strings.Contains(got, "img1") || !strings.Contains(got, "img2") {
		t.Errorf("empty format output should behave like terminal, got: %q", got)
	}
}

// --- json format ---

func TestFormatQuick_JSON_IsValidJSON(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "json")
	var v interface{}
	if err := json.Unmarshal([]byte(got), &v); err != nil {
		t.Errorf("json output is not valid JSON: %v\noutput: %q", err, got)
	}
}

func TestFormatQuick_JSON_Image1Field(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "json")
	var report output.QuickReport
	if err := json.Unmarshal([]byte(got), &report); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if report.Image1 != "img1" {
		t.Errorf("image1 = %q, want %q", report.Image1, "img1")
	}
}

func TestFormatQuick_JSON_ModeField(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "json")
	var report output.QuickReport
	if err := json.Unmarshal([]byte(got), &report); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if report.Mode != "quick" {
		t.Errorf("mode = %q, want %q", report.Mode, "quick")
	}
}

func TestFormatQuick_JSON_LayersSharedField(t *testing.T) {
	summary := testLayerSummary()
	got := output.FormatQuick(summary, "img1", "img2", "json")
	var report output.QuickReport
	if err := json.Unmarshal([]byte(got), &report); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if report.Layers.Shared != summary.SharedCount {
		t.Errorf("layers.shared = %d, want %d", report.Layers.Shared, summary.SharedCount)
	}
}

func TestFormatQuick_JSON_TrailingNewline(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "json")
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("json output should end with newline, got: %q", got)
	}
}

// --- markdown format ---

func TestFormatQuick_Markdown_ContainsHeader(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "markdown")
	if !strings.Contains(got, "## Quick Image Comparison") {
		t.Errorf("markdown output missing header: %q", got)
	}
}

func TestFormatQuick_Markdown_ContainsImageNames(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "markdown")
	if !strings.Contains(got, "img1") {
		t.Errorf("markdown output missing image1: %q", got)
	}
	if !strings.Contains(got, "img2") {
		t.Errorf("markdown output missing image2: %q", got)
	}
}

func TestFormatQuick_Markdown_ContainsBlockquote(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "markdown")
	if !strings.Contains(got, "> ") {
		t.Errorf("markdown output missing blockquote: %q", got)
	}
}

// --- nil summary handling ---

func TestFormatQuick_NilSummary_Terminal_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FormatQuick(nil, ..., terminal) panicked: %v", r)
		}
	}()
	got := output.FormatQuick(nil, "img1", "img2", "terminal")
	if got == "" {
		t.Error("expected non-empty output for nil summary terminal")
	}
}

func TestFormatQuick_NilSummary_JSON_DoesNotPanicAndIsValidJSON(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FormatQuick(nil, ..., json) panicked: %v", r)
		}
	}()
	got := output.FormatQuick(nil, "img1", "img2", "json")
	var v interface{}
	if err := json.Unmarshal([]byte(got), &v); err != nil {
		t.Errorf("nil summary json output is not valid JSON: %v\noutput: %q", err, got)
	}
}

func TestFormatQuick_NilSummary_Markdown_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FormatQuick(nil, ..., markdown) panicked: %v", r)
		}
	}()
	got := output.FormatQuick(nil, "img1", "img2", "markdown")
	if got == "" {
		t.Error("expected non-empty output for nil summary markdown")
	}
}

// --- unknown format ---

func TestFormatQuick_UnknownFormat_ReturnsErrorString(t *testing.T) {
	got := output.FormatQuick(testLayerSummary(), "img1", "img2", "unknown")
	if !strings.Contains(got, "unsupported format") {
		t.Errorf("unknown format should return error string, got: %q", got)
	}
	if !strings.Contains(got, "unknown") {
		t.Errorf("error string should contain the format name, got: %q", got)
	}
}
