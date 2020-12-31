// Copyright 2021 Google Inc.
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
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/draw"
	"image"
	"math"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/heatmap/internal/axes"
)

// HeatMap draws heat map charts.
//
// Heatmap consists of several cells. Each cell represents a value.
// The larger the value, the darker the color of the cell (from white to black).
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
	// lastWidth is the height of the canvas as of the last time when Draw was called.
	lastHeight int

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
	if len(values) == 0 {
		return errors.New("the values cannot be empty")
	}

	hp.mu.Lock()
	defer hp.mu.Unlock()

	var (
		xl []string
		yl []string
		v  [][]float64
	)

	if xLabels == nil {
		// Set initial labels if it is not provided.
		xl = initLabels(len(values[0]))
	} else {
		// Copy to avoid external modifications.
		xl = make([]string, len(xLabels))
		copy(xl, xLabels)
	}

	if yLabels == nil {
		yl = initLabels(len(values))
	} else {
		yl = make([]string, len(yLabels))
		copy(yl, yLabels)
	}

	if len(yl) != len(values) || len(xl) != len(values[0]) {
		return errors.New("the number of labels does not match the shape of the values")
	}

	v = make([][]float64, len(values))
	copy(v, values)

	hp.xLabels = xl
	hp.yLabels = yl
	hp.values = v
	for _, opt := range opts {
		opt.set(hp.opts)
	}
	hp.minValue, hp.maxValue = minMax(hp.values)
	return nil
}

// ClearXLabels clear the X labels.
func (hp *HeatMap) ClearXLabels() {
	hp.xLabels = nil
}

// ClearYLabels clear the Y labels.
func (hp *HeatMap) ClearYLabels() {
	hp.yLabels = nil
}

// ValueCapacity returns the number of rows and columns of the heat map
// that can fit into the canvas.
//
// This is essentially the number of available cells on the canvas as observed
// on the last call to draw. Returns zero if draw wasn't called.
//
// Note that this capacity changes each time the terminal resizes, so there is
// no guarantee this remains the same next time Draw is called.
// Should be used as a hint only.
func (hp *HeatMap) ValueCapacity() (rows, cols int) {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	if hp.lastWidth == 0 || hp.lastHeight == 0 {
		return 0, 0
	}

	rows = hp.lastHeight - 1
	var cw int

	if hp.opts.cellWidth > minCellWidth {
		cw = hp.opts.cellWidth
	} else {
		cw = minCellWidth
	}
	cols = int(math.Floor(float64(hp.lastWidth-axes.LongestString(hp.yLabels)-axes.AxisWidth) / float64(cw)))
	return
}

// axesDetails determines the details about the X and Y axes.
func (hp *HeatMap) axesDetails(cvs *canvas.Canvas) (*axes.XDetails, *axes.YDetails, error) {
	hp.cellWidthAdaptive(cvs)

	yd, err := axes.NewYDetails(hp.yLabels)
	if err != nil {
		return nil, nil, err
	}

	xd, err := axes.NewXDetails(yd.End, hp.xLabels, hp.opts.cellWidth)
	if err != nil {
		return nil, nil, err
	}

	return xd, yd, nil
}

// Draw draws cells, X labels and Y labels as HeatMap.
// Implements widgetapi.Widget.Draw.
func (hp *HeatMap) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	hp.lastWidth = cvs.Area().Dx()
	hp.lastHeight = cvs.Area().Dy()
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
	err = hp.drawCells(cvs, yd)
	if err != nil {
		return err
	}

	return hp.drawLabels(cvs, xd, yd)
}

// drawCells draws m*n cells (rectangles) representing the stored values.
// The height of each cell is 1 and the default width is 3.
func (hp *HeatMap) drawCells(cvs *canvas.Canvas, yd *axes.YDetails) error {
	for i := 0; i < len(hp.values); i++ {
		for j := 0; j < len(hp.values[0]); j++ {
			startX := yd.Start.X + axes.AxisWidth + j*hp.opts.cellWidth
			startY := yd.Labels[i].Pos.Y

			endX := startX + hp.opts.cellWidth
			endY := startY + 1

			rect := image.Rect(startX, startY, endX, endY)
			color := hp.getCellColor(hp.values[i][j])

			if err := cvs.SetAreaCells(rect, hp.opts.cellChar, cell.BgColor(color)); err != nil {
				return err
			}
		}
	}

	return nil
}

// drawAxes draws X labels (under the cells) and Y Labels (on the left side of the cell).
func (hp *HeatMap) drawLabels(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	if !hp.opts.hideYLabels {
		for _, l := range yd.Labels {
			if err := draw.Text(cvs, l.Text, l.Pos,
				draw.TextCellOpts(hp.opts.yLabelCellOpts...),
			); err != nil {
				return fmt.Errorf("failed to draw the Y labels: %v", err)
			}
		}
	}

	if !hp.opts.hideXLabels {
		for _, l := range xd.Labels {
			if err := draw.Text(cvs, l.Text, l.Pos, draw.TextCellOpts(hp.opts.xLabelCellOpts...)); err != nil {
				return fmt.Errorf("failed to draw the X labels: %v", err)
			}
		}
	}
	return nil
}

// minCellWidth is the minimum width of each cell in the heat map.
const minCellWidth = 3

// cellWidthAdaptive determines the width of a single cell (grid) based the canvas.
func (hp *HeatMap) cellWidthAdaptive(cvs *canvas.Canvas) {
	rem := cvs.Area().Dx() - axes.LongestString(hp.yLabels) - axes.AxisWidth
	var cw int
	if len(hp.values) != 0 && len(hp.values[0]) != 0 {
		cw = rem / len(hp.values[0])
	}
	if cw >= minCellWidth {
		hp.opts.cellWidth = cw
	} else {
		hp.opts.cellWidth = minCellWidth
	}
}

// minSize determines the minimum required size to draw HeatMap.
func (hp *HeatMap) minSize() image.Point {
	// At the very least we need:
	// - n unit width for the Y axis and its labels.
	// - m unit width for the graph.
	// cells is the number of cells in a row.
	var cells int
	if len(hp.values) != 0 {
		cells = len(hp.values[0])
	}
	reqWidth := axes.LongestString(hp.yLabels) + axes.AxisWidth + minCellWidth*cells

	// For the height:
	// - 1 unit height for labels on the X axis.
	// - n unit height for the graph.
	reqHeight := 1 + len(hp.values)

	return image.Point{X: reqWidth, Y: reqHeight}
}

// Keyboard input isn't supported on the HeatMap widget.
func (*HeatMap) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return errors.New("the HeatMap widget doesn't support keyboard events")
}

// Mouse input isn't supported on the HeatMap widget.
func (*HeatMap) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return errors.New("the HeatMap widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (hp *HeatMap) Options() widgetapi.Options {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	return widgetapi.Options{}
}

// getCellColor returns the color of the cell according to its value.
// The larger the value, the darker the color.
// The color range is in Xterm color, from 232 to 255.
// Refer to https://jonasjacek.github.io/colors/.
func (hp *HeatMap) getCellColor(value float64) cell.Color {
	const colorNum = 23
	scale := hp.maxValue - hp.minValue
	rgb := int(255 - ((value - hp.minValue) / scale * colorNum))
	return cell.ColorNumber(rgb)
}

// initLabels return initial labels, like '0', '1', '2', ...
func initLabels(l int) []string {
	var ret []string
	for i := 0; i < l; i++ {
		ret = append(ret, fmt.Sprintf("%d", i))
	}
	return ret
}

// minMax returns the min and max values in given integer array.
func minMax(values [][]float64) (min, max float64) {
	min = math.MaxFloat64
	max = math.SmallestNonzeroFloat64

	for i := 0; i < len(values); i++ {
		for j := 0; j < len(values[i]); j++ {
			min = math.Min(min, values[i][j])
			max = math.Max(max, values[i][j])
		}
	}
	return
}
