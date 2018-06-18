// Package barchart implements a widget that displays multiple bars displaying
// values and their relative ratios.
package barchart

import (
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/widgetapi"
)

// BarChart displays multiple bars showing relative ratios of values.
//
// Each bar can have a text label under it explaining the meaning of the value
// it displays and can display the value itself inside the bar.
//
// Implements widgetapi.Widget. This object is thread-safe.
type BarChart struct {
	// values are the values provided on a call to Values(). These are the
	// individual bars that will be drawn.
	values []int
	// max is the maximum value of a bar. A bar having this value takes all the
	// vertical space.
	max int

	// mu protects the BarChart.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new BarChart.
func New(opts ...Option) *BarChart {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	return &BarChart{
		opts: opt,
	}
}

// Draw draws the BarChart widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (bc *BarChart) Draw(cvs *canvas.Canvas) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return nil
}

// Values sets the values to be displayed by the BarChart.
// Each value ends up in its own bar. The values must not be negative and must
// be less or equal the maximum value. A bar displaying the maximum value is a
// full bar, taking all available vertical space.
// Provided options override values set when New() was called.
func (bc *BarChart) Values(values []int, max int, opts ...Option) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if err := validateValues(values, max); err != nil {
		return err
	}

	for _, opt := range opts {
		opt.set(bc.opts)
	}
	bc.values = values
	bc.max = max
	return nil
}

// Options implements widgetapi.Widget.Options.
func (bc *BarChart) Options() widgetapi.Options {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return widgetapi.Options{
		MinimumSize:  bc.minSize(),
		WantKeyboard: false,
		WantMouse:    false,
	}
}

// minSize determines the minimum required size of the canvas.
func (bc *BarChart) minSize() image.Point {
	bars := len(bc.values)
	if bars == 0 {
		return image.Point{1, 1}
	}

	minHeight := 1 // At least one character vertically to display the bar.
	if len(bc.opts.labels) > 0 {
		minHeight++ // One line for the labels.
	}

	var minBarWidth int
	if bc.opts.barWidth < 1 {
		minBarWidth = 1 // At least one char for the bar itself.
	} else {
		minBarWidth = bc.opts.barWidth
	}
	minWidth := bars*minBarWidth + (bars-1)*bc.opts.barGap
	return image.Point{minWidth, minHeight}
}

// validateValues validates the provided values and maximum.
func validateValues(values []int, max int) error {
	if max < 1 {
		return fmt.Errorf("invalid maximum value %d, must be at least 1", max)
	}

	for i, v := range values {
		if v < 0 || v > max {
			return fmt.Errorf("invalid values[%d]: %d, each value must be 0 <= value <= max", i, v)
		}
	}
	return nil
}
