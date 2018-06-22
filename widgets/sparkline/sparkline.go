// Package sparkline is a widget that draws a graph showing a series of values as vertical bars.
package sparkline

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// SparkLine draws a graph showing a series of values as vertical bars.
//
// Bars can have sub-cell height. The graphs scale adjusts dynamically based on
// the largest displayed value or has a statically set maximum.
//
// Implements widgetapi.Widget. This object is thread-safe.
type SparkLine struct {
	// data are the data points the SparkLine displays.
	data []int

	// mu protects the SparkLine.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new SparkLine.
func New(opts ...Option) *SparkLine {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	return &SparkLine{
		opts: opt,
	}
}

// Draw draws the SparkLine widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (sl *SparkLine) Draw(cvs *canvas.Canvas) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return nil
}

// Add adds data points to the SparkLine.
// At least one data point must be provided. All data points must be positive
// integers.
// The last added data point will be the one displayed all the way on the right
// of the SparkLine.
func (sl *SparkLine) Add(data ...int) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	for i, d := range data {
		if d < 0 {
			return fmt.Errorf("data point[%d]: %v must be a positive integer", i, d)
		}
	}
	sl.data = append(sl.data, data...)
	return nil
}

// Clear removes all the data points in the sparkline, effectively returning to
// an empty graph.
func (sl *SparkLine) Clear() {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.data = nil
}

// Keyboard input isn't supported on the SparkLine widget.
func (*SparkLine) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the SparkLine widget doesn't support keyboard events")
}

// Mouse input isn't supported on the SparkLine widget.
func (*SparkLine) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the SparkLine widget doesn't support mouse events")
}

// minSize returns the minimum canvas size for the sparkline based on the options.
func (sl *SparkLine) minSize() image.Point {
	// At least one data point.
	const minWidth = 1
	var minHeight int
	if sl.opts.height > 0 {
		minHeight = sl.opts.height
	} else {
		minHeight = 1 // At least one line of characters.
	}

	if sl.opts.label != "" {
		minHeight++ // One line for the text label.
	}
	return image.Point{minWidth, minHeight}
}

// Options implements widgetapi.Widget.Options.
func (sl *SparkLine) Options() widgetapi.Options {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	min := sl.minSize()
	var max image.Point
	if sl.opts.height > 0 {
		max = min
	}

	return widgetapi.Options{
		MinimumSize:  min,
		MaximumSize:  max,
		WantKeyboard: false,
		WantMouse:    false,
	}
}
