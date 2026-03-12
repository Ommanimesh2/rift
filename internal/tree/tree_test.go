package tree

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// tarEntry describes a single entry to write into an in-memory tar archive.
type tarEntry struct {
	path    string
	content []byte
	mode    int64
	uid     int
	gid     int
	typeflag byte // tar.TypeReg, tar.TypeDir, tar.TypeSymlink
	linkname string
}

// buildTar creates an in-memory tar archive from the given entries.
func buildTar(entries []tarEntry) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.path,
			Mode:     e.mode,
			Uid:      e.uid,
			Gid:      e.gid,
			Typeflag: e.typeflag,
			Linkname: e.linkname,
			Size:     int64(len(e.content)),
		}
		if e.typeflag == tar.TypeDir {
			hdr.Size = 0
		}
		if e.typeflag == tar.TypeSymlink {
			hdr.Size = 0
		}
		_ = tw.WriteHeader(hdr)
		if len(e.content) > 0 {
			_, _ = tw.Write(e.content)
		}
	}
	_ = tw.Close()
	return buf.Bytes()
}

// fakeLayer is a minimal v1.Layer implementation backed by an in-memory tar.
type fakeLayer struct {
	tarData []byte
}

func (f *fakeLayer) Digest() (v1.Hash, error)            { return v1.Hash{}, nil }
func (f *fakeLayer) DiffID() (v1.Hash, error)            { return v1.Hash{}, nil }
func (f *fakeLayer) Compressed() (io.ReadCloser, error)  { return nil, nil }
func (f *fakeLayer) Size() (int64, error)                { return int64(len(f.tarData)), nil }
func (f *fakeLayer) MediaType() (types.MediaType, error) { return "", nil }
func (f *fakeLayer) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.tarData)), nil
}

// makeFakeLayer builds a v1.Layer from a slice of tarEntry values.
func makeFakeLayer(entries []tarEntry) v1.Layer {
	return &fakeLayer{tarData: buildTar(entries)}
}

// --- Tests for ParseLayer ---

func TestParseLayer_RegularFile(t *testing.T) {
	layer := makeFakeLayer([]tarEntry{
		{
			path:     "usr/bin/cat",
			content:  []byte("hello content"),
			mode:     0755,
			uid:      0,
			gid:      0,
			typeflag: tar.TypeReg,
		},
	})

	nodes, err := ParseLayer(layer)
	if err != nil {
		t.Fatalf("ParseLayer returned unexpected error: %v", err)
	}

	node, ok := nodes["usr/bin/cat"]
	if !ok {
		t.Fatalf("expected key %q in result map, got keys: %v", "usr/bin/cat", mapKeys(nodes))
	}

	if node.Path != "usr/bin/cat" {
		t.Errorf("expected Path=%q, got %q", "usr/bin/cat", node.Path)
	}
	if node.Size != 13 {
		t.Errorf("expected Size=13, got %d", node.Size)
	}
	if node.Mode != os.FileMode(0755) {
		t.Errorf("expected Mode=0755, got %v", node.Mode)
	}
	if node.IsDir {
		t.Error("expected IsDir=false for regular file")
	}
	if node.LinkTarget != "" {
		t.Errorf("expected empty LinkTarget, got %q", node.LinkTarget)
	}
}

func TestParseLayer_Directory(t *testing.T) {
	layer := makeFakeLayer([]tarEntry{
		{
			path:     "etc/",
			mode:     0755,
			uid:      0,
			gid:      0,
			typeflag: tar.TypeDir,
		},
	})

	nodes, err := ParseLayer(layer)
	if err != nil {
		t.Fatalf("ParseLayer returned unexpected error: %v", err)
	}

	node, ok := nodes["etc"]
	if !ok {
		t.Fatalf("expected key %q in result map, got keys: %v", "etc", mapKeys(nodes))
	}

	if !node.IsDir {
		t.Error("expected IsDir=true for directory entry")
	}
	if node.Size != 0 {
		t.Errorf("expected Size=0 for directory, got %d", node.Size)
	}
}

func TestParseLayer_Symlink(t *testing.T) {
	layer := makeFakeLayer([]tarEntry{
		{
			path:     "usr/bin/sh",
			mode:     0777,
			typeflag: tar.TypeSymlink,
			linkname: "/bin/bash",
		},
	})

	nodes, err := ParseLayer(layer)
	if err != nil {
		t.Fatalf("ParseLayer returned unexpected error: %v", err)
	}

	node, ok := nodes["usr/bin/sh"]
	if !ok {
		t.Fatalf("expected key %q in result map, got keys: %v", "usr/bin/sh", mapKeys(nodes))
	}

	if node.LinkTarget != "/bin/bash" {
		t.Errorf("expected LinkTarget=%q, got %q", "/bin/bash", node.LinkTarget)
	}
	if node.Size != 0 {
		t.Errorf("expected Size=0 for symlink, got %d", node.Size)
	}
	if node.IsDir {
		t.Error("expected IsDir=false for symlink")
	}
}

func TestParseLayer_MetadataPreserved(t *testing.T) {
	tests := []struct {
		name     string
		entry    tarEntry
		wantMode os.FileMode
		wantUID  int
		wantGID  int
	}{
		{
			name: "executable mode 0755",
			entry: tarEntry{
				path: "bin/tool", content: []byte("x"), mode: 0755, uid: 1000, gid: 1000, typeflag: tar.TypeReg,
			},
			wantMode: 0755,
			wantUID:  1000,
			wantGID:  1000,
		},
		{
			name: "read-only mode 0644",
			entry: tarEntry{
				path: "etc/config", content: []byte("y"), mode: 0644, uid: 0, gid: 500, typeflag: tar.TypeReg,
			},
			wantMode: 0644,
			wantUID:  0,
			wantGID:  500,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			layer := makeFakeLayer([]tarEntry{tc.entry})
			nodes, err := ParseLayer(layer)
			if err != nil {
				t.Fatalf("ParseLayer error: %v", err)
			}
			node, ok := nodes[tc.entry.path]
			if !ok {
				t.Fatalf("key %q not in result", tc.entry.path)
			}
			if node.Mode != tc.wantMode {
				t.Errorf("Mode: want %v, got %v", tc.wantMode, node.Mode)
			}
			if node.UID != tc.wantUID {
				t.Errorf("UID: want %d, got %d", tc.wantUID, node.UID)
			}
			if node.GID != tc.wantGID {
				t.Errorf("GID: want %d, got %d", tc.wantGID, node.GID)
			}
		})
	}
}

func TestParseLayer_MultipleEntries(t *testing.T) {
	layer := makeFakeLayer([]tarEntry{
		{path: "a.txt", content: []byte("aaa"), mode: 0644, typeflag: tar.TypeReg},
		{path: "b.txt", content: []byte("bbbbb"), mode: 0644, typeflag: tar.TypeReg},
		{path: "subdir/", mode: 0755, typeflag: tar.TypeDir},
	})

	nodes, err := ParseLayer(layer)
	if err != nil {
		t.Fatalf("ParseLayer error: %v", err)
	}
	if len(nodes) != 3 {
		t.Errorf("expected 3 entries, got %d: %v", len(nodes), mapKeys(nodes))
	}
	if nodes["a.txt"].Size != 3 {
		t.Errorf("a.txt: expected size 3, got %d", nodes["a.txt"].Size)
	}
	if nodes["b.txt"].Size != 5 {
		t.Errorf("b.txt: expected size 5, got %d", nodes["b.txt"].Size)
	}
	if !nodes["subdir"].IsDir {
		t.Error("subdir: expected IsDir=true")
	}
}

func TestParseLayer_EmptyLayer(t *testing.T) {
	layer := makeFakeLayer([]tarEntry{})

	nodes, err := ParseLayer(layer)
	if err != nil {
		t.Fatalf("ParseLayer returned error for empty layer: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected empty map, got %d entries: %v", len(nodes), mapKeys(nodes))
	}
}

func TestParseLayer_PathNormalization(t *testing.T) {
	tests := []struct {
		name    string
		rawPath string
		wantKey string
	}{
		{name: "leading dot-slash", rawPath: "./etc/passwd", wantKey: "etc/passwd"},
		{name: "leading slash", rawPath: "/var/log/app.log", wantKey: "var/log/app.log"},
		{name: "clean path", rawPath: "usr/local/bin/tool", wantKey: "usr/local/bin/tool"},
		{name: "dot-slash on dir", rawPath: "./home/", wantKey: "home"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			typeflag := byte(tar.TypeReg)
			content := []byte("data")
			if tc.rawPath[len(tc.rawPath)-1] == '/' {
				typeflag = tar.TypeDir
				content = nil
			}
			layer := makeFakeLayer([]tarEntry{
				{path: tc.rawPath, content: content, mode: 0644, typeflag: typeflag},
			})
			nodes, err := ParseLayer(layer)
			if err != nil {
				t.Fatalf("ParseLayer error: %v", err)
			}
			if _, ok := nodes[tc.wantKey]; !ok {
				t.Errorf("expected key %q, got keys: %v", tc.wantKey, mapKeys(nodes))
			}
		})
	}
}

func TestParseLayer_ContentDigest(t *testing.T) {
	content := []byte("known content for hashing")
	layer := makeFakeLayer([]tarEntry{
		{path: "file.txt", content: content, mode: 0644, typeflag: tar.TypeReg},
	})

	nodes, err := ParseLayer(layer)
	if err != nil {
		t.Fatalf("ParseLayer error: %v", err)
	}

	node := nodes["file.txt"]
	if node.Digest.Algorithm != "sha256" {
		t.Errorf("expected digest algorithm %q, got %q", "sha256", node.Digest.Algorithm)
	}
	if node.Digest.Hex == "" {
		t.Error("expected non-empty digest hex for regular file")
	}
}

// mapKeys returns a sorted slice of keys from the map, for test error messages.
func mapKeys(m map[string]*FileNode) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// --- Tests for BuildTree (whiteout handling and multi-layer squashing) ---

func TestBuildTree_TwoLayersNoWhiteout(t *testing.T) {
	// Layer 1: base with two files.
	layer1 := makeFakeLayer([]tarEntry{
		{path: "etc/config.txt", content: []byte("v1 config"), mode: 0644, typeflag: tar.TypeReg},
		{path: "bin/tool", content: []byte("v1 binary"), mode: 0755, typeflag: tar.TypeReg},
	})
	// Layer 2: overrides config.txt, adds new file.
	layer2 := makeFakeLayer([]tarEntry{
		{path: "etc/config.txt", content: []byte("v2 config updated"), mode: 0644, typeflag: tar.TypeReg},
		{path: "var/log/app.log", content: []byte("log data"), mode: 0644, typeflag: tar.TypeReg},
	})

	ft, err := BuildTree([]v1.Layer{layer1, layer2})
	if err != nil {
		t.Fatalf("BuildTree error: %v", err)
	}

	// All three distinct paths should be present.
	if ft.Size() != 3 {
		t.Errorf("expected 3 entries, got %d: %v", ft.Size(), treeKeys(ft))
	}
	// Layer 2 version of config.txt should win (larger content).
	node := ft.Get("etc/config.txt")
	if node == nil {
		t.Fatal("expected etc/config.txt in tree")
	}
	if node.Size != 17 {
		t.Errorf("expected layer2 size 17, got %d", node.Size)
	}
	if ft.Get("bin/tool") == nil {
		t.Error("bin/tool from layer1 should remain")
	}
	if ft.Get("var/log/app.log") == nil {
		t.Error("var/log/app.log from layer2 should be present")
	}
}

func TestBuildTree_FileWhiteout(t *testing.T) {
	layer1 := makeFakeLayer([]tarEntry{
		{path: "etc/foo.txt", content: []byte("to be deleted"), mode: 0644, typeflag: tar.TypeReg},
		{path: "etc/keep.txt", content: []byte("stays"), mode: 0644, typeflag: tar.TypeReg},
	})
	// Layer 2 deletes foo.txt via whiteout marker.
	layer2 := makeFakeLayer([]tarEntry{
		{path: "etc/.wh.foo.txt", content: nil, mode: 0644, typeflag: tar.TypeReg},
	})

	ft, err := BuildTree([]v1.Layer{layer1, layer2})
	if err != nil {
		t.Fatalf("BuildTree error: %v", err)
	}

	if ft.Get("etc/foo.txt") != nil {
		t.Error("etc/foo.txt should have been whited out")
	}
	if ft.Get("etc/keep.txt") == nil {
		t.Error("etc/keep.txt should still be present")
	}
	// Whiteout marker itself should NOT be in the tree.
	if ft.Get("etc/.wh.foo.txt") != nil {
		t.Error("whiteout marker etc/.wh.foo.txt must not appear in final tree")
	}
}

func TestBuildTree_OpaqueWhiteout(t *testing.T) {
	// Layer 1 populates a directory with several files.
	layer1 := makeFakeLayer([]tarEntry{
		{path: "app/", mode: 0755, typeflag: tar.TypeDir},
		{path: "app/old.txt", content: []byte("old"), mode: 0644, typeflag: tar.TypeReg},
		{path: "app/other.txt", content: []byte("other"), mode: 0644, typeflag: tar.TypeReg},
		{path: "keep.txt", content: []byte("root file"), mode: 0644, typeflag: tar.TypeReg},
	})
	// Layer 2 has an opaque whiteout for "app/" and adds one new file inside.
	layer2 := makeFakeLayer([]tarEntry{
		{path: "app/", mode: 0755, typeflag: tar.TypeDir},
		{path: "app/.wh..wh..opq", content: nil, mode: 0644, typeflag: tar.TypeReg},
		{path: "app/new.txt", content: []byte("new content"), mode: 0644, typeflag: tar.TypeReg},
	})

	ft, err := BuildTree([]v1.Layer{layer1, layer2})
	if err != nil {
		t.Fatalf("BuildTree error: %v", err)
	}

	// Old files inside app/ must be gone.
	if ft.Get("app/old.txt") != nil {
		t.Error("app/old.txt should have been removed by opaque whiteout")
	}
	if ft.Get("app/other.txt") != nil {
		t.Error("app/other.txt should have been removed by opaque whiteout")
	}
	// New file from layer 2 inside app/ must be present.
	if ft.Get("app/new.txt") == nil {
		t.Error("app/new.txt from layer2 should be present after opaque whiteout")
	}
	// Opaque whiteout marker must not appear.
	if ft.Get("app/.wh..wh..opq") != nil {
		t.Error("opaque whiteout marker must not appear in final tree")
	}
	// Root-level file must be untouched.
	if ft.Get("keep.txt") == nil {
		t.Error("keep.txt at root should not be affected by opaque whiteout of app/")
	}
}

func TestBuildTree_NestedWhiteout(t *testing.T) {
	layer1 := makeFakeLayer([]tarEntry{
		{path: "a/b/c.txt", content: []byte("deep file"), mode: 0644, typeflag: tar.TypeReg},
	})
	layer2 := makeFakeLayer([]tarEntry{
		{path: "a/b/.wh.c.txt", content: nil, mode: 0644, typeflag: tar.TypeReg},
	})

	ft, err := BuildTree([]v1.Layer{layer1, layer2})
	if err != nil {
		t.Fatalf("BuildTree error: %v", err)
	}

	if ft.Get("a/b/c.txt") != nil {
		t.Error("a/b/c.txt should have been removed by nested whiteout")
	}
}

func TestBuildTree_WhiteoutNonExistentFile(t *testing.T) {
	// Whiteout for a file that never existed — should not error.
	layer1 := makeFakeLayer([]tarEntry{
		{path: "existing.txt", content: []byte("here"), mode: 0644, typeflag: tar.TypeReg},
	})
	layer2 := makeFakeLayer([]tarEntry{
		{path: ".wh.ghost.txt", content: nil, mode: 0644, typeflag: tar.TypeReg},
	})

	ft, err := BuildTree([]v1.Layer{layer1, layer2})
	if err != nil {
		t.Fatalf("BuildTree should not error on whiteout of non-existent file: %v", err)
	}
	if ft.Get("existing.txt") == nil {
		t.Error("existing.txt should still be present")
	}
	if ft.Get(".wh.ghost.txt") != nil {
		t.Error("whiteout marker must not appear in tree")
	}
}

func TestBuildTree_LayerOrderMatters(t *testing.T) {
	layer1 := makeFakeLayer([]tarEntry{
		{path: "file.txt", content: []byte("from layer1"), mode: 0644, typeflag: tar.TypeReg},
	})
	layer2 := makeFakeLayer([]tarEntry{
		{path: "file.txt", content: []byte("from layer2 overrides"), mode: 0644, typeflag: tar.TypeReg},
	})
	layer3 := makeFakeLayer([]tarEntry{
		{path: "file.txt", content: []byte("layer3 wins"), mode: 0755, typeflag: tar.TypeReg},
	})

	ft, err := BuildTree([]v1.Layer{layer1, layer2, layer3})
	if err != nil {
		t.Fatalf("BuildTree error: %v", err)
	}

	node := ft.Get("file.txt")
	if node == nil {
		t.Fatal("file.txt should be in tree")
	}
	// Layer 3 is the top layer: its size and mode should win.
	if node.Size != int64(len("layer3 wins")) {
		t.Errorf("expected layer3 size %d, got %d", len("layer3 wins"), node.Size)
	}
	if node.Mode != os.FileMode(0755) {
		t.Errorf("expected layer3 mode 0755, got %v", node.Mode)
	}
}

func TestBuildTree_FileTreeAccessors(t *testing.T) {
	layer := makeFakeLayer([]tarEntry{
		{path: "a.txt", content: []byte("a"), mode: 0644, typeflag: tar.TypeReg},
		{path: "b/", mode: 0755, typeflag: tar.TypeDir},
		{path: "b/c.txt", content: []byte("c"), mode: 0644, typeflag: tar.TypeReg},
	})

	ft, err := BuildTree([]v1.Layer{layer})
	if err != nil {
		t.Fatalf("BuildTree error: %v", err)
	}

	if ft.Size() != 3 {
		t.Errorf("Size(): expected 3, got %d", ft.Size())
	}
	if ft.Get("a.txt") == nil {
		t.Error("Get(a.txt) should return non-nil")
	}
	if ft.Get("nonexistent") != nil {
		t.Error("Get(nonexistent) should return nil")
	}
}

// treeKeys returns a slice of keys from the FileTree, for test error messages.
func treeKeys(ft *FileTree) []string {
	keys := make([]string, 0, len(ft.Entries))
	for k := range ft.Entries {
		keys = append(keys, k)
	}
	return keys
}

// --- Helpers for IdenticalLeadingLayers and BuildFromImageSkipFirst tests ---

// fakeDiffIDLayer is a v1.Layer whose DiffID returns a configurable hash.
// Used to test IdenticalLeadingLayers without any tar/network I/O.
type fakeDiffIDLayer struct {
	diffID v1.Hash
	errDiffID bool
	// tarData is optional — used when Uncompressed() must work for BuildFromImageSkipFirst tests.
	tarData []byte
}

func (f *fakeDiffIDLayer) Digest() (v1.Hash, error)            { return v1.Hash{}, nil }
func (f *fakeDiffIDLayer) DiffID() (v1.Hash, error) {
	if f.errDiffID {
		return v1.Hash{}, fmt.Errorf("simulated DiffID error")
	}
	return f.diffID, nil
}
func (f *fakeDiffIDLayer) Compressed() (io.ReadCloser, error)  { return nil, nil }
func (f *fakeDiffIDLayer) Size() (int64, error)                { return int64(len(f.tarData)), nil }
func (f *fakeDiffIDLayer) MediaType() (types.MediaType, error) { return "", nil }
func (f *fakeDiffIDLayer) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.tarData)), nil
}

// fakeImageWithLayers is a minimal v1.Image whose Layers() returns a fixed slice.
// All other v1.Image methods return errors (they are not needed for these tests).
type fakeImageWithLayers struct {
	layers []v1.Layer
}

func (f *fakeImageWithLayers) Layers() ([]v1.Layer, error) {
	return f.layers, nil
}

// Satisfy the v1.Image interface — all unused methods return errors.
func (f *fakeImageWithLayers) MediaType() (types.MediaType, error)  { return "", fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) Size() (int64, error)                  { return 0, fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) ConfigName() (v1.Hash, error)          { return v1.Hash{}, fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) ConfigFile() (*v1.ConfigFile, error)   { return nil, fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) RawConfigFile() ([]byte, error)        { return nil, fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) Digest() (v1.Hash, error)              { return v1.Hash{}, fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) Manifest() (*v1.Manifest, error)       { return nil, fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) RawManifest() ([]byte, error)          { return nil, fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) LayerByDigest(v1.Hash) (v1.Layer, error) { return nil, fmt.Errorf("not implemented") }
func (f *fakeImageWithLayers) LayerByDiffID(v1.Hash) (v1.Layer, error) { return nil, fmt.Errorf("not implemented") }

// hashFor returns a v1.Hash with algorithm "sha256" and the given hex string.
// It is used to construct deterministic DiffID values in tests.
func hashFor(hex string) v1.Hash {
	return v1.Hash{Algorithm: "sha256", Hex: hex}
}

// makeImageWithDiffIDs creates a fakeImageWithLayers from a slice of hex strings.
// Each hex string becomes the DiffID for the corresponding layer.
func makeImageWithDiffIDs(hexes []string) *fakeImageWithLayers {
	layers := make([]v1.Layer, len(hexes))
	for i, h := range hexes {
		layers[i] = &fakeDiffIDLayer{diffID: hashFor(h)}
	}
	return &fakeImageWithLayers{layers: layers}
}

// --- Tests for IdenticalLeadingLayers ---

func TestIdenticalLeadingLayers(t *testing.T) {
	tests := []struct {
		name    string
		img1IDs []string
		img2IDs []string
		want    int
	}{
		{
			name:    "both images empty",
			img1IDs: []string{},
			img2IDs: []string{},
			want:    0,
		},
		{
			name:    "all layers identical single",
			img1IDs: []string{"aaa"},
			img2IDs: []string{"aaa"},
			want:    1,
		},
		{
			name:    "all layers identical multiple",
			img1IDs: []string{"aaa", "bbb", "ccc"},
			img2IDs: []string{"aaa", "bbb", "ccc"},
			want:    3,
		},
		{
			name:    "no layers identical",
			img1IDs: []string{"aaa", "bbb"},
			img2IDs: []string{"xxx", "yyy"},
			want:    0,
		},
		{
			name:    "partial prefix first two of four identical",
			img1IDs: []string{"aaa", "bbb", "ccc", "ddd"},
			img2IDs: []string{"aaa", "bbb", "xxx", "yyy"},
			want:    2,
		},
		{
			name:    "prefix only img1=[A,B,X] img2=[A,B,Y]",
			img1IDs: []string{"aaa", "bbb", "xxx"},
			img2IDs: []string{"aaa", "bbb", "yyy"},
			want:    2,
		},
		{
			name:    "non-contiguous shared layers only prefix counted",
			img1IDs: []string{"aaa", "xxx", "bbb"},
			img2IDs: []string{"aaa", "yyy", "bbb"},
			want:    1,
		},
		{
			name:    "different length with full prefix overlap img1 shorter",
			img1IDs: []string{"aaa", "bbb"},
			img2IDs: []string{"aaa", "bbb", "ccc"},
			want:    2,
		},
		{
			name:    "different length with full prefix overlap img2 shorter",
			img1IDs: []string{"aaa", "bbb", "ccc"},
			img2IDs: []string{"aaa", "bbb"},
			want:    2,
		},
		{
			name:    "single layer mismatch",
			img1IDs: []string{"aaa"},
			img2IDs: []string{"bbb"},
			want:    0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			img1 := makeImageWithDiffIDs(tc.img1IDs)
			img2 := makeImageWithDiffIDs(tc.img2IDs)

			got, err := IdenticalLeadingLayers(img1, img2)
			if err != nil {
				t.Fatalf("IdenticalLeadingLayers returned unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("IdenticalLeadingLayers = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestIdenticalLeadingLayers_DiffIDError(t *testing.T) {
	// Layer at index 1 will return a DiffID error.
	// Expectation: return 1 (the count before the error), nil (non-fatal).
	layers1 := []v1.Layer{
		&fakeDiffIDLayer{diffID: hashFor("aaa")},
		&fakeDiffIDLayer{diffID: hashFor("bbb"), errDiffID: true},
		&fakeDiffIDLayer{diffID: hashFor("ccc")},
	}
	layers2 := []v1.Layer{
		&fakeDiffIDLayer{diffID: hashFor("aaa")},
		&fakeDiffIDLayer{diffID: hashFor("bbb")},
		&fakeDiffIDLayer{diffID: hashFor("ccc")},
	}
	img1 := &fakeImageWithLayers{layers: layers1}
	img2 := &fakeImageWithLayers{layers: layers2}

	got, err := IdenticalLeadingLayers(img1, img2)
	if err != nil {
		t.Fatalf("expected nil error on DiffID error, got: %v", err)
	}
	if got != 1 {
		t.Errorf("expected 1 (layers before error), got %d", got)
	}
}

// --- Tests for BuildFromImageSkipFirst ---

func TestBuildFromImageSkipFirst_SkipZero(t *testing.T) {
	// skipFirst=0 must produce the same result as BuildFromImage.
	layer1 := makeFakeLayer([]tarEntry{
		{path: "base.txt", content: []byte("base"), mode: 0644, typeflag: tar.TypeReg},
	})
	layer2 := makeFakeLayer([]tarEntry{
		{path: "top.txt", content: []byte("top"), mode: 0644, typeflag: tar.TypeReg},
	})
	img := &fakeImageWithLayers{layers: []v1.Layer{layer1, layer2}}

	ftSkip, err := BuildFromImageSkipFirst(img, 0)
	if err != nil {
		t.Fatalf("BuildFromImageSkipFirst(img, 0) error: %v", err)
	}
	ftFull, err := BuildFromImage(img)
	if err != nil {
		t.Fatalf("BuildFromImage error: %v", err)
	}

	if ftSkip.Size() != ftFull.Size() {
		t.Errorf("size mismatch: skipFirst=0 got %d, BuildFromImage got %d", ftSkip.Size(), ftFull.Size())
	}
	for path := range ftFull.Entries {
		if ftSkip.Get(path) == nil {
			t.Errorf("path %q present in BuildFromImage but missing in BuildFromImageSkipFirst(0)", path)
		}
	}
}

func TestBuildFromImageSkipFirst_SkipOne(t *testing.T) {
	// A 3-layer image; skip the first layer.
	// Files from layer 1 must be absent; files from layers 2 and 3 must be present.
	layer1 := makeFakeLayer([]tarEntry{
		{path: "layer1_only.txt", content: []byte("l1"), mode: 0644, typeflag: tar.TypeReg},
	})
	layer2 := makeFakeLayer([]tarEntry{
		{path: "layer2_file.txt", content: []byte("l2"), mode: 0644, typeflag: tar.TypeReg},
	})
	layer3 := makeFakeLayer([]tarEntry{
		{path: "layer3_file.txt", content: []byte("l3"), mode: 0644, typeflag: tar.TypeReg},
	})
	img := &fakeImageWithLayers{layers: []v1.Layer{layer1, layer2, layer3}}

	ft, err := BuildFromImageSkipFirst(img, 1)
	if err != nil {
		t.Fatalf("BuildFromImageSkipFirst(img, 1) error: %v", err)
	}

	if ft.Get("layer1_only.txt") != nil {
		t.Error("layer1_only.txt should be absent when skipFirst=1")
	}
	if ft.Get("layer2_file.txt") == nil {
		t.Error("layer2_file.txt should be present when skipFirst=1")
	}
	if ft.Get("layer3_file.txt") == nil {
		t.Error("layer3_file.txt should be present when skipFirst=1")
	}
}

func TestBuildFromImageSkipFirst_SkipAll(t *testing.T) {
	// skipFirst >= len(layers): must return an empty tree without panicking.
	layer1 := makeFakeLayer([]tarEntry{
		{path: "file.txt", content: []byte("data"), mode: 0644, typeflag: tar.TypeReg},
	})
	img := &fakeImageWithLayers{layers: []v1.Layer{layer1}}

	ft, err := BuildFromImageSkipFirst(img, 5)
	if err != nil {
		t.Fatalf("BuildFromImageSkipFirst with skipFirst>=len(layers) error: %v", err)
	}
	if ft == nil {
		t.Fatal("expected non-nil FileTree")
	}
	if ft.Size() != 0 {
		t.Errorf("expected empty tree, got %d entries", ft.Size())
	}
}
