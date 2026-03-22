// Package content extracts and compares file content from container images.
package content

import (
	"archive/tar"
	"bytes"
	"io"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const maxTextFileSize = 1 << 20 // 1 MB

// ExtractFile extracts a single file's content from an image by path.
// It reads layers top-to-bottom and returns the first match.
// Returns nil if the file is not found.
func ExtractFile(img v1.Image, path string) ([]byte, error) {
	layers, err := img.Layers()
	if err != nil {
		return nil, err
	}

	// Read layers top-to-bottom (last layer wins).
	for i := len(layers) - 1; i >= 0; i-- {
		content, found, err := extractFromLayer(layers[i], path)
		if err != nil {
			return nil, err
		}
		if found {
			return content, nil
		}
	}

	return nil, nil
}

// extractFromLayer looks for a file in a single layer tar.
func extractFromLayer(layer v1.Layer, path string) ([]byte, bool, error) {
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, false, err
	}
	defer rc.Close()

	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, false, err
		}

		normalized := normalizePath(hdr.Name)
		if normalized == path {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tr); err != nil {
				return nil, false, err
			}
			return buf.Bytes(), true, nil
		}
	}

	return nil, false, nil
}

// IsText returns true if the content appears to be text (no null bytes in the
// first 512 bytes).
func IsText(data []byte) bool {
	check := data
	if len(check) > 512 {
		check = check[:512]
	}
	return !bytes.Contains(check, []byte{0})
}

// IsDiffable returns true if the file is suitable for content diff
// (text content, under size limit).
func IsDiffable(data []byte) bool {
	return len(data) <= maxTextFileSize && IsText(data)
}

func normalizePath(p string) string {
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")
	return p
}
