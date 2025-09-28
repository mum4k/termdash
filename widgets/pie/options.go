package pie

import (
	"github.com/mum4k/termdash/cell"
)

// Option defines a function that sets a specific option for the Pie widget.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// options stores the provided options.
type options struct {
	colors []cell.Color
}

// validates the provided options
// at the moment no validation is performed cause options are not required
func (o *options) validate() error {
	return nil
}

// ColorOption sets custom colors for the pie chart segments.
func ColorOption(colors []cell.Color) Option {
	return option(func(opts *options) {
		opts.colors = colors
	})
}

// newOptions creates a new options instance.
func newOptions() *options {
	return &options{
		colors: DefaultColors,
	}
}

// DefaultColors defines a default set of colors used for rendering pie chart segments.
// These colors are chosen from the predefined cell.Color constants and include a variety
// of primary and secondary colors to ensure visual distinction between segments.
var DefaultColors = []cell.Color{
	cell.ColorRed,
	cell.ColorGreen,
	cell.ColorBlue,
	cell.ColorYellow,
	cell.ColorMagenta,
	cell.ColorCyan,
	cell.ColorWhite,
}
