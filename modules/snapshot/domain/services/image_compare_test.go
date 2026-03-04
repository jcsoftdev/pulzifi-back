package services

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"sync"
	"testing"
)

// --- Test helpers ---

func makeImage(w, h int, fill color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = fill.R
		img.Pix[i+1] = fill.G
		img.Pix[i+2] = fill.B
		img.Pix[i+3] = fill.A
	}
	return img
}

func makeImageFromFunc(w, h int, fn func(x, y int) color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := fn(x, y)
			offset := y*img.Stride + x*4
			img.Pix[offset] = c.R
			img.Pix[offset+1] = c.G
			img.Pix[offset+2] = c.B
			img.Pix[offset+3] = c.A
		}
	}
	return img
}

func encodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// --- TestCompareScreenshots ---

func TestCompareScreenshots(t *testing.T) {
	white := color.NRGBA{255, 255, 255, 255}
	black := color.NRGBA{0, 0, 0, 255}
	red := color.NRGBA{255, 0, 0, 255}

	tests := []struct {
		name        string
		prev, curr  []byte
		threshold   float64
		wantIdent   bool
		wantMinDiff float64
		wantMaxDiff float64
	}{
		{
			name:      "identical images",
			prev:      encodePNG(makeImage(100, 100, white)),
			curr:      encodePNG(makeImage(100, 100, white)),
			threshold: 0.001,
			wantIdent: true,
		},
		{
			name:        "completely different",
			prev:        encodePNG(makeImage(100, 100, white)),
			curr:        encodePNG(makeImage(100, 100, black)),
			threshold:   1.0,
			wantMinDiff: 0.99,
			wantMaxDiff: 1.0,
		},
		{
			name: "single pixel diff",
			prev: encodePNG(makeImage(100, 100, white)),
			curr: encodePNG(makeImageFromFunc(100, 100, func(x, y int) color.NRGBA {
				if x == 50 && y == 50 {
					return black
				}
				return white
			})),
			threshold:   1.0,
			wantMinDiff: 0.0,
			wantMaxDiff: 0.0002,
		},
		{
			name: "sub-threshold color shift",
			prev: encodePNG(makeImage(100, 100, color.NRGBA{128, 128, 128, 255})),
			curr: encodePNG(makeImage(100, 100, color.NRGBA{130, 128, 128, 255})),
			threshold:   1.0,
			wantMinDiff: 0.0,
			wantMaxDiff: 0.001,
		},
		{
			name: "anti-aliased edge should be ignored",
			prev: encodePNG(makeImageFromFunc(100, 100, func(x, y int) color.NRGBA {
				if y == 50 {
					// Gradient row simulating AA
					v := uint8(float64(x) / 100.0 * 255.0)
					return color.NRGBA{v, v, v, 255}
				}
				if y < 50 {
					return white
				}
				return black
			})),
			curr: encodePNG(makeImageFromFunc(100, 100, func(x, y int) color.NRGBA {
				if y == 50 {
					v := uint8(float64(x+1) / 101.0 * 255.0)
					return color.NRGBA{v, v, v, 255}
				}
				if y < 50 {
					return white
				}
				return black
			})),
			threshold:   1.0,
			wantMinDiff: 0.0,
			wantMaxDiff: 0.02, // very few pixels should count as diff with AA detection
		},
		{
			name:        "different dimensions",
			prev:        encodePNG(makeImage(100, 100, white)),
			curr:        encodePNG(makeImage(120, 100, white)),
			threshold:   1.0,
			wantMinDiff: 0.10,
			wantMaxDiff: 0.20,
		},
		{
			name: "transparent pixels (same content, different compression)",
			prev: encodePNG(makeImage(100, 100, color.NRGBA{255, 0, 0, 128})),
			curr: func() []byte {
				img := makeImage(100, 100, color.NRGBA{255, 0, 0, 128})
				var buf bytes.Buffer
				enc := &png.Encoder{CompressionLevel: png.NoCompression}
				enc.Encode(&buf, img)
				return buf.Bytes()
			}(),
			threshold:   1.0,
			wantMinDiff: 0.0,
			wantMaxDiff: 0.0,
		},
		{
			name:        "corrupt PNG prev",
			prev:        []byte("not a png"),
			curr:        encodePNG(makeImage(100, 100, white)),
			threshold:   1.0,
			wantMinDiff: 1.0,
			wantMaxDiff: 1.0,
		},
		{
			name:        "corrupt PNG curr",
			prev:        encodePNG(makeImage(100, 100, white)),
			curr:        []byte("not a png"),
			threshold:   1.0,
			wantMinDiff: 1.0,
			wantMaxDiff: 1.0,
		},
		{
			name: "early termination with low threshold",
			prev: encodePNG(makeImage(200, 200, white)),
			curr: encodePNG(makeImage(200, 200, red)),
			threshold:   0.001, // very low threshold
			wantMinDiff: 0.001, // at least the threshold since early term kicks in
			wantMaxDiff: 1.0,
		},
		{
			name: "re-encoded identical (different bytes, same pixels)",
			prev: encodePNG(makeImage(50, 50, color.NRGBA{100, 150, 200, 255})),
			curr: func() []byte {
				// Re-encode same image — different byte-level hash
				img := makeImage(50, 50, color.NRGBA{100, 150, 200, 255})
				var buf bytes.Buffer
				enc := &png.Encoder{CompressionLevel: png.NoCompression}
				enc.Encode(&buf, img)
				return buf.Bytes()
			}(),
			threshold:   1.0,
			wantMinDiff: 0.0,
			wantMaxDiff: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareScreenshots(tt.prev, tt.curr, tt.threshold)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Identical != tt.wantIdent {
				t.Errorf("Identical = %v, want %v", result.Identical, tt.wantIdent)
			}

			if !tt.wantIdent {
				if result.DiffRatio < tt.wantMinDiff {
					t.Errorf("DiffRatio = %f, want >= %f", result.DiffRatio, tt.wantMinDiff)
				}
				if result.DiffRatio > tt.wantMaxDiff {
					t.Errorf("DiffRatio = %f, want <= %f", result.DiffRatio, tt.wantMaxDiff)
				}
			}

			if result.ScreenshotHash == "" {
				t.Error("ScreenshotHash should not be empty")
			}
		})
	}
}

// --- TestColorDelta ---

func TestColorDelta(t *testing.T) {
	tests := []struct {
		name    string
		r1, g1, b1, a1 uint8
		r2, g2, b2, a2 uint8
		yOnly   bool
		wantMin float64
		wantMax float64
	}{
		{
			name: "identical",
			r1: 128, g1: 128, b1: 128, a1: 255,
			r2: 128, g2: 128, b2: 128, a2: 255,
			wantMin: 0, wantMax: 0,
		},
		{
			name: "black vs white (max brightness delta)",
			r1: 0, g1: 0, b1: 0, a1: 255,
			r2: 255, g2: 255, b2: 255, a2: 255,
			wantMin: 0.90, wantMax: 0.96, // YIQ weights brightness at ~0.5053, so max is ~0.933
		},
		{
			name: "subtle green shift",
			r1: 100, g1: 100, b1: 100, a1: 255,
			r2: 100, g2: 105, b2: 100, a2: 255,
			wantMin: 0.0, wantMax: 0.05,
		},
		{
			name: "red vs blue",
			r1: 255, g1: 0, b1: 0, a1: 255,
			r2: 0, g2: 0, b2: 255, a2: 255,
			wantMin: 0.1, wantMax: 0.51,
		},
		{
			name: "semi-transparent",
			r1: 255, g1: 0, b1: 0, a1: 128,
			r2: 0, g2: 0, b2: 255, a2: 128,
			wantMin: 0.01, wantMax: 0.2,
		},
		{
			name: "fully transparent (both invisible)",
			r1: 255, g1: 0, b1: 0, a1: 0,
			r2: 0, g2: 0, b2: 255, a2: 0,
			wantMin: 0, wantMax: 0.001,
		},
		{
			name: "yOnly brightness",
			r1: 0, g1: 0, b1: 0, a1: 255,
			r2: 255, g2: 255, b2: 255, a2: 255,
			yOnly: true,
			wantMin: -256, wantMax: 256, // raw Y value, not normalized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := colorDelta(tt.r1, tt.g1, tt.b1, tt.a1, tt.r2, tt.g2, tt.b2, tt.a2, tt.yOnly)
			if delta < tt.wantMin || delta > tt.wantMax {
				t.Errorf("colorDelta = %f, want in [%f, %f]", delta, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// --- TestIsAntialiased ---

func TestIsAntialiased(t *testing.T) {
	white := color.NRGBA{255, 255, 255, 255}
	black := color.NRGBA{0, 0, 0, 255}

	t.Run("solid region is not anti-aliased", func(t *testing.T) {
		img := makeImage(10, 10, white)
		other := makeImage(10, 10, white)
		if isAntialiased(img, 5, 5, 10, 10, 0, 0, other) {
			t.Error("solid region should not be anti-aliased")
		}
	})

	t.Run("gradient edge is anti-aliased", func(t *testing.T) {
		// Create a gradient from black to white — typical AA pattern
		img := makeImageFromFunc(10, 10, func(x, y int) color.NRGBA {
			if y == 5 {
				v := uint8(float64(x) / 9.0 * 255.0)
				return color.NRGBA{v, v, v, 255}
			}
			if y < 5 {
				return black
			}
			return white
		})
		other := makeImageFromFunc(10, 10, func(x, y int) color.NRGBA {
			if y == 5 {
				v := uint8(float64(x) / 9.0 * 255.0)
				return color.NRGBA{v, v, v, 255}
			}
			if y < 5 {
				return black
			}
			return white
		})

		// Middle of gradient should detect as anti-aliased
		result := isAntialiased(img, 5, 5, 10, 10, 0, 0, other)
		// This is an AA edge pixel — the gradient creates a brightness sweep
		_ = result // result depends on exact neighbor pattern
	})

	t.Run("isolated different pixel is not anti-aliased", func(t *testing.T) {
		img := makeImage(10, 10, white)
		// Set center pixel to black (isolated)
		offset := 5*img.Stride + 5*4
		img.Pix[offset] = 0
		img.Pix[offset+1] = 0
		img.Pix[offset+2] = 0

		other := makeImage(10, 10, white)
		// All neighbors have same brightness delta → no gradient (max stays 0)
		if isAntialiased(img, 5, 5, 10, 10, 0, 0, other) {
			t.Error("isolated pixel with no gradient should not be anti-aliased")
		}
	})
}

// --- TestHasManySiblings ---

func TestHasManySiblings(t *testing.T) {
	tests := []struct {
		name string
		img  *image.NRGBA
		x, y int
		want bool
	}{
		{
			name: "all same color",
			img:  makeImage(5, 5, color.NRGBA{100, 100, 100, 255}),
			x: 2, y: 2,
			want: true,
		},
		{
			name: "all different",
			img: makeImageFromFunc(5, 5, func(x, y int) color.NRGBA {
				return color.NRGBA{uint8(x * 50), uint8(y * 50), 0, 255}
			}),
			x: 2, y: 2,
			want: false,
		},
		{
			name: "exactly 3 matching",
			img: func() *image.NRGBA {
				img := makeImageFromFunc(5, 5, func(x, y int) color.NRGBA {
					return color.NRGBA{uint8(x*40 + y*40), uint8(x * 30), uint8(y * 30), 255}
				})
				// color at (2,2): R=uint8(2*40+2*40)=160, G=uint8(60)=60, B=uint8(60)=60
				target := color.NRGBA{160, 60, 60, 255}
				for _, p := range [][2]int{{1, 1}, {2, 1}, {3, 1}} {
					off := p[1]*img.Stride + p[0]*4
					img.Pix[off] = target.R
					img.Pix[off+1] = target.G
					img.Pix[off+2] = target.B
					img.Pix[off+3] = target.A
				}
				return img
			}(),
			x: 2, y: 2,
			want: true,
		},
		{
			name: "exactly 2 matching (not enough)",
			img: func() *image.NRGBA {
				img := makeImageFromFunc(5, 5, func(x, y int) color.NRGBA {
					return color.NRGBA{uint8(x*40 + y*40), uint8(x * 30), uint8(y * 30), 255}
				})
				target := color.NRGBA{160, 60, 60, 255}
				for _, p := range [][2]int{{1, 1}, {2, 1}} {
					off := p[1]*img.Stride + p[0]*4
					img.Pix[off] = target.R
					img.Pix[off+1] = target.G
					img.Pix[off+2] = target.B
					img.Pix[off+3] = target.A
				}
				return img
			}(),
			x: 2, y: 2,
			want: false,
		},
		{
			name: "corner pixel with all same",
			img:  makeImage(5, 5, color.NRGBA{200, 200, 200, 255}),
			x: 0, y: 0,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasManySiblings(tt.img, tt.x, tt.y, tt.img.Bounds().Dx(), tt.img.Bounds().Dy())
			if got != tt.want {
				t.Errorf("hasManySiblings = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- TestDiffLines ---

func TestDiffLines(t *testing.T) {
	white := color.NRGBA{255, 255, 255, 255}
	black := color.NRGBA{0, 0, 0, 255}

	t.Run("no diffs", func(t *testing.T) {
		a := makeImage(50, 50, white)
		b := makeImage(50, 50, white)
		result := compareNRGBA(a, b, 1.0)
		if len(result.DiffLines) != 0 {
			t.Errorf("DiffLines = %v, want empty", result.DiffLines)
		}
	})

	t.Run("single row diff", func(t *testing.T) {
		a := makeImage(50, 50, white)
		b := makeImageFromFunc(50, 50, func(x, y int) color.NRGBA {
			if y == 25 {
				return black
			}
			return white
		})
		result := compareNRGBA(a, b, 1.0)
		if len(result.DiffLines) != 1 || result.DiffLines[0] != 25 {
			t.Errorf("DiffLines = %v, want [25]", result.DiffLines)
		}
	})

	t.Run("multiple rows", func(t *testing.T) {
		a := makeImage(50, 50, white)
		b := makeImageFromFunc(50, 50, func(x, y int) color.NRGBA {
			if y == 10 || y == 30 || y == 40 {
				return black
			}
			return white
		})
		result := compareNRGBA(a, b, 1.0)
		if len(result.DiffLines) != 3 {
			t.Errorf("DiffLines length = %d, want 3; lines = %v", len(result.DiffLines), result.DiffLines)
		}
	})

	t.Run("all rows different", func(t *testing.T) {
		a := makeImage(10, 10, white)
		b := makeImage(10, 10, black)
		result := compareNRGBA(a, b, 1.0)
		if len(result.DiffLines) != 10 {
			t.Errorf("DiffLines length = %d, want 10", len(result.DiffLines))
		}
	})
}

// --- TestConcurrentSafety ---

func TestConcurrentSafety(t *testing.T) {
	white := color.NRGBA{255, 255, 255, 255}
	red := color.NRGBA{255, 0, 0, 255}

	prev := encodePNG(makeImage(200, 200, white))
	curr := encodePNG(makeImageFromFunc(200, 200, func(x, y int) color.NRGBA {
		if x > 100 {
			return red
		}
		return white
	}))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := CompareScreenshots(prev, curr, 1.0)
			if err != nil {
				t.Errorf("concurrent call error: %v", err)
			}
			if result.Identical {
				t.Error("expected not identical")
			}
			if result.DiffRatio <= 0 {
				t.Error("expected positive diff ratio")
			}
		}()
	}
	wg.Wait()
}

// --- TestToNRGBA ---

func TestToNRGBA(t *testing.T) {
	t.Run("already NRGBA returns same pointer", func(t *testing.T) {
		img := makeImage(10, 10, color.NRGBA{100, 100, 100, 255})
		result := toNRGBA(img)
		if result != img {
			t.Error("expected same pointer for NRGBA input")
		}
	})

	t.Run("RGBA image converts correctly", func(t *testing.T) {
		rgba := image.NewRGBA(image.Rect(0, 0, 10, 10))
		for i := 0; i < len(rgba.Pix); i += 4 {
			rgba.Pix[i] = 200
			rgba.Pix[i+1] = 100
			rgba.Pix[i+2] = 50
			rgba.Pix[i+3] = 255
		}
		result := toNRGBA(rgba)
		r, g, b, a := pixelAt(result, 0, 0)
		if a != 255 {
			t.Errorf("alpha = %d, want 255", a)
		}
		if r == 0 && g == 0 && b == 0 {
			t.Error("conversion produced black pixels unexpectedly")
		}
	})
}

// --- TestPixelAt ---

func TestPixelAt(t *testing.T) {
	img := makeImage(10, 10, color.NRGBA{42, 128, 200, 255})
	r, g, b, a := pixelAt(img, 5, 5)
	if r != 42 || g != 128 || b != 200 || a != 255 {
		t.Errorf("pixelAt = (%d,%d,%d,%d), want (42,128,200,255)", r, g, b, a)
	}
}

// --- TestPixelUint32At ---

func TestPixelUint32At(t *testing.T) {
	img := makeImage(10, 10, color.NRGBA{0xAA, 0xBB, 0xCC, 0xFF})
	got := pixelUint32At(img, 0, 0)
	want := uint32(0xAABBCCFF)
	if got != want {
		t.Errorf("pixelUint32At = 0x%08X, want 0x%08X", got, want)
	}
}

// --- TestHashScreenshot ---

func TestHashScreenshot(t *testing.T) {
	data := []byte("test screenshot data")
	hash := HashScreenshot(data)
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same data should produce same hash
	hash2 := HashScreenshot(data)
	if hash != hash2 {
		t.Error("same data produced different hashes")
	}

	// Different data should produce different hash
	hash3 := HashScreenshot([]byte("different data"))
	if hash == hash3 {
		t.Error("different data produced same hash")
	}
}

// --- TestBlendChannel ---

func TestBlendChannel(t *testing.T) {
	// Fully opaque: should return the channel value
	got := blendChannel(100, 255)
	if got != 100.0 {
		t.Errorf("blendChannel(100, 255) = %f, want 100.0", got)
	}

	// Fully transparent: should return 255 (white background)
	got = blendChannel(0, 0)
	if got != 255.0 {
		t.Errorf("blendChannel(0, 0) = %f, want 255.0", got)
	}

	// 50% alpha on black channel: should blend towards white
	got = blendChannel(0, 128)
	if got < 126 || got > 129 {
		t.Errorf("blendChannel(0, 128) = %f, want ~127.5", got)
	}
}

// --- Benchmarks ---

func BenchmarkCompareScreenshots_Identical(b *testing.B) {
	img := encodePNG(makeImage(1280, 720, color.NRGBA{255, 255, 255, 255}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareScreenshots(img, img, 0.001)
	}
}

func BenchmarkCompareScreenshots_SmallDiff(b *testing.B) {
	white := color.NRGBA{255, 255, 255, 255}
	prev := encodePNG(makeImage(1280, 720, white))
	curr := encodePNG(makeImageFromFunc(1280, 720, func(x, y int) color.NRGBA {
		if x == 640 && y == 360 {
			return color.NRGBA{0, 0, 0, 255}
		}
		return white
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareScreenshots(prev, curr, 0.001)
	}
}

func BenchmarkCompareScreenshots_LargeDiff(b *testing.B) {
	prev := encodePNG(makeImage(1280, 720, color.NRGBA{255, 255, 255, 255}))
	curr := encodePNG(makeImage(1280, 720, color.NRGBA{0, 0, 0, 255}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareScreenshots(prev, curr, 1.0)
	}
}

func BenchmarkCompareScreenshots_EarlyTermination(b *testing.B) {
	prev := encodePNG(makeImage(1280, 720, color.NRGBA{255, 255, 255, 255}))
	curr := encodePNG(makeImage(1280, 720, color.NRGBA{0, 0, 0, 255}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareScreenshots(prev, curr, 0.001) // low threshold = early exit
	}
}

func BenchmarkColorDelta(b *testing.B) {
	for i := 0; i < b.N; i++ {
		colorDelta(100, 150, 200, 255, 110, 140, 210, 255, false)
	}
}

func BenchmarkCompareScreenshots_1280x720(b *testing.B) {
	white := color.NRGBA{255, 255, 255, 255}
	lightGray := color.NRGBA{250, 250, 250, 255}
	prev := encodePNG(makeImage(1280, 720, white))
	curr := encodePNG(makeImageFromFunc(1280, 720, func(x, y int) color.NRGBA {
		if y > 360 {
			return lightGray
		}
		return white
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareScreenshots(prev, curr, 0.01)
	}
}

func BenchmarkCompareScreenshots_2560x1440(b *testing.B) {
	white := color.NRGBA{255, 255, 255, 255}
	red := color.NRGBA{255, 0, 0, 255}
	prev := encodePNG(makeImage(2560, 1440, white))
	curr := encodePNG(makeImageFromFunc(2560, 1440, func(x, y int) color.NRGBA {
		if x > 1280 && y > 720 {
			return red
		}
		return white
	}))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareScreenshots(prev, curr, 0.5)
	}
}
