// Copyright 2018 Google Inc.
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

// Package linechart contains a widget that displays line charts.
package linechart

import (
	"errors"
	"fmt"
	"image"
	"math"
	"sort"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/numbers"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/linechart/internal/axes"
	"github.com/mum4k/termdash/widgets/linechart/internal/zoom"
)

const defaultScalePrecision = 2

// seriesValues represent values stored in the series.
type seriesValues struct {
	// values are the values in the series.
	values []float64
	// min is the smallest value, zero if values is empty.
	min float64
	// max is the largest value, zero if values is empty.
	max float64

	seriesCellOpts []cell.Option
	// The custom labels provided on a call to Series and a bool indicating if
	// the labels were provided. This allows resetting them to nil.
	xLabelsSet bool
	xLabels    map[int]string
}

// newSeriesValues returns a new seriesValues instance.
func newSeriesValues(values []float64) *seriesValues {
	// Copy to avoid external modifications. See #174.
	v := make([]float64, len(values))
	copy(v, values)

	min, max := minMax(v)
	return &seriesValues{
		values: v,
		min:    min,
		max:    max,
	}
}

// LineChart draws line charts.
//
// Each line chart has an identifying label and a set of values that are
// plotted.
//
// The size of the two axes is determined from the values.
// The X axis will have a number of evenly distributed data points equal to the
// largest count of values among all the labeled line charts.
// The Y axis will be sized so that it can conveniently accommodate the largest
// value among all the labeled line charts. This determines the used scale.
//
// LineChart supports mouse based zoom, zooming is achieved by either
// highlighting an area on the graph (left mouse clicking and dragging) or by
// using the mouse scroll button.
//
// Implements widgetapi.Widget. This object is thread-safe.
type LineChart struct {
	// mu protects the LineChart widget.
	mu sync.RWMutex

	// series are the series that will be plotted.
	// Keyed by the name of the series and updated by calling Series.
	series map[string]*seriesValues

	// yMin are the min and max values for the Y axis.
	yMin, yMax float64

	// capacity is the last observed value capacity in pixels when Draw was
	// called.
	capacity int

	// opts are the provided options.
	opts *options

	// xLabels that were provided on a call to Series.
	xLabels map[int]string

	// zoom tracks the zooming of the X axis.
	zoom *zoom.Tracker

	// thresholdLine is an optional horizontal guide rendered across the chart.
	thresholdLine *thresholdLine
}

// Sample identifies a visible data point under a rendered line segment.
type Sample struct {
	// X is the one-based visible sample index.
	X int
	// Y is the value at X or the interpolated value on a visible segment.
	Y float64
}

// New returns a new line chart widget.
func New(opts ...Option) (*LineChart, error) {
	opt := newOptions(opts...)
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &LineChart{
		series:        map[string]*seriesValues{},
		opts:          opt,
		thresholdLine: opt.thresholdLine,
	}, nil
}

// SeriesOption is used to provide options to Series.
type SeriesOption interface {
	// set sets the provided option.
	set(*seriesValues)
}

// seriesOption implements SeriesOption.
type seriesOption func(*seriesValues)

// set implements SeriesOption.set.
func (so seriesOption) set(sv *seriesValues) {
	so(sv)
}

// SeriesCellOpts sets the cell options for this series.
// Note that the braille canvas has resolution of 2x4 pixels per cell, but each
// cell can only have one set of cell options set. Meaning that where series
// share a cell, the last drawn series sets the cell options. Series are drawn
// in alphabetical order based on their name.
func SeriesCellOpts(co ...cell.Option) SeriesOption {
	return seriesOption(func(opts *seriesValues) {
		opts.seriesCellOpts = co
	})
}

// SeriesXLabels is used to provide custom labels for the X axis.
// The argument maps the positions in the provided series to the desired label.
// The labels are only used if they fit under the axis.
// Custom labels are property of the line chart, since there is only one X axis,
// providing multiple custom labels overwrites the previous value.
func SeriesXLabels(labels map[int]string) SeriesOption {
	return seriesOption(func(opts *seriesValues) {
		opts.xLabelsSet = true

		// Copy to avoid external modifications. See #174.
		opts.xLabels = make(map[int]string, len(labels))
		for pos, label := range labels {
			opts.xLabels[pos] = label
		}
	})
}

// yMinMax determines the min and max values for the Y axis.
func (lc *LineChart) yMinMax() (float64, float64) {
	var (
		minimums []float64
		maximums []float64
	)
	for _, sv := range lc.series {
		minimums = append(minimums, sv.min)
		maximums = append(maximums, sv.max)
	}

	if lc.opts.yAxisCustomScale != nil {
		minimums = append(minimums, lc.opts.yAxisCustomScale.min)
		maximums = append(maximums, lc.opts.yAxisCustomScale.max)
	}

	min, _ := minMax(minimums)
	_, max := minMax(maximums)

	return min, max
}

// ValueCapacity returns the number of values that could be fit onto the X axis
// without a need to rescale the X axis. This is essentially the number of
// available pixels on the braille canvas based on the width of the LineChart
// as observed on the last call to draw. Returns zero if draw wasn't called.
//
// Note that this capacity changes each time the terminal resizes, so there is
// no guarantee this remains the same next time Draw is called.
// Should be used as a hint only.
func (lc *LineChart) ValueCapacity() int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.capacity
}

// ValueAt returns the visible sample under pos for a canvas of size.
//
// The provided position must be relative to the widget canvas, not the
// terminal. A sample is only returned when the rendered line occupies the
// hovered cell, which keeps hover readouts from appearing in empty graph space.
func (lc *LineChart) ValueAt(size, pos image.Point) (Sample, bool) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if size.X <= 0 || size.Y <= 0 {
		return Sample{}, false
	}

	cvs, err := canvas.New(image.Rect(0, 0, size.X, size.Y))
	if err != nil {
		return Sample{}, false
	}
	xd, _, graphAr, segments, ok := lc.hoverLayout(cvs)
	if !ok || !pos.In(graphAr) {
		return Sample{}, false
	}

	hovered, err := lc.hoveredSeriesCell(graphAr, segments, pos)
	if err != nil || !hovered {
		return Sample{}, false
	}

	sample, ok := hoverSample(graphAr, xd, segments, pos)
	return sample, ok
}

// Series sets the values that should be displayed as the line chart with the
// provided label.
// The values that should not be displayed on the line chart should be represented
// as math.NaN values on the values slice.
// Subsequent calls with the same label replace any previously provided values.
func (lc *LineChart) Series(label string, values []float64, opts ...SeriesOption) error {
	if label == "" {
		return errors.New("the label cannot be empty")
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	series := newSeriesValues(values)
	for _, opt := range opts {
		opt.set(series)
	}
	if series.xLabelsSet {
		for i, t := range series.xLabels {
			if i < 0 {
				return fmt.Errorf("invalid key %d -> %q provided in SeriesXLabels, keys must be positive", i, t)
			}
			if t == "" {
				return fmt.Errorf("invalid label %d -> %q provided in SeriesXLabels, values cannot be empty", i, t)
			}
		}
		lc.xLabels = series.xLabels
	}

	lc.series[label] = series
	yMin, yMax := lc.yMinMax()
	lc.yMin = yMin
	lc.yMax = yMax
	return nil
}

// SetThresholdLine updates the horizontal guide rendered across the plotted
// area at the provided Y value.
func (lc *LineChart) SetThresholdLine(value float64, opts ...cell.Option) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.thresholdLine = &thresholdLine{
		value:    value,
		cellOpts: opts,
	}
}

// ClearThresholdLine removes any previously configured horizontal guide.
func (lc *LineChart) ClearThresholdLine() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.thresholdLine = nil
}

// xDetails returns the details for the X axis given the specified minimum and
// maximum value to display.
func (lc *LineChart) xDetails(cvs *canvas.Canvas, reqYWidth, min, max int) (*axes.XDetails, error) {
	xp := &axes.XProperties{
		Min:          min,
		Max:          max,
		ReqYWidth:    reqYWidth,
		CustomLabels: lc.xLabels,
		LO:           lc.opts.xLabelOrientation,
	}
	xd, err := axes.NewXDetails(cvs.Area(), xp)
	if err != nil {
		return nil, fmt.Errorf("NewXDetails => %v", err)
	}
	return xd, nil
}

// xDetailsForCap adjusts the X details according to the capacity of the
// braille canvas (how many values can it fit).
// If the capacity cannot accommodate all the values, the starting value of the
// X axis is adjusted so that it displays the last n values that fit.
// Returns unadjusted xd if all the values fit.
func (lc *LineChart) xDetailsForCap(cvs *canvas.Canvas, bc *braille.Canvas, xd *axes.XDetails, yd *axes.YDetails) (*axes.XDetails, error) {
	lc.capacity = bc.Area().Dx()
	values := int(xd.Scale.Max.Value) - int(xd.Scale.Min.Value) + 1
	if !lc.opts.xAxisUnscaled || values <= lc.capacity {
		return xd, nil
	}

	diff := values - lc.capacity
	xMin := int(xd.Scale.Min.Value) + diff
	xMax := int(xd.Scale.Max.Value)
	if lc.opts.brailleOnly {
		return lc.brailleOnlyXDetails(cvs, xMin, xMax)
	}
	unscaledXD, err := lc.xDetails(cvs, yd.Start.X, xMin, xMax)
	if err != nil {
		return nil, err
	}
	return unscaledXD, nil
}

// axesDetails determines the details about the X and Y axes.
func (lc *LineChart) axesDetails(cvs *canvas.Canvas) (*axes.XDetails, *axes.YDetails, error) {
	if lc.opts.brailleOnly {
		return lc.brailleOnlyDetails(cvs)
	}
	reqXHeight := axes.RequiredHeight(lc.maxXValue(), lc.xLabels, lc.opts.xLabelOrientation)
	yp := &axes.YProperties{
		Min:            lc.yMin,
		Max:            lc.yMax,
		ReqXHeight:     reqXHeight,
		ScaleMode:      lc.opts.yAxisMode,
		ValueFormatter: lc.opts.yAxisValueFormatter,
	}
	yd, err := axes.NewYDetails(cvs.Area(), yp)
	if err != nil {
		return nil, nil, fmt.Errorf("NewYDetails => %v", err)
	}

	const xMin = 0
	xMax := lc.maxXValue()
	xd, err := lc.xDetails(cvs, yd.Start.X, xMin, xMax)
	if err != nil {
		return nil, nil, err
	}
	return xd, yd, nil
}

// brailleOnlyDetails returns full-canvas scale details without visible axes.
func (lc *LineChart) brailleOnlyDetails(cvs *canvas.Canvas) (*axes.XDetails, *axes.YDetails, error) {
	graphAr := cvs.Area()
	xd, err := lc.brailleOnlyXDetails(cvs, 0, lc.maxXValue())
	if err != nil {
		return nil, nil, err
	}
	ydScale, err := axes.NewYScale(lc.yMin, lc.yMax, graphAr.Dy(), defaultScalePrecision, lc.opts.yAxisMode, lc.opts.yAxisValueFormatter)
	if err != nil {
		return nil, nil, err
	}
	return xd, &axes.YDetails{
		Width: 0,
		Start: image.Point{X: graphAr.Min.X, Y: graphAr.Min.Y},
		End:   image.Point{X: graphAr.Min.X, Y: graphAr.Max.Y - 1},
		Scale: ydScale,
	}, nil
}

// brailleOnlyXDetails returns scale details for full-canvas braille plotting.
func (lc *LineChart) brailleOnlyXDetails(cvs *canvas.Canvas, min, max int) (*axes.XDetails, error) {
	graphAr := cvs.Area()
	scale, err := axes.NewXScale(min, max, graphAr.Dx(), defaultScalePrecision)
	if err != nil {
		return nil, err
	}
	return &axes.XDetails{
		Start: image.Point{X: graphAr.Min.X, Y: graphAr.Max.Y - 1},
		End:   image.Point{X: graphAr.Max.X - 1, Y: graphAr.Max.Y - 1},
		Scale: scale,
		Properties: &axes.XProperties{
			Min: min,
			Max: max,
			LO:  lc.opts.xLabelOrientation,
		},
	}, nil
}

// Draw draws the values as line charts.
// Implements widgetapi.Widget.Draw.
func (lc *LineChart) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	needAr, err := area.FromSize(lc.minSize())
	if err != nil {
		return err
	}
	if !needAr.In(cvs.Area()) {
		return draw.ResizeNeeded(cvs)
	}

	xd, yd, err := lc.axesDetails(cvs)
	if err != nil {
		return err
	}

	adjXD, err := lc.drawSeries(cvs, xd, yd)
	if err != nil {
		return err
	}
	return lc.drawAxes(cvs, adjXD, yd)
}

// drawAxes draws the X,Y axes and their labels.
func (lc *LineChart) drawAxes(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	if lc.opts.brailleOnly {
		return nil
	}
	lines := []draw.HVLine{
		{Start: yd.Start, End: yd.End},
		{Start: xd.Start, End: xd.End},
	}
	if err := draw.HVLines(cvs, lines, draw.HVLineCellOpts(lc.opts.axesCellOpts...)); err != nil {
		return fmt.Errorf("failed to draw the axes: %v", err)
	}

	for _, l := range yd.Labels {
		if err := draw.Text(cvs, l.Value.Text(), l.Pos,
			draw.TextMaxX(yd.Start.X),
			draw.TextOverrunMode(draw.OverrunModeThreeDot),
			draw.TextCellOpts(lc.opts.yLabelCellOpts...),
		); err != nil {
			return fmt.Errorf("failed to draw the Y labels: %v", err)
		}
	}

	for _, l := range xd.Labels {
		switch lc.opts.xLabelOrientation {
		case axes.LabelOrientationHorizontal:
			if err := draw.Text(cvs, l.Value.Text(), l.Pos, draw.TextCellOpts(lc.opts.xLabelCellOpts...)); err != nil {
				return fmt.Errorf("failed to draw the X horizontal labels: %v", err)
			}

		case axes.LabelOrientationVertical:
			if err := draw.VerticalText(cvs, l.Value.Text(), l.Pos,
				draw.VerticalTextCellOpts(lc.opts.xLabelCellOpts...),
				draw.VerticalTextOverrunMode(draw.OverrunModeThreeDot),
			); err != nil {
				return fmt.Errorf("failed to draw the vertical X labels: %v", err)
			}
		}
	}
	return nil
}

// graphAr returns the area available for the graph itself sized so that it
// fits between the axes and the canvas borders.
func (lc *LineChart) graphAr(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) image.Rectangle {
	if lc.opts.brailleOnly {
		return cvs.Area()
	}
	return image.Rect(yd.Start.X+1, yd.Start.Y, cvs.Area().Max.X, xd.End.Y)
}

// drawSeries draws the graph representing the stored series.
// Returns XDetails that might be adjusted to not start at zero value if some
// of the series didn't fit the graphs and XAxisUnscaled was provided.
// If the series has NaN values they will be ignored and not draw on the graph.
func (lc *LineChart) drawSeries(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) (*axes.XDetails, error) {
	graphAr := lc.graphAr(cvs, xd, yd)
	bc, err := braille.New(graphAr)
	if err != nil {
		return nil, err
	}

	xdForCap, err := lc.xDetailsForCap(cvs, bc, xd, yd)
	if err != nil {
		return nil, err
	}

	if lc.zoom == nil {
		z, err := zoom.New(xdForCap, cvs.Area(), graphAr, zoom.ScrollStep(lc.opts.zoomStepPercent))
		if err != nil {
			return nil, err
		}
		lc.zoom = z
	} else {
		if err := lc.zoom.Update(xdForCap, cvs.Area(), graphAr); err != nil {
			return nil, err
		}
	}

	xdZoomed := lc.zoom.Zoom()
	var names []string
	for name := range lc.series {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		sv := lc.series[name]
		for _, segment := range sampleVisibleSeries(
			sv.values,
			int(xdZoomed.Scale.Min.Value),
			int(xdZoomed.Scale.Max.Value),
			lc.capacity,
			lc.opts.downsampler,
		) {
			if len(segment) <= 1 {
				continue
			}
			for i := 1; i < len(segment); i++ {
				start, end := segment[i-1], segment[i]
				startX, err := xdZoomed.Scale.ValueToPixel(start.index)
				if err != nil {
					return nil, fmt.Errorf("failure for series %v[%d] on scale %v, xdZoomed.Scale.ValueToPixel(%v) => %v", name, start.index, xdZoomed.Scale, start.index, err)
				}
				endX, err := xdZoomed.Scale.ValueToPixel(end.index)
				if err != nil {
					return nil, fmt.Errorf("failure for series %v[%d] on scale %v, xdZoomed.Scale.ValueToPixel(%v) => %v", name, end.index, xdZoomed.Scale, end.index, err)
				}

				startY, err := yd.Scale.ValueToPixel(start.value)
				if err != nil {
					return nil, fmt.Errorf("failure for series %v[%d] on scale %v, yd.Scale.ValueToPixel(%v) => %v", name, start.index, yd.Scale, start.value, err)
				}

				endY, err := yd.Scale.ValueToPixel(end.value)
				if err != nil {
					return nil, fmt.Errorf("failure for series %v[%d] on scale %v, yd.Scale.ValueToPixel(%v) => %v", name, end.index, yd.Scale, end.value, err)
				}

				if err := draw.BrailleLine(bc,
					image.Point{startX, startY},
					image.Point{endX, endY},
					draw.BrailleLineCellOpts(sv.seriesCellOpts...),
				); err != nil {
					return nil, fmt.Errorf("draw.BrailleLine => %v", err)
				}
			}
		}
	}

	if highlight, hRange := lc.zoom.Highlight(); highlight {
		if err := lc.highlightRange(bc, hRange); err != nil {
			return nil, err
		}
	}

	if err := lc.drawThresholdLine(bc, yd); err != nil {
		return nil, err
	}

	if err := bc.CopyTo(cvs); err != nil {
		return nil, fmt.Errorf("bc.Apply => %v", err)
	}
	return xdZoomed, nil
}

// hoverSegment stores one visible plotted segment in braille pixel space.
type hoverSegment struct {
	startX int
	startY int
	endX   int
	endY   int
	start  samplePoint
	end    samplePoint
}

// hoverLayout determines the visible graph area, zoomed scales, and visible
// line segments for hover hit-testing.
func (lc *LineChart) hoverLayout(cvs *canvas.Canvas) (*axes.XDetails, *axes.YDetails, image.Rectangle, []hoverSegment, bool) {
	needAr, err := area.FromSize(lc.minSize())
	if err != nil || !needAr.In(cvs.Area()) {
		return nil, nil, image.Rectangle{}, nil, false
	}

	xd, yd, err := lc.axesDetails(cvs)
	if err != nil {
		return nil, nil, image.Rectangle{}, nil, false
	}

	graphAr := lc.graphAr(cvs, xd, yd)
	bc, err := braille.New(graphAr)
	if err != nil {
		return nil, nil, image.Rectangle{}, nil, false
	}

	xdForCap, err := lc.xDetailsForCap(cvs, bc, xd, yd)
	if err != nil {
		return nil, nil, image.Rectangle{}, nil, false
	}

	xdZoomed := xdForCap
	if lc.zoom != nil {
		xdZoomed = lc.zoom.Zoom()
	}

	segments, ok := lc.hoverSegments(xdZoomed, yd)
	if !ok {
		return nil, nil, image.Rectangle{}, nil, false
	}
	return xdZoomed, yd, graphAr, segments, true
}

// hoverSegments returns the visible plotted segments used for hover hit-testing.
func (lc *LineChart) hoverSegments(xd *axes.XDetails, yd *axes.YDetails) ([]hoverSegment, bool) {
	var names []string
	for name := range lc.series {
		names = append(names, name)
	}
	sort.Strings(names)

	var segments []hoverSegment
	for _, name := range names {
		sv := lc.series[name]
		for _, segment := range sampleVisibleSeries(
			sv.values,
			int(xd.Scale.Min.Value),
			int(xd.Scale.Max.Value),
			lc.capacity,
			lc.opts.downsampler,
		) {
			if len(segment) <= 1 {
				continue
			}
			for i := 1; i < len(segment); i++ {
				start, end := segment[i-1], segment[i]
				startX, err := xd.Scale.ValueToPixel(start.index)
				if err != nil {
					return nil, false
				}
				endX, err := xd.Scale.ValueToPixel(end.index)
				if err != nil {
					return nil, false
				}
				startY, err := yd.Scale.ValueToPixel(start.value)
				if err != nil {
					return nil, false
				}
				endY, err := yd.Scale.ValueToPixel(end.value)
				if err != nil {
					return nil, false
				}
				segments = append(segments, hoverSegment{
					startX: startX,
					startY: startY,
					endX:   endX,
					endY:   endY,
					start:  start,
					end:    end,
				})
			}
		}
	}
	return segments, true
}

// hoveredSeriesCell reports whether pos lands on a rendered series cell.
func (lc *LineChart) hoveredSeriesCell(graphAr image.Rectangle, segments []hoverSegment, pos image.Point) (bool, error) {
	cvs, err := canvas.New(image.Rect(0, 0, graphAr.Dx(), graphAr.Dy()))
	if err != nil {
		return false, err
	}
	bc, err := braille.New(cvs.Area())
	if err != nil {
		return false, err
	}

	for _, segment := range segments {
		if err := draw.BrailleLine(
			bc,
			image.Point{X: segment.startX, Y: segment.startY},
			image.Point{X: segment.endX, Y: segment.endY},
		); err != nil {
			return false, err
		}
	}
	if err := bc.CopyTo(cvs); err != nil {
		return false, err
	}

	cellPoint := image.Point{X: pos.X - graphAr.Min.X, Y: pos.Y - graphAr.Min.Y}
	cellValue, err := cvs.Cell(cellPoint)
	if err != nil {
		return false, err
	}
	return cellValue.Rune != 0 && cellValue.Rune != ' ', nil
}

// hoverSample returns the nearest visible sample represented at pos.
func hoverSample(graphAr image.Rectangle, xd *axes.XDetails, segments []hoverSegment, pos image.Point) (Sample, bool) {
	if len(segments) == 0 {
		return Sample{}, false
	}

	cellPixelLeft := (pos.X - graphAr.Min.X) * braille.ColMult
	cellPixelRight := cellPixelLeft + braille.ColMult - 1
	cellPixelTop := (pos.Y - graphAr.Min.Y) * braille.RowMult
	cellPixelBottom := cellPixelTop + braille.RowMult - 1
	centerX := float64(cellPixelLeft+cellPixelRight) / 2
	centerY := float64(cellPixelTop+cellPixelBottom) / 2

	bestDistance := math.MaxFloat64
	bestSample := Sample{}
	bestFound := false

	for _, segment := range segments {
		minX := min(segment.startX, segment.endX)
		maxX := max(segment.startX, segment.endX)
		if cellPixelRight < minX || cellPixelLeft > maxX {
			continue
		}

		for hoverPixelX := cellPixelLeft; hoverPixelX <= cellPixelRight; hoverPixelX++ {
			t := 0.0
			if segment.startX != segment.endX {
				t = float64(hoverPixelX-segment.startX) / float64(segment.endX-segment.startX)
			}
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}

			pixelY := float64(segment.startY) + float64(segment.endY-segment.startY)*t
			if pixelY < float64(cellPixelTop) || pixelY > float64(cellPixelBottom) {
				continue
			}

			xValue, err := xd.Scale.PixelToValue(hoverPixelX)
			if err != nil {
				continue
			}
			yValue := segment.start.value + (segment.end.value-segment.start.value)*t
			distance := math.Hypot(float64(hoverPixelX)-centerX, pixelY-centerY)
			if !bestFound || distance < bestDistance {
				bestDistance = distance
				bestSample = Sample{
					X: int(math.Round(xValue)) + 1,
					Y: yValue,
				}
				bestFound = true
			}
		}
	}

	return bestSample, bestFound
}

// min returns the smaller of the provided integers.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// max returns the larger of the provided integers.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// drawThresholdLine renders the configured horizontal guide across the plotted
// area when one is enabled.
func (lc *LineChart) drawThresholdLine(bc *braille.Canvas, yd *axes.YDetails) error {
	if lc.thresholdLine == nil {
		return nil
	}

	y, err := yd.Scale.ValueToPixel(lc.thresholdLine.value)
	if err != nil {
		return fmt.Errorf("yd.Scale.ValueToPixel(%v) => %v", lc.thresholdLine.value, err)
	}

	maxX := bc.Area().Dx() - 1
	if maxX < 0 {
		return nil
	}

	if err := draw.BrailleLine(
		bc,
		image.Point{X: 0, Y: y},
		image.Point{X: maxX, Y: y},
		draw.BrailleLineCellOpts(lc.thresholdLine.cellOpts...),
	); err != nil {
		return fmt.Errorf("draw.BrailleLine(threshold) => %v", err)
	}
	return nil
}

// highlightRange highlights the range of X columns on the braille canvas.
func (lc *LineChart) highlightRange(bc *braille.Canvas, hRange *zoom.Range) error {
	cellAr := bc.CellArea()
	ar := image.Rect(hRange.Start, cellAr.Min.Y, hRange.End, cellAr.Max.Y)
	return bc.SetAreaCellOpts(ar, cell.BgColor(lc.opts.zoomHightlightColor))
}

// Keyboard implements widgetapi.Widget.Keyboard.
func (lc *LineChart) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return errors.New("the LineChart widget doesn't support keyboard events")
}

// Mouse implements widgetapi.Widget.Mouse.
func (lc *LineChart) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.zoom == nil {
		return nil
	}
	return lc.zoom.Mouse(m)
}

// minSize determines the minimum required size to draw the line chart.
func (lc *LineChart) minSize() image.Point {
	if lc.opts.brailleOnly {
		return image.Point{1, 1}
	}
	// At the very least we need:
	// - n cells width for the Y axis and its labels as reported by it.
	// - at least 1 cell width for the graph.
	reqWidth := axes.RequiredWidth(lc.yMin, lc.yMax) + 1

	// And for the height:
	// - n cells width for the X axis and its labels as reported by it.
	// - at least 2 cell height for the graph.
	reqHeight := axes.RequiredHeight(lc.maxXValue(), lc.xLabels, lc.opts.xLabelOrientation) + 2
	return image.Point{reqWidth, reqHeight}
}

// Options implements widgetapi.Widget.Options.
func (lc *LineChart) Options() widgetapi.Options {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	return widgetapi.Options{
		MinimumSize: lc.minSize(),
		WantMouse:   widgetapi.MouseScopeGlobal,
	}
}

// maxXValue returns the maximum value on the X axis among all the series.
// lc.mu must be held when calling this method.
func (lc *LineChart) maxXValue() int {
	maxLen := 0
	for _, sv := range lc.series {
		if l := len(sv.values); l > maxLen {
			maxLen = l
		}
	}
	if maxLen == 0 {
		return 0
	}
	return maxLen - 1
}

// minMax is a wrapper around numbers.MinMax that controls
// the output if the values are NaN and sets defaults if it's
// the case.
func minMax(values []float64) (x, y float64) {
	min, max := numbers.MinMax(values)
	if math.IsNaN(min) {
		min = 0
	}
	if math.IsNaN(max) {
		max = 0
	}
	return min, max
}
