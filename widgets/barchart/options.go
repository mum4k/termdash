package barchart

// options.go contains configurable options for BarChart.

import (
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// options holds the provided options.
type options struct {
	barChar     rune
	barWidth    int
	barGap      int
	showValues  bool
	barColors   []cell.Color
	labelColors []cell.Color
	valueColors []cell.Color
	labels      []string
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		barChar: DefaultChar,
		barGap:  DefaultBarGap,
	}
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// DefaultChar is the default value for the Char option.
const DefaultChar = draw.DefaultRectChar

// Char sets the rune that is used when drawing the rectangle representing the
// bars.
func Char(ch rune) Option {
	return option(func(opts *options) {
		opts.barChar = ch
	})
}

// BarWidth sets the width of the bars. If not set, the bars use all the space
// available to the widget.
func BarWidth(width int) Option {
	return option(func(opts *options) {
		opts.barWidth = width
	})
}

// DefaultBarGap is the default value for the BarGap option.
const DefaultBarGap = 1

// BarGap sets the width of the space between the bars.
// Defaults to DefaultBarGap.
func BarGap(width int) Option {
	return option(func(opts *options) {
		opts.barGap = width
	})
}

// ShowValues tells the bar chart to display the actual values inside each of the bars.
func ShowValues() Option {
	return option(func(opts *options) {
		opts.showValues = true
	})
}

// DefaultBarColor is the default color of a bar, unless specified otherwise
// via the BarColors option.
const DefaultBarColor = cell.ColorRed

// BarColors sets the colors of each of the bars.
// Bars are created on a call to Values(), each value ends up in its own Bar.
// The first supplied color applies to the bar displaying the first value.
// Any bars that don't have a color specified use the DefaultBarColor.
func BarColors(colors []cell.Color) Option {
	return option(func(opts *options) {
		opts.barColors = colors
	})
}

// DefaultLabelColor is the default color of a bar label, unless specified
// otherwise via the LabelColors option.
const DefaultLabelColor = cell.ColorGreen

// LabelColors sets the colors of each of the labels under the bars.
// Bars are created on a call to Values(), each value ends up in its own Bar.
// The first supplied color applies to the label of the bar displaying the
// first value. Any labels that don't have a color specified use the
// DefaultLabelColor.
func LabelColors(colors []cell.Color) Option {
	return option(func(opts *options) {
		opts.labelColors = colors
	})
}

// Labels sets the labels displayed under each bar,
// Bars are created on a call to Values(), each value ends up in its own Bar.
// The first supplied label applies to the bar displaying the first value.
// If not specified, the corresponding bar (or all the bars) don't have a
// label.
func Labels(labels []string) Option {
	return option(func(opts *options) {
		opts.labels = labels
	})
}

// DefaultValueColor is the default color of a bar value, unless specified
// otherwise via the ValueColors option.
const DefaultValueColor = cell.ColorYellow

// ValueColors sets the colors of each of the values in the bars. Bars are
// created on a call to Values(), each value ends up in its own Bar. The first
// supplied color applies to the bar displaying the first value. Any values
// that don't have a color specified use the DefaultValueColor.
func ValueColors(colors []cell.Color) Option {
	return option(func(opts *options) {
		opts.valueColors = colors
	})
}
