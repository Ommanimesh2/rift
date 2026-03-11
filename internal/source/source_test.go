package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSourceType(t *testing.T) {
	// Create a temporary tar file for os.Stat-based detection
	tmpDir := t.TempDir()
	tmpTar := filepath.Join(tmpDir, "image.tar")
	if err := os.WriteFile(tmpTar, []byte("fake tar content"), 0644); err != nil {
		t.Fatalf("failed to create temp tar file: %v", err)
	}

	tests := []struct {
		name string
		ref  string
		want SourceType
	}{
		// Remote registry references
		{name: "short tag", ref: "nginx:latest", want: SourceRemote},
		{name: "fully qualified", ref: "docker.io/library/nginx:1.25", want: SourceRemote},
		{name: "ghcr reference", ref: "ghcr.io/owner/repo:v1.0", want: SourceRemote},

		// Daemon references
		{name: "daemon prefix", ref: "daemon://nginx:latest", want: SourceDaemon},
		{name: "daemon custom image", ref: "daemon://myapp:v2", want: SourceDaemon},

		// Tarball references
		{name: "existing tar file", ref: tmpTar, want: SourceTarball},
		{name: "tar.gz extension", ref: "/tmp/backup.tar.gz", want: SourceTarball},
		{name: "tgz extension", ref: "archive.tgz", want: SourceTarball},
		{name: "tar extension no file", ref: "nonexistent.tar", want: SourceTarball},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectSourceType(tt.ref)
			if got != tt.want {
				t.Errorf("DetectSourceType(%q) = %v, want %v", tt.ref, got, tt.want)
			}
		})
	}
}

func TestSourceTypeString(t *testing.T) {
	tests := []struct {
		st   SourceType
		want string
	}{
		{SourceRemote, "remote"},
		{SourceDaemon, "daemon"},
		{SourceTarball, "tarball"},
		{SourceType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.st.String()
			if got != tt.want {
				t.Errorf("SourceType(%d).String() = %q, want %q", tt.st, got, tt.want)
			}
		})
	}
}

func TestParsePlatform(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOS  string
		wantArch string
		wantErr bool
	}{
		{name: "linux amd64", input: "linux/amd64", wantOS: "linux", wantArch: "amd64"},
		{name: "linux arm64", input: "linux/arm64", wantOS: "linux", wantArch: "arm64"},
		{name: "empty string", input: "", wantOS: "", wantArch: ""},
		{name: "no slash", input: "linux", wantErr: true},
		{name: "too many parts", input: "linux/amd64/v8", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform, err := parsePlatform(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parsePlatform(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parsePlatform(%q) unexpected error: %v", tt.input, err)
				return
			}
			if platform.OS != tt.wantOS {
				t.Errorf("parsePlatform(%q).OS = %q, want %q", tt.input, platform.OS, tt.wantOS)
			}
			if platform.Architecture != tt.wantArch {
				t.Errorf("parsePlatform(%q).Architecture = %q, want %q", tt.input, platform.Architecture, tt.wantArch)
			}
		})
	}
}
