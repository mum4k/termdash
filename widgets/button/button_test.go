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

package button

import (
	"errors"
	"image"
	"sync"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// callbackTracker tracks whether callback was called.
type callbackTracker struct {
	// wantErr when set to true, makes callback return an error.
	wantErr bool

	// called asserts whether the callback was called.
	called bool

	// count is the number of times the callback was called.
	count int

	// mu protects the tracker.
	mu sync.Mutex
}

// callback is the callback function.
func (ct *callbackTracker) callback() error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if ct.wantErr {
		return errors.New("ct.wantErr set to true")
	}

	ct.count++
	ct.called = true
	return nil
}

func TestButton(t *testing.T) {
	tests := []struct {
		desc         string
		text         string
		opts         []Option
		events       []terminalapi.Event
		canvas       image.Rectangle
		want         func(size image.Point) *faketerm.Terminal
		wantCallback *callbackTracker
		wantNewErr   bool
		wantDrawErr  bool
	}{}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotCallback := &callbackTracker{}
			b, err := New(tc.text, gotCallback.callback, tc.opts...)
			if (err != nil) != tc.wantNewErr {
				t.Errorf("New => unexpected error: %v, wantNewErr: %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			for _, ev := range tc.events {
				switch ev.(type) {
				default:
					t.Fatalf("unsupported event type: %T", ev)
				}
			}

			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			err = b.Draw(c)
			if (err != nil) != tc.wantDrawErr {
				t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
			}
			if err != nil {
				return
			}

			got, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			var want *faketerm.Terminal
			if tc.want != nil {
				want = tc.want(c.Size())
			} else {
				want = faketerm.MustNew(c.Size())
			}

			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if diff := pretty.Compare(tc.wantCallback, gotCallback); diff != "" {
				t.Errorf("CallbackFn => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestKeyboard(t *testing.T) {
	ct := &callbackTracker{}
	b, err := New("text", ct.callback)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := b.Keyboard(&terminalapi.Keyboard{}); err == nil {
		t.Errorf("Keyboard => got nil err, wanted one")
	}
}

func TestMouse(t *testing.T) {
	ct := &callbackTracker{}
	b, err := New("text", ct.callback)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := b.Mouse(&terminalapi.Mouse{}); err == nil {
		t.Errorf("Mouse => got nil err, wanted one")
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc string
		text string
		opts []Option
		want widgetapi.Options
	}{
		{
			desc: "width is based on the text width by default",
			text: "hello world",
			want: widgetapi.Options{
				MinimumSize:  image.Point{14, 4},
				MaximumSize:  image.Point{14, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    true,
			},
		},
		{
			desc: "width supports full-width unicode characters",
			text: "■㈱の世界①",
			want: widgetapi.Options{
				MinimumSize:  image.Point{13, 4},
				MaximumSize:  image.Point{13, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    true,
			},
		},
		{
			desc: "width specified via WidthFor",
			text: "hello",
			opts: []Option{
				WidthFor("■㈱の世界①"),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{13, 4},
				MaximumSize:  image.Point{13, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    true,
			},
		},
		{
			desc: "custom height specified",
			text: "hello",
			opts: []Option{
				Height(10),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 11},
				MaximumSize:  image.Point{8, 11},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    true,
			},
		},
		{
			desc: "custom width specified",
			text: "hello",
			opts: []Option{
				Width(10),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{11, 4},
				MaximumSize:  image.Point{11, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    true,
			},
		},

		{
			desc: "doesn't want keyboard by default",
			text: "hello",
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 4},
				MaximumSize:  image.Point{8, 4},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    true,
			},
		},
		{
			desc: "registers for focused keyboard events",
			text: "hello",
			opts: []Option{
				Key(keyboard.KeyEnter),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 4},
				MaximumSize:  image.Point{8, 4},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    true,
			},
		},
		{
			desc: "registers for global keyboard events",
			text: "hello",
			opts: []Option{
				GlobalKey(keyboard.KeyEnter),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 4},
				MaximumSize:  image.Point{8, 4},
				WantKeyboard: widgetapi.KeyScopeGlobal,
				WantMouse:    true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ct := &callbackTracker{}
			b, err := New(tc.text, ct.callback, tc.opts...)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			got := b.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}

}
