// Package tree provides in-memory file tree construction from OCI image layers.
// It handles tar parsing, OCI whiteout semantics, and multi-layer squashing
// to produce a single merged filesystem view from a container image.
package tree

import (
	"archive/tar"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// FileNode represents a single filesystem entry extracted from an image layer.
// It captures all metadata needed for file-level diffing and security analysis.
type FileNode struct {
	// Path is the clean, normalized path of the entry (no leading ./ or /).
	Path string

	// Size is the file size in bytes. Directories and symlinks have size 0.
	Size int64

	// Mode contains the file permission bits.
	Mode os.FileMode

	// UID is the numeric user owner.
	UID int

	// GID is the numeric group owner.
	GID int

	// IsDir is true when this entry is a directory.
	IsDir bool

	// LinkTarget holds the target path for symbolic links. Empty for non-symlinks.
	LinkTarget string

	// Digest is the sha256 hash of the file content for regular files.
	// Algorithm is always "sha256". Hex is empty for directories and symlinks.
	Digest v1.Hash

	// LayerIndex is the zero-based index of the layer that last wrote this file.
	// Set by BuildTree; defaults to 0.
	LayerIndex int
}

// normalizePath strips leading "./" and "/" from a tar header name and returns
// the clean relative path. The trailing slash is removed so directories are
// stored without it (matching the key used for lookups).
func normalizePath(name string) string {
	// Strip leading "./" prefix characters and "/" characters.
	name = strings.TrimLeft(name, "./")
	// Also strip any leading slash that remains.
	name = strings.TrimLeft(name, "/")
	// Remove trailing slash (directories are keyed without it).
	name = strings.TrimRight(name, "/")
	return name
}

// ParseLayer reads a single OCI layer and returns a map of normalized path →
// *FileNode containing metadata for every tar entry. The v1.Layer's
// Uncompressed() reader is used so the caller does not need to handle
// decompression.
//
// OCI whiteout entries (.wh.*) are included as-is; callers that build a
// complete squashed tree should use BuildTree instead.
func ParseLayer(layer v1.Layer) (map[string]*FileNode, error) {
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, fmt.Errorf("open layer uncompressed reader: %w", err)
	}
	defer rc.Close()

	result := make(map[string]*FileNode)
	tr := tar.NewReader(rc)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar entry: %w", err)
		}

		clean := normalizePath(hdr.Name)
		if clean == "" {
			// Skip the root "." entry (common in Docker-built layers).
			continue
		}

		node := &FileNode{
			Path: clean,
			Mode: os.FileMode(hdr.Mode),
			UID:  hdr.Uid,
			GID:  hdr.Gid,
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			node.IsDir = true

		case tar.TypeSymlink, tar.TypeLink:
			node.LinkTarget = hdr.Linkname

		default:
			// Regular file (TypeReg, TypeRegA, and any other file-like type).
			node.Size = hdr.Size

			// Compute sha256 digest of content.
			h := sha256.New()
			if _, err := io.Copy(h, tr); err != nil {
				return nil, fmt.Errorf("compute digest for %q: %w", clean, err)
			}
			node.Digest = v1.Hash{
				Algorithm: "sha256",
				Hex:       fmt.Sprintf("%x", h.Sum(nil)),
			}
		}

		result[clean] = node
	}

	return result, nil
}

// FileTree is a squashed, merged filesystem view of a container image.
// It wraps the final set of FileNode entries after all layers have been
// applied in order with whiteout semantics.
type FileTree struct {
	// Entries maps normalized path → *FileNode.
	Entries map[string]*FileNode
}

// Size returns the total number of entries in the tree (files + directories).
func (t *FileTree) Size() int {
	return len(t.Entries)
}

// Get returns the FileNode for the given path, or nil if not present.
func (t *FileTree) Get(path string) *FileNode {
	return t.Entries[path]
}

// String returns a human-readable summary of the tree.
// Format: "{N} files, {M} directories, {S} total bytes"
func (t *FileTree) String() string {
	var files, dirs int
	var totalBytes int64
	for _, n := range t.Entries {
		if n.IsDir {
			dirs++
		} else {
			files++
			totalBytes += n.Size
		}
	}
	return fmt.Sprintf("%d files, %d directories, %s total", files, dirs, formatBytes(totalBytes))
}

// formatBytes converts a byte count to a human-readable string using SI
// suffixes (KB, MB, GB). Values below 1 KB are rendered as "N bytes".
func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d bytes", b)
	}
}

// BuildTree squashes an ordered slice of layers (index 0 = bottom, last = top)
// into a single FileTree, applying OCI whiteout semantics at each layer.
//
// OCI whiteout rules:
//   - ".wh.FILENAME" → delete FILENAME from lower layers
//   - ".wh..wh..opq" in a directory → delete ALL entries under that directory
//     from lower layers, then add current layer's entries for that directory
//
// Whiteout marker entries themselves are never included in the final tree.
func BuildTree(layers []v1.Layer) (*FileTree, error) {
	return BuildTreeWithAttribution(layers, 0)
}

// BuildTreeWithAttribution squashes layers like BuildTree but also sets
// FileNode.LayerIndex to baseIndex + the layer's position in the slice.
func BuildTreeWithAttribution(layers []v1.Layer, baseIndex int) (*FileTree, error) {
	tree := &FileTree{
		Entries: make(map[string]*FileNode),
	}

	for i, layer := range layers {
		layerIdx := baseIndex + i
		nodes, err := ParseLayer(layer)
		if err != nil {
			return nil, fmt.Errorf("parse layer: %w", err)
		}

		// Set layer index on all parsed nodes.
		for _, node := range nodes {
			node.LayerIndex = layerIdx
		}

		// Collect opaque whiteout directories from this layer first.
		// Opaque whiteout (.wh..wh..opq) means: discard all entries from
		// lower layers that live under the same directory.
		opaqueDirs := make(map[string]bool)
		for path, node := range nodes {
			if !node.IsDir && filepath.Base(path) == ".wh..wh..opq" {
				dir := filepath.Dir(path)
				if dir == "." {
					dir = ""
				}
				opaqueDirs[dir] = true
			}
		}

		// Apply opaque whiteouts: remove all tree entries under the opaque dirs.
		for opaqueDir := range opaqueDirs {
			prefix := opaqueDir
			if prefix != "" {
				prefix += "/"
			}
			for key := range tree.Entries {
				if prefix == "" || strings.HasPrefix(key, prefix) {
					delete(tree.Entries, key)
				}
			}
			// Also remove the directory entry itself from lower layers
			// (it will be re-added from the current layer's entries if present).
			if opaqueDir != "" {
				delete(tree.Entries, opaqueDir)
			}
		}

		// Apply normal entries and file whiteouts.
		for path, node := range nodes {
			base := filepath.Base(path)

			// Skip opaque whiteout marker — it's not a real file.
			if base == ".wh..wh..opq" {
				continue
			}

			// File whiteout: ".wh.FILENAME" removes FILENAME from the tree.
			if strings.HasPrefix(base, ".wh.") {
				target := strings.TrimPrefix(base, ".wh.")
				dir := filepath.Dir(path)
				var targetPath string
				if dir == "." {
					targetPath = target
				} else {
					targetPath = filepath.Join(dir, target)
				}
				delete(tree.Entries, targetPath)
				continue
			}

			// Normal entry: add or override in the tree.
			tree.Entries[path] = node
		}
	}

	return tree, nil
}

// BuildFromImage builds a squashed FileTree from all layers of a v1.Image.
// Layers are applied in image order (index 0 = bottom layer, last = top layer).
func BuildFromImage(img v1.Image) (*FileTree, error) {
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("get image layers: %w", err)
	}
	return BuildTree(layers)
}

// IdenticalLeadingLayers returns the number of contiguous leading layers that
// are identical between img1 and img2. It compares layers using DiffID (the
// uncompressed content digest stored in the image config), which requires no
// network download.
//
// The comparison stops at the first mismatch. Only the contiguous prefix is
// counted — non-prefix identical layers (e.g., a shared layer at index 2 when
// index 1 differs) are not counted. This is required for whiteout safety: a
// whiteout in a skipped layer could delete a file that is present in the diff.
//
// If a DiffID call returns an error for layer i, the function treats it
// conservatively as a mismatch and returns i, nil (non-fatal).
func IdenticalLeadingLayers(img1, img2 v1.Image) (int, error) {
	layers1, err := img1.Layers()
	if err != nil {
		return 0, fmt.Errorf("get layers for image1: %w", err)
	}
	layers2, err := img2.Layers()
	if err != nil {
		return 0, fmt.Errorf("get layers for image2: %w", err)
	}

	limit := len(layers1)
	if len(layers2) < limit {
		limit = len(layers2)
	}

	for i := 0; i < limit; i++ {
		id1, err := layers1[i].DiffID()
		if err != nil {
			// Conservative: cannot confirm equality, stop here.
			return i, nil
		}
		id2, err := layers2[i].DiffID()
		if err != nil {
			// Conservative: cannot confirm equality, stop here.
			return i, nil
		}
		if id1 != id2 {
			return i, nil
		}
	}
	return limit, nil
}

// BuildFromImageSkipFirst builds a squashed FileTree from a v1.Image, skipping
// the first skipFirst layers. This is used together with IdenticalLeadingLayers
// to avoid downloading layers that are provably identical between two images.
//
// If skipFirst is 0, the result is identical to BuildFromImage.
// If skipFirst >= len(layers), an empty FileTree is returned with no error.
func BuildFromImageSkipFirst(img v1.Image, skipFirst int) (*FileTree, error) {
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("get image layers: %w", err)
	}
	if skipFirst >= len(layers) {
		return &FileTree{Entries: make(map[string]*FileNode)}, nil
	}
	return BuildTreeWithAttribution(layers[skipFirst:], skipFirst)
}
