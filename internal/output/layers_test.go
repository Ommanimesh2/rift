package output_test

import (
	"io"
	"strings"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"

	"github.com/Ommanimesh2/rift/internal/output"
)

// --- Test helpers ---

// testLayer is a minimal v1.Layer implementation for layer comparison tests.
// It has a configurable DiffID and compressed Size; all other methods panic.
type testLayer struct {
	diffID v1.Hash
	size   int64
}

func (l *testLayer) DiffID() (v1.Hash, error)            { return l.diffID, nil }
func (l *testLayer) Size() (int64, error)                { return l.size, nil }
func (l *testLayer) Digest() (v1.Hash, error)            { panic("not implemented") }
func (l *testLayer) Compressed() (io.ReadCloser, error)  { panic("not implemented") }
func (l *testLayer) Uncompressed() (io.ReadCloser, error) { panic("not implemented") }
func (l *testLayer) MediaType() (types.MediaType, error) { panic("not implemented") }

// makeTestLayer creates a testLayer with the given algorithm, hex, and size.
func makeTestLayer(algorithm, hex string, size int64) v1.Layer {
	return &testLayer{
		diffID: v1.Hash{Algorithm: algorithm, Hex: hex},
		size:   size,
	}
}

// testImage is a minimal v1.Image implementation for layer comparison tests.
// Only Layers() is implemented; all other methods panic.
type testImage struct {
	layers []v1.Layer
}

func (img *testImage) Layers() ([]v1.Layer, error) { return img.layers, nil }

// Satisfy the full v1.Image interface with panics for unused methods.
func (img *testImage) MediaType() (types.MediaType, error)               { panic("not implemented") }
func (img *testImage) Size() (int64, error)                              { panic("not implemented") }
func (img *testImage) ConfigName() (v1.Hash, error)                      { panic("not implemented") }
func (img *testImage) ConfigFile() (*v1.ConfigFile, error)               { panic("not implemented") }
func (img *testImage) RawConfigFile() ([]byte, error)                    { panic("not implemented") }
func (img *testImage) Digest() (v1.Hash, error)                          { panic("not implemented") }
func (img *testImage) Manifest() (*v1.Manifest, error)                   { panic("not implemented") }
func (img *testImage) RawManifest() ([]byte, error)                      { panic("not implemented") }
func (img *testImage) LayerByDigest(v1.Hash) (v1.Layer, error)           { panic("not implemented") }
func (img *testImage) LayerByDiffID(v1.Hash) (v1.Layer, error)           { panic("not implemented") }

// makeTestImage creates a testImage from a slice of v1.Layer values.
func makeTestImage(layers ...v1.Layer) v1.Image {
	return &testImage{layers: layers}
}

// hash builds a v1.Hash for use in test comparisons.
func hash(algorithm, hex string) v1.Hash {
	return v1.Hash{Algorithm: algorithm, Hex: hex}
}

// --- CompareLayers tests ---

func TestCompareLayers_IdenticalImages(t *testing.T) {
	layerA := makeTestLayer("sha256", "aaaa", 1000)
	layerB := makeTestLayer("sha256", "bbbb", 2000)

	imgA := makeTestImage(layerA, layerB)
	imgB := makeTestImage(layerA, layerB)

	summary, err := output.CompareLayers(imgA, imgB)
	if err != nil {
		t.Fatalf("CompareLayers error: %v", err)
	}

	if summary.SharedCount != 2 {
		t.Errorf("SharedCount: want 2, got %d", summary.SharedCount)
	}
	if summary.OnlyInACount != 0 {
		t.Errorf("OnlyInACount: want 0, got %d", summary.OnlyInACount)
	}
	if summary.OnlyInBCount != 0 {
		t.Errorf("OnlyInBCount: want 0, got %d", summary.OnlyInBCount)
	}
	if summary.TotalA != 2 {
		t.Errorf("TotalA: want 2, got %d", summary.TotalA)
	}
	if summary.TotalB != 2 {
		t.Errorf("TotalB: want 2, got %d", summary.TotalB)
	}
	// SharedBytes = sum of shared layer sizes = 1000 + 2000
	if summary.SharedBytes != 3000 {
		t.Errorf("SharedBytes: want 3000, got %d", summary.SharedBytes)
	}
}

func TestCompareLayers_CompletelyDifferentImages(t *testing.T) {
	layerA1 := makeTestLayer("sha256", "aaaa", 500)
	layerA2 := makeTestLayer("sha256", "bbbb", 600)
	layerB1 := makeTestLayer("sha256", "cccc", 700)
	layerB2 := makeTestLayer("sha256", "dddd", 800)

	imgA := makeTestImage(layerA1, layerA2)
	imgB := makeTestImage(layerB1, layerB2)

	summary, err := output.CompareLayers(imgA, imgB)
	if err != nil {
		t.Fatalf("CompareLayers error: %v", err)
	}

	if summary.SharedCount != 0 {
		t.Errorf("SharedCount: want 0, got %d", summary.SharedCount)
	}
	if summary.OnlyInACount != 2 {
		t.Errorf("OnlyInACount: want 2, got %d", summary.OnlyInACount)
	}
	if summary.OnlyInBCount != 2 {
		t.Errorf("OnlyInBCount: want 2, got %d", summary.OnlyInBCount)
	}
	if summary.OnlyInABytes != 1100 {
		t.Errorf("OnlyInABytes: want 1100, got %d", summary.OnlyInABytes)
	}
	if summary.OnlyInBBytes != 1500 {
		t.Errorf("OnlyInBBytes: want 1500, got %d", summary.OnlyInBBytes)
	}
}

func TestCompareLayers_PartiallyShared(t *testing.T) {
	sharedLayer := makeTestLayer("sha256", "shared", 1024)
	onlyA := makeTestLayer("sha256", "onlya", 512)
	onlyB := makeTestLayer("sha256", "onlyb", 768)

	imgA := makeTestImage(sharedLayer, onlyA)
	imgB := makeTestImage(sharedLayer, onlyB)

	summary, err := output.CompareLayers(imgA, imgB)
	if err != nil {
		t.Fatalf("CompareLayers error: %v", err)
	}

	if summary.SharedCount != 1 {
		t.Errorf("SharedCount: want 1, got %d", summary.SharedCount)
	}
	if summary.SharedBytes != 1024 {
		t.Errorf("SharedBytes: want 1024, got %d", summary.SharedBytes)
	}
	if summary.OnlyInACount != 1 {
		t.Errorf("OnlyInACount: want 1, got %d", summary.OnlyInACount)
	}
	if summary.OnlyInABytes != 512 {
		t.Errorf("OnlyInABytes: want 512, got %d", summary.OnlyInABytes)
	}
	if summary.OnlyInBCount != 1 {
		t.Errorf("OnlyInBCount: want 1, got %d", summary.OnlyInBCount)
	}
	if summary.OnlyInBBytes != 768 {
		t.Errorf("OnlyInBBytes: want 768, got %d", summary.OnlyInBBytes)
	}
	if summary.TotalA != 2 {
		t.Errorf("TotalA: want 2, got %d", summary.TotalA)
	}
	if summary.TotalB != 2 {
		t.Errorf("TotalB: want 2, got %d", summary.TotalB)
	}
}

func TestCompareLayers_EmptyImageA(t *testing.T) {
	layerB := makeTestLayer("sha256", "bbbb", 500)

	imgA := makeTestImage()
	imgB := makeTestImage(layerB)

	summary, err := output.CompareLayers(imgA, imgB)
	if err != nil {
		t.Fatalf("CompareLayers error: %v", err)
	}

	if summary.SharedCount != 0 {
		t.Errorf("SharedCount: want 0, got %d", summary.SharedCount)
	}
	if summary.OnlyInACount != 0 {
		t.Errorf("OnlyInACount: want 0, got %d", summary.OnlyInACount)
	}
	if summary.OnlyInBCount != 1 {
		t.Errorf("OnlyInBCount: want 1, got %d", summary.OnlyInBCount)
	}
	if summary.OnlyInBBytes != 500 {
		t.Errorf("OnlyInBBytes: want 500, got %d", summary.OnlyInBBytes)
	}
	if summary.TotalA != 0 {
		t.Errorf("TotalA: want 0, got %d", summary.TotalA)
	}
	if summary.TotalB != 1 {
		t.Errorf("TotalB: want 1, got %d", summary.TotalB)
	}
}

func TestCompareLayers_SingleLayerImages(t *testing.T) {
	shared := makeTestLayer("sha256", "same", 2048)

	imgA := makeTestImage(shared)
	imgB := makeTestImage(shared)

	summary, err := output.CompareLayers(imgA, imgB)
	if err != nil {
		t.Fatalf("CompareLayers error: %v", err)
	}

	if summary.SharedCount != 1 {
		t.Errorf("SharedCount: want 1, got %d", summary.SharedCount)
	}
	if summary.SharedBytes != 2048 {
		t.Errorf("SharedBytes: want 2048, got %d", summary.SharedBytes)
	}
	if summary.OnlyInACount != 0 {
		t.Errorf("OnlyInACount: want 0, got %d", summary.OnlyInACount)
	}
	if summary.OnlyInBCount != 0 {
		t.Errorf("OnlyInBCount: want 0, got %d", summary.OnlyInBCount)
	}
}

// --- FormatLayerSummary tests ---

func TestFormatLayerSummary_AllShared(t *testing.T) {
	summary := &output.LayerSummary{
		SharedCount: 3,
		SharedBytes: 3 * 1024,
		TotalA:      3,
		TotalB:      3,
	}

	got := output.FormatLayerSummary(summary)

	if !strings.Contains(got, "3 → 3") {
		t.Errorf("FormatLayerSummary = %q, want it to contain layer counts %q", got, "3 → 3")
	}
	if !strings.Contains(got, "3 layers") {
		t.Errorf("FormatLayerSummary = %q, want it to contain shared layers count", got)
	}
	// No "Only in" lines when counts are zero
	if strings.Contains(got, "Only in A") {
		t.Errorf("FormatLayerSummary = %q, should not contain 'Only in A' when count is 0", got)
	}
	if strings.Contains(got, "Only in B") {
		t.Errorf("FormatLayerSummary = %q, should not contain 'Only in B' when count is 0", got)
	}
}

func TestFormatLayerSummary_WithUniqueLayersOnBothSides(t *testing.T) {
	summary := &output.LayerSummary{
		SharedCount:  2,
		SharedBytes:  2048,
		OnlyInACount: 1,
		OnlyInABytes: 512,
		OnlyInBCount: 3,
		OnlyInBBytes: 3072,
		TotalA:       3,
		TotalB:       5,
	}

	got := output.FormatLayerSummary(summary)

	if !strings.Contains(got, "3 → 5") {
		t.Errorf("FormatLayerSummary = %q, want %q", got, "3 → 5")
	}
	if !strings.Contains(got, "Only in A") {
		t.Errorf("FormatLayerSummary = %q, want 'Only in A' section", got)
	}
	if !strings.Contains(got, "Only in B") {
		t.Errorf("FormatLayerSummary = %q, want 'Only in B' section", got)
	}
}

func TestFormatLayerSummary_OnlyInAOmittedWhenZero(t *testing.T) {
	summary := &output.LayerSummary{
		SharedCount:  1,
		SharedBytes:  1024,
		OnlyInACount: 0,
		OnlyInBCount: 2,
		OnlyInBBytes: 2048,
		TotalA:       1,
		TotalB:       3,
	}

	got := output.FormatLayerSummary(summary)

	if strings.Contains(got, "Only in A") {
		t.Errorf("FormatLayerSummary = %q, 'Only in A' line should be omitted when count is 0", got)
	}
	if !strings.Contains(got, "Only in B") {
		t.Errorf("FormatLayerSummary = %q, want 'Only in B' section", got)
	}
}

func TestFormatLayerSummary_OnlyInBOmittedWhenZero(t *testing.T) {
	summary := &output.LayerSummary{
		SharedCount:  1,
		SharedBytes:  1024,
		OnlyInACount: 2,
		OnlyInABytes: 2048,
		OnlyInBCount: 0,
		TotalA:       3,
		TotalB:       1,
	}

	got := output.FormatLayerSummary(summary)

	if strings.Contains(got, "Only in B") {
		t.Errorf("FormatLayerSummary = %q, 'Only in B' line should be omitted when count is 0", got)
	}
	if !strings.Contains(got, "Only in A") {
		t.Errorf("FormatLayerSummary = %q, want 'Only in A' section", got)
	}
}

// Ensure the hash helper compiles correctly.
var _ = hash("sha256", "abc")
