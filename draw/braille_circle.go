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

package draw

// braille_circle.go contains code that draws circles on a braille canvas.

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/cell"
)

// BrailleCircleOption is used to provide options to BrailleCircle().
type BrailleCircleOption interface {
	// set sets the provided option.
	set(*brailleCircleOptions)
}

// brailleCircleOptions stores the provided options.
type brailleCircleOptions struct {
	cellOpts []cell.Option
	filled   bool
}

// newBrailleCircleOptions returns a new brailleCircleOptions instance.
func newBrailleCircleOptions() *brailleCircleOptions {
	return &brailleCircleOptions{}
}

// brailleCircleOption implements BrailleCircleOption.
type brailleCircleOption func(*brailleCircleOptions)

// set implements BrailleCircleOption.set.
func (o brailleCircleOption) set(opts *brailleCircleOptions) {
	o(opts)
}

// BrailleCircleCellOpts sets options on the cells that contain the line.
func BrailleCircleCellOpts(cOpts ...cell.Option) BrailleCircleOption {
	return brailleCircleOption(func(opts *brailleCircleOptions) {
		opts.cellOpts = cOpts
	})
}

// BrailleCircleFilled indicates that the drawn circle should be filled.
func BrailleCircleFilled() BrailleCircleOption {
	return brailleCircleOption(func(opts *brailleCircleOptions) {
		opts.filled = true
	})
}

// BrailleCircle draws an approximated circle with the specified mid point and radius.
// The mid point must be a valid pixel within the canvas.
// All the points that form the circle must fit into the canvas.
// The smallest valid radius is one which draws exactly one pixel.
func BrailleCircle(bc *braille.Canvas, mid image.Point, radius int, opts ...BrailleCircleOption) error {
	if ar := bc.Area(); !mid.In(ar) {
		return fmt.Errorf("unable to draw circle with mid point %v which is outside of the braille canvas area %v", mid, ar)
	}
	if min := 1; radius < min {
		return fmt.Errorf("unable to draw circle with radius %d, must be in range %d <= radius", radius, min)
	}

	opt := newBrailleCircleOptions()
	for _, o := range opts {
		o.set(opt)
	}

	// Midpoint circle algorithm
	// https://en.wikipedia.org/wiki/Midpoint_circle_algorithm
	x := radius - 1
	y := 0
	dx := 1
	dy := 1
	diff := dx - (radius << 1) // Cheap multiplication by two.

	for x >= y {
		// Pairs of points with the same Y coordinate.
		// Structure used when drawing a filled circle.
		pointGroups := [][]image.Point{
			{
				{mid.X + x, mid.Y + y},
				{mid.X - x, mid.Y + y},
			},
			{
				{mid.X + y, mid.Y + x},
				{mid.X - y, mid.Y + x},
			},
			{
				{mid.X - x, mid.Y - y},
				{mid.X + x, mid.Y - y},
			},
			{
				{mid.X - y, mid.Y - x},
				{mid.X + y, mid.Y - x},
			},
		}

		for _, group := range pointGroups {
			if opt.filled {
				// Draw a line for each pair of points with the same Y coordinate.
				if err := BrailleLine(bc, group[0], group[1], BrailleLineCellOpts(opt.cellOpts...)); err != nil {
					return fmt.Errorf("circle with mid:%v, radius:%d, BrailleLine => %v", mid, radius, err)
				}
			} else {
				// Just set all the pixels.
				for _, p := range group {
					if err := bc.SetPixel(p, opt.cellOpts...); err != nil {
						return fmt.Errorf("circle with mid:%v, radius:%d, SetPixel => %v", mid, radius, err)
					}
				}
			}

		}

		if diff <= 0 {
			y++
			diff += dy
			dy += 2
		}

		if diff > 0 {
			x--
			dx += 2
			diff += dx - (radius << 1)
		}
	}
	return nil
}
