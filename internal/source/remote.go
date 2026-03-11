package source

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// openRemote pulls an image from a remote container registry.
func openRemote(ref string, opts Options) (v1.Image, error) {
	parsedRef, err := name.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	remoteOpts := []remote.Option{
		remote.WithAuthFromKeychain(authn.DefaultKeychain),
	}

	if opts.Platform != "" {
		platform, err := parsePlatform(opts.Platform)
		if err != nil {
			return nil, err
		}
		remoteOpts = append(remoteOpts, remote.WithPlatform(platform))
	}

	return remote.Image(parsedRef, remoteOpts...)
}
