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

package radio

import (
	"image"
	"strings"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

func TestNew(t *testing.T) {
	tests := []struct {
		desc    string
		items   []Item
		opts    []Option
		wantErr bool
	}{
		{desc: "fails without items", wantErr: true},
		{desc: "fails on empty label", items: []Item{{Label: "ON"}, {Label: ""}}, wantErr: true},
		{desc: "fails on invalid selected index", items: []Item{{Label: "ON"}}, opts: []Option{Selected(1)}, wantErr: true},
		{desc: "fails on invalid gap", items: []Item{{Label: "ON"}}, opts: []Option{Gap(-1)}, wantErr: true},
		{desc: "accepts defaults", items: []Item{{Label: "ON"}, {Label: "OFF"}}},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			r, err := New(tc.items, tc.opts...)
			if tc.wantErr {
				if err == nil {
					t.Fatal("New => nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}
			if r == nil {
				t.Fatal("New => nil radio")
			}
		})
	}
}

func TestDraw(t *testing.T) {
	r, err := New([]Item{
		{Label: "ON"},
		{Label: "OFF", UnselectedText: "◎"},
	}, Selected(1), Gap(2), IndicatorGap(2))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	ft := drawRadio(t, r, image.Point{X: 16, Y: 1}, true)
	if got := strings.TrimRight(strings.Split(ft.String(), "\n")[0], " "); got != "○  ON  ◉  OFF" {
		t.Fatalf("draw = %q, want %q", got, "○  ON  ◉  OFF")
	}
}

func TestDrawResizeMarker(t *testing.T) {
	r, err := New([]Item{{Label: "ON"}, {Label: "OFF"}})
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 1, Y: 1})
	cvs := testcanvas.MustNew(ft.Area())
	if err := r.Draw(cvs, &widgetapi.Meta{Focused: true}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:1]); got != "⇄" {
		t.Fatalf("resize marker = %q, want %q", got, "⇄")
	}
}

func TestKeyboardAndMouse(t *testing.T) {
	var got []int
	r, err := New([]Item{{Label: "ON"}, {Label: "OFF"}}, OnChange(func(index int, label string) error {
		_ = label
		got = append(got, index)
		return nil
	}))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := r.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowRight}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard right => unexpected error: %v", err)
	}
	if err := r.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard left => unexpected error: %v", err)
	}
	if err := r.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEnd}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard end => unexpected error: %v", err)
	}
	if err := r.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyHome}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard home => unexpected error: %v", err)
	}
	if err := r.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 7, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse left => unexpected error: %v", err)
	}
	if err := r.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: -1, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse out of bounds => unexpected error: %v", err)
	}
	if err := r.Mouse(&terminalapi.Mouse{Button: mouse.ButtonRight, Position: image.Point{X: 0, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse right => unexpected error: %v", err)
	}

	if want := []int{1, 0, 1, 0, 1}; len(got) != len(want) {
		t.Fatalf("callback count = %d, want %d", len(got), len(want))
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("callback[%d] = %d, want %d", i, got[i], want[i])
			}
		}
	}
}

func TestStateAndHelpers(t *testing.T) {
	r, err := New([]Item{
		{Label: "ON", CellOpts: []cell.Option{cell.FgColor(cell.ColorGreen)}},
		{Label: "OFF", SelectedText: "■", UnselectedText: "□"},
	}, IndicatorGap(1))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if got := r.SelectedText(); got != "ON" {
		t.Fatalf("SelectedText = %q, want %q", got, "ON")
	}
	if err := r.SetSelected(1); err != nil {
		t.Fatalf("SetSelected => unexpected error: %v", err)
	}
	if err := r.SetSelected(2); err == nil {
		t.Fatal("SetSelected invalid => nil error, want error")
	}
	if got := r.Options().MinimumSize.Y; got != 1 {
		t.Fatalf("Options.MinimumSize.Y = %d, want 1", got)
	}
	if got := renderItem(Item{Label: "OFF", SelectedText: "■", UnselectedText: "□"}, true, 1); got != "■ OFF" {
		t.Fatalf("renderItem selected = %q, want %q", got, "■ OFF")
	}
	if got := renderItem(Item{Label: "OFF", SelectedText: "■", UnselectedText: "□"}, false, 1); got != "□ OFF" {
		t.Fatalf("renderItem unselected = %q, want %q", got, "□ OFF")
	}
	if got := widthFor([]Item{{Label: "ON", SelectedText: "◉", UnselectedText: "○"}, {Label: "OFF", SelectedText: "◉", UnselectedText: "○"}}, 2, 1); got <= 0 {
		t.Fatalf("widthFor = %d, want > 0", got)
	}
	if got := r.Text(); got != "○ ON   ■ OFF" {
		t.Fatalf("Text = %q, want %q", got, "○ ON   ■ OFF")
	}
	if got := r.ItemText(1); got != "■ OFF" {
		t.Fatalf("ItemText = %q, want %q", got, "■ OFF")
	}
	if got := r.itemAtX(100); got != -1 {
		t.Fatalf("itemAtX out of bounds = %d, want -1", got)
	}
}

func drawRadio(t *testing.T, r *Radio, size image.Point, focused bool) *faketerm.Terminal {
	t.Helper()

	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())
	if err := r.Draw(cvs, &widgetapi.Meta{Focused: focused}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	return ft
}
