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

package runewidth

import (
	"testing"

	runewidth "github.com/mattn/go-runewidth"
)

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		desc      string
		runes     []rune
		opts      []Option
		eastAsian bool
		want      int
	}{
		{
			desc:  "ascii characters",
			runes: []rune{'a', 'f', '#'},
			want:  1,
		},
		{
			desc:  "non-printable characters from mattn/runewidth/runewidth_test",
			runes: []rune{'\x00', '\x01', '\u0300', '\u2028', '\u2029', '\n'},
			want:  0,
		},
		{
			desc:  "override rune width with an option",
			runes: []rune{'\n'},
			opts: []Option{
				CountAsWidth('\n', 3),
			},
			want: 3,
		},
		{
			desc:  "half-width runes from mattn/runewidth/runewidth_test",
			runes: []rune{'ｾ', 'ｶ', 'ｲ', '☆'},
			want:  1,
		},
		{
			desc:  "full-width runes from mattn/runewidth/runewidth_test",
			runes: []rune{'世', '界'},
			want:  2,
		},
		{
			desc:      "ambiguous so double-width in eastAsian from mattn/runewidth/runewidth_test",
			runes:     []rune{'☆'},
			eastAsian: true,
			want:      2,
		},
		{
			desc:  "braille runes",
			runes: []rune{'⠀', '⠴', '⠷', '⣿'},
			want:  1,
		},
		{
			desc:      "braille runes in eastAsian",
			runes:     []rune{'⠀', '⠴', '⠷', '⣿'},
			eastAsian: true,
			want:      1,
		},
		{
			desc:  "termdash special runes",
			runes: []rune{'⇄', '…', '⇧', '⇩', '⇦', '⇨'},
			want:  1,
		},
		{
			desc:      "termdash special runes in eastAsian",
			runes:     []rune{'⇄', '…', '⇧', '⇩', '⇦', '⇨'},
			eastAsian: true,
			want:      1,
		},
		{
			desc:  "termdash sparks",
			runes: []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'},
			want:  1,
		},
		{
			desc:      "termdash sparks in eastAsian",
			runes:     []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'},
			eastAsian: true,
			want:      1,
		},
		{
			desc:  "termdash line styles",
			runes: []rune{'─', '═', '─', '┼', '╬', '┼'},
			want:  1,
		},
		{
			desc:      "termdash line styles in eastAsian",
			runes:     []rune{'─', '═', '─', '┼', '╬', '┼'},
			eastAsian: true,
			want:      1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			runewidth.DefaultCondition.EastAsianWidth = tc.eastAsian
			defer func() {
				runewidth.DefaultCondition.EastAsianWidth = false
			}()

			for _, r := range tc.runes {
				if got := RuneWidth(r, tc.opts...); got != tc.want {
					t.Errorf("RuneWidth(%c, %#x) => %v, want %v", r, r, got, tc.want)
				}
			}
		})
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		desc      string
		str       string
		opts      []Option
		eastAsian bool
		want      int
	}{
		{
			desc: "ascii characters",
			str:  "hello",
			want: 5,
		},
		{
			desc: "override rune widths with an option",
			str:  "hello",
			opts: []Option{
				CountAsWidth('h', 5),
				CountAsWidth('e', 5),
			},
			want: 13,
		},
		{
			desc: "string from mattn/runewidth/runewidth_test",
			str:  "■㈱の世界①",
			want: 10,
		},
		{
			desc:      "string in eastAsian from mattn/runewidth/runewidth_test",
			str:       "■㈱の世界①",
			eastAsian: true,
			want:      12,
		},
		{
			desc: "string using termdash characters",
			str:  "⇄…⇧⇩",
			want: 4,
		},
		{
			desc:      "string in eastAsien using termdash characters",
			str:       "⇄…⇧⇩",
			eastAsian: true,
			want:      4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			runewidth.DefaultCondition.EastAsianWidth = tc.eastAsian
			defer func() {
				runewidth.DefaultCondition.EastAsianWidth = false
			}()

			if got := StringWidth(tc.str, tc.opts...); got != tc.want {
				t.Errorf("StringWidth(%q) => %v, want %v", tc.str, got, tc.want)
			}
		})
	}
}
