// Copyright 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package segdisp provides utilities used by all segment display types.
package segdisp

import (
	"fmt"
	"image"
	"math"

	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
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

// SegmentSize given an area for the display segment determines the size of
// individual segments, i.e. the width of a vertical or the height of a
// horizontal segment.
func SegmentSize(ar image.Rectangle) int {
	// widthPerc is the relative width of a segment to the width of the canvas.
	const widthPerc = 9
	s := int(math.Round(float64(ar.Dx()) * widthPerc / 100))
	if s > 3 && s%2 == 0 {
		// Segments with odd number of pixels in their width/height look
		// better, since the spike at the top of their slopes has only one
		// pixel.
		s++
	}
	return s
}
