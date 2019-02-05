// Package testsixteen provides helpers for tests that use the sixteen package.
package testsixteen

import (
	"fmt"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/draw/segdisp/sixteen"
)

// MustSetCharacter sets the character on the display or panics.
func MustSetCharacter(d *sixteen.Display, c rune) {
	if err := d.SetCharacter(c); err != nil {
		panic(fmt.Errorf("sixteen.Display.SetCharacter => unexpected error: %v", err))
	}
}

// MustDraw draws the display onto the canvas or panics.
func MustDraw(d *sixteen.Display, cvs *canvas.Canvas, opts ...sixteen.Option) {
	if err := d.Draw(cvs, opts...); err != nil {
		panic(fmt.Errorf("sixteen.Display.Draw => unexpected error: %v", err))
	}
}
