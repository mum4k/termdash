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

package textinput

import (
	"errors"
	"image"
	"sync"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// callbackTracker tracks whether callback was called.
type callbackTracker struct {
	// wantErr when set to true, makes callback return an error.
	wantErr bool

	// text is the text received OnSubmit.
	text string

	// count is the number of times the callback was called.
	count int

	// mu protects the tracker.
	mu sync.Mutex
}

// submit is the callback function called OnSubmit.
func (ct *callbackTracker) submit(text string) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if ct.wantErr {
		return errors.New("ct.wantErr set to true")
	}

	ct.count++
	ct.text = text
	return nil
}

func TestTextInput(t *testing.T) {
	tests := []struct {
		desc         string
		text         string
		callback     *callbackTracker
		opts         []Option
		events       []terminalapi.Event
		canvas       image.Rectangle
		meta         *widgetapi.Meta
		want         func(size image.Point) *faketerm.Terminal
		wantCallback *callbackTracker
		wantNewErr   bool
		wantDrawErr  bool
		wantEventErr bool
	}{}

	textFieldRune = 'x'
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotCallback := tc.callback
			if gotCallback != nil {
				tc.opts = append(tc.opts, OnSubmit(gotCallback.submit))
			}

			ti, err := New(tc.opts...)
			if (err != nil) != tc.wantNewErr {
				t.Errorf("New => unexpected error: %v, wantNewErr: %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			for _, ev := range tc.events {
				switch e := ev.(type) {
				case *terminalapi.Mouse:
					err := ti.Mouse(e)
					if (err != nil) != tc.wantEventErr {
						t.Errorf("Mouse => unexpected error: %v, wantEventErr: %v", err, tc.wantEventErr)
					}
					if err != nil {
						return
					}

				case *terminalapi.Keyboard:
					err := ti.Keyboard(e)
					if (err != nil) != tc.wantEventErr {
						t.Errorf("Keyboard => unexpected error: %v, wantEventErr: %v", err, tc.wantEventErr)
					}
					if err != nil {
						return
					}

				default:
					t.Fatalf("unsupported event type: %T", ev)
				}
			}

			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			{
				err = ti.Draw(c, tc.meta)
				if (err != nil) != tc.wantDrawErr {
					t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
				}
				if err != nil {
					return
				}
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

func TestOptions(t *testing.T) {
	tests := []struct {
		desc string
		text string
		opts []Option
		want widgetapi.Options
	}{}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ti, err := New(tc.opts...)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			got := ti.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}