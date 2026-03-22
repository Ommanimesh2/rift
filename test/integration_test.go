//go:build integration

package test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// riftBin is the path to the rift binary. Build before running:
//   go build -o rift . && go test -tags integration ./test/...
const riftBin = "../rift"

// Use small, stable, pinned images for deterministic tests.
const (
	alpine318 = "alpine:3.18"
	alpine319 = "alpine:3.19"
)

func runRift(t *testing.T, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(riftBin, args...)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run rift: %v", err)
		}
	}
	return string(out), exitCode
}

func TestIntegration_BasicComparison(t *testing.T) {
	output, code := runRift(t, alpine318, alpine319)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "added") || !strings.Contains(output, "removed") || !strings.Contains(output, "modified") {
		t.Errorf("expected diff summary in output, got: %s", output[:min(len(output), 200)])
	}
}

func TestIntegration_JSONOutput(t *testing.T) {
	output, code := runRift(t, "--format", "json", alpine318, alpine319)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nOutput: %s", err, output[:min(len(output), 500)])
	}

	if _, ok := result["summary"]; !ok {
		t.Error("expected 'summary' key in JSON output")
	}
	if _, ok := result["changes"]; !ok {
		t.Error("expected 'changes' key in JSON output")
	}
	if _, ok := result["security_events"]; !ok {
		t.Error("expected 'security_events' key in JSON output")
	}
}

func TestIntegration_MarkdownOutput(t *testing.T) {
	output, code := runRift(t, "--format", "markdown", alpine318, alpine319)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "## Image Diff") {
		t.Error("expected markdown header in output")
	}
}

func TestIntegration_QuickMode(t *testing.T) {
	output, code := runRift(t, "--quick", alpine318, alpine319)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if len(output) == 0 {
		t.Error("expected non-empty quick mode output")
	}
}

func TestIntegration_ExitCodeOnChanges(t *testing.T) {
	_, code := runRift(t, "--exit-code", alpine318, alpine319)
	if code != 2 {
		t.Errorf("expected exit code 2 (changes detected), got %d", code)
	}
}

func TestIntegration_SameImage(t *testing.T) {
	_, code := runRift(t, "--exit-code", alpine318, alpine318)
	if code != 0 {
		t.Errorf("expected exit code 0 for identical images, got %d", code)
	}
}

func TestIntegration_SecurityOnly(t *testing.T) {
	output, code := runRift(t, "--security-only", alpine318, alpine319)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	// Should either show findings or "No security findings."
	if len(output) == 0 {
		t.Error("expected non-empty security output")
	}
}

func TestIntegration_VerboseMode(t *testing.T) {
	output, code := runRift(t, "-v", alpine318, alpine319)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	// Verbose output goes to stderr which is captured by CombinedOutput
	if !strings.Contains(output, "Opening image") {
		t.Error("expected verbose step messages in output")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
