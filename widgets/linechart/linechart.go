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
	"sync"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille"
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

	// reqWidth is the last reported required with on a call to Options.
	reqWidth int

	// opts are the provided options.
	opts *options
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

// Series sets the values that should be displayed as the line chart with the
// provided label.
// Subsequent calls with the same label replace any previously provided values.
func (lc *LineChart) Series(label string, values []float64) error {
	if label == "" {
		return errors.New("the label cannot be empty")
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	series := newSeriesValues(values)
	lc.series[label] = series
	lc.yAxis = axes.NewY(series.min, series.max)
	return nil
}

// Draw draws the values as line charts.
// Implements widgetapi.Widget.Draw.
func (lc *LineChart) Draw(cvs *canvas.Canvas) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// The Y axis and its labels cannot take the entire canvas width, reserve one column for the line chart.
	maxWidth := cvs.Area().Dx() - 1
	yd, err := lc.yAxis.Details(cvs.Area().Dy()-2, maxWidth)
	if err != nil {
		return fmt.Errorf("lc.yAxis.Details => %v", err)
	}

	yStart := image.Point{yd.Width - 1, 0}
	xStart := image.Point{yStart.X, cvs.Area().Max.Y - 2} // One for X labels.
	yEnd := image.Point{yStart.X, xStart.Y}

	xWidth := cvs.Area().Dx() - yStart.X - 1
	xEnd := image.Point{xStart.X + xWidth - 1, xStart.Y}
	xd, err := axes.NewXDetails(lc.maxPoints(), xStart, xWidth)
	if err != nil {
		return fmt.Errorf("NewXDetails => %v", err)
	}

	lines := []draw.HVLine{
		{Start: yStart, End: yEnd},
		{Start: xStart, End: xEnd},
	}
	if err := draw.HVLines(cvs, lines); err != nil {
		return fmt.Errorf("failed to draw the axes: %v", err)
	}

	for _, l := range yd.Labels {
		if err := draw.Text(cvs, l.Value.Text(), l.Pos,
			draw.TextMaxX(yStart.X),
			draw.TextOverrunMode(draw.OverrunModeThreeDot),
		); err != nil {
			return fmt.Errorf("failed to draw the Y labels: %v", err)
		}
	}

	for _, l := range xd.Labels {
		if err := draw.Text(cvs, l.Value.Text(), l.Pos); err != nil {
			return fmt.Errorf("failed to draw the X labels: %v", err)
		}
	}

	ba := image.Rect(yStart.X+1, yStart.Y, cvs.Area().Max.X, xEnd.Y)
	bc, err := braille.New(ba)
	if err != nil {
		return fmt.Errorf("braille.New => %v", err)
	}
	for name, sv := range lc.series {
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

			if err := draw.BrailleLine(bc, image.Point{startX, startY}, image.Point{endX, endY}); err != nil {
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

// Implements widgetapi.Widget.Keyboard.
func (lc *LineChart) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the LineChart widget doesn't support keyboard events")
}

// Implements widgetapi.Widget.Mouse.
func (lc *LineChart) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the LineChart widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (lc *LineChart) Options() widgetapi.Options {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// At the very least we need:
	// - n columns for the Y axis and its values as reported by it.
	// - 2 rows for the X axis and its values.
	lc.reqWidth = lc.yAxis.RequiredWidth() + 1
	const reqHeight = 3
	return widgetapi.Options{
		// - 1 row and column for the line chart.
		MinimumSize: image.Point{lc.reqWidth, reqHeight},
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
