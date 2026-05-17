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

package emojikeyboard

import (
	"image"
	"strings"
	"testing"
	"time"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// TestDrawReturnsResizeNeededMarker verifies undersized canvases produce the
// standard resize-needed indicator instead of silently rendering nothing.
func TestDrawReturnsResizeNeededMarker(t *testing.T) {
	kbd := New(Emojis("😀", "😎"))

	ft := faketerm.MustNew(image.Point{X: 2, Y: 1})
	cvs := testcanvas.MustNew(ft.Area())
	if err := kbd.Draw(cvs, &widgetapi.Meta{Focused: true}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:1]); got != "⇄" {
		t.Fatalf("resize marker = %q, want %q", got, "⇄")
	}
}

// TestDrawUsesCustomGlyphSets verifies the widget uses caller-supplied framing
// and pagination glyphs when rendering selections and the page row.
func TestDrawUsesCustomGlyphSets(t *testing.T) {
	kbd := New(
		Emojis("😀", "😎"),
		InitialSelection("😀"),
		UseSelectionGlyphSet(SelectionGlyphSets.Bracket),
		UsePaginationGlyphSet(PaginationGlyphSets.Minimal),
	)

	ft := drawEmojiKeyboard(t, kbd, image.Point{X: 10, Y: 3}, true)
	lines := strings.Split(ft.String(), "\n")
	if got := lines[0]; !strings.Contains(got, "[😀") || !strings.Contains(got, "]") {
		t.Fatalf("selected cell row = %q, want row containing %q", got, "[😀]")
	}
	if got := string([]rune(lines[2])[:1]); got != "‹" {
		t.Fatalf("prev pagination glyph = %q, want %q", got, "‹")
	}
	if got := string([]rune(lines[2])[len([]rune(lines[2]))-1:]); got != "›" {
		t.Fatalf("next pagination glyph = %q, want %q", got, "›")
	}
}

// TestKeyboardAndMouseSelection verifies both focused keyboard navigation and
// mouse selection update the widget and fire the callback.
func TestKeyboardAndMouseSelection(t *testing.T) {
	selected := make(chan string, 2)
	kbd := New(
		Emojis("😀", "😎", "🚀"),
		OnSelectFunc(func(emoji string) {
			selected <- emoji
		}),
	)

	if err := kbd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowRight}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard arrow right => unexpected error: %v", err)
	}
	if err := kbd.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyEnter}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Keyboard enter => unexpected error: %v", err)
	}
	if got := waitForSelection(t, selected); got != "😎" {
		t.Fatalf("keyboard selection = %q, want %q", got, "😎")
	}

	_ = drawEmojiKeyboard(t, kbd, image.Point{X: 10, Y: 3}, true)
	if err := kbd.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 6, Y: 0}}, &widgetapi.EventMeta{Focused: true}); err != nil {
		t.Fatalf("Mouse select => unexpected error: %v", err)
	}
	if got := waitForSelection(t, selected); got != "😎" {
		t.Fatalf("mouse selection = %q, want %q", got, "😎")
	}
	if got := kbd.SelectedEmoji(); got != "😎" {
		t.Fatalf("SelectedEmoji = %q, want %q", got, "😎")
	}
}

// TestSetEmojisResetsState verifies the catalog replacement API resets the page
// state and clears selections that no longer exist.
func TestSetEmojisResetsState(t *testing.T) {
	kbd := New(
		Emojis("😀", "😎", "🚀"),
		InitialSelection("🚀"),
	)
	kbd.page = 1
	kbd.cursorIdx = 2

	kbd.SetEmojis([]string{"🍎", "🍐"})
	if got := kbd.SelectedEmoji(); got != "" {
		t.Fatalf("SelectedEmoji after SetEmojis = %q, want empty", got)
	}
	if got := kbd.page; got != 0 {
		t.Fatalf("page after SetEmojis = %d, want 0", got)
	}
	if got := kbd.cursorIdx; got != 0 {
		t.Fatalf("cursorIdx after SetEmojis = %d, want 0", got)
	}
}

// TestOptionsExposeMinimumGeometry verifies the widget reports geometry that
// matches its cell sizing.
func TestOptionsExposeMinimumGeometry(t *testing.T) {
	kbd := New(CellWidth(7), CellHeight(2))
	opts := kbd.Options()
	if got, want := opts.MinimumSize, (image.Point{X: 7, Y: 3}); got != want {
		t.Fatalf("Options.MinimumSize = %v, want %v", got, want)
	}
	if got, want := opts.WantKeyboard, widgetapi.KeyScopeFocused; got != want {
		t.Fatalf("Options.WantKeyboard = %v, want %v", got, want)
	}
	if got, want := opts.WantMouse, widgetapi.MouseScopeWidget; got != want {
		t.Fatalf("Options.WantMouse = %v, want %v", got, want)
	}
}

// drawEmojiKeyboard renders the widget into a fake terminal for snapshot-style assertions.
func drawEmojiKeyboard(t *testing.T, kbd *EmojiKeyboard, size image.Point, focused bool) *faketerm.Terminal {
	t.Helper()

	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())
	if err := kbd.Draw(cvs, &widgetapi.Meta{Focused: focused}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	return ft
}

// waitForSelection waits briefly for the asynchronous selection callback.
func waitForSelection(t *testing.T, ch <-chan string) string {
	t.Helper()

	select {
	case got := <-ch:
		return got
	case <-time.After(500 * time.Millisecond):
		t.Fatal("selection callback did not fire")
		return ""
	}
}
