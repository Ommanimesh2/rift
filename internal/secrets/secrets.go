// Package secrets detects potential secrets in container image files.
package secrets

import (
	"regexp"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/security"
)

// SecretFinding represents a detected secret in an image file.
type SecretFinding struct {
	Kind   SecretKind
	Path   string
	Detail string
}

// SecretKind classifies the type of secret found.
type SecretKind string

const (
	KindPrivateKey   SecretKind = "private_key"
	KindAWSAccessKey SecretKind = "aws_access_key"
	KindAPIToken     SecretKind = "generic_api_token"
	KindSecretFile   SecretKind = "secret_file_path"
)

// Known secret file paths and patterns.
var secretPathPatterns = []string{
	"**/.env",
	"**/.env.*",
	"**/*.pem",
	"**/*.key",
	"**/id_rsa",
	"**/id_ed25519",
	"**/id_ecdsa",
	"**/id_dsa",
	"**/.ssh/authorized_keys",
	"**/.ssh/known_hosts",
	"**/.ssh/config",
	"**/credentials.json",
	"**/service-account*.json",
	"**/.aws/credentials",
	"**/.aws/config",
	"**/.npmrc",
	"**/.pypirc",
	"**/.docker/config.json",
	"**/.gitconfig",
	"**/.netrc",
	"**/htpasswd",
	"**/*.p12",
	"**/*.pfx",
	"**/*.jks",
}

// Content patterns compiled at init.
var contentPatterns []contentRule

type contentRule struct {
	kind    SecretKind
	detail  string
	pattern *regexp.Regexp
}

func init() {
	contentPatterns = []contentRule{
		{KindPrivateKey, "Private key header", regexp.MustCompile(`-----BEGIN\s+(RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----`)},
		{KindAWSAccessKey, "AWS access key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
		{KindAPIToken, "Generic API token/secret", regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?secret|access[_-]?token|auth[_-]?token|secret[_-]?key)\s*[:=]\s*['"]?[A-Za-z0-9/+=_-]{20,}`)},
	}
}

// AnalyzePaths checks added/modified file paths against known secret file patterns.
// This is a pure function — no I/O, no file content needed.
func AnalyzePaths(result *diff.DiffResult) []SecretFinding {
	var findings []SecretFinding

	for _, entry := range result.Entries {
		if entry.Type == diff.Removed {
			continue
		}
		for _, pattern := range secretPathPatterns {
			matched, _ := doublestar.Match(pattern, entry.Path)
			if matched {
				findings = append(findings, SecretFinding{
					Kind:   KindSecretFile,
					Path:   entry.Path,
					Detail: "matches known secret file pattern",
				})
				break // one finding per path
			}
		}
	}

	return findings
}

// AnalyzeContent scans file content for secret patterns.
// Returns findings for a single file's content.
func AnalyzeContent(path string, data []byte) []SecretFinding {
	var findings []SecretFinding

	for _, rule := range contentPatterns {
		if rule.pattern.Match(data) {
			findings = append(findings, SecretFinding{
				Kind:   rule.kind,
				Path:   path,
				Detail: rule.detail,
			})
		}
	}

	return findings
}

// ToSecurityEvents converts SecretFindings to SecurityEvents for unified output.
func ToSecurityEvents(findings []SecretFinding) []security.SecurityEvent {
	events := make([]security.SecurityEvent, 0, len(findings))
	for _, f := range findings {
		var kind security.SecurityEventKind
		switch f.Kind {
		case KindPrivateKey:
			kind = security.KindSecretPrivateKey
		case KindAWSAccessKey:
			kind = security.KindSecretAWSKey
		case KindAPIToken:
			kind = security.KindSecretAPIToken
		case KindSecretFile:
			kind = security.KindSecretFilePath
		default:
			kind = security.KindSecretFilePath
		}
		events = append(events, security.SecurityEvent{
			Kind: kind,
			Path: f.Path,
		})
	}
	return events
}
