package source

import (
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
)

// openDaemon reads an image from the local Docker daemon.
func openDaemon(ref string, opts Options) (v1.Image, error) {
	// Strip "daemon://" prefix
	stripped := strings.TrimPrefix(ref, "daemon://")

	tag, err := name.NewTag(stripped)
	if err != nil {
		return nil, err
	}

	return daemon.Image(tag)
}
