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

package attrrange

import (
	"log"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
)

func Example() {
	// Caller has a slice of some attributes, like a cell color that applies
	// to a portion of text.
	attrs := []cell.Color{cell.ColorRed, cell.ColorBlue}
	redIdx := 0
	blueIdx := 1

	// This is the text the colors apply to.
	const text = "HelloWorld"

	// Assuming that we want the word "Hello" in red and the word "World" in
	// green, we can set our ranges as follows:
	tr := NewTracker()
	if err := tr.Add(0, len("Hello"), redIdx); err != nil {
		panic(err)
	}
	if err := tr.Add(len("Hello")+1, len(text), blueIdx); err != nil {
		panic(err)
	}

	// Now to get the index into attrs (i.e. the color) for a particular
	// character, we can do:
	for i, c := range text {
		ar, err := tr.ForPosition(i)
		if err != nil {
			panic(err)
		}
		log.Printf("character at text[%d] = %q, color index %d = %v, range low:%d, high:%d", i, c, ar.AttrIdx, attrs[ar.AttrIdx], ar.Low, ar.High)
	}
}

func TestForPosition(t *testing.T) {
	tests := []struct {
		desc string
		// if not nil, called before calling ForPosition.
		// Can add ranges.
		update        func(*Tracker) error
		pos           int
		want          *AttrRange
		wantErr       error
		wantUpdateErr bool
	}{
		{
			desc:    "fails when no ranges given",
			pos:     0,
			wantErr: ErrNotFound,
		},
		{
			desc: "fails to add a duplicate",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(2, 3, 41)
			},
			wantUpdateErr: true,
		},
		{
			desc: "fails when multiple given ranges, position falls before them",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:     1,
			wantErr: ErrNotFound,
		},
		{
			desc: "multiple given options, position falls on the lower",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:  2,
			want: newAttrRange(2, 5, 40),
		},
		{
			desc: "multiple given options, position falls between them",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:  4,
			want: newAttrRange(2, 5, 40),
		},
		{
			desc: "multiple given options, position falls on the higher",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:  5,
			want: newAttrRange(5, 10, 41),
		},
		{
			desc: "multiple given options, position falls after them",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:     10,
			wantErr: ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tr := NewTracker()
			if tc.update != nil {
				err := tc.update(tr)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("tc.update => unexpected error:%v, wantUpdateErr:%v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}
			}

			got, err := tr.ForPosition(tc.pos)
			if err != tc.wantErr {
				t.Errorf("ForPosition => unexpected error:%v, wantErr:%v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("ForPosition => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
