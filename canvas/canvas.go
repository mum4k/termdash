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

// Package canvas defines the canvas that the widgets draw on.
package canvas

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminalapi"
)

// Canvas is where a widget draws its output for display on the terminal.
type Canvas struct {
	// area is the area the buffer was created for.
	// Contains absolute coordinates on the target terminal, while the buffer
	// contains relative zero-based coordinates for this canvas.
	area image.Rectangle

	// buffer is where the drawing happens.
	buffer cell.Buffer
}

// New returns a new Canvas with a buffer for the provided area.
func New(ar image.Rectangle) (*Canvas, error) {
	if ar.Min.X < 0 || ar.Min.Y < 0 || ar.Max.X < 0 || ar.Max.Y < 0 {
		return nil, fmt.Errorf("area cannot start or end on the negative axis, got: %+v", ar)
	}

	b, err := cell.NewBuffer(area.Size(ar))
	if err != nil {
		return nil, err
	}
	return &Canvas{
		area:   ar,
		buffer: b,
	}, nil
}

// Size returns the size of the 2-D canvas.
func (c *Canvas) Size() image.Point {
	return c.buffer.Size()
}

// Area returns the area of the 2-D canvas.
func (c *Canvas) Area() image.Rectangle {
	s := c.buffer.Size()
	return image.Rect(0, 0, s.X, s.Y)
}

// Clear clears all the content on the canvas.
func (c *Canvas) Clear() error {
	b, err := cell.NewBuffer(c.Size())
	if err != nil {
		return err
	}
	c.buffer = b
	return nil
}

// SetCell sets the value of the specified cell on the canvas.
// Use the options to specify which attributes to modify, if an attribute
// option isn't specified, the attribute retains its previous value.
func (c *Canvas) SetCell(p image.Point, r rune, opts ...cell.Option) error {
	ar, err := area.FromSize(c.buffer.Size())
	if err != nil {
		return err
	}
	if !p.In(ar) {
		return fmt.Errorf("cell at point %+v falls out of the canvas area %+v", p, ar)
	}

	cell := c.buffer[p.X][p.Y]
	cell.Rune = r
	cell.Apply(opts...)
	return nil
}

// Apply applies the canvas to the corresponding area of the terminal.
// Guarantees to stay within limits of the area the canvas was created with.
func (c *Canvas) Apply(t terminalapi.Terminal) error {
	termArea, err := area.FromSize(t.Size())
	if err != nil {
		return err
	}

	bufArea, err := area.FromSize(c.buffer.Size())
	if err != nil {
		return err
	}

	if !bufArea.In(termArea) {
		return fmt.Errorf("the canvas area %+v doesn't fit onto the terminal %+v", bufArea, termArea)
	}

	for col := range c.buffer {
		for row := range c.buffer[col] {
			cell := c.buffer[col][row]
			// The image.Point{0, 0} of this canvas isn't always exactly at
			// image.Point{0, 0} on the terminal.
			// Depends on area assigned by the container.
			offset := c.area.Min
			p := image.Point{col, row}.Add(offset)
			if err := t.SetCell(p, cell.Rune, cell.Opts); err != nil {
				return fmt.Errorf("terminal.SetCell(%+v) => error: %v", p, err)
			}
		}
	}
	return nil
}
