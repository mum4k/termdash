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

package dotseg

// attributes.go calculates attributes needed when determining placement of
// segments.

import (
	"fmt"
	"image"
	"math"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/internal/alignfor"
	"github.com/mum4k/termdash/internal/area"
	"github.com/mum4k/termdash/internal/segdisp"
)

// segmentSize given an area for the display determines the size of individual
// segments, i.e. the width of a vertical or the height of a horizontal
// segment.
func segmentSize(ar image.Rectangle) int {
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

// attributes contains attributes needed to draw the segment display.
// Refer to doc/segment_placement.svg for a visual aid and explanation of the
// usage of the square roots.
type attributes struct {
	// bcAr is the area the attributes were created for.
	bcAr image.Rectangle

	// segSize is the width of a vertical or height of a horizontal segment.
	segSize int
}

// newAttributes calculates attributes needed to place the segments for the
// provided pixel area.
func newAttributes(bcAr image.Rectangle) *attributes {
	// Dots have double width of normal segments to fill more space in the
	// segment display.
	segSize := segdisp.SegmentSize(bcAr) * 2

	return &attributes{
		bcAr:    bcAr,
		segSize: segSize,
	}
}

// segArea returns the area for the specified segment.
func (a *attributes) segArea(seg Segment) (image.Rectangle, error) {
	// An area representing the dot which gets aligned and moved into position
	// below.
	dotAr := image.Rect(
		a.bcAr.Min.X,
		a.bcAr.Min.Y,
		a.bcAr.Min.X+a.segSize,
		a.bcAr.Min.Y+a.segSize,
	)
	mid, err := alignfor.Rectangle(a.bcAr, dotAr, align.HorizontalCenter, align.VerticalMiddle)
	if err != nil {
		return image.ZR, err
	}

	// moveBySize is the multiplier of segment size to determine by how many
	// pixels to move D1 and D2 up and down from the center.
	const moveBySize = 1.5
	moveBy := int(math.Round(moveBySize * float64(a.segSize)))
	switch seg {
	case D1:
		moved, err := area.MoveUp(mid, moveBy)
		if err != nil {
			return image.ZR, err
		}
		return moved, nil

	case D2:
		moved, err := area.MoveDown(mid, moveBy)
		if err != nil {
			return image.ZR, err
		}
		return moved, nil

	case D3:
		bot, err := alignfor.Rectangle(a.bcAr, dotAr, align.HorizontalCenter, align.VerticalBottom)
		if err != nil {
			return image.ZR, err
		}

		return bot, nil

	default:
		return image.ZR, fmt.Errorf("cannot calculate area for %v(%d)", seg, seg)
	}
}
