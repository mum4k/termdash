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
	"fmt"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/widgetapi"
	"github.com/mum4k/termdash/linestyle"
)

// applyOptions applies the options to the container.
func applyOptions(c *Container, opts ...Option) error {
	for _, opt := range opts {
		if err := opt.set(c); err != nil {
			return err
		}
	}
	return nil
}

// Option is used to provide options to a container.
type Option interface {
	// set sets the provided option.
	set(*Container) error
}

// options stores the options provided to the container.
type options struct {
	// inherited are options that are inherited by child containers.
	inherited inherited

	// split identifies how is this container split.
	split        splitType
	splitPercent int

	// widget is the widget in the container.
	// A container can have either two sub containers (left and right) or a
	// widget. But not both.
	widget widgetapi.Widget

	// Alignment of the widget if present.
	hAlign align.Horizontal
	vAlign align.Vertical

	// border is the border around the container.
	border            linestyle.LineStyle
	borderTitle       string
	borderTitleHAlign align.Horizontal
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
		hAlign:       align.HorizontalCenter,
		vAlign:       align.VerticalMiddle,
		splitPercent: DefaultSplitPercent,
	}
	if parent != nil {
		opts.inherited = parent.inherited
	}
	return opts
}

// option implements Option.
type option func(*Container) error

// set implements Option.set.
func (o option) set(c *Container) error {
	return o(c)
}

// SplitOption is used when splitting containers.
type SplitOption interface {
	// setSplit sets the provided split option.
	setSplit(*options) error
}

// splitOption implements SplitOption.
type splitOption func(*options) error

// setSplit implements SplitOption.setSplit.
func (so splitOption) setSplit(opts *options) error {
	return so(opts)
}

// DefaultSplitPercent is the default value for the SplitPercent option.
const DefaultSplitPercent = 50

// SplitPercent sets the relative size of the split as percentage of the available space.
// When using SplitVertical, the provided size is applied to the new left
// container, the new right container gets the reminder of the size.
// When using SplitHorizontal, the provided size is applied to the new top
// container, the new bottom container gets the reminder of the size.
// The provided value must be a positive number in the range 0 < p < 100.
// If not provided, defaults to DefaultSplitPercent.
func SplitPercent(p int) SplitOption {
	return splitOption(func(opts *options) error {
		if min, max := 0, 100; p <= min || p >= max {
			return fmt.Errorf("invalid split percentage %d, must be in range %d < p < %d", p, min, max)
		}
		opts.splitPercent = p
		return nil
	})
}

// SplitVertical splits the container along the vertical axis into two sub
// containers. The use of this option removes any widget placed at this
// container, containers with sub containers cannot contain widgets.
func SplitVertical(l LeftOption, r RightOption, opts ...SplitOption) Option {
	return option(func(c *Container) error {
		c.opts.split = splitTypeVertical
		c.opts.widget = nil
		for _, opt := range opts {
			if err := opt.setSplit(c.opts); err != nil {
				return err
			}
		}

		f, err := c.createFirst()
		if err != nil {
			return err
		}
		if err := applyOptions(f, l.lOpts()...); err != nil {
			return err
		}

		s, err := c.createSecond()
		if err != nil {
			return err
		}
		return applyOptions(s, r.rOpts()...)
	})
}

// SplitHorizontal splits the container along the horizontal axis into two sub
// containers. The use of this option removes any widget placed at this
// container, containers with sub containers cannot contain widgets.
func SplitHorizontal(t TopOption, b BottomOption, opts ...SplitOption) Option {
	return option(func(c *Container) error {
		c.opts.split = splitTypeHorizontal
		c.opts.widget = nil
		for _, opt := range opts {
			if err := opt.setSplit(c.opts); err != nil {
				return err
			}
		}

		f, err := c.createFirst()
		if err != nil {
			return err
		}
		if err := applyOptions(f, t.tOpts()...); err != nil {
			return err
		}

		s, err := c.createSecond()
		if err != nil {
			return err
		}
		return applyOptions(s, b.bOpts()...)
	})
}

// PlaceWidget places the provided widget into the container.
// The use of this option removes any sub containers. Containers with sub
// containers cannot have widgets.
func PlaceWidget(w widgetapi.Widget) Option {
	return option(func(c *Container) error {
		c.opts.widget = w
		c.first = nil
		c.second = nil
		return nil
	})
}

// AlignHorizontal sets the horizontal alignment for the widget placed in the
// container. Has no effect if the container contains no widget.
// Defaults to alignment in the center.
func AlignHorizontal(h align.Horizontal) Option {
	return option(func(c *Container) error {
		c.opts.hAlign = h
		return nil
	})
}

// AlignVertical sets the vertical alignment for the widget placed in the container.
// Has no effect if the container contains no widget.
// Defaults to alignment in the middle.
func AlignVertical(v align.Vertical) Option {
	return option(func(c *Container) error {
		c.opts.vAlign = v
		return nil
	})
}

// Border configures the container to have a border of the specified style.
func Border(ls linestyle.LineStyle) Option {
	return option(func(c *Container) error {
		c.opts.border = ls
		return nil
	})
}

// BorderTitle sets a text title within the border.
func BorderTitle(title string) Option {
	return option(func(c *Container) error {
		c.opts.borderTitle = title
		return nil
	})
}

// BorderTitleAlignLeft aligns the border title on the left.
func BorderTitleAlignLeft() Option {
	return option(func(c *Container) error {
		c.opts.borderTitleHAlign = align.HorizontalLeft
		return nil
	})
}

// BorderTitleAlignCenter aligns the border title in the center.
func BorderTitleAlignCenter() Option {
	return option(func(c *Container) error {
		c.opts.borderTitleHAlign = align.HorizontalCenter
		return nil
	})
}

// BorderTitleAlignRight aligns the border title on the right.
func BorderTitleAlignRight() Option {
	return option(func(c *Container) error {
		c.opts.borderTitleHAlign = align.HorizontalRight
		return nil
	})
}

// BorderColor sets the color of the border around the container.
// This option is inherited to sub containers created by container splits.
func BorderColor(color cell.Color) Option {
	return option(func(c *Container) error {
		c.opts.inherited.borderColor = color
		return nil
	})
}

// FocusedColor sets the color of the border around the container when it has
// keyboard focus.
// This option is inherited to sub containers created by container splits.
func FocusedColor(color cell.Color) Option {
	return option(func(c *Container) error {
		c.opts.inherited.focusedColor = color
		return nil
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
