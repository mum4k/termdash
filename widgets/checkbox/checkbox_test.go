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

package checkbox

import (
	"image"
	"strings"
	"testing"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

func TestNewAndState(t *testing.T) {
	cb, err := New("Cloak", Checked(true))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if !cb.Checked() {
		t.Fatal("Checked = false, want true")
	}
	cb.SetChecked(false)
	if cb.Checked() {
		t.Fatal("Checked = true after SetChecked(false), want false")
	}
	if checked := cb.Toggle(); !checked {
		t.Fatal("Toggle => false, want true")
	}
}

func TestDraw(t *testing.T) {
	cb, err := New("Cloak")
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	ft := drawCheckbox(t, cb, image.Point{X: 12, Y: 2}, false)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:9]); got != "[ ] Cloak" {
		t.Fatalf("unchecked draw = %q, want %q", got, "[ ] Cloak")
	}

	cb.SetChecked(true)
	ft = drawCheckbox(t, cb, image.Point{X: 12, Y: 2}, true)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:9]); got != "[x] Cloak" {
		t.Fatalf("checked draw = %q, want %q", got, "[x] Cloak")
	}
}

func TestDrawResizeMarker(t *testing.T) {
	cb, err := New("Cloak")
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 2, Y: 1})
	cvs := testcanvas.MustNew(ft.Area())
	if err := cb.Draw(cvs, &widgetapi.Meta{Focused: true}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:1]); got != "⇄" {
		t.Fatalf("resize marker = %q, want %q", got, "⇄")
	}
}

func TestKeyboardAndMouse(t *testing.T) {
	var got []bool
	cb, err := New("Cloak", OnChange(func(checked bool) error {
		got = append(got, checked)
		return nil
	}))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := cb.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEnter}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard enter => unexpected error: %v", err)
	}
	if err := cb.Keyboard(&terminalapi.Keyboard{Key: ' '}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard space => unexpected error: %v", err)
	}
	if err := cb.Keyboard(&terminalapi.Keyboard{Key: 'x'}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard noop => unexpected error: %v", err)
	}
	if err := cb.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse left => unexpected error: %v", err)
	}
	if err := cb.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: -1, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse out of bounds => unexpected error: %v", err)
	}
	if err := cb.Mouse(&terminalapi.Mouse{Button: mouse.ButtonRight, Position: image.Point{}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse right => unexpected error: %v", err)
	}

	if gotWant := []bool{true, false, true}; len(got) != len(gotWant) {
		t.Fatalf("callback count = %d, want %d", len(got), len(gotWant))
	} else {
		for i := range gotWant {
			if got[i] != gotWant[i] {
				t.Fatalf("callback[%d] = %v, want %v", i, got[i], gotWant[i])
			}
		}
	}
}

func TestOptionsAndHelpers(t *testing.T) {
	cb, err := New("Cloak",
		UseIndicatorSet(IndicatorSets.Heavy),
		LabelGap(2),
		CellOpts(),
		FocusedCellOpts(),
		CheckedCellOpts(),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if got, want := cb.Options().MinimumSize, (image.Point{X: widthFor(cb.opts, "Cloak"), Y: 1}); got != want {
		t.Fatalf("Options.MinimumSize = %v, want %v", got, want)
	}
	if got := textFor(cb.opts, "", false); got != "☐" {
		t.Fatalf("textFor empty unchecked = %q, want %q", got, "☐")
	}
	if got := textFor(cb.opts, "", true); got != "☑" {
		t.Fatalf("textFor empty checked = %q, want %q", got, "☑")
	}
	if got := cb.TextFor(true); got != "☑  Cloak" {
		t.Fatalf("TextFor checked = %q, want %q", got, "☑  Cloak")
	}
	if got := cb.Text(); got != "☐  Cloak" {
		t.Fatalf("Text = %q, want %q", got, "☐  Cloak")
	}
}

func drawCheckbox(t *testing.T, cb *Checkbox, size image.Point, focused bool) *faketerm.Terminal {
	t.Helper()

	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())
	if err := cb.Draw(cvs, &widgetapi.Meta{Focused: focused}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	return ft
}
