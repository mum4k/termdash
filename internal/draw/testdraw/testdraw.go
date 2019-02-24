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

// Package testdraw provides helpers for tests that use the draw package.
package testdraw

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/braille"
	"github.com/mum4k/termdash/internal/draw"
)

// MustBorder draws border on the canvas or panics.
func MustBorder(c *canvas.Canvas, border image.Rectangle, opts ...draw.BorderOption) {
	if err := draw.Border(c, border, opts...); err != nil {
		panic(fmt.Sprintf("draw.Border => unexpected error: %v", err))
	}
}

// MustText draws the text on the canvas or panics.
func MustText(c *canvas.Canvas, text string, start image.Point, opts ...draw.TextOption) {
	if err := draw.Text(c, text, start, opts...); err != nil {
		panic(fmt.Sprintf("draw.Text => unexpected error: %v", err))
	}
}

// MustVerticalText draws the vertical text on the canvas or panics.
func MustVerticalText(c *canvas.Canvas, text string, start image.Point, opts ...draw.VerticalTextOption) {
	if err := draw.VerticalText(c, text, start, opts...); err != nil {
		panic(fmt.Sprintf("draw.VerticalText => unexpected error: %v", err))
	}
}

// MustRectangle draws the rectangle on the canvas or panics.
func MustRectangle(c *canvas.Canvas, r image.Rectangle, opts ...draw.RectangleOption) {
	if err := draw.Rectangle(c, r, opts...); err != nil {
		panic(fmt.Sprintf("draw.Rectangle => unexpected error: %v", err))
	}
}

// MustHVLines draws the vertical / horizontal lines or panics.
func MustHVLines(c *canvas.Canvas, lines []draw.HVLine, opts ...draw.HVLineOption) {
	if err := draw.HVLines(c, lines, opts...); err != nil {
		panic(fmt.Sprintf("draw.HVLines => unexpected error: %v", err))
	}
}

// MustBrailleLine draws the braille line or panics.
func MustBrailleLine(bc *braille.Canvas, start, end image.Point, opts ...draw.BrailleLineOption) {
	if err := draw.BrailleLine(bc, start, end, opts...); err != nil {
		panic(fmt.Sprintf("draw.BrailleLine => unexpected error: %v", err))
	}
}

// MustBrailleCircle draws the braille circle or panics.
func MustBrailleCircle(bc *braille.Canvas, mid image.Point, radius int, opts ...draw.BrailleCircleOption) {
	if err := draw.BrailleCircle(bc, mid, radius, opts...); err != nil {
		panic(fmt.Sprintf("draw.BrailleCircle => unexpected error: %v", err))
	}
}

// MustResizeNeeded draws the character or panics.
func MustResizeNeeded(cvs *canvas.Canvas) {
	if err := draw.ResizeNeeded(cvs); err != nil {
		panic(fmt.Sprintf("draw.ResizeNeeded => unexpected error: %v", err))
	}
}
