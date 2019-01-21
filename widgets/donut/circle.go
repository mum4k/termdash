package donut

// circle.go assists in calculation of points and angles on a circle.

import (
	"image"
	"math"

	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/numbers"
)

// startEndAngles given progress indicators and the desired start angle and
// direction, returns the starting and the ending angle of the partial circle
// that represents this progress.
func startEndAngles(current, total, startAngle, direction int) (start, end int) {
	const fullCircle = 360
	if total == 0 {
		return startAngle, startAngle
	}

	mult := float64(current) / float64(total)
	angleSize := numbers.Round(float64(360) * mult)

	if angleSize == fullCircle {
		return 0, fullCircle
	}
	end = startAngle + int(math.Round(float64(direction)*angleSize))

	if end < 0 {
		end += fullCircle
		if startAngle == 0 {
			startAngle = fullCircle
		}
		return end, startAngle
	}

	if end < startAngle {
		return end, startAngle
	}
	if end > fullCircle {
		end = end % fullCircle
	}
	return startAngle, end
}

// midAndRadius given an area of a braille canvas, determines the mid point in
// pixels and radius to draw the largest circle that fits.
// The circle's mid point is always positioned on the {0,1} pixel in the chosen
// cell so that any text inside of it can be visually centered.
func midAndRadius(ar image.Rectangle) (image.Point, int) {
	mid := image.Point{ar.Dx() / 2, ar.Dy() / 2}
	if mid.X%2 != 0 {
		mid.X -= 1
	}
	switch mid.Y % 4 {
	case 0:
		mid.Y += 1
	case 1:
	case 2:
		mid.Y -= 1
	case 3:
		mid.Y -= 2

	}

	// Calculate radius based on the smaller axis.
	var radius int
	if ar.Dx() < ar.Dy() {
		if mid.X < ar.Dx()/2 {
			radius = mid.X
		} else {
			radius = ar.Dx() - mid.X - 1
		}
	} else {
		if mid.Y < ar.Dy()/2 {
			radius = mid.Y
		} else {
			radius = ar.Dy() - mid.Y - 1
		}
	}
	return mid, radius
}

// availableCells given a radius returns the number of cells that are available
// within the circle and the coordinates of the first cell.
// These coordinates are for a normal (non-braille) canvas.
// That is the cells that do not contain any of the circle points. This is
// important since normal characters and braille characters cannot share the
// same cell.
func availableCells(mid image.Point, radius int) (int, image.Point) {
	if radius < 3 {
		return 0, image.Point{0, 0}
	}
	// Pixels available for the text only.
	// Subtract one for the circle itself.
	pixels := radius*2 - 1

	startPixel := image.Point{mid.X - pixels/2, mid.Y}
	startCell := image.Point{
		startPixel.X / braille.ColMult,
		mid.Y / braille.RowMult,
	}
	return pixels / braille.ColMult, startCell
}
