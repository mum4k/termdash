// Copyright 2026 Google Inc.
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

package modal

import (
	"image"

	"github.com/mum4k/termdash/cell"
)

// Option configures modal widget options.
type Option interface {
	// set applies the option to the provided options.
	set(*Options)
}

// Options holds modal configuration.
type Options struct {
	// Border controls whether draggable items draw a rounded border.
	Border bool

	// MinimumSize is the smallest canvas that the modal will request.
	MinimumSize image.Point

	// TitleBarCellOpts styles the draggable window title bar background.
	TitleBarCellOpts []cell.Option

	// TitleCellOpts styles the draggable window title text.
	TitleCellOpts []cell.Option

	// TitleControlCellOpts styles the title-bar control glyphs.
	TitleControlCellOpts []cell.Option

	// MinimizeGlyph is drawn in the title bar while the window is expanded.
	MinimizeGlyph rune

	// RestoreGlyph is drawn in the title bar while the window is minimized.
	RestoreGlyph rune

	// DockGap is the spacing used between minimized windows in the bottom dock.
	DockGap int
}

// option implements Option.
type option func(*Options)

// set implements Option.set.
func (o option) set(opts *Options) {
	o(opts)
}

// newOptions returns modal options with defaults applied.
func newOptions() *Options {
	return &Options{
		Border:      true,
		MinimumSize: image.Point{X: 10, Y: 5},
		TitleBarCellOpts: []cell.Option{
			cell.BgColor(cell.ColorNumber(238)),
			cell.FgColor(cell.ColorNumber(252)),
		},
		TitleCellOpts: []cell.Option{
			cell.FgColor(cell.ColorNumber(252)),
			cell.Bold(),
		},
		TitleControlCellOpts: []cell.Option{
			cell.FgColor(cell.ColorNumber(159)),
			cell.Bold(),
		},
		MinimizeGlyph: '▁',
		RestoreGlyph:  '▢',
		DockGap:       1,
	}
}

// NewOptions creates an Options value with the provided options applied.
func NewOptions(opts ...Option) *Options {
	o := newOptions()
	for _, opt := range opts {
		opt.set(o)
	}
	if o.MinimumSize.X < 1 {
		o.MinimumSize.X = 1
	}
	if o.MinimumSize.Y < 1 {
		o.MinimumSize.Y = 1
	}
	if o.MinimizeGlyph == 0 {
		o.MinimizeGlyph = '▁'
	}
	if o.RestoreGlyph == 0 {
		o.RestoreGlyph = '▢'
	}
	if o.DockGap < 0 {
		o.DockGap = 0
	}
	return o
}

// Border sets whether draggable items should draw rounded borders.
func Border(border bool) Option {
	return option(func(o *Options) {
		o.Border = border
	})
}

// MinimumSize sets the minimum canvas size requested by the modal widget.
func MinimumSize(size image.Point) Option {
	return option(func(o *Options) {
		o.MinimumSize = size
	})
}

// TitleBarCellOpts sets the cell styling used for the title bar background.
func TitleBarCellOpts(opts ...cell.Option) Option {
	return option(func(o *Options) {
		o.TitleBarCellOpts = append([]cell.Option(nil), opts...)
	})
}

// TitleCellOpts sets the cell styling used for the window title text.
func TitleCellOpts(opts ...cell.Option) Option {
	return option(func(o *Options) {
		o.TitleCellOpts = append([]cell.Option(nil), opts...)
	})
}

// TitleControlCellOpts sets the cell styling used for title bar control glyphs.
func TitleControlCellOpts(opts ...cell.Option) Option {
	return option(func(o *Options) {
		o.TitleControlCellOpts = append([]cell.Option(nil), opts...)
	})
}

// MinimizeGlyphs sets the glyphs used for minimize and restore controls.
func MinimizeGlyphs(minimize, restore rune) Option {
	return option(func(o *Options) {
		if minimize != 0 {
			o.MinimizeGlyph = minimize
		}
		if restore != 0 {
			o.RestoreGlyph = restore
		}
	})
}

// DockGap sets the spacing used between minimized windows in the bottom dock.
func DockGap(gap int) Option {
	return option(func(o *Options) {
		o.DockGap = gap
	})
}

// EnableLogging is retained as a compatibility no-op.
//
// Deprecated: modal logging has been removed. This option has no effect.
func EnableLogging(enable bool) Option {
	return option(func(o *Options) {
		_ = enable
	})
}

// LogPrefix is retained as a compatibility no-op.
//
// Deprecated: modal logging has been removed. This option has no effect.
func LogPrefix(prefix string) Option {
	return option(func(o *Options) {
		_ = prefix
	})
}
