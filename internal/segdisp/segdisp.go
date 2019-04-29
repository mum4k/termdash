// Package segdisp provides utilities used by all segment display types.
package segdisp

import (
	"fmt"
	"image"
	"math"

	"github.com/mum4k/termdash/internal/area"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/braille"
)

// Minimum valid size of a cell canvas in order to draw a segment display.
const (
	// MinCols is the smallest valid amount of columns in a cell area.
	MinCols = 6
	// MinRowPixels is the smallest valid amount of rows in a cell area.
	MinRows = 5
)

// aspectRatio is the desired aspect ratio of a single segment display.
var aspectRatio = image.Point{3, 5}

// Required when given an area of cells, returns either an area of the same
// size or a smaller area that is required to draw one segment display (i.e.
// one character).
// Returns a smaller area when the provided area didn't have the required
// aspect ratio.
// Returns an error if the area is too small to draw a segment display, i.e.
// smaller than MinCols x MinRows.
func Required(cellArea image.Rectangle) (image.Rectangle, error) {
	if cols, rows := cellArea.Dx(), cellArea.Dy(); cols < MinCols || rows < MinRows {
		return image.ZR, fmt.Errorf("cell area %v is too small to draw the segment display, has %dx%d cells, need at least %dx%d cells",
			cellArea, cols, rows, MinCols, MinRows)
	}

	bcAr := image.Rect(cellArea.Min.X, cellArea.Min.Y, cellArea.Max.X*braille.ColMult, cellArea.Max.Y*braille.RowMult)
	bcArAdj := area.WithRatio(bcAr, aspectRatio)

	needCols := int(math.Ceil(float64(bcArAdj.Dx()) / braille.ColMult))
	needRows := int(math.Ceil(float64(bcArAdj.Dy()) / braille.RowMult))
	needAr := image.Rect(cellArea.Min.X, cellArea.Min.Y, cellArea.Min.X+needCols, cellArea.Min.Y+needRows)
	return needAr, nil
}

// ToBraille converts the canvas into a braille canvas and returns a pixel area
// with aspect ratio adjusted for the segment display.
func ToBraille(cvs *canvas.Canvas) (*braille.Canvas, image.Rectangle, error) {
	ar, err := Required(cvs.Area())
	if err != nil {
		return nil, image.ZR, fmt.Errorf("Required => %v", err)
	}

	bc, err := braille.New(ar)
	if err != nil {
		return nil, image.ZR, fmt.Errorf("braille.New => %v", err)
	}
	return bc, area.WithRatio(bc.Area(), aspectRatio), nil
}
