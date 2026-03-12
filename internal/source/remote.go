package source

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// openRemote pulls an image from a remote container registry.
// When opts.Username is non-empty, explicit Basic auth credentials are used.
// Otherwise, the Docker DefaultKeychain (reads ~/.docker/config.json) is used.
func openRemote(ref string, opts Options) (v1.Image, error) {
	parsedRef, err := name.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	var remoteOpts []remote.Option
	if opts.Username != "" {
		remoteOpts = append(remoteOpts, remote.WithAuth(&authn.Basic{
			Username: opts.Username,
			Password: opts.Password,
		}))
	} else {
		remoteOpts = append(remoteOpts, remote.WithAuthFromKeychain(authn.DefaultKeychain))
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
