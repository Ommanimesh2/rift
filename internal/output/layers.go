package output

import (
	"fmt"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// LayerSummary holds the result of comparing layers between two images.
type LayerSummary struct {
	// SharedCount is the number of layers with matching DiffIDs in both images.
	SharedCount int
	// SharedBytes is the combined compressed size of shared layers.
	SharedBytes int64

	// OnlyInACount is the number of layers unique to image A.
	OnlyInACount int
	// OnlyInABytes is the combined compressed size of A-unique layers.
	OnlyInABytes int64

	// OnlyInBCount is the number of layers unique to image B.
	OnlyInBCount int
	// OnlyInBBytes is the combined compressed size of B-unique layers.
	OnlyInBBytes int64

	// TotalA is the total number of layers in image A.
	TotalA int
	// TotalB is the total number of layers in image B.
	TotalB int
}

// CompareLayers compares the layers of two v1.Images by DiffID and returns a
// *LayerSummary describing which layers are shared and which are unique.
//
// DiffID is the digest of the uncompressed layer content — identical layers
// in two images share the same DiffID regardless of compression.
//
// Size is collected via layer.Size() (compressed size). This matches what
// registries report and gives a realistic download-cost metric.
func CompareLayers(imgA, imgB v1.Image) (*LayerSummary, error) {
	layersA, err := imgA.Layers()
	if err != nil {
		return nil, fmt.Errorf("get layers for image A: %w", err)
	}
	layersB, err := imgB.Layers()
	if err != nil {
		return nil, fmt.Errorf("get layers for image B: %w", err)
	}

	// Build map[DiffID string → compressed size] for image A.
	aMap := make(map[v1.Hash]int64, len(layersA))
	for _, layer := range layersA {
		diffID, err := layer.DiffID()
		if err != nil {
			return nil, fmt.Errorf("get DiffID for A layer: %w", err)
		}
		size, err := layer.Size()
		if err != nil {
			return nil, fmt.Errorf("get size for A layer: %w", err)
		}
		aMap[diffID] = size
	}

	// Build map[DiffID string → compressed size] for image B.
	bMap := make(map[v1.Hash]int64, len(layersB))
	for _, layer := range layersB {
		diffID, err := layer.DiffID()
		if err != nil {
			return nil, fmt.Errorf("get DiffID for B layer: %w", err)
		}
		size, err := layer.Size()
		if err != nil {
			return nil, fmt.Errorf("get size for B layer: %w", err)
		}
		bMap[diffID] = size
	}

	summary := &LayerSummary{
		TotalA: len(layersA),
		TotalB: len(layersB),
	}

	// Classify A layers as shared or only-in-A.
	for diffID, size := range aMap {
		if _, inB := bMap[diffID]; inB {
			summary.SharedCount++
			summary.SharedBytes += size
		} else {
			summary.OnlyInACount++
			summary.OnlyInABytes += size
		}
	}

	// Classify B layers that are not in A as only-in-B.
	for diffID, size := range bMap {
		if _, inA := aMap[diffID]; !inA {
			summary.OnlyInBCount++
			summary.OnlyInBBytes += size
		}
	}

	return summary, nil
}

// FormatLayerSummary renders a LayerSummary as a multi-line plain-text string.
//
// Format:
//
//	Layers: {totalA} → {totalB}
//	  Shared:    {sharedCount} layers ({sharedBytes})
//	  Only in A: {onlyInACount} layers ({onlyInABytes})   [omitted if 0]
//	  Only in B: {onlyInBCount} layers ({onlyInBBytes})   [omitted if 0]
func FormatLayerSummary(summary *LayerSummary) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Layers: %d → %d\n", summary.TotalA, summary.TotalB))
	sb.WriteString(fmt.Sprintf("  Shared:    %d layers (%s)\n",
		summary.SharedCount, FormatBytes(summary.SharedBytes)))

	if summary.OnlyInACount > 0 {
		sb.WriteString(fmt.Sprintf("  Only in A: %d layers (%s)\n",
			summary.OnlyInACount, FormatBytes(summary.OnlyInABytes)))
	}
	if summary.OnlyInBCount > 0 {
		sb.WriteString(fmt.Sprintf("  Only in B: %d layers (%s)\n",
			summary.OnlyInBCount, FormatBytes(summary.OnlyInBBytes)))
	}

	return strings.TrimRight(sb.String(), "\n")
}
