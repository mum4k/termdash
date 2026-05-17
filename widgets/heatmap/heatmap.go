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
	"fmt"
	"image"
	"math"
	"strconv"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/heatmap/internal/axes"
)

// HeatMap draws heat map charts.
//
// Heatmap consists of several cells. Each cell represents a value.
// By default, larger values appear darker in a grayscale ramp. Call Palette to
// provide a custom low-to-high color ramp instead.
//
// The two dimensions of the values (cells) array are determined by the length of
// the xLabels and yLabels arrays respectively.
//
// HeatMap does not support mouse based zoom.
//
// Implements widgetapi.Widget. This object is thread-safe.
type HeatMap struct {
	// values are the values in the heat map.
	values [][]float64

	// xLabels are the labels on the X axis in an increasing order.
	xLabels []string
	// yLabels are the labels on the Y axis in an increasing order.
	yLabels []string

	// minValue and maxValue are the Min and Max values in the values,
	// which will be used to calculate the color of each cell.
	minValue, maxValue float64

	// lastWidth is the width of the canvas as of the last time when Draw was called.
	lastWidth int
	// lastCapacity is the most recent visible value capacity derived from Draw.
	lastCapacity int

	// opts are the provided options.
	opts *options

	// mu protects the HeatMap widget.
	mu sync.RWMutex
}

// New returns a new HeatMap widget.
func New(opts ...Option) (*HeatMap, error) {
	opt := newOptions(opts...)
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &HeatMap{
		opts: opt,
	}, nil
}

// Values sets the values to be displayed by the HeatMap.
//
// Each value in values has a xLabel and a yLabel, which means
// len(yLabels) == len(values) and len(xLabels) == len(values[i]).
// But labels could be empty strings.
// When no labels are provided, labels will be "0", "1", "2"...
//
// Each call to Values overwrites any previously provided values.
// Provided options override values set when New() was called.
func (hp *HeatMap) Values(xLabels []string, yLabels []string, values [][]float64, opts ...Option) error {
	clonedValues, minValue, maxValue, err := cloneValues(values)
	if err != nil {
		return err
	}

	nextOpts := *hp.opts
	for _, opt := range opts {
		opt.set(&nextOpts)
	}
	if err := nextOpts.validate(); err != nil {
		return err
	}

	rows := len(clonedValues)
	cols := len(clonedValues[0])
	xs, err := normalizeLabels(xLabels, cols)
	if err != nil {
		return fmt.Errorf("invalid x labels: %w", err)
	}
	ys, err := normalizeLabels(yLabels, rows)
	if err != nil {
		return fmt.Errorf("invalid y labels: %w", err)
	}

	hp.mu.Lock()
	defer hp.mu.Unlock()
	hp.values = clonedValues
	hp.xLabels = xs
	hp.yLabels = ys
	hp.minValue = minValue
	hp.maxValue = maxValue
	hp.opts = &nextOpts
	return nil
}

// ClearXLabels clear the X labels.
func (hp *HeatMap) ClearXLabels() {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	hp.xLabels = nil
}

// ClearYLabels clear the Y labels.
func (hp *HeatMap) ClearYLabels() {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	hp.yLabels = nil
}

// ValueCapacity returns the number of values that can fit into the canvas.
// This is essentially the number of available cells on the canvas as observed
// on the last call to draw. Returns zero if draw wasn't called.
//
// Note that this capacity changes each time the terminal resizes, so there is
// no guarantee this remains the same next time Draw is called.
// Should be used as a hint only.
func (hp *HeatMap) ValueCapacity() int {
	hp.mu.RLock()
	defer hp.mu.RUnlock()
	return hp.lastCapacity
}

// axesDetails determines the details about the X and Y axes.
func (hp *HeatMap) axesDetails(cvs *canvas.Canvas) (*axes.XDetails, *axes.YDetails, error) {
	yd, err := axes.NewYDetails(hp.yLabels)
	if err != nil {
		return nil, nil, err
	}
	xOriginY := len(hp.values) - 1
	if xOriginY < 0 {
		xOriginY = 0
	}
	xOriginX := yd.Width - 1
	if yd.Width == 0 {
		xOriginX = -1
	}
	xd, err := axes.NewXDetails(cvs.Area(), image.Point{X: xOriginX, Y: xOriginY}, hp.xLabels, hp.opts.cellWidth)
	if err != nil {
		return nil, nil, err
	}
	return xd, yd, nil
}

// Draw draws cells, X labels and Y labels as HeatMap.
// Implements widgetapi.Widget.Draw.
func (hp *HeatMap) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	_ = meta

	hp.mu.Lock()
	defer hp.mu.Unlock()

	hp.lastWidth = cvs.Area().Dx()
	hp.lastCapacity = hp.visibleCapacity(cvs.Area().Size())

	if len(hp.values) == 0 {
		return nil
	}
	if need := hp.minSize(); need.X > cvs.Area().Dx() || need.Y > cvs.Area().Dy() {
		return draw.ResizeNeeded(cvs)
	}

	xd, yd, err := hp.axesDetails(cvs)
	if err != nil {
		return err
	}
	if err := hp.drawCells(cvs, xd, yd); err != nil {
		return err
	}
	return hp.drawLabels(cvs, xd, yd)
}

// drawCells draws m*n cells (rectangles) representing the stored values.
// The height of each cell is 1 and the default width is 3.
func (hp *HeatMap) drawCells(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	_ = xd
	xOffset := yd.Width
	for row, vals := range hp.values {
		for col, value := range vals {
			cellArea := image.Rect(
				xOffset+col*hp.opts.cellWidth,
				row,
				xOffset+(col+1)*hp.opts.cellWidth,
				row+1,
			)
			if err := cvs.SetAreaCells(cellArea, ' ', cell.BgColor(hp.getCellColor(value))); err != nil {
				return err
			}
		}
	}
	return nil
}

// drawAxes draws X labels (under the cells) and Y Labels (on the left side of the cell).
func (hp *HeatMap) drawLabels(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	for _, label := range yd.Labels {
		if err := draw.Text(cvs, label.Text, label.Pos, draw.TextOverrunMode(draw.OverrunModeTrim), draw.TextCellOpts(hp.opts.yLabelCellOpts...)); err != nil {
			return err
		}
	}
	if yd.Width > 0 {
		for row := yd.Start.Y; row <= yd.End.Y; row++ {
			if _, err := cvs.SetCell(image.Point{X: yd.Start.X, Y: row}, '│', hp.opts.axisCellOpts...); err != nil {
				return err
			}
		}
	}

	for _, label := range xd.Labels {
		if err := draw.Text(cvs, label.Text, label.Pos, draw.TextOverrunMode(draw.OverrunModeTrim), draw.TextCellOpts(hp.opts.xLabelCellOpts...)); err != nil {
			return err
		}
	}
	return nil
}

// minSize determines the minimum required size to draw HeatMap.
func (hp *HeatMap) minSize() image.Point {
	if len(hp.values) == 0 {
		return image.Point{}
	}
	yDetails, _ := axes.NewYDetails(hp.yLabels)
	cols := 0
	if len(hp.values) > 0 {
		cols = len(hp.values[0])
	}
	height := len(hp.values)
	if len(hp.xLabels) > 0 {
		height++
	}
	return image.Point{
		X: yDetails.Width + cols*hp.opts.cellWidth,
		Y: height,
	}
}

// Keyboard input isn't supported on the HeatMap widget.
func (*HeatMap) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return nil
}

// Mouse input isn't supported on the HeatMap widget.
func (*HeatMap) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return nil
}

// Options implements widgetapi.Widget.Options.
func (hp *HeatMap) Options() widgetapi.Options {
	hp.mu.RLock()
	defer hp.mu.RUnlock()
	return widgetapi.Options{
		MinimumSize: hp.minSize(),
	}
}

// getCellColor returns the color of the cell according to its value.
//
// When a palette is configured, values are projected across that palette from
// low to high. Otherwise the original grayscale mapping is used.
func (hp *HeatMap) getCellColor(value float64) cell.Color {
	ratio := hp.valueRatio(value)
	if hp.opts != nil && len(hp.opts.palette) > 0 {
		return hp.paletteColor(ratio)
	}
	return hp.grayscaleColor(ratio)
}

// valueRatio normalizes a value into the inclusive range [0, 1].
func (hp *HeatMap) valueRatio(value float64) float64 {
	if hp.maxValue <= hp.minValue {
		return 0
	}
	ratio := (value - hp.minValue) / (hp.maxValue - hp.minValue)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	return ratio
}

// paletteColor maps a normalized ratio to the configured palette.
func (hp *HeatMap) paletteColor(ratio float64) cell.Color {
	if hp.opts == nil || len(hp.opts.palette) == 0 {
		return cell.ColorNumber(244)
	}
	if len(hp.opts.palette) == 1 {
		return hp.opts.palette[0]
	}
	index := int(math.Round(ratio * float64(len(hp.opts.palette)-1)))
	if index < 0 {
		index = 0
	}
	if index >= len(hp.opts.palette) {
		index = len(hp.opts.palette) - 1
	}
	return hp.opts.palette[index]
}

// grayscaleColor keeps the original grayscale low-to-high mapping.
func (hp *HeatMap) grayscaleColor(ratio float64) cell.Color {
	// Higher values should appear darker, so walk the xterm grayscale range
	// from 255 (light) down to 232 (dark).
	shade := 255 - int(math.Round(ratio*23))
	if shade < 232 {
		shade = 232
	}
	if shade > 255 {
		shade = 255
	}
	return cell.ColorNumber(shade)
}

func cloneValues(values [][]float64) ([][]float64, float64, float64, error) {
	if len(values) == 0 {
		return nil, 0, 0, fmt.Errorf("values must contain at least one row")
	}

	cloned := make([][]float64, len(values))
	cols := len(values[0])
	if cols == 0 {
		return nil, 0, 0, fmt.Errorf("values must contain at least one column")
	}

	minValue := values[0][0]
	maxValue := values[0][0]
	for row, vals := range values {
		if len(vals) != cols {
			return nil, 0, 0, fmt.Errorf("values[%d] length %d, want %d", row, len(vals), cols)
		}
		cloned[row] = append([]float64(nil), vals...)
		for _, value := range vals {
			if value < minValue {
				minValue = value
			}
			if value > maxValue {
				maxValue = value
			}
		}
	}
	return cloned, minValue, maxValue, nil
}

func normalizeLabels(labels []string, count int) ([]string, error) {
	if len(labels) == 0 {
		out := make([]string, count)
		for i := 0; i < count; i++ {
			out[i] = strconv.Itoa(i)
		}
		return out, nil
	}
	if len(labels) != count {
		return nil, fmt.Errorf("label count %d, want %d", len(labels), count)
	}
	return append([]string(nil), labels...), nil
}

func (hp *HeatMap) visibleCapacity(size image.Point) int {
	if len(hp.values) == 0 || hp.opts == nil {
		return 0
	}
	yDetails, _ := axes.NewYDetails(hp.yLabels)
	graphWidth := size.X - yDetails.Width
	graphHeight := size.Y
	if len(hp.xLabels) > 0 {
		graphHeight--
	}
	if graphWidth <= 0 || graphHeight <= 0 {
		return 0
	}
	cols := graphWidth / hp.opts.cellWidth
	if cols < 0 {
		cols = 0
	}
	return cols * graphHeight
}
