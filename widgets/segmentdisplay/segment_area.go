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

package segmentdisplay

// segment_area.go contains code that determines how many segments we can fit
// in the canvas.

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/private/segdisp"
)

// segArea contains information about the area that will contain the segments.
type segArea struct {
	// segment is the area for one segment.
	segment image.Rectangle
	// canFit is the number of segments we can fit on the canvas.
	canFit int
	// gapPixels is the size of gaps between segments in pixels.
	gapPixels int
	// gaps is the number of gaps that will be drawn.
	gaps int
}

// needArea returns the complete area required for all the segments that we can
// fit and any gaps.
func (sa *segArea) needArea() image.Rectangle {
	return image.Rect(
		0,
		0,
		sa.segment.Dx()*sa.canFit+sa.gaps*sa.gapPixels,
		sa.segment.Dy(),
	)
}

// newSegArea calculates the area for segments given available canvas area,
// length of the text to be displayed and the size of gap between segments
func newSegArea(cvsAr image.Rectangle, textLen, gapPercent int) (*segArea, error) {
	segAr, err := segdisp.Required(cvsAr)
	if err != nil {
		return nil, fmt.Errorf("sixteen.Required => %v", err)
	}
	gapPixels := segAr.Dy() * gapPercent / 100

	var (
		gaps   int
		canFit int
		taken  int
	)
	for i := 0; ; {
		taken += segAr.Dx()

		if taken > cvsAr.Dx() {
			break
		}
		canFit++

		// Don't insert gaps after the last segment in the text or the last
		// segment we can fit.
		if gapPixels == 0 || i == textLen-1 {
			continue
		}

		remaining := cvsAr.Dx() - taken
		// Only insert gaps if we can still fit one more segment with the gap.
		if remaining >= gapPixels+segAr.Dx() {
			taken += gapPixels
			gaps++
		} else {
			// Gap is needed but doesn't fit together with the next segment.
			// So insert neither.
			break
		}
		i++
	}
	return &segArea{
		segment:   segAr,
		canFit:    canFit,
		gapPixels: gapPixels,
		gaps:      gaps,
	}, nil
}

// maximizeFit finds the largest individual segment size that enables us to fit
// the most characters onto a canvas with the provided area. Returns the area
// required for a single segment and the number of segments we can fit.
func maximizeFit(cvsAr image.Rectangle, textLen, gapPercent int) (*segArea, error) {
	var bestSegAr *segArea
	for height := cvsAr.Dy(); height >= segdisp.MinRows; height-- {
		cvsAr := image.Rect(cvsAr.Min.X, cvsAr.Min.Y, cvsAr.Max.X, cvsAr.Min.Y+height)
		segAr, err := newSegArea(cvsAr, textLen, gapPercent)
		if err != nil {
			return nil, err
		}

		if textLen > 0 && segAr.canFit >= textLen {
			return segAr, nil
		}
		bestSegAr = segAr
	}
	return bestSegAr, nil
}
