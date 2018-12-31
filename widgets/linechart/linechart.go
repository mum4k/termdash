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
	"image"
	"sync"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

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

	// opts are the provided options.
	opts *options
}

// New returns a new line chart widget.
func New(opts ...Option) *LineChart {
	opt := newOptions(opts...)
	return &LineChart{
		opts: opt,
	}
}

// Series sets the values that should be displayed as the line chart with the
// provided label.
// Subsequent calls with the same label replace any previously provided values.
func (lc *LineChart) Series(label string, values []float64) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	return errors.New("unimplemented")
}

// Draw draws the values as line charts.
// Implements widgetapi.Widget.Draw.
func (lc *LineChart) Draw(cvs *canvas.Canvas) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	return errors.New("unimplemented")
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
	return widgetapi.Options{
		// At the very least we need:
		// - 2 columns for the Y axis and its values.
		// - 2 rows for the X axis and its values.
		// - 1 row for the line chart.
		MinimumSize: image.Point{2, 3},
	}
}
