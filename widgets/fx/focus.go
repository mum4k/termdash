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
	"sync"
	"time"

	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// FocusEffectWidget wraps any widgetapi.Widget and plays different effect
// sequences when the container gains or loses keyboard focus.
//
// Create with FocusNew, then optionally set OnFocusChange to be notified of
// each focus transition.
//
// Example:
//
//	fw, _ := fx.FocusNew(myWidget,
//	    []fx.Effect{fx.FadeIn(300*time.Millisecond)},   // gained
//	    []fx.Effect{fx.FadeOut(300*time.Millisecond)},  // lost
//	)
//	fw.OnFocusChange = func(gained bool) {
//	    // update the container border or widget border here
//	}
//
// Implements widgetapi.Widget.  Thread-safe.
type FocusEffectWidget struct {
	mu       sync.Mutex
	inner    widgetapi.Widget
	focusIn  []Effect // played when focus is gained
	focusOut []Effect // played when focus is lost

	knownFocused *bool // nil until the first Draw call
	startTime    time.Time
	currentFx    []Effect // currently active sequence; nil = pass-through

	// OnFocusChange, if non-nil, is called in a new goroutine on each focus
	// transition.  gained=true means the container just received focus.
	OnFocusChange func(gained bool)

	tmp      *canvas.Canvas
	lastSize image.Point
}

// FocusNew wraps inner, playing focusIn when the container gains focus and
// focusOut when it loses focus.  Either slice may be nil or empty to skip the
// corresponding animation.
func FocusNew(inner widgetapi.Widget, focusIn, focusOut []Effect) (*FocusEffectWidget, error) {
	return &FocusEffectWidget{
		inner:    inner,
		focusIn:  focusIn,
		focusOut: focusOut,
	}, nil
}

// Draw implements widgetapi.Widget.
func (w *FocusEffectWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	size := cvs.Size()

	// Recreate the scratch canvas on resize and reset focus tracking so the
	// next Draw reliably detects the current state.
	if w.tmp == nil || w.lastSize != size {
		var err error
		w.tmp, err = canvas.New(cvs.Area())
		if err != nil {
			return err
		}
		w.lastSize = size
		w.knownFocused = nil
	}

	// Detect focus transitions.
	nowFocused := meta.Focused
	if w.knownFocused == nil || *w.knownFocused != nowFocused {
		if w.knownFocused != nil {
			// A real transition occurred — start the matching effect sequence.
			if nowFocused {
				w.currentFx = w.focusIn
			} else {
				w.currentFx = w.focusOut
			}
			w.startTime = time.Now()
			if w.OnFocusChange != nil {
				cb := w.OnFocusChange
				go cb(nowFocused)
			}
		}
		f := nowFocused
		w.knownFocused = &f
	}

	// Paint the inner widget onto the scratch canvas.
	if err := w.tmp.Clear(); err != nil {
		return err
	}
	if err := w.inner.Draw(w.tmp, meta); err != nil {
		return err
	}

	// No active effect — pass through unchanged.
	if len(w.currentFx) == 0 {
		return copyCanvas(w.tmp, cvs)
	}

	elapsed := time.Since(w.startTime)
	active, effectElapsed := findEffect(w.currentFx, elapsed)
	if active == nil {
		// Sequence finished.
		w.currentFx = nil
		return copyCanvas(w.tmp, cvs)
	}
	return applyEffect(active, effectElapsed, w.tmp, cvs)
}

// Keyboard implements widgetapi.Widget.
func (w *FocusEffectWidget) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return w.inner.Keyboard(k, meta)
}

// Mouse implements widgetapi.Widget.
func (w *FocusEffectWidget) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return w.inner.Mouse(m, meta)
}

// Options implements widgetapi.Widget.
func (w *FocusEffectWidget) Options() widgetapi.Options {
	return w.inner.Options()
}
