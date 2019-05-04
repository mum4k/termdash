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

package table

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

func TestTable(t *testing.T) {
	tests := []struct {
		desc          string
		opts          []Option
		update        func(*Table) error // update gets called before drawing of the widget.
		canvas        image.Rectangle
		meta          *widgetapi.Meta
		want          func(size image.Point) *faketerm.Terminal
		wantNewErr    bool
		wantUpdateErr bool // whether to expect an error on a call to the update function
		wantDrawErr   bool
	}{}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tbl, err := New(tc.opts...)
			if (err != nil) != tc.wantNewErr {
				t.Errorf("New => unexpected error: %v, wantNewErr: %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			if tc.update != nil {
				err = tc.update(tbl)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}
			}

			err = tbl.Draw(c, tc.meta)
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
		})
	}
}

func TestKeyboard(t *testing.T) {
	tbl, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := tbl.Keyboard(&terminalapi.Keyboard{}); err == nil {
		t.Errorf("Keyboard => got nil err, wanted one")
	}
}

func TestMouse(t *testing.T) {
	tbl, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := tbl.Mouse(&terminalapi.Mouse{}); err == nil {
		t.Errorf("Mouse => got nil err, wanted one")
	}
}

func TestOptions(t *testing.T) {
	tbl, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	got := tbl.Options()
	want := widgetapi.Options{
		MinimumSize:  image.Point{1, 1},
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
	}

}
