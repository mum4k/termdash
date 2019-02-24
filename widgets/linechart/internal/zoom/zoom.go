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

// Package zoom contains code that tracks the current zoom level.
package zoom

import (
	"fmt"
	"image"
	"reflect"

	"github.com/mum4k/termdash/internal/button"
	"github.com/mum4k/termdash/internal/numbers"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart/internal/axes"
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// options stores the provided options.
type options struct {
	scrollStepPerc int
}

// newOptions creates new options instance and applies the provided options.
func newOptions(opts ...Option) *options {
	o := &options{
		scrollStepPerc: DefaultScrollStep,
	}
	for _, opt := range opts {
		opt.set(o)
	}
	return o
}

// validate validates the provided options.
func (o *options) validate() error {
	if min, max := 1, 100; o.scrollStepPerc < min || o.scrollStepPerc > max {
		return fmt.Errorf("invalid ScrollStep %d, must be a value in the range %d <= value <= %d", o.scrollStepPerc, min, max)
	}
	return nil
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// DefaultScrollStep is the default value for the ScrollStep option.
const DefaultScrollStep = 10

// ScrollStep sets the amount of zoom in or out on a single mouse scroll event.
// This is set as a percentage of the current value size of the X axis.
// Must be a value in range 0 < value <= 100.
// Defaults to DefaultScrollStep.
func ScrollStep(perc int) Option {
	return option(func(opts *options) {
		opts.scrollStepPerc = perc
	})
}

// Tracker tracks the state of mouse selection on the linechart and stores
// requests for zoom.
// This object is not thread-safe.
type Tracker struct {
	// baseX is the base X axis without any zoom applied.
	baseX *axes.XDetails
	// zoomX is the zoomed X axis or nil if zoom isn't applied.
	zoomX *axes.XDetails

	// cvsAr is the entire canvas available to the linechart widget.
	cvsAr image.Rectangle

	// graphAr is a smaller part of the cvsAr that contains the linechart
	// itself. I.e. an area between the axis and the borders of cvsAr.
	graphAr image.Rectangle

	// fsm is the state machine tracking the state of mouse left button.
	fsm *button.FSM

	// highlight is the currently highlighted area.
	highlight *Range

	// opts are the provided options.
	opts *options
}

// New returns a new zoom tracker that tracks zoom requests within
// the provided graph area. The cvsAr argument indicates size of the entire
// canvas available to the widget.
func New(baseX *axes.XDetails, cvsAr, graphAr image.Rectangle, opts ...Option) (*Tracker, error) {
	o := newOptions(opts...)
	if err := o.validate(); err != nil {
		return nil, err
	}

	t := &Tracker{
		fsm:       button.NewFSM(mouse.ButtonLeft, graphAr),
		highlight: &Range{},
		opts:      o,
	}
	if err := t.Update(baseX, cvsAr, graphAr); err != nil {
		return nil, err
	}
	return t, nil
}

// Update is used to inform the zoom tracker about the base X axis and the
// graph area.
// Should be called each time the widget redraws.
func (t *Tracker) Update(baseX *axes.XDetails, cvsAr, graphAr image.Rectangle) error {
	if !graphAr.In(cvsAr) {
		return fmt.Errorf("the graphAr %v doesn't fit inside the cvsAr %v", graphAr, cvsAr)
	}
	// If any of these parameters changed, we need to reset the FSM and ensure
	// the current zoom is still within the range of the new X axis.
	ac, sc := t.axisChanged(baseX), t.sizeChanged(cvsAr, graphAr)
	if sc {
		t.highlight.reset()
		t.fsm.UpdateArea(graphAr)
	}
	if ac || sc {
		if t.zoomX != nil {
			// Input data changed and we have an existing zoom in place.
			// We need to normalize it again, since it might be outside of the
			// currently visible values (e.g. if the terminal size decreased).
			zoomMin := int(t.zoomX.Scale.Min.Value)
			zoomMax := int(t.zoomX.Scale.Max.Value)
			opt := &normalizeOptions{
				oldBaseMin: t.baseX.Scale.Min,
				oldBaseMax: t.baseX.Scale.Max,
			}
			min, max := normalize(baseX.Scale.Min, baseX.Scale.Max, zoomMin, zoomMax, opt)
			if !hasMinMax(min, max, baseX) {
				zoom, err := newZoomedFromBase(min, max, baseX, cvsAr)
				if err != nil {
					return err
				}
				t.zoomX = zoom
			} else {
				// Fully unzoom.
				t.zoomX = nil
			}
		}
	}

	t.baseX = baseX
	t.cvsAr = cvsAr
	t.graphAr = graphAr
	return nil
}

// sizeChanged asserts whether the physical layout of the terminal changed.
func (t *Tracker) sizeChanged(cvsAr, graphAr image.Rectangle) bool {
	return !cvsAr.Eq(t.cvsAr) || !graphAr.Eq(t.graphAr)
}

// axisChanged asserts whether the axis scale changed.
func (t *Tracker) axisChanged(baseX *axes.XDetails) bool {
	return !reflect.DeepEqual(baseX, t.baseX)
}

// baseForZoom returns the base axis before zooming.
// This is either the base provided to New or Update if no zoom was performed
// yet, or the previously zoomed axis.
func (t *Tracker) baseForZoom() *axes.XDetails {
	if t.zoomX == nil {
		return t.baseX
	}
	return t.zoomX
}

// Mouse is used to forward mouse events to the zoom tracker.
func (t *Tracker) Mouse(m *terminalapi.Mouse) error {
	if m.Position.In(t.graphAr) {
		switch m.Button {
		case mouse.ButtonWheelUp, mouse.ButtonWheelDown:
			zoom, err := zoomToScroll(m, t.cvsAr, t.graphAr, t.baseForZoom(), t.baseX, t.opts)
			if err != nil {
				return err
			}
			t.zoomX = zoom
		}
	}

	clicked, bs := t.fsm.Event(m)
	switch {
	case bs == button.Down:
		cellX := m.Position.X - t.graphAr.Min.X
		t.highlight.addX(cellX)

	case clicked && bs == button.Up:
		if t.highlight.length() >= 2 {
			zoom, err := zoomToHighlight(t.baseForZoom(), t.highlight, t.cvsAr)
			if err != nil {
				return err
			}
			t.zoomX = zoom
		}
		t.highlight.reset()

	default:
		t.highlight.reset()
	}
	return nil
}

// Range represents a range of values.
// The range includes all values x such that Start <= x < End.
type Range struct {
	// Start is the start of the range.
	Start int
	// End is the end of the range.
	End int

	// last is the last coordinate that was added to the range.
	last int
}

// length returns the length of the range.
func (r *Range) length() int {
	return numbers.Abs(r.End - r.Start)
}

// empty asserts if the range is empty.
func (r *Range) empty() bool {
	return r.Start == r.End
}

// reset resets the range back to zero.
func (r *Range) reset() {
	r.Start, r.End, r.last = 0, 0, 0
}

// addX adds the provided X coordinate to the range.
func (r *Range) addX(x int) {
	switch {
	case r.empty():
		r.Start = x
		r.End = x + 1

	case x < r.Start:
		if r.last == r.End-1 {
			// Handles fast mouse move to the left across Start.
			// If we don't adjust the end, we would extend both ends of the
			// range.
			r.End = r.Start + 1
		}
		r.Start = x

	case x >= r.End:
		if r.last == r.Start {
			// Handles fast mouse move to the right across End.
			// If we don't adjust the start, we would extend both ends of the
			// range.
			r.Start = r.End - 1
		}
		r.End = x + 1

	case x > r.last:
		// Handles change of direction from left to right.
		r.Start = x

	case x < r.last:
		// Handles change of direction from right to left.
		r.End = x + 1
	}
	r.last = x
}

// Highlight returns true if a range on the graph area should be highlighted
// because the user is holding down the left mouse button and dragging mouse
// across the graph area. The returned range indicates the range of X cell
// coordinates within the graph area provided to New or Update. These are the
// columns that should be highlighted.
// Returns false of no area should be highlighted, in which case the state of
// the Range return value is undefined.
func (t *Tracker) Highlight() (bool, *Range) {
	if t.highlight.empty() {
		return false, nil
	}
	return true, t.highlight
}

// Zoom returns an adjusted X axis if zoom is applied, or the same axis as was
// provided to New or Update.
func (t *Tracker) Zoom() *axes.XDetails {
	if t.zoomX == nil {
		return t.baseX
	}
	return t.zoomX
}

// normalizeOptions are optional parameters for zoom normalization.
type normalizeOptions struct {
	// oldBaseMin is the previous minimum value before an Update was called.
	oldBaseMin *axes.Value
	// oldBaseMax is the previous maximum value before an Update was called.
	oldBaseMax *axes.Value
}

// rolledBy returns the number of values by which the current base axis
// provided to Update rolled as compared to the previous one.
// The axis rolls if the linechart runs with the XAxisUnscaled option and runs
// out of capacity.
// Returns zero if the axis didn't role or if the call didn't provide the old
// axis boundaries.
// Returns a positive number of the axis rolled to the left or negative if it
// rolled to the right.
// A roll by one is identified if both the minimum and the maximum changed by
// one in the same direction.
func (co *normalizeOptions) rolledBy(baseMin, baseMax *axes.Value) int {
	if co == nil || co.oldBaseMin == nil || co.oldBaseMax == nil {
		return 0
	}

	minDiff := int(baseMin.Value) - int(co.oldBaseMin.Value)
	maxDiff := int(baseMax.Value) - int(co.oldBaseMax.Value)
	if minDiff != maxDiff {
		// The axis didn't roll, just the layout or values changed.
		return 0
	}
	return minDiff
}

// normalize normalizes the zoom range.
// This handles cases where zoom out would happen above the base axis or
// when the base axis itself changes (user provided new values) or when the
// graph areas change (terminal size changed).
// Argument opts can be nil.
func normalize(baseMin, baseMax *axes.Value, min, max int, opts *normalizeOptions) (int, int) {
	bMin := int(baseMin.Value)
	bMax := int(baseMax.Value)

	if rolled := opts.rolledBy(baseMin, baseMax); rolled != 0 {
		min += rolled
		max += rolled
	}

	var newMin, newMax int
	// Don't zoom-out above or below the base axis.
	switch {
	case min < bMin:
		newMin = bMin
	case min > bMax:
		newMin = bMax
	default:
		newMin = min
	}

	switch {
	case max < bMin:
		newMax = bMin
	case max > bMax:
		newMax = bMax
	default:
		newMax = max
	}

	if newMin > newMax {
		newMin, newMax = newMax, newMin
	}

	if newMin == newMax {
		return findValuePair(newMin, newMax, baseMin, baseMax)
	}
	return newMin, newMax
}

// newZoomedFromBase returns a new X axis zoomed to the provided min and max.
func newZoomedFromBase(min, max int, base *axes.XDetails, cvsAr image.Rectangle) (*axes.XDetails, error) {
	zp := *base.Properties // Shallow copy.
	zp.Min = min
	zp.Max = max

	zoom, err := axes.NewXDetails(cvsAr, &zp)
	if err != nil {
		return nil, fmt.Errorf("failed to create zoomed X axis: %v", err)
	}
	return zoom, nil
}

// findValuePair given two values on the base X axis returns the closest
// possible distinct values  that are still within the range pf base X.
// Returns the min and max of the base X of no such values exist.
func findValuePair(min, max int, baseMin, baseMax *axes.Value) (int, int) {
	bMin := int(baseMin.Value)
	bMax := int(baseMax.Value)

	// Try above the max.
	for v := max; v <= bMax; v++ {
		if v > min {
			return min, v
		}
	}

	// Try below the min.
	for v := min; v >= bMin; v-- {
		if v < max {
			return v, max
		}
	}

	return bMin, bMax
}

// findCellPair given two cells on the base X axis returns the values of the
// closest or the same cells such that the values are distinct.
// Useful while zooming, if the zoom targets a view that would only have one
// value, this function adjusts the view to the closest two cells with distinct
// values.
func findCellPair(base *axes.XDetails, minCell, maxCell int) (*axes.Value, *axes.Value, error) {
	minL, err := base.Scale.CellLabel(minCell)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to determine min label for cell %d: %v", minCell, err)
	}
	maxL, err := base.Scale.CellLabel(maxCell)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to determine max label for cell %d: %v", maxCell, err)
	}

	diff := maxL.Value - minL.Value
	if diff > 1 {
		return minL, maxL, nil
	}

	// Try above the max.
	for cellNum := maxCell; cellNum < base.Scale.GraphWidth; cellNum++ {
		l, err := base.Scale.CellLabel(cellNum)
		if err != nil {
			return nil, nil, err
		}
		if l.Value > minL.Value {
			return minL, l, nil
		}
	}

	// Try below the min.
	for cellNum := minCell; cellNum >= 0; cellNum-- {
		l, err := base.Scale.CellLabel(cellNum)
		if err != nil {
			return nil, nil, err
		}
		if l.Value < maxL.Value {
			return l, maxL, nil
		}
	}

	// Give up and use the first and the last cells.
	firstL, err := base.Scale.CellLabel(0)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to determine label for the first cell: %v", err)
	}
	lastL, err := base.Scale.CellLabel(base.Scale.GraphWidth - 1)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to determine label for the last cell: %v", err)
	}
	return firstL, lastL, nil
}

// zoomToHighlight zooms the base X axis according to the highlighted range.
func zoomToHighlight(base *axes.XDetails, hr *Range, cvsAr image.Rectangle) (*axes.XDetails, error) {
	minL, maxL, err := findCellPair(base, hr.Start, hr.End-1)
	if err != nil {
		return nil, err
	}

	zoom, err := newZoomedFromBase(int(minL.Value), int(maxL.Value), base, cvsAr)
	if err != nil {
		return nil, err
	}
	return zoom, nil
}

// hasMinMax asserts whether the provided min and max values represent the
// boundary values of the base axis.
func hasMinMax(min, max int, base *axes.XDetails) bool {
	return min == int(base.Scale.Min.Value) && max == int(base.Scale.Max.Value)
}

// zoomToScroll zooms or unzooms the current X axis in or out depending on the
// direction of the scroll. Doesn't zoom out above the base X axis view.
// Can return nil, which indicates that we are at 0% zoom (fully unzoomed).
func zoomToScroll(m *terminalapi.Mouse, cvsAr, graphAr image.Rectangle, curr, base *axes.XDetails, opts *options) (*axes.XDetails, error) {
	var direction int         // Positive on zoom in, negative on zoom out.
	var limits *axes.XDetails // Limit values for the zooming operation.
	switch m.Button {
	case mouse.ButtonWheelUp:
		direction = 1
		limits = curr

	case mouse.ButtonWheelDown:
		direction = -1
		limits = base
	}

	cellX := m.Position.X - graphAr.Min.X
	tgtVal, err := curr.Scale.CellLabel(cellX)
	if err != nil {
		return nil, fmt.Errorf("unable to determine value at the point where scrolling occurred: %v", err)
	}

	currMin := int(curr.Scale.Min.Value)
	currMax := int(curr.Scale.Max.Value)
	baseMin := int(base.Scale.Min.Value)
	baseMax := int(base.Scale.Max.Value)
	size := baseMax - baseMin
	step := size * opts.scrollStepPerc / 100
	_, left := numbers.MinMaxInts([]int{
		1,
		int(tgtVal.Value) - currMin,
	})
	_, right := numbers.MinMaxInts([]int{
		1,
		currMax - int(tgtVal.Value),
	})

	splitStep := numbers.SplitByRatio(step, image.Point{left, right})
	newMin := currMin + (direction * splitStep.X)
	newMax := currMax - (direction * splitStep.Y)

	min, max := normalize(limits.Scale.Min, limits.Scale.Max, newMin, newMax, nil)
	if m.Button == mouse.ButtonWheelDown && hasMinMax(min, max, limits) {
		// Fully unzoom.
		return nil, nil
	}

	minCell, err := limits.Scale.ValueToCell(min)
	if err != nil {
		return nil, err
	}
	maxCell, err := limits.Scale.ValueToCell(max)
	if err != nil {
		return nil, err
	}
	minL, maxL, err := findCellPair(limits, minCell, maxCell)
	if err != nil {
		return nil, err
	}

	zoom, err := newZoomedFromBase(int(minL.Value), int(maxL.Value), curr, cvsAr)
	if err != nil {
		return nil, err
	}
	return zoom, nil
}
