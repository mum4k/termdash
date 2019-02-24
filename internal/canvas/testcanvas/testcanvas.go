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

// Package testcanvas provides helpers for tests that use the canvas package.
package testcanvas

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/faketerm"
)

// MustNew returns a new canvas or panics.
func MustNew(area image.Rectangle) *canvas.Canvas {
	cvs, err := canvas.New(area)
	if err != nil {
		panic(fmt.Sprintf("canvas.New => unexpected error: %v", err))
	}
	return cvs
}

// MustApply applies the canvas on the terminal or panics.
func MustApply(c *canvas.Canvas, t *faketerm.Terminal) {
	if err := c.Apply(t); err != nil {
		panic(fmt.Sprintf("canvas.Apply => unexpected error: %v", err))
	}
}

// MustSetCell sets the cell value or panics. Returns the number of cells the
// rune occupies, wide runes can occupy multiple cells when printed on the
// terminal. See http://www.unicode.org/reports/tr11/.
func MustSetCell(c *canvas.Canvas, p image.Point, r rune, opts ...cell.Option) int {
	cells, err := c.SetCell(p, r, opts...)
	if err != nil {
		panic(fmt.Sprintf("canvas.SetCell => unexpected error: %v", err))
	}
	return cells
}

// MustSetAreaCells sets the cells in the area  or panics.
func MustSetAreaCells(c *canvas.Canvas, cellArea image.Rectangle, r rune, opts ...cell.Option) {
	if err := c.SetAreaCells(cellArea, r, opts...); err != nil {
		panic(fmt.Sprintf("canvas.SetAreaCells => unexpected error: %v", err))
	}
}

// MustCell returns the cell or panics.
func MustCell(c *canvas.Canvas, p image.Point) *cell.Cell {
	cell, err := c.Cell(p)
	if err != nil {
		panic(fmt.Sprintf("canvas.Cell => unexpected error: %v", err))
	}
	return cell
}

// MustCopyTo copies the content of the source canvas onto the destination
// canvas or panics.
func MustCopyTo(src, dst *canvas.Canvas) {
	if err := src.CopyTo(dst); err != nil {
		panic(fmt.Sprintf("canvas.CopyTo => unexpected error: %v", err))
	}
}
