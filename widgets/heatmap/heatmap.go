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
	"image"
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

	// MinValue and MaxValue are the Min and Max values in the values,
	// which will be used to calculate the color of each cell.
	MinValue, MaxValue float64

	// opts are the provided options.
	opts *options

	// mu protects the HeatMap widget.
	mu sync.RWMutex
}

// New returns a new HeatMap widget.
func New(opts ...Option) (*HeatMap, error) {
	return nil, errors.New("not implemented")
}

// Values sets the values to be displayed by the HeatMap.
// Each value in values has a xLabel and a yLabel, which means
// len(xLabels) == len(values) and len(yLabels) == len(values[i]).
// Provided options override values set when New() was called.
func (hp *HeatMap) Values(xLabels []string, yLabels []string, values [][]float64, opts ...Option) error {
	return errors.New("not implemented")
}

// axesDetails determines the details about the X and Y axes.
func (hp *HeatMap) axesDetails(cvs *canvas.Canvas) (*axes.XDetails, *axes.YDetails, error) {
	return nil, nil, errors.New("not implemented")
}

// Draw draws cells, X labels and Y labels as HeatMap.
// Implements widgetapi.Widget.Draw.
func (hp *HeatMap) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	return errors.New("not implemented")
}

// drawCells draws m*n cells (rectangles) representing the stored values.
// The height of each cell is 1 and the default width is 3.
func (hp *HeatMap) drawCells(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	return errors.New("not implemented")
}

// drawAxes draws X labels (under the cells) and Y Labels (on the left side of the cell).
func (hp *HeatMap) drawLabels(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	return errors.New("not implemented")
}

// minSize determines the minimum required size to draw HeatMap.
func (hp *HeatMap) minSize() image.Point {
	return image.Point{}
}

// Keyboard input isn't supported on the HeatMap widget.
func (*HeatMap) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the HeatMap widget doesn't support keyboard events")
}

// Mouse input isn't supported on the HeatMap widget.
func (*HeatMap) Mouse(m *terminalapi.Mouse) error {
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
	return cell.ColorDefault
}
