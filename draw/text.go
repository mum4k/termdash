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

// text.go contains code that prints UTF-8 encoded strings on the canvas.

import (
	"fmt"
	"image"
	"unicode/utf8"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
)

// OverrunMode represents
type OverrunMode int

// String implements fmt.Stringer()
func (om OverrunMode) String() string {
	if n, ok := overrunModeNames[om]; ok {
		return n
	}
	return "OverrunModeUnknown"
}

// overrunModeNames maps OverrunMode values to human readable names.
var overrunModeNames = map[OverrunMode]string{
	OverrunModeStrict:   "OverrunModeStrict",
	OverrunModeTrim:     "OverrunModeTrim",
	OverrunModeThreeDot: "OverrunModeThreeDot",
}

const (
	// OverrunModeStrict verifies that the drawn value fits the canvas and
	// returns an error if it doesn't.
	OverrunModeStrict OverrunMode = iota

	// OverrunModeTrim trims the part of the text that doesn't fit.
	OverrunModeTrim

	// OverrunModeThreeDot trims the text and places the horizontal ellipsis
	// '…' character at the end.
	OverrunModeThreeDot
)

// TextBounds specifies the limits (start and end cells) that the text must
// fall into and the overrun mode when it doesn't.
type TextBounds struct {
	// Start is the starting point of the drawn text.
	Start image.Point

	// MaxX sets a limit on the X coordinate (column) of the drawn text.
	// The X coordinate of all cells used by the text must be within
	// start.X <= X < MaxX.
	// This is optional, if set to zero, the width of the canvas is used as MaxX.
	// This cannot be negative or greater than the width of the canvas.
	MaxX int

	// Om indicates what to do with text that overruns the MaxX or the width of
	// the canvas if MaxX isn't specified.
	Overrun OverrunMode
}

// bounds enforces the text bounds based on the specified overrun mode.
// Returns test that can be safely drawn within the bounds.
func bounds(text string, maxRunes int, om OverrunMode) (string, error) {
	runes := utf8.RuneCountInString(text)
	if runes <= maxRunes {
		return text, nil
	}

	switch om {
	case OverrunModeStrict:
		return "", fmt.Errorf("the requested text %q takes %d runes to draw, space is available for only %d runes and overrun mode is %v", text, runes, maxRunes, om)
	case OverrunModeTrim:
		return text[:maxRunes], nil

	case OverrunModeThreeDot:
		return fmt.Sprintf("%s…", text[:maxRunes-1]), nil
	default:
		return "", fmt.Errorf("unsupported overrun mode %v", om)
	}
}

// Text prints the provided text on the canvas.
func Text(c *canvas.Canvas, text string, tb TextBounds, opts ...cell.Option) error {
	ar := c.Area()
	if !tb.Start.In(ar) {
		return fmt.Errorf("the requested start point %v falls outside of the provided canvas %v", tb.Start, ar)
	}

	if tb.MaxX < 0 || tb.MaxX > ar.Max.X {
		return fmt.Errorf("invalid TextBouds.MaxX %v, must be a positive number that is <= canvas.width %v", tb.MaxX, ar.Dx())
	}

	var wantMaxX int
	if tb.MaxX == 0 {
		wantMaxX = ar.Max.X
	} else {
		wantMaxX = tb.MaxX
	}

	maxRunes := wantMaxX - tb.Start.X
	trimmed, err := bounds(text, maxRunes, tb.Overrun)
	if err != nil {
		return err
	}

	cur := tb.Start
	for _, r := range trimmed {
		if err := c.SetCell(cur, r, opts...); err != nil {
			return err
		}
		cur = image.Point{cur.X + 1, cur.Y}
	}
	return nil
}
