package draw

// text.go contains code that prints UTF-8 encoded strings on the canvas.

import (
	"fmt"
	"image"

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
	OverrunModeStrict: "OverrunModeStrict",
}

const (
	// OverrunModeStrict verifies that the drawn value fits the canvas and
	// returns an error if it doesn't.
	OverrunModeStrict OverrunMode = iota

	// TODO(mum4k): Support other overrun modes, like Trim, ThreeDot or LineWrap.
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
	if maxX := tb.Start.X + len(text); maxX > wantMaxX && tb.Overrun == OverrunModeStrict {
		return fmt.Errorf("the requested text %q would end at X coordinate %v which falls outside of the maximum %v", text, maxX, wantMaxX)
	}

	cur := tb.Start
	for _, r := range text {
		if err := c.SetCell(cur, r, opts...); err != nil {
			return err
		}
		cur = image.Point{cur.X + 1, cur.Y}
	}
	return nil
}
