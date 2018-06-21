package sparkline

import "github.com/mum4k/termdash/cell"

// options.go contains configurable options for SparkLine.

// Option is used to provide options.
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

// options holds the provided options.
type options struct {
	label         string
	labelCellOpts []cell.Option
	height        int
	color         cell.Color
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{}
}

// Label adds a label above the SparkLine.
func Label(text string, cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.label = text
		opts.labelCellOpts = cOpts
	})
}

// Height sets a fixed height for the SparkLine.
// If not provided, the SparkLine takes all the available vertical space in the
// container.
func Height(h int) Option {
	return option(func(opts *options) {
		opts.height = h
	})
}

// DefaultColor is the default value for the Color option.
const DefaultColor = cell.ColorBlue

// Color sets the color of the SparkLine.
// Defaults to DefaultColor if not set.
func Color(c cell.Color) Option {
	return option(func(opts *options) {
		opts.color = c
	})
}
