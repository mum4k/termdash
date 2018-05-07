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

package draw

// border.go contains code that draws borders.

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
)

// BorderOption is used to provide options to Border().
type BorderOption interface {
	// set sets the provided option.
	set(*borderOptions)
}

// borderOptions stores the provided options.
type borderOptions struct {
	cellOpts  []cell.Option
	lineStyle LineStyle
}

// borderOption implements BorderOption.
type borderOption func(bOpts *borderOptions)

// set implements BorderOption.set.
func (bo borderOption) set(bOpts *borderOptions) {
	bo(bOpts)
}

// DefaultBorderLineStyle is the default value for the BorderLineStyle option.
const DefaultBorderLineStyle = LineStyleLight

// BorderLineStyle sets the style of the line used to draw the border.
func BorderLineStyle(ls LineStyle) BorderOption {
	return borderOption(func(bOpts *borderOptions) {
		bOpts.lineStyle = ls
	})
}

// BorderCellOpts sets options on the cells that create the border.
func BorderCellOpts(opts ...cell.Option) BorderOption {
	return borderOption(func(bOpts *borderOptions) {
		bOpts.cellOpts = opts
	})
}

// borderChar returns the correct border character from the parts for the use
// at the specified point of the border. Returns -1 if no character should be at
// this point.
func borderChar(p image.Point, border image.Rectangle, parts map[linePart]rune) rune {
	switch {
	case p.X == border.Min.X && p.Y == border.Min.Y:
		return parts[topLeftCorner]
	case p.X == border.Max.X-1 && p.Y == border.Min.Y:
		return parts[topRightCorner]
	case p.X == border.Min.X && p.Y == border.Max.Y-1:
		return parts[bottomLeftCorner]
	case p.X == border.Max.X-1 && p.Y == border.Max.Y-1:
		return parts[bottomRightCorner]
	case p.X == border.Min.X || p.X == border.Max.X-1:
		return parts[vLine]
	case p.Y == border.Min.Y || p.Y == border.Max.Y-1:
		return parts[hLine]
	}
	return -1
}

// Border draws a border on the canvas.
func Border(c *canvas.Canvas, border image.Rectangle, opts ...BorderOption) error {
	if ar := c.Area(); !border.In(ar) {
		return fmt.Errorf("the requested border %+v falls outside of the provided canvas %+v", border, ar)
	}

	const minSize = 2
	if border.Dx() < minSize || border.Dy() < minSize {
		return fmt.Errorf("the smallest supported border is %dx%d, got: %dx%d", minSize, minSize, border.Dx(), border.Dy())
	}

	opt := &borderOptions{
		lineStyle: DefaultBorderLineStyle,
	}
	for _, o := range opts {
		o.set(opt)
	}

	parts, err := lineParts(opt.lineStyle)
	if err != nil {
		return err
	}

	for col := border.Min.X; col < border.Max.X; col++ {
		for row := border.Min.Y; row < border.Max.Y; row++ {
			p := image.Point{col, row}
			r := borderChar(p, border, parts)
			if r == -1 {
				continue
			}

			if err := c.SetCell(p, r, opt.cellOpts...); err != nil {
				return err
			}
		}
	}
	return nil
}
