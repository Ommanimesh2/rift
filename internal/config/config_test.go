package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoFile(t *testing.T) {
	// Run in a temp dir with no config file
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != "" {
		t.Errorf("expected empty format, got %q", cfg.Format)
	}
	if cfg.Include != nil {
		t.Errorf("expected nil include, got %v", cfg.Include)
	}
}

func TestLoad_WithFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	content := `
format: json
security-only: true
platform: linux/arm64
include:
  - "etc/**"
exclude:
  - "var/cache/**"
  - "**/*.pyc"
size-threshold: 10MB
verbose: true
content-diff: true
`
	os.WriteFile(".rift.yml", []byte(content), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Format != "json" {
		t.Errorf("expected format=json, got %q", cfg.Format)
	}
	if cfg.SecurityOnly == nil || !*cfg.SecurityOnly {
		t.Error("expected security-only=true")
	}
	if cfg.Platform != "linux/arm64" {
		t.Errorf("expected platform=linux/arm64, got %q", cfg.Platform)
	}
	if len(cfg.Include) != 1 || cfg.Include[0] != "etc/**" {
		t.Errorf("expected include=[etc/**], got %v", cfg.Include)
	}
	if len(cfg.Exclude) != 2 {
		t.Errorf("expected 2 excludes, got %d", len(cfg.Exclude))
	}
	if cfg.SizeThreshold != "10MB" {
		t.Errorf("expected size-threshold=10MB, got %q", cfg.SizeThreshold)
	}
	if cfg.Verbose == nil || !*cfg.Verbose {
		t.Error("expected verbose=true")
	}
	if cfg.ContentDiff == nil || !*cfg.ContentDiff {
		t.Error("expected content-diff=true")
	}
}

func TestLoad_YamlExtension(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile(".rift.yaml", []byte("format: markdown\n"), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != "markdown" {
		t.Errorf("expected format=markdown, got %q", cfg.Format)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile(".rift.yml", []byte("{{invalid yaml"), 0o644)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestDefaultTemplate(t *testing.T) {
	tmpl := DefaultTemplate()
	if len(tmpl) == 0 {
		t.Error("expected non-empty template")
	}
	// Should contain commented-out examples
	if !contains(tmpl, "# format:") {
		t.Error("expected template to contain '# format:'")
	}
	if !contains(tmpl, "# exclude:") {
		t.Error("expected template to contain '# exclude:'")
	}
}

func TestConfigPaths(t *testing.T) {
	paths := configPaths()
	if len(paths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(paths))
	}
	if paths[0] != ".rift.yml" {
		t.Errorf("expected first path .rift.yml, got %s", paths[0])
	}

	home, err := os.UserHomeDir()
	if err == nil {
		expected := filepath.Join(home, ".config", "rift", "config.yml")
		found := false
		for _, p := range paths {
			if p == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s in config paths", expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
