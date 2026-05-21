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

func TestActionClickRunsCallback(t *testing.T) {
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

	called := false
	tm.Notify("Actions", "Click footer", Sticky(), WithAction("dock", func() error {
		called = true
		return nil
	}))
	cvs := testcanvas.MustNew(image.Rect(0, 0, 40, 12))
	if err := tm.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	if err := tm.Mouse(&terminalapi.Mouse{
		Position: image.Point{X: 22, Y: 5},
		Button:   mouse.ButtonLeft,
	}, &widgetapi.EventMeta{}); err != nil {
		t.Fatalf("Mouse => unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("action callback was not called")
	}
	if got := tm.Count(); got != 1 {
		t.Fatalf("Count after non-dismissing action = %d, want 1", got)
	}
}

func TestSurfaceActionClickRunsCallback(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	surface, err := NewSurface(
		DefaultToastOptions(
			Clock(func() time.Time { return now }),
			AnimationMode(AnimationNone),
			Width(20),
			Margin(1, 1),
			Shadow(false),
		),
	)
	if err != nil {
		t.Fatalf("NewSurface => unexpected error: %v", err)
	}

	called := false
	surface.Notify("Actions", "Click footer", Sticky(), WithAction("open", func() error {
		called = true
		return nil
	}))
	cvs := testcanvas.MustNew(image.Rect(0, 0, 40, 12))
	if err := surface.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	if err := surface.Mouse(&terminalapi.Mouse{
		Position: image.Point{X: 22, Y: 5},
		Button:   mouse.ButtonLeft,
	}, &widgetapi.EventMeta{}); err != nil {
		t.Fatalf("Mouse => unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("surface action callback was not called")
	}
}

func TestSurfaceActionClickWorksInsideFramedWidget(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	surface, err := NewSurface(
		DefaultToastOptions(
			Clock(func() time.Time { return now }),
			AnimationMode(AnimationNone),
			Width(20),
			Margin(1, 1),
			Shadow(false),
		),
	)
	if err != nil {
		t.Fatalf("NewSurface => unexpected error: %v", err)
	}
	framed, err := fx.FramedNew(surface)
	if err != nil {
		t.Fatalf("FramedNew => unexpected error: %v", err)
	}

	called := false
	surface.Notify("Actions", "Click footer", Sticky(), WithAction("open", func() error {
		called = true
		return nil
	}))
	cvs := testcanvas.MustNew(image.Rect(0, 0, 42, 14))
	if err := framed.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	if err := framed.Mouse(&terminalapi.Mouse{
		Position: image.Point{X: 23, Y: 6},
		Button:   mouse.ButtonLeft,
	}, &widgetapi.EventMeta{}); err != nil {
		t.Fatalf("Mouse => unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("framed surface action callback was not called")
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

func TestActionClickWorksInsideModal(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	tm, err := New(
		Clock(func() time.Time { return now }),
		AnimationMode(AnimationNone),
		Width(20),
		MinWidth(20),
		Margin(1, 1),
		Shadow(false),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	called := false
	tm.Notify("Modal action", "Click footer", Sticky(), WithAction("dock", func() error {
		called = true
		return nil
	}))
	item := modal.NewDraggableWidget("toast-panel", tm, 1, 1, 28, 10, modal.NewOptions())
	mod := modal.NewModal("notifications", []*modal.DraggableWidget{item}, modal.NewOptions())
	cvs := testcanvas.MustNew(image.Rect(0, 0, 40, 16))
	if err := mod.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("modal Draw => unexpected error: %v", err)
	}

	// The child widget canvas starts at modal coordinates (item.X+1,item.Y+2).
	// The action sits at child-relative y=5 and x=7 for this geometry.
	if err := mod.Mouse(&terminalapi.Mouse{
		Position: image.Point{X: item.X + 1 + 7, Y: item.Y + 2 + 5},
		Button:   mouse.ButtonLeft,
	}, &widgetapi.EventMeta{}); err != nil {
		t.Fatalf("modal Mouse => unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("modal-hosted toast action callback was not called")
	}
}

func TestSurfaceNotifyAtAndClear(t *testing.T) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	surface, err := NewSurface(
		DefaultToastOptions(
			Clock(func() time.Time { return now }),
			AnimationMode(AnimationNone),
			Width(12),
			MinWidth(12),
			Shadow(false),
		),
		SurfacePlacement(PlacementBottomLeft),
	)
	if err != nil {
		t.Fatalf("NewSurface => unexpected error: %v", err)
	}

	surface.Notify("Default", "top right")
	surface.NotifyAt(PlacementBottomLeft, "Bottom", "left")
	if got := surface.Count(); got != 2 {
		t.Fatalf("Count = %d, want 2", got)
	}

	cvs := testcanvas.MustNew(image.Rect(0, 0, 40, 12))
	if err := surface.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	if got, want := testcanvas.MustCell(cvs, image.Point{X: 26, Y: 1}).Rune, '╭'; got != want {
		t.Fatalf("default placement border = %q, want %q", got, want)
	}
	if got, want := testcanvas.MustCell(cvs, image.Point{X: 2, Y: 5}).Rune, '╭'; got != want {
		t.Fatalf("bottom-left placement border = %q, want %q", got, want)
	}

	surface.ClearAt(PlacementBottomLeft)
	if got := surface.Count(); got != 1 {
		t.Fatalf("Count after ClearAt = %d, want 1", got)
	}
	surface.Clear()
	if got := surface.Count(); got != 0 {
		t.Fatalf("Count after Clear = %d, want 0", got)
	}
}

func TestSurfaceCustomPlacementMustBeRegistered(t *testing.T) {
	surface, err := NewSurface()
	if err != nil {
		t.Fatalf("NewSurface => unexpected error: %v", err)
	}
	if got := surface.NotifyAt(PlacementCustom, "Custom", "missing position"); got != "" {
		t.Fatalf("NotifyAt unregistered custom = %q, want empty ID", got)
	}

	err = surface.Register(PlacementCustom, CustomPosition(func(canvas image.Rectangle, size image.Point, index int) image.Point {
		return image.Point{X: 2, Y: 3}
	}))
	if err != nil {
		t.Fatalf("Register custom => unexpected error: %v", err)
	}
	id := surface.NotifyAt(PlacementCustom, "Custom", "registered")
	if id == "" {
		t.Fatalf("NotifyAt registered custom returned empty ID")
	}
}
