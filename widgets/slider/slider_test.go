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

package slider

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

func TestNew(t *testing.T) {
	tests := []struct {
		desc    string
		opts    []Option
		wantErr bool
	}{
		{desc: "fails on inverted range", opts: []Option{Min(10), Max(1)}, wantErr: true},
		{desc: "fails on zero step", opts: []Option{Step(0)}, wantErr: true},
		{desc: "fails on zero width", opts: []Option{Width(0)}, wantErr: true},
		{desc: "accepts defaults"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := New(tc.opts...)
			if tc.wantErr {
				if err == nil {
					t.Fatal("New => nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}
			if s == nil {
				t.Fatal("New => nil slider")
			}
		})
	}
}

func TestDraw(t *testing.T) {
	s, err := New(Min(1), Max(100), Value(50), Width(5))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	ft := drawSlider(t, s, image.Point{X: 5, Y: 1}, false)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:5]); got != "█●░░░" {
		t.Fatalf("draw = %q, want %q", got, "█●░░░")
	}
}

func TestDrawResizeMarker(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 1, Y: 1})
	cvs := testcanvas.MustNew(ft.Area())
	if err := s.Draw(cvs, &widgetapi.Meta{Focused: true}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:1]); got != "⇄" {
		t.Fatalf("resize marker = %q, want %q", got, "⇄")
	}
}

func TestKeyboardAndMouse(t *testing.T) {
	var got []int
	s, err := New(
		Min(1),
		Max(100),
		Value(50),
		Width(10),
		Step(5),
		OnChange(func(value int) error {
			got = append(got, value)
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := s.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowRight}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard right => unexpected error: %v", err)
	}
	if err := s.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard left => unexpected error: %v", err)
	}
	if err := s.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEnd}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard end => unexpected error: %v", err)
	}
	if err := s.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyHome}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard home => unexpected error: %v", err)
	}
	if err := s.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 9, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse left => unexpected error: %v", err)
	}
	if err := s.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: -1, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse out of bounds => unexpected error: %v", err)
	}
	if err := s.Mouse(&terminalapi.Mouse{Button: mouse.ButtonRight, Position: image.Point{X: 0, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse right => unexpected error: %v", err)
	}

	if want := []int{55, 50, 100, 1, 100}; len(got) != len(want) {
		t.Fatalf("callback count = %d, want %d", len(got), len(want))
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("callback[%d] = %d, want %d", i, got[i], want[i])
			}
		}
	}
}

func TestValueHelpersAndOptions(t *testing.T) {
	s, err := New(
		Min(1),
		Max(100),
		Value(150),
		Width(18),
		FillRune('='),
		TrackRune('-'),
		KnobRune('o'),
		FillCellOpts(),
		TrackCellOpts(),
		KnobCellOpts(),
		FocusedKnobCellOpts(),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if got := s.Value(); got != 100 {
		t.Fatalf("Value = %d, want 100", got)
	}
	s.SetValue(-5)
	if got := s.Value(); got != 1 {
		t.Fatalf("Value after SetValue clamp = %d, want 1", got)
	}
	if got, want := s.Options().MinimumSize, (image.Point{X: 18, Y: 1}); got != want {
		t.Fatalf("Options.MinimumSize = %v, want %v", got, want)
	}
	if got := s.valueAtX(0); got != 1 {
		t.Fatalf("valueAtX(0) = %d, want 1", got)
	}
	if got := s.valueAtX(17); got != 100 {
		t.Fatalf("valueAtX(max) = %d, want 100", got)
	}
}

func drawSlider(t *testing.T, s *Slider, size image.Point, focused bool) *faketerm.Terminal {
	t.Helper()

	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())
	if err := s.Draw(cvs, &widgetapi.Meta{Focused: focused}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	return ft
}
