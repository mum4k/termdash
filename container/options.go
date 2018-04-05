package container

// options.go defines container options.

import (
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/widget"
)

// applyOptions applies the options to the container.
func applyOptions(c *Container, opts ...Option) {
	for _, opt := range opts {
		opt.set(c)
	}
}

// Option is used to provide options to a container.
type Option interface {
	// set sets the provided option.
	set(*Container)
}

// options stores the options provided to the container.
type options struct {
	// split identifies how is this container split.
	split splitType

	// widget is the widget in the container.
	// A container can have either two sub containers (left and right) or a
	// widget. But not both.
	widget widget.Widget

	// Alignment of the widget if present.
	hAlign hAlignType
	vAlign vAlignType

	// border is the border around the container.
	border draw.LineStyle
	// borderColor is the color used for the border.
	borderColor cell.Color
}

// option implements Option.
type option func(*Container)

// set implements Option.set.
func (o option) set(c *Container) {
	o(c)
}

// SplitVertical splits the container along the vertical axis into two sub
// containers. The use of this option removes any widget placed at this
// container, containers with sub containers cannot contain widgets.
func SplitVertical(l LeftOption, r RightOption) Option {
	return option(func(c *Container) {
		c.opts.split = splitTypeVertical
		c.opts.widget = nil
		applyOptions(c.createFirst(), l.lOpts()...)
		applyOptions(c.createSecond(), r.rOpts()...)
	})
}

// SplitHorizontal splits the container along the horizontal axis into two sub
// containers. The use of this option removes any widget placed at this
// container, containers with sub containers cannot contain widgets.
func SplitHorizontal(t TopOption, b BottomOption) Option {
	return option(func(c *Container) {
		c.opts.split = splitTypeHorizontal
		c.opts.widget = nil
		applyOptions(c.createFirst(), t.tOpts()...)
		applyOptions(c.createSecond(), b.bOpts()...)
	})
}

// PlaceWidget places the provided widget into the container.
// The use of this option removes any sub containers. Containers with sub
// containers cannot have widgets.
func PlaceWidget(w widget.Widget) Option {
	return option(func(c *Container) {
		c.opts.widget = w
		c.first = nil
		c.second = nil
	})
}

// HorizontalAlignLeft aligns the placed widget on the left of the
// container along the horizontal axis. Has no effect if the container contains
// no widget. This is the default horizontal alignment if no other is specified.
func HorizontalAlignLeft() Option {
	return option(func(c *Container) {
		c.opts.hAlign = hAlignTypeLeft
	})
}

// HorizontalAlignCenter aligns the placed widget in the center of the
// container along the horizontal axis. Has no effect if the container contains
// no widget.
func HorizontalAlignCenter() Option {
	return option(func(c *Container) {
		c.opts.hAlign = hAlignTypeCenter
	})
}

// HorizontalAlignRight aligns the placed widget on the right of the
// container along the horizontal axis. Has no effect if the container contains
// no widget.
func HorizontalAlignRight() Option {
	return option(func(c *Container) {
		c.opts.hAlign = hAlignTypeRight
	})
}

// VerticalAlignTop aligns the placed widget on the top of the
// container along the vertical axis. Has no effect if the container contains
// no widget. This is the default vertical alignment if no other is specified.
func VerticalAlignTop() Option {
	return option(func(c *Container) {
		c.opts.vAlign = vAlignTypeTop
	})
}

// VerticalAlignMiddle aligns the placed widget in the middle of the
// container along the vertical axis. Has no effect if the container contains
// no widget.
func VerticalAlignMiddle() Option {
	return option(func(c *Container) {
		c.opts.vAlign = vAlignTypeMiddle
	})
}

// VerticalAlignBottom aligns the placed widget at the bottom of the
// container along the vertical axis. Has no effect if the container contains
// no widget.
func VerticalAlignBottom() Option {
	return option(func(c *Container) {
		c.opts.vAlign = vAlignTypeBottom
	})
}

// Border configures the container to have a border of the specified style.
func Border(ls draw.LineStyle) Option {
	return option(func(c *Container) {
		c.opts.border = ls
	})
}

// BorderColor sets the color of the border.
func BorderColor(color cell.Color) Option {
	return option(func(c *Container) {
		c.opts.borderColor = color
	})
}

// splitType identifies how a container is split.
type splitType int

// String implements fmt.Stringer()
func (st splitType) String() string {
	if n, ok := splitTypeNames[st]; ok {
		return n
	}
	return "splitTypeUnknown"
}

// splitTypeNames maps splitType values to human readable names.
var splitTypeNames = map[splitType]string{
	splitTypeVertical:   "splitTypeVertical",
	splitTypeHorizontal: "splitTypeHorizontal",
}

const (
	splitTypeVertical splitType = iota
	splitTypeHorizontal
)

// hAlignType indicates the horizontal alignment of the widget in the container.
type hAlignType int

// String implements fmt.Stringer()
func (hat hAlignType) String() string {
	if n, ok := hAlignTypeNames[hat]; ok {
		return n
	}
	return "hAlignTypeUnknown"
}

// hAlignTypeNames maps hAlignType values to human readable names.
var hAlignTypeNames = map[hAlignType]string{
	hAlignTypeLeft:   "hAlignTypeLeft",
	hAlignTypeCenter: "hAlignTypeCenter",
	hAlignTypeRight:  "hAlignTypeRight",
}

const (
	hAlignTypeLeft hAlignType = iota
	hAlignTypeCenter
	hAlignTypeRight
)

// vAlignType represents
type vAlignType int

// String implements fmt.Stringer()
func (vat vAlignType) String() string {
	if n, ok := vAlignTypeNames[vat]; ok {
		return n
	}
	return "vAlignTypeUnknown"
}

// vAlignTypeNames maps vAlignType values to human readable names.
var vAlignTypeNames = map[vAlignType]string{
	vAlignTypeTop:    "vAlignTypeTop",
	vAlignTypeMiddle: "vAlignTypeMiddle",
	vAlignTypeBottom: "vAlignTypeBottom",
}

const (
	vAlignTypeTop vAlignType = iota
	vAlignTypeMiddle
	vAlignTypeBottom
)

// LeftOption is used to provide options to the left sub container after a
// vertical split of the parent.
type LeftOption interface {
	// lOpts returns the options.
	lOpts() []Option
}

// leftOption implements LeftOption.
type leftOption func() []Option

// lOpts implements LeftOption.lOpts.
func (lo leftOption) lOpts() []Option {
	if lo == nil {
		return nil
	}
	return lo()
}

// Left applies options to the left sub container after a vertical split of the parent.
func Left(opts ...Option) LeftOption {
	return leftOption(func() []Option {
		return opts
	})
}

// RightOption is used to provide options to the right sub container after a
// vertical split of the parent.
type RightOption interface {
	// rOpts returns the options.
	rOpts() []Option
}

// rightOption implements RightOption.
type rightOption func() []Option

// rOpts implements RightOption.rOpts.
func (lo rightOption) rOpts() []Option {
	if lo == nil {
		return nil
	}
	return lo()
}

// Right applies options to the right sub container after a vertical split of the parent.
func Right(opts ...Option) RightOption {
	return rightOption(func() []Option {
		return opts
	})
}

// TopOption is used to provide options to the top sub container after a
// horizontal split of the parent.
type TopOption interface {
	// tOpts returns the options.
	tOpts() []Option
}

// topOption implements TopOption.
type topOption func() []Option

// tOpts implements TopOption.tOpts.
func (lo topOption) tOpts() []Option {
	if lo == nil {
		return nil
	}
	return lo()
}

// Top applies options to the top sub container after a horizontal split of the parent.
func Top(opts ...Option) TopOption {
	return topOption(func() []Option {
		return opts
	})
}

// BottomOption is used to provide options to the bottom sub container after a
// horizontal split of the parent.
type BottomOption interface {
	// bOpts returns the options.
	bOpts() []Option
}

// bottomOption implements BottomOption.
type bottomOption func() []Option

// bOpts implements BottomOption.bOpts.
func (lo bottomOption) bOpts() []Option {
	if lo == nil {
		return nil
	}
	return lo()
}

// Bottom applies options to the bottom sub container after a horizontal split of the parent.
func Bottom(opts ...Option) BottomOption {
	return bottomOption(func() []Option {
		return opts
	})
}
