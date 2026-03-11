// Package source provides image access abstraction for remote registries,
// Docker daemon, and OCI tarball archives. All sources return a unified
// v1.Image interface for downstream consumption.
package source

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// SourceType classifies the origin of a container image reference.
type SourceType int

const (
	// SourceRemote indicates a remote registry image (e.g., "nginx:latest").
	SourceRemote SourceType = iota
	// SourceDaemon indicates a local Docker daemon image (e.g., "daemon://nginx:latest").
	SourceDaemon
	// SourceTarball indicates an OCI tarball archive (e.g., "./image.tar").
	SourceTarball
)

// String returns the human-readable name of the source type.
func (s SourceType) String() string {
	switch s {
	case SourceRemote:
		return "remote"
	case SourceDaemon:
		return "daemon"
	case SourceTarball:
		return "tarball"
	default:
		return "unknown"
	}
}

// Options configures image source behavior.
type Options struct {
	// Platform specifies the target platform for multi-arch images (e.g., "linux/amd64").
	// Empty string means use the default platform.
	Platform string
}

// DetectSourceType classifies an image reference string into one of the
// supported source types based on prefix, file existence, and extension.
func DetectSourceType(ref string) SourceType {
	// daemon:// prefix → Docker daemon
	if strings.HasPrefix(ref, "daemon://") {
		return SourceDaemon
	}

	// File exists on disk → tarball
	if _, err := os.Stat(ref); err == nil {
		return SourceTarball
	}

	// Known tarball extensions → tarball (even if file doesn't exist yet)
	lower := strings.ToLower(ref)
	if strings.HasSuffix(lower, ".tar") ||
		strings.HasSuffix(lower, ".tar.gz") ||
		strings.HasSuffix(lower, ".tgz") {
		return SourceTarball
	}

	// Default → remote registry
	return SourceRemote
}

// parsePlatform parses a platform string in "os/arch" format into a v1.Platform.
// An empty string returns a zero Platform (use default).
func parsePlatform(s string) (v1.Platform, error) {
	if s == "" {
		return v1.Platform{}, nil
	}

	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return v1.Platform{}, fmt.Errorf("invalid platform %q: expected os/arch (e.g., linux/amd64)", s)
	}

	return v1.Platform{OS: parts[0], Architecture: parts[1]}, nil
}

// Open resolves an image reference and returns a v1.Image from the appropriate
// source. The source type is auto-detected from the reference string.
// Images are loaded lazily — layers are not downloaded until accessed.
func Open(ref string, opts Options) (v1.Image, error) {
	sourceType := DetectSourceType(ref)

	var img v1.Image
	var err error

	switch sourceType {
	case SourceRemote:
		img, err = openRemote(ref, opts)
	case SourceDaemon:
		img, err = openDaemon(ref, opts)
	case SourceTarball:
		img, err = openTarball(ref, opts)
	default:
		return nil, fmt.Errorf("unsupported source type for %q", ref)
	}

	if err != nil {
		return nil, fmt.Errorf("open %s image %q: %w", sourceType, ref, err)
	}

	return img, nil
}
