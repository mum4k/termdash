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
	"sort"
	"sync"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/numbers"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/linechart/axes"
)

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
	min, max := numbers.MinMax(values)
	return &seriesValues{
		values: values,
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
// Implements widgetapi.Widget. This object is thread-safe.
type LineChart struct {
	// mu protects the LineChart widget.
	mu sync.Mutex

	// series are the series that will be plotted.
	// Keyed by the name of the series and updated by calling Series.
	series map[string]*seriesValues

	// yAxis is the Y axis of the line chart.
	yAxis *axes.Y

	// opts are the provided options.
	opts *options

	// xLabels that were provided on a call to Series.
	xLabels map[int]string
}

// New returns a new line chart widget.
func New(opts ...Option) *LineChart {
	opt := newOptions(opts...)
	return &LineChart{
		series: map[string]*seriesValues{},
		yAxis:  axes.NewY(0, 0),
		opts:   opt,
	}
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
		opts.xLabels = labels
	})
}

// Series sets the values that should be displayed as the line chart with the
// provided label.
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
	lc.yAxis = axes.NewY(series.min, series.max)
	return nil
}

// Draw draws the values as line charts.
// Implements widgetapi.Widget.Draw.
func (lc *LineChart) Draw(cvs *canvas.Canvas) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	yd, err := lc.yAxis.Details(cvs.Area(), lc.opts.yAxisMode)
	if err != nil {
		return fmt.Errorf("lc.yAxis.Details => %v", err)
	}

	xd, err := axes.NewXDetails(lc.maxPoints(), yd.Start, cvs.Area(), lc.xLabels)
	if err != nil {
		return fmt.Errorf("NewXDetails => %v", err)
	}

	if err := lc.drawAxes(cvs, xd, yd); err != nil {
		return err
	}
	return lc.drawSeries(cvs, xd, yd)
}

// drawAxes draws the X,Y axes and their labels.
func (lc *LineChart) drawAxes(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
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
		if err := draw.Text(cvs, l.Value.Text(), l.Pos, draw.TextCellOpts(lc.opts.xLabelCellOpts...)); err != nil {
			return fmt.Errorf("failed to draw the X labels: %v", err)
		}
	}
	return nil
}

// drawSeries draws the graph representing the stored series.
func (lc *LineChart) drawSeries(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	// The area available to the graph.
	graphAr := image.Rect(yd.Start.X+1, yd.Start.Y, cvs.Area().Max.X, xd.End.Y)
	bc, err := braille.New(graphAr)
	if err != nil {
		return fmt.Errorf("braille.New => %v", err)
	}

	var names []string
	for name := range lc.series {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		sv := lc.series[name]
		if len(sv.values) <= 1 {
			continue
		}

		prev := sv.values[0]
		for i := 1; i < len(sv.values); i++ {
			startX, err := xd.Scale.ValueToPixel(i - 1)
			if err != nil {
				return fmt.Errorf("failure for series %v[%d], xd.Scale.ValueToPixel => %v", name, i-1, err)
			}
			endX, err := xd.Scale.ValueToPixel(i)
			if err != nil {
				return fmt.Errorf("failure for series %v[%d], xd.Scale.ValueToPixel => %v", name, i, err)
			}

			startY, err := yd.Scale.ValueToPixel(prev)
			if err != nil {
				return fmt.Errorf("failure for series %v[%d], yd.Scale.ValueToPixel => %v", name, i-1, err)
			}
			v := sv.values[i]
			endY, err := yd.Scale.ValueToPixel(v)
			if err != nil {
				return fmt.Errorf("failure for series %v[%d], yd.Scale.ValueToPixel => %v", name, i, err)
			}

			if err := draw.BrailleLine(bc,
				image.Point{startX, startY},
				image.Point{endX, endY},
				draw.BrailleLineCellOpts(sv.seriesCellOpts...),
			); err != nil {
				return fmt.Errorf("draw.BrailleLine => %v", err)
			}
			prev = v
		}
	}
	if err := bc.CopyTo(cvs); err != nil {
		return fmt.Errorf("bc.Apply => %v", err)
	}
	return nil
}

// Keyboard implements widgetapi.Widget.Keyboard.
func (lc *LineChart) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the LineChart widget doesn't support keyboard events")
}

// Mouse implements widgetapi.Widget.Mouse.
func (lc *LineChart) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the LineChart widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (lc *LineChart) Options() widgetapi.Options {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// At the very least we need:
	// - n cells width for the Y axis and its labels as reported by it.
	// - at least 1 cell width for the graph.
	reqWidth := lc.yAxis.RequiredWidth() + 1
	// - 2 cells height the X axis and its values and 2 for min and max labels on Y.
	const reqHeight = 4
	return widgetapi.Options{
		MinimumSize: image.Point{reqWidth, reqHeight},
	}
}

// maxPoints returns the largest number of points among all the series.
// lc.mu must be held when calling this method.
func (lc *LineChart) maxPoints() int {
	max := 0
	for _, sv := range lc.series {
		if num := len(sv.values); num > max {
			max = num
		}
	}
	return max
}
