// Copyright 2026 Google Inc.
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

package fx

import (
	"image"
	"strings"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/buffer"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

func TestFramedDrawsBorderTitleAndInner(t *testing.T) {
	inner := &framedTestWidget{
		draw: func(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
			if _, err := cvs.SetCell(image.Point{X: 0, Y: 0}, 'X', cell.FgColor(cell.ColorRed)); err != nil {
				return err
			}
			if _, err := cvs.SetCell(image.Point{X: 1, Y: 1}, 'Y'); err != nil {
				return err
			}
			return nil
		},
	}
	fw, err := FramedNew(inner, FramedTitle("Demo"), FramedBorderOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))
	if err != nil {
		t.Fatalf("FramedNew => unexpected error: %v", err)
	}

	cvs, err := canvas.New(image.Rect(0, 0, 14, 5))
	if err != nil {
		t.Fatalf("canvas.New => unexpected error: %v", err)
	}
	if err := fw.Draw(cvs, &widgetapi.Meta{Focused: true}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	for _, tc := range []struct {
		p    image.Point
		want rune
	}{
		{image.Point{X: 0, Y: 0}, '╭'},
		{image.Point{X: 13, Y: 0}, '╮'},
		{image.Point{X: 0, Y: 4}, '╰'},
		{image.Point{X: 13, Y: 4}, '╯'},
		{image.Point{X: 1, Y: 1}, 'X'},
		{image.Point{X: 2, Y: 2}, 'Y'},
	} {
		if got := cellAt(t, cvs, tc.p).Rune; got != tc.want {
			t.Errorf("cell %v = %q, want %q", tc.p, got, tc.want)
		}
	}

	if got := rowString(t, cvs, 0); !strings.Contains(got, " Demo ") {
		t.Errorf("top row %q does not contain title", got)
	}
	if got := cellAt(t, cvs, image.Point{X: 0, Y: 0}).Opts; got.FgColor != cell.ColorGreen || !got.Bold {
		t.Errorf("border options = %+v, want green bold", got)
	}
}

func TestFramedSetBorderColorOverridesForeground(t *testing.T) {
	fw, err := FramedNew(&framedTestWidget{}, FramedBorderOpts(cell.FgColor(cell.ColorGreen), cell.Bold()))
	if err != nil {
		t.Fatalf("FramedNew => unexpected error: %v", err)
	}
	fw.SetBorderColor(cell.ColorNumber(201))

	cvs, err := canvas.New(image.Rect(0, 0, 6, 4))
	if err != nil {
		t.Fatalf("canvas.New => unexpected error: %v", err)
	}
	if err := fw.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	if got := cellAt(t, cvs, image.Point{X: 0, Y: 0}).Opts; got.FgColor != cell.ColorNumber(201) || !got.Bold {
		t.Errorf("border options = %+v, want overridden foreground with base bold preserved", got)
	}
}

func TestFramedOptionsInflateInnerMinimumSize(t *testing.T) {
	fw, err := FramedNew(&framedTestWidget{
		opts: widgetapi.Options{
			MinimumSize: image.Point{X: 8, Y: 2},
			WantMouse:   widgetapi.MouseScopeWidget,
		},
	})
	if err != nil {
		t.Fatalf("FramedNew => unexpected error: %v", err)
	}

	opts := fw.Options()
	if got, want := opts.MinimumSize, (image.Point{X: 10, Y: 4}); got != want {
		t.Errorf("MinimumSize = %v, want %v", got, want)
	}
	if got, want := opts.WantMouse, widgetapi.MouseScopeWidget; got != want {
		t.Errorf("WantMouse = %v, want %v", got, want)
	}
}

func TestFramedOptionsMinimumIsDrawableFrame(t *testing.T) {
	fw, err := FramedNew(&framedTestWidget{})
	if err != nil {
		t.Fatalf("FramedNew => unexpected error: %v", err)
	}

	if got, want := fw.Options().MinimumSize, (image.Point{X: 3, Y: 3}); got != want {
		t.Errorf("MinimumSize = %v, want %v", got, want)
	}
}

func TestFramedDrawOnTinyCanvasDoesNotCallInner(t *testing.T) {
	called := false
	fw, err := FramedNew(&framedTestWidget{
		draw: func(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
			called = true
			return nil
		},
	})
	if err != nil {
		t.Fatalf("FramedNew => unexpected error: %v", err)
	}

	cvs, err := canvas.New(image.Rect(0, 0, 2, 2))
	if err != nil {
		t.Fatalf("canvas.New => unexpected error: %v", err)
	}
	if err := fw.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	if called {
		t.Fatal("inner widget was drawn even though the frame had no drawable interior")
	}
}

func TestFramedDelegatesKeyboardAndMouse(t *testing.T) {
	var gotKey *terminalapi.Keyboard
	var gotMouse *terminalapi.Mouse
	fw, err := FramedNew(&framedTestWidget{
		keyboard: func(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
			gotKey = k
			return nil
		},
		mouse: func(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
			gotMouse = m
			return nil
		},
	})
	if err != nil {
		t.Fatalf("FramedNew => unexpected error: %v", err)
	}

	key := &terminalapi.Keyboard{Key: keyboard.KeyEnter}
	if err := fw.Keyboard(key, &widgetapi.EventMeta{}); err != nil {
		t.Fatalf("Keyboard => unexpected error: %v", err)
	}
	if gotKey != key {
		t.Fatalf("Keyboard was not delegated")
	}

	mouseEvent := &terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 1, Y: 2}}
	if err := fw.Mouse(mouseEvent, &widgetapi.EventMeta{}); err != nil {
		t.Fatalf("Mouse => unexpected error: %v", err)
	}
	if gotMouse != mouseEvent {
		t.Fatalf("Mouse was not delegated")
	}
}

func cellAt(t *testing.T, cvs *canvas.Canvas, p image.Point) *buffer.Cell {
	t.Helper()
	c, err := cvs.Cell(p)
	if err != nil {
		t.Fatalf("Cell(%v) => unexpected error: %v", p, err)
	}
	return c
}

func rowString(t *testing.T, cvs *canvas.Canvas, y int) string {
	t.Helper()
	var b strings.Builder
	for x := 0; x < cvs.Size().X; x++ {
		r := cellAt(t, cvs, image.Point{X: x, Y: y}).Rune
		if r == 0 {
			r = ' '
		}
		b.WriteRune(r)
	}
	return b.String()
}

type framedTestWidget struct {
	draw     func(*canvas.Canvas, *widgetapi.Meta) error
	keyboard func(*terminalapi.Keyboard, *widgetapi.EventMeta) error
	mouse    func(*terminalapi.Mouse, *widgetapi.EventMeta) error
	opts     widgetapi.Options
}

func (w *framedTestWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	if w.draw != nil {
		return w.draw(cvs, meta)
	}
	return nil
}

func (w *framedTestWidget) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	if w.keyboard != nil {
		return w.keyboard(k, meta)
	}
	return nil
}

func (w *framedTestWidget) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	if w.mouse != nil {
		return w.mouse(m, meta)
	}
	return nil
}

func (w *framedTestWidget) Options() widgetapi.Options {
	return w.opts
}
