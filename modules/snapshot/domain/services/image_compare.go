package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/draw"
	"image/png"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)

// ImageCompareResult holds the result of comparing two screenshots.
type ImageCompareResult struct {
	Identical      bool    // Byte-level identical (same SHA-256)
	ScreenshotHash string  // SHA-256 hex of the current screenshot bytes
	DiffRatio      float64 // Fraction of pixels that differ (0.0 = identical, 1.0 = all different)
	DiffCount      int     // Absolute count of different pixels
	TotalPixels    int     // Total comparison area
	DiffLines      []int   // Sorted row indices with diffs (for Vision AI focus)
}

// maxYIQDelta is the maximum possible YIQ color distance (black vs white).
const maxYIQDelta = 35215.0

// CompareScreenshots performs a multi-stage comparison:
// 1. SHA-256 hash comparison (fast, byte-level identity)
// 2. Parallel PNG decode + YIQ perceptual pixel comparison with anti-aliasing detection
func CompareScreenshots(prevBytes, currBytes []byte, diffThreshold float64) (*ImageCompareResult, error) {
	currHash := sha256.Sum256(currBytes)
	currHashStr := hex.EncodeToString(currHash[:])

	// Stage 1: Hash comparison — if identical, no change
	prevHash := sha256.Sum256(prevBytes)
	if prevHash == currHash {
		return &ImageCompareResult{
			Identical:      true,
			ScreenshotHash: currHashStr,
		}, nil
	}

	// Stage 2: Parallel PNG decode
	var (
		prevImg, currImg image.Image
		prevErr, currErr error
		wg               sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		prevImg, prevErr = png.Decode(bytes.NewReader(prevBytes))
	}()
	go func() {
		defer wg.Done()
		currImg, currErr = png.Decode(bytes.NewReader(currBytes))
	}()
	wg.Wait()

	if prevErr != nil || currErr != nil {
		return &ImageCompareResult{
			Identical:      false,
			ScreenshotHash: currHashStr,
			DiffRatio:      1.0,
		}, nil
	}

	// Convert to NRGBA for direct .Pix access
	a := toNRGBA(prevImg)
	b := toNRGBA(currImg)

	result := compareNRGBA(a, b, diffThreshold)
	result.ScreenshotHash = currHashStr

	return result, nil
}

// HashScreenshot returns the SHA-256 hex hash of screenshot bytes.
func HashScreenshot(imgBytes []byte) string {
	h := sha256.Sum256(imgBytes)
	return hex.EncodeToString(h[:])
}

// compareNRGBA performs concurrent band-parallel YIQ perceptual comparison
// with anti-aliasing detection and early termination.
func compareNRGBA(a, b *image.NRGBA, diffThreshold float64) *ImageCompareResult {
	boundsA := a.Bounds()
	boundsB := b.Bounds()

	// Use intersection of both images
	minX := max(boundsA.Min.X, boundsB.Min.X)
	minY := max(boundsA.Min.Y, boundsB.Min.Y)
	maxX := min(boundsA.Max.X, boundsB.Max.X)
	maxY := min(boundsA.Max.Y, boundsB.Max.Y)

	w := maxX - minX
	h := maxY - minY

	if w <= 0 || h <= 0 {
		totalArea := max(boundsA.Dx()*boundsA.Dy(), boundsB.Dx()*boundsB.Dy())
		if totalArea <= 0 {
			return &ImageCompareResult{Identical: true}
		}
		return &ImageCompareResult{
			DiffRatio:   1.0,
			DiffCount:   totalArea,
			TotalPixels: totalArea,
		}
	}

	overlapPixels := w * h
	totalAreaA := boundsA.Dx() * boundsA.Dy()
	totalAreaB := boundsB.Dx() * boundsB.Dy()
	totalArea := max(totalAreaA, totalAreaB)
	nonOverlapping := totalArea - overlapPixels

	// Determine band count
	numBands := runtime.NumCPU()
	if numBands > 16 {
		numBands = 16
	}
	minBandHeight := 64
	if h/numBands < minBandHeight {
		numBands = h / minBandHeight
		if numBands < 1 {
			numBands = 1
		}
	}

	maxAllowedDiffs := int64(math.Ceil(diffThreshold * float64(totalArea)))
	var globalDiffCount int64
	var terminated int32 // 1 = early termination triggered

	type bandResult struct {
		diffs     int64
		diffLines []int
	}

	bandResults := make([]bandResult, numBands)

	var wg sync.WaitGroup
	wg.Add(numBands)

	bandHeight := h / numBands
	for band := 0; band < numBands; band++ {
		band := band
		startY := minY + band*bandHeight
		endY := startY + bandHeight
		if band == numBands-1 {
			endY = maxY // last band gets remaining rows
		}

		go func() {
			defer wg.Done()
			var localDiffs int64
			var localDiffLines []int

			for y := startY; y < endY; y++ {
				if atomic.LoadInt32(&terminated) != 0 {
					break
				}

				rowHasDiff := false
				for x := minX; x < maxX; x++ {
					r1, g1, b1, a1 := pixelAt(a, x, y)
					r2, g2, b2, a2 := pixelAt(b, x, y)

					// Fast path: identical bytes
					if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
						continue
					}

					delta := colorDelta(r1, g1, b1, a1, r2, g2, b2, a2, false)
					if delta <= 0.1 { // per-pixel threshold (matches pixelmatch default)
						continue
					}

					// Check anti-aliasing in both images
					if isAntialiased(a, x, y, w, h, minX, minY, b) ||
						isAntialiased(b, x, y, w, h, minX, minY, a) {
						continue
					}

					localDiffs++
					rowHasDiff = true
				}

				if rowHasDiff {
					localDiffLines = append(localDiffLines, y)
				}

				// Check early termination at row boundaries
				current := atomic.AddInt64(&globalDiffCount, localDiffs)
				localDiffs = 0
				if current+int64(nonOverlapping) > maxAllowedDiffs {
					atomic.StoreInt32(&terminated, 1)
					break
				}
			}

			// Flush any remaining local diffs
			if localDiffs > 0 {
				atomic.AddInt64(&globalDiffCount, localDiffs)
			}

			bandResults[band] = bandResult{
				diffLines: localDiffLines,
			}
		}()
	}

	wg.Wait()

	totalDiffs := int(atomic.LoadInt64(&globalDiffCount)) + nonOverlapping

	// Merge diff lines from all bands (already in row order since bands are sequential)
	var allDiffLines []int
	for _, br := range bandResults {
		allDiffLines = append(allDiffLines, br.diffLines...)
	}

	diffRatio := float64(totalDiffs) / float64(totalArea)
	if diffRatio > 1.0 {
		diffRatio = 1.0
	}

	return &ImageCompareResult{
		Identical:   false,
		DiffRatio:   diffRatio,
		DiffCount:   totalDiffs,
		TotalPixels: totalArea,
		DiffLines:   allDiffLines,
	}
}

// colorDelta computes the YIQ NTSC perceptual color distance between two pixels.
// Returns a normalized value in [0, 1]. If yOnly is true, only the Y (brightness)
// component is computed (used for anti-aliasing neighbor checks).
func colorDelta(r1, g1, b1, a1, r2, g2, b2, a2 uint8, yOnly bool) float64 {
	// Alpha blending against white (common case: both opaque → skip)
	var fr1, fg1, fb1, fr2, fg2, fb2 float64

	if a1 < 255 {
		fr1 = blendChannel(r1, a1)
		fg1 = blendChannel(g1, a1)
		fb1 = blendChannel(b1, a1)
	} else {
		fr1 = float64(r1)
		fg1 = float64(g1)
		fb1 = float64(b1)
	}

	if a2 < 255 {
		fr2 = blendChannel(r2, a2)
		fg2 = blendChannel(g2, a2)
		fb2 = blendChannel(b2, a2)
	} else {
		fr2 = float64(r2)
		fg2 = float64(g2)
		fb2 = float64(b2)
	}

	dr := fr1 - fr2
	dg := fg1 - fg2
	db := fb1 - fb2

	// Y component (brightness)
	y := dr*0.29889531 + dg*0.58662247 + db*0.11448223

	if yOnly {
		return y
	}

	// I and Q components (chrominance)
	i := dr*0.59597799 - dg*0.27417610 - db*0.32180189
	q := dr*0.21147017 - dg*0.52261711 + db*0.31114694

	delta := 0.5053*y*y + 0.299*i*i + 0.1957*q*q
	return delta / maxYIQDelta
}

// blendChannel alpha-premultiplies a color channel against a white background.
func blendChannel(ch, alpha uint8) float64 {
	a := float64(alpha) / 255.0
	return 255.0 + (float64(ch)-255.0)*a
}

// isAntialiased checks if a pixel is part of an anti-aliased edge by examining
// 8 neighbors in the source image. Uses the pixelmatch/odiff algorithm:
// brightness gradient analysis + sibling count.
func isAntialiased(img *image.NRGBA, x, y, w, h, offsetX, offsetY int, other *image.NRGBA) bool {
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()

	// Initialize min/max to 0 (matching pixelmatch): if all neighbor deltas are
	// on the same side of zero, the unset variable stays 0 → "no gradient" exit.
	var (
		zeroes              int
		minDelta, maxDelta  float64
		minX, minY2        int
		maxX2, maxY2       int
	)

	r0, g0, b0, a0 := pixelAt(img, x, y)

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			nx, ny := x+dx, y+dy
			if nx < 0 || ny < 0 || nx >= imgW || ny >= imgH {
				continue
			}

			r1, g1, b1, a1 := pixelAt(img, nx, ny)

			// Brightness-only delta
			delta := colorDelta(r0, g0, b0, a0, r1, g1, b1, a1, true)

			if delta == 0 {
				zeroes++
				if zeroes >= 3 {
					return false // flat region, not anti-aliased
				}
			} else if delta < minDelta {
				minDelta = delta
				minX = nx
				minY2 = ny
			} else if delta > maxDelta {
				maxDelta = delta
				maxX2 = nx
				maxY2 = ny
			}
		}
	}

	// No gradient (one side stayed at 0) → not anti-aliased
	if minDelta == 0 || maxDelta == 0 {
		return false
	}

	// Check if darkest/brightest neighbor has many siblings in both images
	return (hasManySiblings(img, minX, minY2, imgW, imgH) && hasManySiblings(other, minX, minY2, other.Bounds().Dx(), other.Bounds().Dy())) ||
		(hasManySiblings(img, maxX2, maxY2, imgW, imgH) && hasManySiblings(other, maxX2, maxY2, other.Bounds().Dx(), other.Bounds().Dy()))
}

// hasManySiblings checks if a pixel has >= 3 neighbors with the same color,
// using packed uint32 comparison for speed.
func hasManySiblings(img *image.NRGBA, x, y, w, h int) bool {
	target := pixelUint32At(img, x, y)
	count := 0

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			nx, ny := x+dx, y+dy
			if nx < 0 || ny < 0 || nx >= w || ny >= h {
				continue
			}

			if pixelUint32At(img, nx, ny) == target {
				count++
				if count >= 3 {
					return true
				}
			}
		}
	}

	return false
}

// toNRGBA converts any image.Image to *image.NRGBA. Zero-copy if already NRGBA.
func toNRGBA(img image.Image) *image.NRGBA {
	if nrgba, ok := img.(*image.NRGBA); ok {
		return nrgba
	}

	bounds := img.Bounds()
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, img, bounds.Min, draw.Src)
	return nrgba
}

// pixelAt reads RGBA values directly from the NRGBA .Pix slice, avoiding
// the At() interface dispatch overhead.
func pixelAt(img *image.NRGBA, x, y int) (r, g, b, a uint8) {
	offset := (y-img.Rect.Min.Y)*img.Stride + (x-img.Rect.Min.X)*4
	return img.Pix[offset], img.Pix[offset+1], img.Pix[offset+2], img.Pix[offset+3]
}

// pixelUint32At reads 4 RGBA bytes as a packed uint32 for fast equality checks.
func pixelUint32At(img *image.NRGBA, x, y int) uint32 {
	offset := (y-img.Rect.Min.Y)*img.Stride + (x-img.Rect.Min.X)*4
	return uint32(img.Pix[offset])<<24 |
		uint32(img.Pix[offset+1])<<16 |
		uint32(img.Pix[offset+2])<<8 |
		uint32(img.Pix[offset+3])
}
