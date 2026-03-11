package source

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// openTarball reads an image from an OCI tarball archive file.
func openTarball(ref string, opts Options) (v1.Image, error) {
	// nil tag selects the first (and usually only) image in the tarball
	return tarball.ImageFromPath(ref, nil)
}
