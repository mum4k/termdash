// Package area provides functions working with image areas.
package area

import (
	"fmt"
	"image"
)

// Size returns the size of the provided area.
func Size(area image.Rectangle) image.Point {
	return image.Point{
		area.Dx(),
		area.Dy(),
	}
}

// FromSize returns the corresponding area for the provided size.
func FromSize(size image.Point) (image.Rectangle, error) {
	if size.X < 0 || size.Y < 0 {
		return image.Rectangle{}, fmt.Errorf("cannot convert zero or negative size to an area, got: %+v", size)
	}
	return image.Rect(0, 0, size.X, size.Y), nil
}

// HSplit returns two new areas created by splitting the provided area in the
// middle along the horizontal axis. Can return zero size areas.
func HSplit(area image.Rectangle) (image.Rectangle, image.Rectangle) {
	height := area.Dy() / 2
	if height == 0 {
		return image.ZR, image.ZR
	}
	return image.Rect(area.Min.X, area.Min.Y, area.Max.X, area.Min.Y+height),
		image.Rect(area.Min.X, area.Min.Y+height, area.Max.X, area.Max.Y)
}

// VSplit returns two new areas created by splitting the provided area in the
// middle along the vertical axis. Can return zero size areas.
func VSplit(area image.Rectangle) (image.Rectangle, image.Rectangle) {
	width := area.Dx() / 2
	if width == 0 {
		return image.ZR, image.ZR
	}
	return image.Rect(area.Min.X, area.Min.Y, area.Min.X+width, area.Max.Y),
		image.Rect(area.Min.X+width, area.Min.Y, area.Max.X, area.Max.Y)
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ExcludeBorder returns a new area created by subtracting a border around the
// provided area. Return the zero area if there isn't enough space to exclude
// the border.
func ExcludeBorder(area image.Rectangle) image.Rectangle {
	// If the area dimensions are smaller than this, subtracting a point for the
	// border on each of its sides results in a zero area.
	const minDim = 2
	if area.Dx() < minDim || area.Dy() < minDim {
		return image.ZR
	}
	return image.Rect(
		abs(area.Min.X+1),
		abs(area.Min.Y+1),
		abs(area.Max.X-1),
		abs(area.Max.Y-1),
	)
}

// findGCF finds the greatest common factor of two integers.
func findGCF(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}

	// https://en.wikipedia.org/wiki/Euclidean_algorithm
	for {
		rem := a % b
		a = b
		b = rem

		if b == 0 {
			break
		}
	}
	return a
}

// simplifyRatio simplifies the given ratio.
func simplifyRatio(ratio image.Point) image.Point {
	gcf := findGCF(ratio.X, ratio.Y)
	if gcf == 0 {
		return image.ZP
	}
	return image.Point{
		X: ratio.X / gcf,
		Y: ratio.Y / gcf,
	}
}

// WithRatio returns the largest area that has the requested ratio but is
// either equal or smaller than the provided area. Returns zero area if the
// area or the ratio are zero, or if there is no such area.
func WithRatio(area image.Rectangle, ratio image.Point) image.Rectangle {
	ratio = simplifyRatio(ratio)
	if area == image.ZR || ratio == image.ZP {
		return image.ZR
	}

	wFact := area.Dx() / ratio.X
	hFact := area.Dy() / ratio.Y

	var fact int
	if wFact < hFact {
		fact = wFact
	} else {
		fact = hFact
	}
	return image.Rect(
		area.Min.X,
		area.Min.Y,
		ratio.X*fact+area.Min.X,
		ratio.Y*fact+area.Min.Y,
	)
}
