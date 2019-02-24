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

// Package testbraille provides helpers for tests that use the braille package.
package testbraille

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/braille"
	"github.com/mum4k/termdash/internal/cell"
	"github.com/mum4k/termdash/terminal/faketerm"
)

// MustNew returns a new canvas or panics.
func MustNew(area image.Rectangle) *braille.Canvas {
	cvs, err := braille.New(area)
	if err != nil {
		panic(fmt.Sprintf("braille.New => unexpected error: %v", err))
	}
	return cvs
}

// MustApply applies the canvas on the terminal or panics.
func MustApply(bc *braille.Canvas, t *faketerm.Terminal) {
	if err := bc.Apply(t); err != nil {
		panic(fmt.Sprintf("braille.Apply => unexpected error: %v", err))
	}
}

// MustSetPixel sets the specified pixel or panics.
func MustSetPixel(bc *braille.Canvas, p image.Point, opts ...cell.Option) {
	if err := bc.SetPixel(p, opts...); err != nil {
		panic(fmt.Sprintf("braille.SetPixel => unexpected error: %v", err))
	}
}

// MustClearPixel clears the specified pixel or panics.
func MustClearPixel(bc *braille.Canvas, p image.Point, opts ...cell.Option) {
	if err := bc.ClearPixel(p, opts...); err != nil {
		panic(fmt.Sprintf("braille.ClearPixel => unexpected error: %v", err))
	}
}

// MustCopyTo copies the braille canvas onto the provided canvas or panics.
func MustCopyTo(bc *braille.Canvas, dst *canvas.Canvas) {
	if err := bc.CopyTo(dst); err != nil {
		panic(fmt.Sprintf("bc.CopyTo => unexpected error: %v", err))
	}
}

// MustSetCellOpts sets the cell options or panics.
func MustSetCellOpts(bc *braille.Canvas, cellPoint image.Point, opts ...cell.Option) {
	if err := bc.SetCellOpts(cellPoint, opts...); err != nil {
		panic(fmt.Sprintf("bc.SetCellOpts => unexpected error: %v", err))
	}
}

// MustSetAreaCellOpts sets the cell options in the area or panics.
func MustSetAreaCellOpts(bc *braille.Canvas, cellArea image.Rectangle, opts ...cell.Option) {
	if err := bc.SetAreaCellOpts(cellArea, opts...); err != nil {
		panic(fmt.Sprintf("bc.SetAreaCellOpts => unexpected error: %v", err))
	}
}
