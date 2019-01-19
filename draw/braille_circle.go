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
	"github.com/mum4k/termdash/numbers"
)

// BrailleCircleOption is used to provide options to BrailleCircle.
type BrailleCircleOption interface {
	// set sets the provided option.
	set(*brailleCircleOptions)
}

// brailleCircleOptions stores the provided options.
type brailleCircleOptions struct {
	cellOpts []cell.Option
	filled   bool

	arcOnly     bool
	startDegree int
	endDegree   int
}

// validate validates the provided options.
func (opts *brailleCircleOptions) validate() error {
	if !opts.arcOnly {
		return nil
	}

	const (
		min = 0
		max = 360
	)
	if got := opts.startDegree; got < min || got > max {
		return fmt.Errorf("invalid starting degree for the arc %d, must be in range %d <= degree <= %d", got, min, max)
	}
	if got := opts.endDegree; got < min || got > max {
		return fmt.Errorf("invalid ending degree for the arc %d, must be in range %d <= degree <= %d", got, min, max)
	}
	if opts.startDegree == opts.endDegree {
		return fmt.Errorf("invalid degree range, start %d and end %d cannot be equal", opts.startDegree, opts.endDegree)
	}
	return nil
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
// Cell options on a braille canvas can only be set on the entire cell, not per
// pixel.
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

// BrailleCircleArcOnly indicates that only a portion of the circle should be drawn.
// The arc will be between the two provided angles in degrees.
// Each angle must be in range 0 <= angle <= 360. Start and end must not be equal.
// The zero angle is on the X axis, angles grow counter-clockwise.
func BrailleCircleArcOnly(startDegree, endDegree int) BrailleCircleOption {
	return brailleCircleOption(func(opts *brailleCircleOptions) {
		opts.arcOnly = true
		opts.startDegree = startDegree
		opts.endDegree = endDegree

	})
}

// BrailleCircle draws an approximated circle with the specified mid point and radius.
// The mid point must be a valid pixel within the canvas.
// All the points that form the circle must fit into the canvas.
// The smallest valid radius is one.
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

	if err := opt.validate(); err != nil {
		return err
	}

	points := circlePoints(mid, radius)
	if opt.arcOnly {
		points = arcPoints(points, mid, radius, opt)
	}

	if opt.filled {
		for _, pts := range groupByY(points) {
			if err := BrailleLine(bc, pts[0], pts[1], BrailleLineCellOpts(opt.cellOpts...)); err != nil {
				return fmt.Errorf("failed to fill circle with mid:%v, start:%d degrees end:%d degrees, BrailleLine => %v", mid, opt.startDegree, opt.endDegree, err)
			}
		}
		return nil
	}

	for _, p := range points {
		if err := bc.SetPixel(p, opt.cellOpts...); err != nil {
			return fmt.Errorf("failed to draw circle with mid:%v, start:%d degrees end:%d degrees, SetPixel => %v", mid, opt.startDegree, opt.endDegree, err)
		}
	}
	return nil
}

// groupByY groups the points by their Y coordinate.
// Creates a map of Y coordinates to two points on that Y row.
// The points are the point with the smallest and the largest X coordinate.
// This is used to fill a circle or an arc - by drawing lines between these
// points.
func groupByY(points []image.Point) map[int][]image.Point {
	groupped := map[int][]int{} // maps y -> x
	for _, p := range points {
		groupped[p.Y] = append(groupped[p.Y], p.X)
	}

	res := map[int][]image.Point{}
	for y, pts := range groupped {
		min, max := numbers.MinMaxInts(pts)
		res[y] = []image.Point{
			{min, y},
			{max, y},
		}
	}
	return res
}

// filterByAngle filters the provided points, returning only those that fall
// within the starting and the ending angle.
func filterByAngle(points []image.Point, mid image.Point, start, end int) []image.Point {
	var res []image.Point
	for _, p := range points {
		angle := numbers.CircleAngleAtPoint(p, mid)

		// Edge case, this might mean 0 or 360.
		// Decide based on where we are starting.
		if angle == 0 && start > 0 {
			angle = 360
		}

		ranges := toDegreeRanges(start, end)
		for _, r := range ranges {
			if r.in(angle) {
				res = append(res, p)
				break
			}
		}
	}
	return res
}

// intRange represents a range of integers.
type intRange struct {
	start int
	end   int
}

// in asserts whether the integer is in the range.
func (ir *intRange) in(i int) bool {
	return i >= ir.start && i <= ir.end
}

// toDegreeRanges converts the start and end angles in degrees into ranges of
// angles. Solves cases where the 0/360 point falls within the range.
func toDegreeRanges(start, end int) []*intRange {
	if start == 360 && end == 0 {
		start, end = end, start
	}

	if start < end {
		return []*intRange{
			{start, end},
		}
	}

	// The range is crossing the 0/360 degree point.
	// Break it into multiple ranges.
	return []*intRange{
		{start, 360},
		{0, end},
	}
}

// arcPoints returns only those points that belong to an incomplete circle.
func arcPoints(points []image.Point, mid image.Point, radius int, opt *brailleCircleOptions) []image.Point {
	points = filterByAngle(points, mid, opt.startDegree, opt.endDegree)

	if opt.filled {
		// If we are filling the angle - add points representing the lines from
		// the mid point to the start and end point of the arc.
		startP := numbers.CirclePointAtAngle(opt.startDegree, mid, radius)
		endP := numbers.CirclePointAtAngle(opt.endDegree, mid, radius)
		points = append(points, brailleLinePoints(mid, startP)...)
		points = append(points, brailleLinePoints(mid, endP)...)
	}
	return points
}

// circlePoints returns a list of points that represent a circle with
// the specified mid point and radius.
func circlePoints(mid image.Point, radius int) []image.Point {
	var points []image.Point

	// Bresenham algorithm.
	// https://en.wikipedia.org/wiki/Midpoint_circle_algorithm
	x := radius
	y := 0
	dx := 1
	dy := 1
	diff := dx - (radius << 1) // Cheap multiplication by two.

	for x >= y {
		points = append(
			points,
			image.Point{mid.X + x, mid.Y + y},
			image.Point{mid.X + y, mid.Y + x},
			image.Point{mid.X - y, mid.Y + x},
			image.Point{mid.X - x, mid.Y + y},
			image.Point{mid.X - x, mid.Y - y},
			image.Point{mid.X - y, mid.Y - x},
			image.Point{mid.X + y, mid.Y - x},
			image.Point{mid.X + x, mid.Y - y},
		)

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
	return points
}
