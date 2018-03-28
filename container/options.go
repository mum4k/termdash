package container

// options.go defines container options.

import "github.com/mum4k/termdash/widget"

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
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
	border borderType
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// PlaceWidget places the provided widget into the container.
func PlaceWidget(w widget.Widget) Option {
	return option(func(opts *options) {
		opts.widget = w
	})
}

// SplitHorizontal configures the container for a horizontal split.
func SplitHorizontal() Option {
	return option(func(opts *options) {
		opts.split = splitTypeHorizontal
	})
}

// SplitVertical configures the container for a vertical split.
// This is the default split type if neither if SplitHorizontal() or
// SplitVertical() is specified.
func SplitVertical() Option {
	return option(func(opts *options) {
		opts.split = splitTypeVertical
	})
}

// HorizontalAlignLeft aligns the placed widget on the left of the
// container along the horizontal axis. Has no effect if the container contains
// no widget. This is the default horizontal alignment if no other is specified.
func HorizontalAlignLeft() Option {
	return option(func(opts *options) {
		opts.hAlign = hAlignTypeLeft
	})
}

// HorizontalAlignCenter aligns the placed widget in the center of the
// container along the horizontal axis. Has no effect if the container contains
// no widget.
func HorizontalAlignCenter() Option {
	return option(func(opts *options) {
		opts.hAlign = hAlignTypeCenter
	})
}

// HorizontalAlignRight aligns the placed widget on the right of the
// container along the horizontal axis. Has no effect if the container contains
// no widget.
func HorizontalAlignRight() Option {
	return option(func(opts *options) {
		opts.hAlign = hAlignTypeRight
	})
}

// VerticalAlignTop aligns the placed widget on the top of the
// container along the vertical axis. Has no effect if the container contains
// no widget. This is the default vertical alignment if no other is specified.
func VerticalAlignTop() Option {
	return option(func(opts *options) {
		opts.vAlign = vAlignTypeTop
	})
}

// VerticalAlignMiddle aligns the placed widget in the middle of the
// container along the vertical axis. Has no effect if the container contains
// no widget.
func VerticalAlignMiddle() Option {
	return option(func(opts *options) {
		opts.vAlign = vAlignTypeMiddle
	})
}

// VerticalAlignBottom aligns the placed widget at the bottom of the
// container along the vertical axis. Has no effect if the container contains
// no widget.
func VerticalAlignBottom() Option {
	return option(func(opts *options) {
		opts.vAlign = vAlignTypeBottom
	})
}

// BorderNone configures the container to have no border.
// This is the default if none of the Border options is specified.
func BorderNone() Option {
	return option(func(opts *options) {
		opts.border = borderTypeNone
	})
}

// BorderSolid configures the container to have a border made with a solid
// line.
func BorderSolid() Option {
	return option(func(opts *options) {
		opts.border = borderTypeSolid
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

// borderType represents
type borderType int

// String implements fmt.Stringer()
func (bt borderType) String() string {
	if n, ok := borderTypeNames[bt]; ok {
		return n
	}
	return "borderTypeUnknown"
}

// borderTypeNames maps borderType values to human readable names.
var borderTypeNames = map[borderType]string{
	borderTypeNone:  "borderTypeNone",
	borderTypeSolid: "borderTypeSolid",
}

const (
	borderTypeNone borderType = iota
	borderTypeSolid
)
