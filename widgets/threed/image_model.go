package threed

import (
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"math"
	"os"
)

// LoadImageModel reads an image file from disk and converts it into an extruded
// 3D model that can be spun in the ThreeD widget.
//
// PNG images with a transparent background work best — the alpha channel
// determines which pixels become geometry. For fully-opaque images (JPEG, flat
// PNG) dark/saturated pixels are treated as filled, which works well for dark
// logos on white backgrounds.
//
// Returns nil, err if the file cannot be opened or decoded.
func LoadImageModel(path string) (*Model, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return ImageToModel(img), nil
}

// ImageToModel converts any image.Image into an extruded 3D model suitable for
// the ThreeD widget spinner.
//
// The model resolution is defaultSymbolModelResolution so it matches the detail
// level used by the emoji-based spinner.  Pixel colors from the source image are
// preserved on the face geometry.
//
// Returns nil when the image contains no renderable pixels.
func ImageToModel(img image.Image) *Model {
	mask := smartImageMask(img, defaultSymbolModelResolution)
	trimmed, ok := trimGlyphMask(mask)
	if !ok {
		return nil
	}
	return buildGlyphMaskModel(trimmed)
}

// ImageToBrailleLines converts an image.Image into a slice of braille-encoded
// text lines forming a pixel-art preview.
//
// cols and rows control the output size in braille characters; each braille
// cell covers 2×4 pixels, so the preview is rendered at cols*2 × rows*4 pixel
// resolution.
//
// Returns nil when the image is nil or contains no renderable pixels.
func ImageToBrailleLines(img image.Image, cols, rows int) []string {
	if img == nil || cols <= 0 || rows <= 0 {
		return nil
	}
	mask := smartImageMask(img, defaultSymbolMaskResolution)
	trimmed, ok := trimGlyphMask(mask)
	if !ok {
		return nil
	}
	return maskToBrailleLines(trimmed, cols, rows)
}

// smartImageMask builds a glyphMask from an image by auto-detecting whether to
// use alpha-channel transparency or luminance-based thresholding.
//
// Images that contain at least one semi-transparent pixel use alpha thresholding
// (same path as the emoji pipeline).  Fully-opaque images fall back to
// luminance thresholding so dark logos on white backgrounds are handled cleanly.
func smartImageMask(img image.Image, resolution int) glyphMask {
	if imageHasTransparency(img) {
		return rasterImageMask(img, resolution)
	}
	return rasterOpaqueMask(img, resolution)
}

// imageHasTransparency reports whether the image contains any pixel whose alpha
// is less than fully-opaque (0xFFFF in 16-bit linear space).
//
// To keep this fast on large images we sample every N-th pixel where N grows
// with the image dimensions.
func imageHasTransparency(img image.Image) bool {
	b := img.Bounds()
	step := maxInt(1, maxInt(b.Dx(), b.Dy())/64)
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 0xFFFF {
				return true
			}
		}
	}
	return false
}

// rasterOpaqueMask converts a fully-opaque image into a glyphMask by treating
// pixels whose luminance is below ~80 % of full-white as filled geometry.
//
// This threshold handles typical logo artwork (dark ink on a white or light
// background) without clipping bright-colored marks.
func rasterOpaqueMask(img image.Image, resolution int) glyphMask {
	src := img.Bounds()
	srcW, srcH := src.Dx(), src.Dy()
	if srcW <= 0 || srcH <= 0 || resolution <= 0 {
		return glyphMask{}
	}

	scale := math.Max(float64(srcW), float64(srcH)) / float64(resolution)
	if scale < 1 {
		scale = 1
	}
	outW := maxInt(1, int(math.Ceil(float64(srcW)/scale)))
	outH := maxInt(1, int(math.Ceil(float64(srcH)/scale)))
	offsetX := (resolution - outW) / 2
	offsetY := (resolution - outH) / 2

	mask := glyphMask{
		Width:  resolution,
		Height: resolution,
		Filled: make([]bool, resolution*resolution),
		Colors: make([]Color, resolution*resolution),
	}

	for y := 0; y < outH; y++ {
		srcMinY := src.Min.Y + int(math.Floor(float64(y)*scale))
		srcMaxY := src.Min.Y + int(math.Ceil(float64(y+1)*scale))
		if srcMaxY <= srcMinY {
			srcMaxY = srcMinY + 1
		}
		if srcMaxY > src.Max.Y {
			srcMaxY = src.Max.Y
		}

		for x := 0; x < outW; x++ {
			srcMinX := src.Min.X + int(math.Floor(float64(x)*scale))
			srcMaxX := src.Min.X + int(math.Ceil(float64(x+1)*scale))
			if srcMaxX <= srcMinX {
				srcMaxX = srcMinX + 1
			}
			if srcMaxX > src.Max.X {
				srcMaxX = src.Max.X
			}

			var redSum, greenSum, blueSum uint64
			var samples, darkCount uint64

			for sy := srcMinY; sy < srcMaxY; sy++ {
				for sx := srcMinX; sx < srcMaxX; sx++ {
					r, g, b, _ := img.At(sx, sy).RGBA()
					redSum += uint64(r)
					greenSum += uint64(g)
					blueSum += uint64(b)
					samples++
					// Rec.601 luma in 16-bit linear space.
					// 52428 ≈ 0xCCCC ≈ 80% of 0xFFFF (full white).
					lum := (r*299 + g*587 + b*114) / 1000
					if lum < 52428 {
						darkCount++
					}
				}
			}
			if samples == 0 {
				continue
			}

			// Fill the cell when at least 25 % of its source pixels are non-white.
			if darkCount*4 < samples {
				continue
			}

			idx := (offsetY+y)*mask.Width + (offsetX + x)
			mask.Filled[idx] = true
			mask.Colors[idx] = Color{
				R: clampFloat(float64(redSum)/float64(samples)/65535.0, 0, 1),
				G: clampFloat(float64(greenSum)/float64(samples)/65535.0, 0, 1),
				B: clampFloat(float64(blueSum)/float64(samples)/65535.0, 0, 1),
			}
		}
	}

	return mask
}
