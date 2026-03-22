// Package config provides configuration file loading for rift.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the rift configuration file structure.
// All fields are optional; CLI flags override config values.
type Config struct {
	Format        string   `yaml:"format,omitempty"`
	SecurityOnly  *bool    `yaml:"security-only,omitempty"`
	Platform      string   `yaml:"platform,omitempty"`
	ExitCode      *bool    `yaml:"exit-code,omitempty"`
	FailSecurity  *bool    `yaml:"fail-on-security,omitempty"`
	SizeThreshold string   `yaml:"size-threshold,omitempty"`
	Include       []string `yaml:"include,omitempty"`
	Exclude       []string `yaml:"exclude,omitempty"`
	Verbose       *bool    `yaml:"verbose,omitempty"`
	ContentDiff   *bool    `yaml:"content-diff,omitempty"`
}

// Load reads the rift config file from standard locations.
// Search order: .rift.yml in cwd, then ~/.config/rift/config.yml.
// Returns an empty Config (not an error) if no config file exists.
func Load() (*Config, error) {
	paths := configPaths()

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	return &Config{}, nil
}

// configPaths returns the ordered list of config file paths to search.
func configPaths() []string {
	paths := []string{".rift.yml", ".rift.yaml"}

	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths,
			filepath.Join(home, ".config", "rift", "config.yml"),
			filepath.Join(home, ".config", "rift", "config.yaml"),
		)
	}

	return paths
}

// DefaultTemplate returns a commented YAML template for rift init.
func DefaultTemplate() string {
	return `# rift configuration file
# CLI flags override these values.

# Output format: terminal, json, markdown, sarif
# format: terminal

# Show only security-relevant changes
# security-only: false

# Target platform for multi-arch images
# platform: linux/amd64

# Show unified diff for modified text files
# content-diff: false

# Glob patterns to include (supports **)
# include:
#   - "etc/**"
#   - "usr/bin/**"

# Glob patterns to exclude (supports **)
# exclude:
#   - "var/cache/**"
#   - "**/*.pyc"

# Enable verbose logging to stderr
# verbose: false

# CI/CD: exit 2 if any file changes found
# exit-code: false

# CI/CD: exit 2 if security events detected
# fail-on-security: false

# CI/CD: exit 2 if net size increase exceeds threshold
# size-threshold: ""
`
}
