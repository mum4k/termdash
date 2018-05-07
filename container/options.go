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

package container

// options.go defines container options.

import (
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/widgetapi"
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
	// inherited are options that are inherited by child containers.
	inherited inherited

	// split identifies how is this container split.
	split splitType

	// widget is the widget in the container.
	// A container can have either two sub containers (left and right) or a
	// widget. But not both.
	widget widgetapi.Widget

	// Alignment of the widget if present.
	hAlign align.Horizontal
	vAlign align.Vertical

	// border is the border around the container.
	border draw.LineStyle
}

// inherited contains options that are inherited by child containers.
type inherited struct {
	// borderColor is the color used for the border.
	borderColor cell.Color
	// focusedColor is the color used for the border when focused.
	focusedColor cell.Color
}

// newOptions returns a new options instance with the default values.
// Parent are the inherited options from the parent container or nil if these
// options are for a container with no parent (the root).
func newOptions(parent *options) *options {
	opts := &options{
		inherited: inherited{
			focusedColor: cell.ColorYellow,
		},
	}
	if parent != nil {
		opts.inherited = parent.inherited
	}
	return opts
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
func PlaceWidget(w widgetapi.Widget) Option {
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
		c.opts.hAlign = align.HorizontalLeft
	})
}

// HorizontalAlignCenter aligns the placed widget in the center of the
// container along the horizontal axis. Has no effect if the container contains
// no widget.
func HorizontalAlignCenter() Option {
	return option(func(c *Container) {
		c.opts.hAlign = align.HorizontalCenter
	})
}

// HorizontalAlignRight aligns the placed widget on the right of the
// container along the horizontal axis. Has no effect if the container contains
// no widget.
func HorizontalAlignRight() Option {
	return option(func(c *Container) {
		c.opts.hAlign = align.HorizontalRight
	})
}

// VerticalAlignTop aligns the placed widget on the top of the
// container along the vertical axis. Has no effect if the container contains
// no widget. This is the default vertical alignment if no other is specified.
func VerticalAlignTop() Option {
	return option(func(c *Container) {
		c.opts.vAlign = align.VerticalTop
	})
}

// VerticalAlignMiddle aligns the placed widget in the middle of the
// container along the vertical axis. Has no effect if the container contains
// no widget.
func VerticalAlignMiddle() Option {
	return option(func(c *Container) {
		c.opts.vAlign = align.VerticalMiddle
	})
}

// VerticalAlignBottom aligns the placed widget at the bottom of the
// container along the vertical axis. Has no effect if the container contains
// no widget.
func VerticalAlignBottom() Option {
	return option(func(c *Container) {
		c.opts.vAlign = align.VerticalBottom
	})
}

// Border configures the container to have a border of the specified style.
func Border(ls draw.LineStyle) Option {
	return option(func(c *Container) {
		c.opts.border = ls
	})
}

// BorderColor sets the color of the border around the container.
// This option is inherited to sub containers created by container splits.
func BorderColor(color cell.Color) Option {
	return option(func(c *Container) {
		c.opts.inherited.borderColor = color
	})
}

// FocusedColor sets the color of the border around the container when it has
// keyboard focus.
// This option is inherited to sub containers created by container splits.
func FocusedColor(color cell.Color) Option {
	return option(func(c *Container) {
		c.opts.inherited.focusedColor = color
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
