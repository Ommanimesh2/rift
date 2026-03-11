package tree

import (
	"archive/tar"
	"bytes"
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
