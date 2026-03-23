package output

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Ommanimesh2/rift/internal/security"
)

func TestFormatSARIF_EmptyEvents(t *testing.T) {
	data, err := FormatSARIF(nil, "img1", "img2", "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report SarifReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if report.Version != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %s", report.Version)
	}
	if len(report.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(report.Runs))
	}
	if len(report.Runs[0].Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(report.Runs[0].Results))
	}
	if report.Runs[0].Tool.Driver.Name != "rift" {
		t.Errorf("expected tool name rift, got %s", report.Runs[0].Tool.Driver.Name)
	}
	if len(report.Runs[0].Tool.Driver.Rules) != 11 {
		t.Errorf("expected 11 rules, got %d", len(report.Runs[0].Tool.Driver.Rules))
	}
}

func TestFormatSARIF_WithEvents(t *testing.T) {
	events := []security.SecurityEvent{
		{
			Kind:      security.KindNewSUID,
			Path:      "usr/bin/sudo",
			After: 0o4755,
		},
		{
			Kind:      security.KindWorldWritable,
			Path:      "tmp/scratch",
			After: 0o777,
		},
		{
			Kind:      security.KindPermEscalation,
			Path:      "etc/shadow",
			Before: 0o600,
			After:  0o644,
		},
	}

	data, err := FormatSARIF(events, "alpine:3.18", "alpine:3.19", "v1.3.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report SarifReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	results := report.Runs[0].Results
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Check SUID finding
	if results[0].RuleID != "rift/new_suid" {
		t.Errorf("expected ruleId rift/new_suid, got %s", results[0].RuleID)
	}
	if results[0].Level != "error" {
		t.Errorf("expected level error, got %s", results[0].Level)
	}
	if results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI != "usr/bin/sudo" {
		t.Errorf("expected URI usr/bin/sudo, got %s", results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI)
	}

	// Check world-writable is warning level
	if results[1].Level != "warning" {
		t.Errorf("expected level warning for world_writable, got %s", results[1].Level)
	}

	// Check perm_escalation
	if results[2].RuleID != "rift/perm_escalation" {
		t.Errorf("expected ruleId rift/perm_escalation, got %s", results[2].RuleID)
	}

	// Check version is set
	if report.Runs[0].Tool.Driver.Version != "v1.3.0" {
		t.Errorf("expected version v1.3.0, got %s", report.Runs[0].Tool.Driver.Version)
	}
}

func TestFormatSARIF_ValidSchema(t *testing.T) {
	data, err := FormatSARIF(nil, "a", "b", "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "https://json.schemastore.org/sarif-2.1.0.json") {
		t.Error("expected SARIF schema URL in output")
	}
}

func TestFormatSARIF_MessageContainsImages(t *testing.T) {
	events := []security.SecurityEvent{
		{
			Kind:      security.KindNewExecutable,
			Path:      "usr/local/bin/app",
			After: 0o755,
		},
	}

	data, err := FormatSARIF(events, "myapp:v1", "myapp:v2", "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "myapp:v1") || !strings.Contains(s, "myapp:v2") {
		t.Error("expected image names in result message")
	}
}

func TestFormatSARIF_AllRuleIDs(t *testing.T) {
	events := []security.SecurityEvent{
		{Kind: security.KindNewSUID, Path: "a", After: 0o4755},
		{Kind: security.KindNewSGID, Path: "b", After: 0o2755},
		{Kind: security.KindSUIDAdded, Path: "c", Before: 0o755, After: 0o4755},
		{Kind: security.KindSGIDAdded, Path: "d", Before: 0o755, After: 0o2755},
		{Kind: security.KindNewExecutable, Path: "e", After: 0o755},
		{Kind: security.KindWorldWritable, Path: "f", After: 0o777},
		{Kind: security.KindPermEscalation, Path: "g", Before: 0o600, After: 0o644},
	}

	data, err := FormatSARIF(events, "a", "b", "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report SarifReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(report.Runs[0].Results) != 7 {
		t.Errorf("expected 7 results, got %d", len(report.Runs[0].Results))
	}

	// Check each has the correct level
	errorRules := map[string]bool{
		"rift/new_suid":   true,
		"rift/new_sgid":   true,
		"rift/suid_added": true,
		"rift/sgid_added": true,
	}
	for _, r := range report.Runs[0].Results {
		if errorRules[r.RuleID] && r.Level != "error" {
			t.Errorf("rule %s should be error, got %s", r.RuleID, r.Level)
		}
		if !errorRules[r.RuleID] && r.Level != "warning" {
			t.Errorf("rule %s should be warning, got %s", r.RuleID, r.Level)
		}
	}
}

func TestFormatSARIF_OutputIsValidJSON(t *testing.T) {
	events := []security.SecurityEvent{
		{Kind: security.KindNewSUID, Path: "usr/bin/test", After: 0o4755},
	}

	data, err := FormatSARIF(events, "img1", "img2", "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it can be written and re-parsed
	if !json.Valid(data) {
		t.Error("output is not valid JSON")
	}

	// Verify it's indented (pretty-printed)
	if !strings.Contains(string(data), "\n  ") {
		t.Error("expected indented JSON output")
	}

	_ = os.Stdout // suppress unused import
}
