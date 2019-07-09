// Copyright 2019 Google Inc.
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

// Package testdotseg provides helpers for tests that use the dotseg package.
package testdotseg

import (
	"fmt"

	"termdash/internal/canvas"
	"termdash/internal/segdisp/dotseg"
)

// MustSetCharacter sets the character on the display or panics.
func MustSetCharacter(d *dotseg.Display, c rune) {
	if err := d.SetCharacter(c); err != nil {
		panic(fmt.Errorf("dotseg.Display.SetCharacter => unexpected error: %v", err))
	}
}

// MustDraw draws the display onto the canvas or panics.
func MustDraw(d *dotseg.Display, cvs *canvas.Canvas, opts ...dotseg.Option) {
	if err := d.Draw(cvs, opts...); err != nil {
		panic(fmt.Errorf("dotseg.Display.Draw => unexpected error: %v", err))
	}
}
