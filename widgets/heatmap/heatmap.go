// Copyright 2020 Google Inc.
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

// Package heatmap contains a widget that displays heat maps.
package heatmap

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
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/heatmap/internal/axes"
)

// columnValues represent values stored in a column.
type columnValues struct {
	// values are the values in a column.
	values []int64
	// Min is the smallest value in the column, zero if values is empty.
	Min int64
	// Max is the largest value in the column, zero if values is empty.
	Max int64
}

// newColumnValues returns a new columnValues instance.
func newColumnValues(values []int64) *columnValues {
	// Copy to avoid external modifications.
	v := make([]int64, len(values))
	copy(v, values)

	min, max := minMax(values)

	return &columnValues{
		values: v,
		Min:    min,
		Max:    max,
	}
}

// HeatMap draws heatmap charts.
// Implements widgetapi.Widget. This object is thread-safe.
type HeatMap struct {
	columns map[string]*columnValues

	// XLabels are the labels on the X axis in an increasing order.
	XLabels []string
	// YLabels are the labels on the Y axis in an increasing order.
	YLabels []string

	// MinValue and MaxValue are the Min and Max values in the columns.
	MinValue, MaxValue int64

	// opts are the provided options.
	opts *options

	// mu protects the HeatMap widget.
	mu sync.RWMutex
}

// NewHeatMap returns a new HeatMap widget.
func NewHeatMap(opts ...Option) (*HeatMap, error) {
	opt := newOptions(opts...)
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &HeatMap{
		columns: map[string]*columnValues{},
		opts:    opt,
	}, nil
}

// SetColumns sets the HeatMap's values, min and max values.
func (hp *HeatMap) SetColumns(values map[string][]int64) {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	var minMaxValues []int64

	// The iteration order of map is uncertain, so the keys must be sorted explicitly.
	var names []string
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	// Clear XLabels and columns.
	if len(hp.XLabels) > 0 {
		hp.XLabels = hp.XLabels[:0]
	}
	hp.columns = make(map[string]*columnValues)

	for _, name := range names {
		cv := newColumnValues(values[name])
		hp.columns[name] = cv
		hp.XLabels = append(hp.XLabels, name)

		minMaxValues = append(minMaxValues, cv.Min)
		minMaxValues = append(minMaxValues, cv.Max)
	}

	hp.MinValue, hp.MaxValue = minMax(minMaxValues)
}

// SetYLabels sets HeatMap's Y-Labels.
func (hp *HeatMap) SetYLabels(labels []string) {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	// Clear YLabels.
	if len(hp.YLabels) > 0 {
		hp.YLabels = hp.YLabels[:0]
	}

	hp.YLabels = append(hp.YLabels, labels...)

	// Reverse the array.
	for i, j := 0, len(hp.YLabels)-1; i < j; i, j = i+1, j-1 {
		hp.YLabels[i], hp.YLabels[j] = hp.YLabels[j], hp.YLabels[i]
	}
}

// axesDetails determines the details about the X and Y axes.
func (hp *HeatMap) axesDetails(cvs *canvas.Canvas) (*axes.XDetails, *axes.YDetails, error) {
	yd, err := axes.NewYDetails(hp.YLabels)
	if err != nil {
		return nil, nil, err
	}

	xd, err := axes.NewXDetails(cvs.Area(), yd.End, hp.XLabels, hp.opts.cellWidth)
	if err != nil {
		return nil, nil, err
	}

	return xd, yd, nil
}

// Draw draws the values as HeatMap.
// Implements widgetapi.Widget.Draw.
func (hp *HeatMap) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	// Check if the canvas has enough area to draw HeatMap.
	needAr, err := area.FromSize(hp.minSize())
	if err != nil {
		return err
	}
	if !needAr.In(cvs.Area()) {
		return draw.ResizeNeeded(cvs)
	}

	xd, yd, err := hp.axesDetails(cvs)
	if err != nil {
		return err
	}

	err = hp.drawColumns(cvs, xd, yd)
	if err != nil {
		return err
	}

	return hp.drawAxes(cvs, xd, yd)
}

// drawColumns draws the graph representing the stored series.
// Returns XDetails that might be adjusted to not start at zero value if some
// of the series didn't fit the graphs and XAxisUnscaled was provided.
// If the series has NaN values they will be ignored and not draw on the graph.
func (hp *HeatMap) drawColumns(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	for i, xl := range hp.XLabels {
		cv := hp.columns[xl]

		for j := 0; j < len(cv.values); j++ {
			v := cv.values[j]

			startX := xd.Start.X + 1 + i*hp.opts.cellWidth
			startY := yd.Labels[j].Pos.Y

			endX := startX + hp.opts.cellWidth
			endY := startY + 1

			rect := image.Rect(startX, startY, endX, endY)
			color := hp.getBlockColor(v)

			if err := cvs.SetAreaCells(rect, ' ', cell.BgColor(color)); err != nil {
				return err
			}
		}
	}

	return nil
}

// drawAxes draws the X,Y axes and their labels.
func (hp *HeatMap) drawAxes(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	for _, l := range yd.Labels {
		if err := draw.Text(cvs, l.Text, l.Pos,
			draw.TextMaxX(yd.Start.X),
			draw.TextOverrunMode(draw.OverrunModeThreeDot),
			draw.TextCellOpts(hp.opts.yLabelCellOpts...),
		); err != nil {
			return fmt.Errorf("failed to draw the Y labels: %v", err)
		}
	}

	for _, l := range xd.Labels {
		if err := draw.Text(cvs, l.Text, l.Pos, draw.TextCellOpts(hp.opts.xLabelCellOpts...)); err != nil {
			return fmt.Errorf("failed to draw the X horizontal labels: %v", err)
		}
	}
	return nil
}

// minSize determines the minimum required size to draw HeatMap.
func (hp *HeatMap) minSize() image.Point {
	// At the very least we need:
	// - n cells width for the Y axis and its labels.
	// - m cells width for the graph.
	reqWidth := axes.LongestString(hp.YLabels) + axes.AxisWidth + hp.opts.cellWidth*len(hp.columns)

	// For the height:
	// - 1 cells height for labels on the X axis.
	// - n cell height for the graph.
	reqHeight := 1 + len(hp.YLabels)

	return image.Point{X: reqWidth, Y: reqHeight}
}

// Keyboard input isn't supported on the SparkLine widget.
func (*HeatMap) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the HeatMap widget doesn't support keyboard events")
}

// Mouse input isn't supported on the SparkLine widget.
func (*HeatMap) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the HeatMap widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (hp *HeatMap) Options() widgetapi.Options {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	return widgetapi.Options{}
}

// getBlockColor returns the color of the block according to the value.
// The larger the value, the darker the color.
func (hp *HeatMap) getBlockColor(value int64) cell.Color {
	const colorNum = 23
	scale := float64(hp.MaxValue - hp.MinValue)
	fv := float64(value)

	// Refer to https://jonasjacek.github.io/colors/.
	// The color range is in Xterm color [232, 255].
	rgb := int(255 - (fv / scale * colorNum))
	return cell.ColorNumber(rgb)
}

// minMax returns the min and max values in given integer array.
func minMax(values []int64) (min, max int64) {
	min = math.MaxInt64
	max = math.MinInt64

	for _, v := range values {
		min = int64(math.Min(float64(min), float64(v)))
		max = int64(math.Max(float64(max), float64(v)))
	}
	return
}
