package secrets

import (
	"testing"

	"github.com/Ommanimesh2/rift/internal/diff"
	"github.com/Ommanimesh2/rift/internal/security"
	"github.com/Ommanimesh2/rift/internal/tree"
)

func TestAnalyzePaths_DetectsSecretFiles(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{
			{Path: "app/.env", Type: diff.Added, After: &tree.FileNode{Size: 100}},
			{Path: "app/main.go", Type: diff.Added, After: &tree.FileNode{Size: 500}},
			{Path: "home/user/.ssh/id_rsa", Type: diff.Added, After: &tree.FileNode{Size: 200}},
			{Path: "etc/ssl/cert.pem", Type: diff.Modified, Before: &tree.FileNode{Size: 100}, After: &tree.FileNode{Size: 110}},
		},
	}

	findings := AnalyzePaths(result)

	paths := make(map[string]bool)
	for _, f := range findings {
		paths[f.Path] = true
	}

	if !paths["app/.env"] {
		t.Error("expected .env to be detected")
	}
	if paths["app/main.go"] {
		t.Error("did not expect main.go to be detected")
	}
	if !paths["home/user/.ssh/id_rsa"] {
		t.Error("expected id_rsa to be detected")
	}
	if !paths["etc/ssl/cert.pem"] {
		t.Error("expected .pem to be detected")
	}
}

func TestAnalyzePaths_IgnoresRemoved(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{
			{Path: "app/.env", Type: diff.Removed, Before: &tree.FileNode{Size: 100}},
		},
	}

	findings := AnalyzePaths(result)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for removed files, got %d", len(findings))
	}
}

func TestAnalyzePaths_NoSecrets(t *testing.T) {
	result := &diff.DiffResult{
		Entries: []*diff.DiffEntry{
			{Path: "usr/bin/app", Type: diff.Added, After: &tree.FileNode{Size: 100}},
			{Path: "etc/nginx/nginx.conf", Type: diff.Modified, Before: &tree.FileNode{Size: 50}, After: &tree.FileNode{Size: 55}},
		},
	}

	findings := AnalyzePaths(result)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestAnalyzeContent_PrivateKey(t *testing.T) {
	content := []byte(`some config
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWep4PAtGoLFt0Y
-----END RSA PRIVATE KEY-----
`)
	findings := AnalyzeContent("app/key.pem", content)
	if len(findings) == 0 {
		t.Error("expected private key finding")
	}
	if findings[0].Kind != KindPrivateKey {
		t.Errorf("expected KindPrivateKey, got %s", findings[0].Kind)
	}
}

func TestAnalyzeContent_AWSKey(t *testing.T) {
	content := []byte(`AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE`)
	findings := AnalyzeContent(".env", content)

	hasAWS := false
	for _, f := range findings {
		if f.Kind == KindAWSAccessKey {
			hasAWS = true
		}
	}
	if !hasAWS {
		t.Error("expected AWS key finding")
	}
}

func TestAnalyzeContent_GenericToken(t *testing.T) {
	content := []byte(`api_key=test_token_abcdefghijklmnopqrstuvwxyz1234567890`)
	findings := AnalyzeContent("config.yml", content)

	hasToken := false
	for _, f := range findings {
		if f.Kind == KindAPIToken {
			hasToken = true
		}
	}
	if !hasToken {
		t.Error("expected generic API token finding")
	}
}

func TestAnalyzeContent_NoSecrets(t *testing.T) {
	content := []byte(`server {
    listen 80;
    server_name example.com;
}`)
	findings := AnalyzeContent("nginx.conf", content)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestAnalyzeContent_OpenSSHKey(t *testing.T) {
	content := []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAA
-----END OPENSSH PRIVATE KEY-----`)
	findings := AnalyzeContent("id_ed25519", content)
	if len(findings) == 0 {
		t.Error("expected OPENSSH private key finding")
	}
}

func TestToSecurityEvents(t *testing.T) {
	findings := []SecretFinding{
		{Kind: KindPrivateKey, Path: "key.pem", Detail: "Private key"},
		{Kind: KindAWSAccessKey, Path: ".env", Detail: "AWS key"},
		{Kind: KindSecretFile, Path: "credentials.json", Detail: "secret path"},
	}

	events := ToSecurityEvents(findings)
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].Kind != security.KindSecretPrivateKey {
		t.Errorf("expected KindSecretPrivateKey, got %s", events[0].Kind)
	}
	if events[1].Kind != security.KindSecretAWSKey {
		t.Errorf("expected KindSecretAWSKey, got %s", events[1].Kind)
	}
	if events[2].Kind != security.KindSecretFilePath {
		t.Errorf("expected KindSecretFilePath, got %s", events[2].Kind)
	}
}
