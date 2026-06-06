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

package dropdown

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
		items   []string
		opts    []Option
		wantErr bool
	}{
		{
			desc:    "fails without items",
			wantErr: true,
		},
		{
			desc:    "fails on empty item",
			items:   []string{"01", ""},
			wantErr: true,
		},
		{
			desc:    "fails on invalid selected index",
			items:   []string{"01", "02"},
			opts:    []Option{Selected(2)},
			wantErr: true,
		},
		{
			desc:    "fails on invalid width",
			items:   []string{"01"},
			opts:    []Option{Width(3)},
			wantErr: true,
		},
		{
			desc:  "accepts defaults",
			items: []string{"01", "02"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			dd, err := New(tc.items, tc.opts...)
			if tc.wantErr {
				if err == nil {
					t.Fatal("New => nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}
			if dd == nil {
				t.Fatal("New => nil widget")
			}
		})
	}
}

func TestIntRange(t *testing.T) {
	if got, want := IntRange(1, 5, 2, "%02d"), []string{"01", "03", "05"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("IntRange => %v, want %v", got, want)
	}
	if got := IntRange(5, 1, 1, ""); got != nil {
		t.Fatalf("IntRange invalid range => %v, want nil", got)
	}
}

func TestDrawClosed(t *testing.T) {
	dd, err := New([]string{"01", "02", "03"}, Selected(1))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	ft := drawDropdown(t, dd, image.Point{X: 6, Y: 4}, true)
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[0])[:6]); got != "[02 ▼]" {
		t.Fatalf("closed trigger = %q, want %q", got, "[02 ▼]")
	}
}

func TestDrawOpen(t *testing.T) {
	dd, err := New([]string{"01", "02", "03"}, Selected(1))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	dd.Open()

	ft := drawDropdown(t, dd, image.Point{X: 6, Y: 6}, true)
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[0])[:6]); got != "[02 ▲]" {
		t.Fatalf("open trigger = %q, want %q", got, "[02 ▲]")
	}
	if got := string([]rune(lines[1])[:6]); got != "┌────┐" {
		t.Fatalf("top border = %q, want %q", got, "┌────┐")
	}
	if got := string([]rune(lines[3])[:6]); got != "│>02 │" {
		t.Fatalf("selected row = %q, want %q", got, "│>02 │")
	}
	if got := string([]rune(lines[5])[:6]); got != "└────┘" {
		t.Fatalf("bottom border = %q, want %q", got, "└────┘")
	}
}

func TestKeyboardSelection(t *testing.T) {
	var (
		gotIndex int
		gotLabel string
	)
	dd, err := New([]string{"01", "02", "03"}, OnSelect(func(index int, label string) error {
		gotIndex = index
		gotLabel = label
		return nil
	}))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEnter}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard open => unexpected error: %v", err)
	}
	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard down => unexpected error: %v", err)
	}
	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEnter}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard commit => unexpected error: %v", err)
	}

	if got := dd.SelectedText(); got != "02" {
		t.Fatalf("SelectedText = %q, want %q", got, "02")
	}
	if gotIndex != 1 || gotLabel != "02" {
		t.Fatalf("callback got index=%d label=%q, want index=1 label=%q", gotIndex, gotLabel, "02")
	}
}

func TestMouseSelection(t *testing.T) {
	var (
		gotIndex int
		gotLabel string
	)
	dd, err := New([]string{"01", "02", "03"}, OnSelect(func(index int, label string) error {
		gotIndex = index
		gotLabel = label
		return nil
	}))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	_ = drawDropdown(t, dd, image.Point{X: 6, Y: 6}, true)
	if err := dd.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 1, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse open => unexpected error: %v", err)
	}
	_ = drawDropdown(t, dd, image.Point{X: 6, Y: 6}, true)
	if err := dd.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 2, Y: 4}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse select => unexpected error: %v", err)
	}

	if got := dd.SelectedText(); got != "03" {
		t.Fatalf("SelectedText = %q, want %q", got, "03")
	}
	if gotIndex != 2 || gotLabel != "03" {
		t.Fatalf("callback got index=%d label=%q, want index=2 label=%q", gotIndex, gotLabel, "03")
	}
}

func TestDrawOpenClampsVisibleRows(t *testing.T) {
	dd, err := New([]string{"01", "02", "03", "04", "05"}, Selected(4))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	dd.Open()

	ft := drawDropdown(t, dd, image.Point{X: 6, Y: 6}, true)
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[2])[:6]); got != "│ 03 │" {
		t.Fatalf("first visible row = %q, want %q", got, "│ 03 │")
	}
	if got := string([]rune(lines[4])[:6]); got != "│>05 │" {
		t.Fatalf("last visible row = %q, want %q", got, "│>05 │")
	}
}

func TestStateSettersAndOptions(t *testing.T) {
	dd, err := New([]string{"alpha", "beta"},
		Selected(1),
		Width(10),
		GlyphSet(GlyphProfiles.Minimal),
		CellOpts(),
		FocusedCellOpts(),
		SelectedCellOpts(),
		BorderCellOpts(),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if got := dd.SelectedIndex(); got != 1 {
		t.Fatalf("SelectedIndex = %d, want 1", got)
	}
	if got := dd.SelectedText(); got != "beta" {
		t.Fatalf("SelectedText = %q, want %q", got, "beta")
	}

	opts := dd.Options()
	if got, want := opts.MinimumSize, (image.Point{X: 10, Y: 1}); got != want {
		t.Fatalf("Options.MinimumSize = %v, want %v", got, want)
	}
	if got, want := opts.WantKeyboard, widgetapi.KeyScopeFocused; got != want {
		t.Fatalf("Options.WantKeyboard = %v, want %v", got, want)
	}
	if got, want := opts.WantMouse, widgetapi.MouseScopeContainer; got != want {
		t.Fatalf("Options.WantMouse = %v, want %v", got, want)
	}

	if err := dd.SetSelected(0); err != nil {
		t.Fatalf("SetSelected => unexpected error: %v", err)
	}
	if err := dd.SetSelected(3); err == nil {
		t.Fatal("SetSelected => nil error, want error")
	}

	if err := dd.SetItems([]string{"one", "two", "three"}); err != nil {
		t.Fatalf("SetItems => unexpected error: %v", err)
	}
	if got := dd.SelectedText(); got != "one" {
		t.Fatalf("SelectedText after SetItems = %q, want %q", got, "one")
	}
	if err := dd.SetItems(nil); err == nil {
		t.Fatal("SetItems(nil) => nil error, want error")
	}

	dd.Open()
	dd.Close()
	if got := dd.TriggerTextFor("beta"); got != "[beta   ▾]" {
		t.Fatalf("TriggerTextFor = %q, want %q", got, "[beta   ▾]")
	}
}

func TestKeyboardNavigationBranches(t *testing.T) {
	dd, err := New([]string{"01", "02", "03"})
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := dd.Keyboard(&terminalapi.Keyboard{Key: ' '}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard space => unexpected error: %v", err)
	}
	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEnd}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard end => unexpected error: %v", err)
	}
	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyHome}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard home => unexpected error: %v", err)
	}
	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEsc}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard esc => unexpected error: %v", err)
	}
	if got := dd.SelectedText(); got != "01" {
		t.Fatalf("SelectedText after esc = %q, want %q", got, "01")
	}
}

func TestMouseOutsideOptionClosesOpenList(t *testing.T) {
	dd, err := New([]string{"01", "02", "03"})
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	_ = drawDropdown(t, dd, image.Point{X: 6, Y: 6}, true)
	if err := dd.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 1, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse open => unexpected error: %v", err)
	}
	_ = drawDropdown(t, dd, image.Point{X: 6, Y: 6}, true)
	if err := dd.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 0, Y: 5}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse outside => unexpected error: %v", err)
	}

	ft := drawDropdown(t, dd, image.Point{X: 6, Y: 6}, true)
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[0])[:6]); got != "[01 ▼]" {
		t.Fatalf("trigger after outside click = %q, want %q", got, "[01 ▼]")
	}
}

func TestDrawOpenWithoutRoomForList(t *testing.T) {
	dd, err := New([]string{"01", "02", "03"})
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	dd.Open()

	ft := drawDropdown(t, dd, image.Point{X: 6, Y: 3}, true)
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[0])[:6]); got != "[01 ▲]" {
		t.Fatalf("trigger in short canvas = %q, want %q", got, "[01 ▲]")
	}
	if got := string([]rune(lines[1])[:6]); got != "      " {
		t.Fatalf("unexpected list content in short canvas = %q, want blank row", got)
	}
}

func TestFormattingHelpers(t *testing.T) {
	if got := formatTrigger("HELLO", GlyphProfiles.Classic.OpenArrow, 4); got != "[ ▲]" {
		t.Fatalf("formatTrigger narrow = %q, want %q", got, "[ ▲]")
	}
	if got := formatOption("ABCDE", true, 4, ">", " "); got != ">ABC" {
		t.Fatalf("formatOption = %q, want %q", got, ">ABC")
	}
	if got := fitText("ABCDEFG", 4); got != "ABCD" {
		t.Fatalf("fitText = %q, want %q", got, "ABCD")
	}
}

func TestSetItemsPreservesSelectionAndExpandsWidth(t *testing.T) {
	dd, err := New([]string{"one", "two"}, Selected(1))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := dd.SetItems([]string{"one", "a much longer label", "three"}); err != nil {
		t.Fatalf("SetItems => unexpected error: %v", err)
	}
	if got := dd.SelectedText(); got != "a much longer label" {
		t.Fatalf("SelectedText after preserving selection = %q, want %q", got, "a much longer label")
	}
	if got, want := dd.Options().MinimumSize.X >= len("a much longer label")+4, true; got != want {
		t.Fatalf("Options.MinimumSize.X did not expand enough")
	}
}

func TestDrawReturnsResizeNeededMarker(t *testing.T) {
	dd, err := New([]string{"01", "02"})
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 5, Y: 1})
	cvs := testcanvas.MustNew(ft.Area())
	if err := dd.Draw(cvs, &widgetapi.Meta{Focused: true}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:1]); got != "⇄" {
		t.Fatalf("resize marker = %q, want %q", got, "⇄")
	}
}

func TestInternalGeometryHelpers(t *testing.T) {
	dd, err := New([]string{"01", "02", "03", "04", "05"})
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	dd.Open()
	dd.lastSize = image.Point{X: 6, Y: 6}
	dd.cursor = 4

	if got := dd.visibleRows(dd.lastSize.Y); got != 3 {
		t.Fatalf("visibleRows = %d, want 3", got)
	}
	if got := dd.viewStart(3); got != 2 {
		t.Fatalf("viewStart = %d, want 2", got)
	}
	if got := dd.optionIndexAt(image.Point{X: 2, Y: 2}); got != 2 {
		t.Fatalf("optionIndexAt first visible row = %d, want 2", got)
	}
	if got := dd.optionIndexAt(image.Point{X: 0, Y: 5}); got != -1 {
		t.Fatalf("optionIndexAt border = %d, want -1", got)
	}
	if got := dd.CanvasSize(6); got != (image.Point{X: 6, Y: 6}) {
		t.Fatalf("CanvasSize open = %v, want %v", got, image.Point{X: 6, Y: 6})
	}
	dd.Close()
	if got := dd.CanvasSize(6); got != (image.Point{X: 6, Y: 1}) {
		t.Fatalf("CanvasSize closed = %v, want %v", got, image.Point{X: 6, Y: 1})
	}
}

func TestKeyboardArrowBoundsAndMouseIgnore(t *testing.T) {
	dd, err := New([]string{"01", "02"})
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowUp}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard arrow up => unexpected error: %v", err)
	}
	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard arrow down => unexpected error: %v", err)
	}
	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard arrow down again => unexpected error: %v", err)
	}
	if err := dd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEnter}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard commit => unexpected error: %v", err)
	}
	if got := dd.SelectedText(); got != "02" {
		t.Fatalf("SelectedText after arrow selection = %q, want %q", got, "02")
	}

	if err := dd.Mouse(&terminalapi.Mouse{Button: mouse.ButtonRight, Position: image.Point{X: 1, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse right click => unexpected error: %v", err)
	}
}

func drawDropdown(t *testing.T, dd *Dropdown, size image.Point, focused bool) *faketerm.Terminal {
	t.Helper()

	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())
	if err := dd.Draw(cvs, &widgetapi.Meta{Focused: focused}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	return ft
}
