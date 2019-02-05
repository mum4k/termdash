package segmentdisplay

// write_options.go contains options used when writing content to the widget.

import "github.com/mum4k/termdash/cell"

// WriteOption is used to provide options to Write().
type WriteOption interface {
	// set sets the provided option.
	set(*writeOptions)
}

// writeOptions stores the provided options.
type writeOptions struct {
	cellOpts         []cell.Option
	errOnUnsupported bool
}

// newWriteOptions returns new writeOptions instance.
func newWriteOptions(wOpts ...WriteOption) *writeOptions {
	wo := &writeOptions{}
	for _, o := range wOpts {
		o.set(wo)
	}
	return wo
}

// writeOption implements WriteOption.
type writeOption func(*writeOptions)

// set implements WriteOption.set.
func (wo writeOption) set(wOpts *writeOptions) {
	wo(wOpts)
}

// WriteCellOpts sets options on the cells that contain the text.
func WriteCellOpts(opts ...cell.Option) WriteOption {
	return writeOption(func(wOpts *writeOptions) {
		wOpts.cellOpts = opts
	})
}

// WriteSanitize instructs Write to sanitize the text, replacing all characters
// the display doesn't support with a space ' ' character.
// This is the default behavior.
func WriteSanitize(opts ...cell.Option) WriteOption {
	return writeOption(func(wOpts *writeOptions) {
		wOpts.errOnUnsupported = false
	})
}

// WriteErrOnUnsupported instructs Write to return an error when the text
// contains a character the display doesn't support.
// The default behavior is to sanitize the text, see WriteSanitize().
func WriteErrOnUnsupported(opts ...cell.Option) WriteOption {
	return writeOption(func(wOpts *writeOptions) {
		wOpts.errOnUnsupported = true
	})
}
