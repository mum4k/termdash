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
// provided area. Can return a zero area.
func ExcludeBorder(area image.Rectangle) image.Rectangle {
	// If the area dimensions are smaller than this, subtracting a point for the
	// border on each of its sides results in a zero area.
	const minDim = 3
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
