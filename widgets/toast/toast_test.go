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

package toast

import (
	"image"
	"testing"
	"time"

	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/fx"
	"github.com/mum4k/termdash/widgets/modal"
)

func TestNewRejectsInvalidOptions(t *testing.T) {
	tests := []struct {
		desc string
		opts []Option
	}{
		{
			desc: "zero width",
			opts: []Option{Width(0)},
		},
		{
			desc: "negative margin",
			opts: []Option{Margin(-1, 0)},
		},
		{
			desc: "custom anchor without position function",
			opts: []Option{Anchor(PlacementCustom)},
		},
		{
			desc: "nil clock",
			opts: []Option{Clock(nil)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if _, err := New(tc.opts...); err == nil {
				t.Fatalf("New => nil error, want error")
			}
		})
	}
}

func TestNotifyDrawsAndDismissesOnClick(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	tm, err := New(
		Clock(func() time.Time { return now }),
		AnimationMode(AnimationNone),
		Width(20),
		Margin(1, 1),
		Shadow(false),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	id := tm.Notify("Saved", "Profile updated", WithSeverity(SeveritySuccess))
	cvs := testcanvas.MustNew(image.Rect(0, 0, 40, 12))
	if err := tm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	if got, want := testcanvas.MustCell(cvs, image.Point{X: 19, Y: 1}).Rune, '╭'; got != want {
		t.Fatalf("top-left border = %q, want %q", got, want)
	}
	if got, want := testcanvas.MustCell(cvs, image.Point{X: 23, Y: 3}).Rune, 'S'; got != want {
		t.Fatalf("title rune = %q, want %q", got, want)
	}

	if err := tm.Mouse(&terminalapi.Mouse{
		Position: image.Point{X: 20, Y: 2},
		Button:   mouse.ButtonLeft,
	}, &widgetapi.EventMeta{}); err != nil {
		t.Fatalf("Mouse => unexpected error: %v", err)
	}
	if got := tm.Count(); got != 0 {
		t.Fatalf("Count after click = %d, want 0", got)
	}
	if tm.Dismiss(id) {
		t.Fatalf("Dismiss succeeded for already removed notification")
	}
}

func TestTTLExpiresOnDraw(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	tm, err := New(
		Clock(func() time.Time { return now }),
		AnimationMode(AnimationNone),
		Shadow(false),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	tm.Notify("Done", "Short lived", WithTTL(time.Second))
	cvs := testcanvas.MustNew(image.Rect(0, 0, 60, 12))
	if err := tm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	if got := tm.Count(); got != 1 {
		t.Fatalf("Count before expiry = %d, want 1", got)
	}

	now = now.Add(time.Second)
	if err := tm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw after expiry => unexpected error: %v", err)
	}
	if got := tm.Count(); got != 0 {
		t.Fatalf("Count after expiry = %d, want 0", got)
	}
}

func TestSlideAnimationMovesFromRightEdge(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	tm, err := New(
		Clock(func() time.Time { return now }),
		AnimationMode(AnimationSlide),
		AnimationDuration(time.Second),
		Width(10),
		MinWidth(10),
		Margin(0, 0),
		Shadow(false),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	id := tm.Notify("Build", "Queued", Sticky())
	cvs := testcanvas.MustNew(image.Rect(0, 0, 30, 10))
	if err := tm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw initial => unexpected error: %v", err)
	}
	if got, want := tm.lastRects[id].Min.X, 30; got != want {
		t.Fatalf("initial rect X = %d, want %d", got, want)
	}

	now = now.Add(time.Second)
	if err := tm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw settled => unexpected error: %v", err)
	}
	if got, want := tm.lastRects[id].Min.X, 20; got != want {
		t.Fatalf("settled rect X = %d, want %d", got, want)
	}
}

func TestCustomPosition(t *testing.T) {
	tm, err := New(
		AnimationMode(AnimationNone),
		Width(12),
		MinWidth(12),
		Shadow(false),
		CustomPosition(func(canvas image.Rectangle, size image.Point, index int) image.Point {
			return image.Point{X: 2 + index, Y: 3 + index}
		}),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	tm.Notify("Custom", "Position")
	cvs := testcanvas.MustNew(image.Rect(0, 0, 40, 12))
	if err := tm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	if got, want := testcanvas.MustCell(cvs, image.Point{X: 2, Y: 3}).Rune, '╭'; got != want {
		t.Fatalf("custom border = %q, want %q", got, want)
	}
}

func TestWorksInsideFXAndModal(t *testing.T) {
	tm, err := New(
		AnimationMode(AnimationNone),
		Shadow(false),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	tm.Notify("Modal", "Toast manager hosted inside a draggable modal")

	fxWidget, err := fx.New(tm, fx.FadeIn(time.Millisecond))
	if err != nil {
		t.Fatalf("fx.New => unexpected error: %v", err)
	}
	mod := modal.NewModal("notifications", []*modal.DraggableWidget{
		modal.NewDraggableWidget("toast-panel", fxWidget, 1, 1, 34, 10, modal.NewOptions()),
	}, modal.NewOptions())

	cvs := testcanvas.MustNew(image.Rect(0, 0, 60, 18))
	if err := mod.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("modal Draw => unexpected error: %v", err)
	}
}
