/*
Package cell implements cell options and attributes.

A cell is the smallest point on the terminal.
*/
package cell

// Option is used to provide options for cells on a 2-D terminal.
type Option interface {
	// set sets the provided option.
	set(*Options)
}

// Options stores the provided options.
type Options struct {
	FgColor Color
	BgColor Color
}

// option implements Option.
type option func(*Options)

// set implements Option.set.
func (co option) set(opts *Options) {
	co(opts)
}

// FgColor sets the foreground color of the cell.
func FgColor(color Color) Option {
	return option(func(co *Options) {
		co.FgColor = color
	})
}

// BgColor sets the background color of the cell.
func BgColor(color Color) Option {
	return option(func(co *Options) {
		co.BgColor = color
	})
}
