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

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/faketerm"
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

// MustSetCell sets the cell value or panics.
func MustSetCell(c *canvas.Canvas, p image.Point, r rune, opts ...cell.Option) {
	if err := c.SetCell(p, r, opts...); err != nil {
		panic(fmt.Sprintf("canvas.SetCell => unexpected error: %v", err))
	}
}
