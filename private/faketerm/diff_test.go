// Copyright 2022 Google LLC
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

package faketerm

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/cell"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		desc     string
		term1    *Terminal
		term2    *Terminal
		wantDiff bool
	}{
		{
			desc: "no diff on equal terminals",
			term1: func() *Terminal {
				t := MustNew(image.Point{2, 2})
				t.SetCell(image.Point{0, 0}, 'a')
				return t
			}(),
			term2: func() *Terminal {
				t := MustNew(image.Point{2, 2})
				t.SetCell(image.Point{0, 0}, 'a')
				return t
			}(),
			wantDiff: false,
		},
		{
			desc: "reports diff on when cell runes differ",
			term1: func() *Terminal {
				t := MustNew(image.Point{2, 2})
				t.SetCell(image.Point{0, 0}, 'a')
				return t
			}(),
			term2: func() *Terminal {
				t := MustNew(image.Point{2, 2})
				t.SetCell(image.Point{1, 1}, 'a')
				return t
			}(),
			wantDiff: true,
		},
		{
			desc: "reports diff on when cell options differ",
			term1: func() *Terminal {
				t := MustNew(image.Point{2, 2})
				t.SetCell(image.Point{0, 0}, 'a', cell.Bold())
				return t
			}(),
			term2: func() *Terminal {
				t := MustNew(image.Point{2, 2})
				t.SetCell(image.Point{0, 0}, 'a')
				return t
			}(),
			wantDiff: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotDiff := Diff(tc.term1, tc.term2)
			if (gotDiff != "") != tc.wantDiff {
				t.Errorf("Diff -> unexpected diff while wantDiff:%v, the diff:\n%s", tc.wantDiff, gotDiff)
			}
		})
	}
}
