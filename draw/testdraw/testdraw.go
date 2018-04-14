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
	"image"
	"log"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
)

// MustBox draws box on the canvas or panics.
func MustBox(c *canvas.Canvas, box image.Rectangle, ls draw.LineStyle, opts ...cell.Option) {
	if err := draw.Box(c, box, ls, opts...); err != nil {
		log.Fatalf("draw.Box => unexpected error: %v", err)
	}
}

// MustText draws the text on the canvas or panics.
func MustText(c *canvas.Canvas, text string, tb draw.TextBounds, opts ...cell.Option) {
	if err := draw.Text(c, text, tb, opts...); err != nil {
		log.Fatalf("draw.Text => unexpected error: %v", err)
	}
}
