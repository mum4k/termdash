// Copyright 2018 Google Inc.
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

package wrap

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/kylelemons/godebug/pretty"
	"termdash/cell"
	"termdash/internal/canvas/buffer"
)

func TestValidTextAndCells(t *testing.T) {
	tests := []struct {
		desc    string
		text    string // All runes checked individually.
		wantErr bool
	}{
		{
			desc:    "empty text is not valid",
			wantErr: true,
		},
		{
			desc: "digits are allowed",
			text: "0123456789",
		},
		{
			desc: "all printable ASCII characters are allowed",
			text: func() string {
				var b strings.Builder
				for i := 0; i < unicode.MaxASCII; i++ {
					r := rune(i)
					if unicode.IsPrint(r) {
						b.WriteRune(r)
					}
				}
				return b.String()
			}(),
		},
		{
			desc: "all printable Unicode characters in the Latin-1 space are allowed",
			text: func() string {
				var b strings.Builder
				for i := 0; i < unicode.MaxLatin1; i++ {
					r := rune(i)
					if unicode.IsPrint(r) {
						b.WriteRune(r)
					}
				}
				return b.String()
			}(),
		},
		{
			desc: "sample of half-width unicode runes that are allowed",
			text: "ｾｶｲ☆",
		},
		{
			desc: "sample of full-width unicode runes that are allowed",
			text: "世界",
		},
		{
			desc: "spaces are allowed",
			text: "   ",
		},
		{
			desc:    "no other space characters",
			text:    fmt.Sprintf("\t\v\f\r%c%c", 0x85, 0xA0),
			wantErr: true,
		},
		{
			desc:    "no control characters",
			text:    fmt.Sprintf("%c%c", 0x0000, 0x007f),
			wantErr: true,
		},

		{
			desc: "newlines are allowed",
			text: "   ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			{
				err := ValidText(tc.text)
				if (err != nil) != tc.wantErr {
					t.Errorf("ValidText => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
			}

			// All individual runes must give the same result.
			for _, r := range tc.text {
				err := ValidText(string(r))
				if (err != nil) != tc.wantErr {
					t.Errorf("ValidText => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
			}

			var cells []*buffer.Cell
			for _, r := range tc.text {
				cells = append(cells, buffer.NewCell(r))
			}
			{
				err := ValidCells(cells)
				if (err != nil) != tc.wantErr {
					t.Errorf("ValidCells => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
			}
		})
	}
}

func TestCells(t *testing.T) {
	tests := []struct {
		desc  string
		cells []*buffer.Cell
		// width is the width of the canvas.
		width   int
		mode    Mode
		want    [][]*buffer.Cell
		wantErr bool
	}{
		{
			desc:    "fails with zero text",
			width:   1,
			wantErr: true,
		},
		{
			desc:    "fails with invalid runes (tabs)",
			cells:   buffer.NewCells("hello\t"),
			width:   1,
			wantErr: true,
		},
		{
			desc:    "fails with unsupported wrap mode",
			cells:   buffer.NewCells("hello"),
			width:   1,
			mode:    Mode(-1),
			wantErr: true,
		},
		{
			desc:  "zero canvas width",
			cells: buffer.NewCells("hello"),
			width: 0,
			want:  nil,
		},
		{
			desc:  "wrapping disabled, no newlines, fits in canvas width",
			cells: buffer.NewCells("hello"),
			width: 5,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
			},
		},
		{
			desc:  "wrapping disabled, no newlines, doesn't fits in canvas width",
			cells: buffer.NewCells("hello"),
			width: 4,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
			},
		},
		{
			desc:  "wrapping disabled, newlines, fits in canvas width",
			cells: buffer.NewCells("hello\nworld"),
			width: 5,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
				buffer.NewCells("world"),
			},
		},
		{
			desc:  "wrapping disabled, newlines, doesn't fit in canvas width",
			cells: buffer.NewCells("hello\nworld"),
			width: 4,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
				buffer.NewCells("world"),
			},
		},
		{
			desc:  "wrapping enabled, no newlines, fits in canvas width",
			cells: buffer.NewCells("hello"),
			width: 5,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
			},
		},
		{
			desc:  "wrapping enabled, no newlines, doesn't fit in canvas width",
			cells: buffer.NewCells("hello"),
			width: 4,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hell"),
				buffer.NewCells("o"),
			},
		},
		{
			desc:  "wrapping enabled, newlines, fits in canvas width",
			cells: buffer.NewCells("hello\nworld"),
			width: 5,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
				buffer.NewCells("world"),
			},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in canvas width",
			cells: buffer.NewCells("hello\nworld"),
			width: 4,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hell"),
				buffer.NewCells("o"),
				buffer.NewCells("worl"),
				buffer.NewCells("d"),
			},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in canvas width, unicode characters",
			cells: buffer.NewCells("⇧\n…\n⇩"),
			width: 1,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("⇧"),
				buffer.NewCells("…"),
				buffer.NewCells("⇩"),
			},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in width, full-width unicode characters",
			cells: buffer.NewCells("你好\n世界"),
			width: 2,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("你"),
				buffer.NewCells("好"),
				buffer.NewCells("世"),
				buffer.NewCells("界"),
			},
		},
		{
			desc:  "wraps before a full-width character that starts in and falls out",
			cells: buffer.NewCells("a你b"),
			width: 2,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("a"),
				buffer.NewCells("你"),
				buffer.NewCells("b"),
			},
		},
		{
			desc:  "wraps before a full-width character that falls out",
			cells: buffer.NewCells("ab你b"),
			width: 2,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("ab"),
				buffer.NewCells("你"),
				buffer.NewCells("b"),
			},
		},
		{
			desc:  "handles leading and trailing newlines",
			cells: buffer.NewCells("\n\n\nhello\n\n\n"),
			width: 4,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells("hell"),
				buffer.NewCells("o"),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
			},
		},
		{
			desc:  "handles multiple newlines in the middle",
			cells: buffer.NewCells("hello\n\n\nworld"),
			width: 5,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells("world"),
			},
		},
		{
			desc:  "handles multiple newlines in the middle and wraps",
			cells: buffer.NewCells("hello\n\n\nworld"),
			width: 2,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("he"),
				buffer.NewCells("ll"),
				buffer.NewCells("o"),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells("wo"),
				buffer.NewCells("rl"),
				buffer.NewCells("d"),
			},
		},
		{
			desc:  "contains only newlines",
			cells: buffer.NewCells("\n\n\n"),
			width: 4,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
			},
		},
		{
			desc:  "wraps at words, no need to wrap",
			cells: buffer.NewCells("aaa bb cc"),
			width: 9,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aaa bb cc"),
			},
		},
		{
			desc:  "wraps at words, all fit individually, wrap falls on space",
			cells: buffer.NewCells("aaa bb cc"),
			width: 6,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aaa bb"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, all fit individually, each word on its own line",
			cells: buffer.NewCells("aaa bb cc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aaa"),
				buffer.NewCells("bb"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, respects newline characters with spaces between words",
			cells: buffer.NewCells("aaa \n bb cc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aaa"),
				buffer.NewCells(" "),
				buffer.NewCells(" bb"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, respects newline characters between words",
			cells: buffer.NewCells("aaa\nbb cc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aaa"),
				buffer.NewCells("bb"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, respects multiple spaces between words",
			cells: buffer.NewCells("aa   bb cc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa "),
				buffer.NewCells(" "),
				buffer.NewCells("bb"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, handles leading spaces",
			cells: buffer.NewCells("   aa bb cc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("  "),
				buffer.NewCells("aa"),
				buffer.NewCells("bb"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, handles trailing spaces",
			cells: buffer.NewCells("aa bb cc   "),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa"),
				buffer.NewCells("bb"),
				buffer.NewCells("cc "),
				buffer.NewCells("  "),
			},
		},
		{
			desc:  "wraps at words, handles leading newlines",
			cells: buffer.NewCells("\n\n\naa bb cc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells("aa"),
				buffer.NewCells("bb"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, handles trailing newlines",
			cells: buffer.NewCells("aa bb cc\n\n\n"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa"),
				buffer.NewCells("bb"),
				buffer.NewCells("cc"),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
			},
		},
		{
			desc:  "wraps at words, handles continuous newlines",
			cells: buffer.NewCells("aa\n\n\ncc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa"),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, punctuation is wrapped with words",
			cells: buffer.NewCells("aa. bb! cc?"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa."),
				buffer.NewCells("bb!"),
				buffer.NewCells("cc?"),
			},
		},
		{
			desc:  "wraps at words, quotes are wrapped with words",
			cells: buffer.NewCells("'aa' 'bb' 'cc'"),
			width: 4,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("'aa'"),
				buffer.NewCells("'bb'"),
				buffer.NewCells("'cc'"),
			},
		},
		{
			desc:  "wraps at words, begins with a word too long for one line",
			cells: buffer.NewCells("aabbcc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa-"),
				buffer.NewCells("bbc"),
				buffer.NewCells("c"),
			},
		},
		{
			desc:  "wraps at words, begins with a word too long for one line, width is one",
			cells: buffer.NewCells("abcd"),
			width: 1,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("a"),
				buffer.NewCells("b"),
				buffer.NewCells("c"),
				buffer.NewCells("d"),
			},
		},
		{
			desc:  "wraps at words, begins with a word too long for multiple lines",
			cells: buffer.NewCells("aabbccaabbcc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa-"),
				buffer.NewCells("bbc"),
				buffer.NewCells("caa"),
				buffer.NewCells("bbc"),
				buffer.NewCells("c"),
			},
		},
		{
			desc:  "wraps at words, a word doesn't fit on one line",
			cells: buffer.NewCells("aa bbbb cc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa"),
				buffer.NewCells("bb-"),
				buffer.NewCells("bb"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, a word doesn't fit on multiple line",
			cells: buffer.NewCells("aa bbbbbb cc"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa"),
				buffer.NewCells("bb-"),
				buffer.NewCells("bbb"),
				buffer.NewCells("b"),
				buffer.NewCells("cc"),
			},
		},
		{
			desc:  "wraps at words, a word doesn't fit on multiple line, width is one so no dash",
			cells: buffer.NewCells("a bbb"),
			width: 1,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("a"),
				buffer.NewCells("b"),
				buffer.NewCells("b"),
				buffer.NewCells("b"),
			},
		},
		{
			desc:  "wraps at words, starts with half-width runes word, fits exactly",
			cells: buffer.NewCells("aaa"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aaa"),
			},
		},
		{
			desc:  "wraps at words, starts with half-width runes word, wraps",
			cells: buffer.NewCells("abc"),
			width: 2,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("a-"),
				buffer.NewCells("bc"),
			},
		},
		{
			desc:  "wraps at words, starts with full-width runes word, fits exactly",
			cells: buffer.NewCells("世世"),
			width: 4,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("世世"),
			},
		},
		{
			desc:  "wraps at words, starts with full-width runes word, wraps",
			cells: buffer.NewCells("世世"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("世"),
				buffer.NewCells("世"),
			},
		},
		{
			desc:  "wraps at words, a full-width rune word in the middle, fits exactly",
			cells: buffer.NewCells("aaaa 世世"),
			width: 4,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aaaa"),
				buffer.NewCells("世世"),
			},
		},
		{
			desc:  "wraps at words, a full-width rune word in the middle, one cell left, wraps",
			cells: buffer.NewCells("aaa 世世"),
			width: 3,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aaa"),
				buffer.NewCells("世"),
				buffer.NewCells("世"),
			},
		},
		{
			desc:  "wraps at words, a full-width rune word in the middle, no cell left, wraps",
			cells: buffer.NewCells("aa 世世"),
			width: 2,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa"),
				buffer.NewCells("世"),
				buffer.NewCells("世"),
			},
		},
		{
			desc:  "wraps of words with half-width runes preserves cell options",
			cells: buffer.NewCells("a bc", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
			width: 2,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("a", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
				buffer.NewCells("bc", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
			},
		},
		{
			desc:  "wraps of words with full-width runes preserves cell options",
			cells: buffer.NewCells("aa 世世", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
			width: 2,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("aa", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
				buffer.NewCells("世", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
				buffer.NewCells("世", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
			},
		},
		{
			desc:  "inserted dash inherits cell options",
			cells: buffer.NewCells("abc", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
			width: 2,
			mode:  AtWords,
			want: [][]*buffer.Cell{
				buffer.NewCells("a-", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
				buffer.NewCells("bc", cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Logf(fmt.Sprintf("Mode: %v", tc.mode))
			got, err := Cells(tc.cells, tc.width, tc.mode)
			if (err != nil) != tc.wantErr {
				t.Errorf("Cells => unexpected error %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Cells =>\n got:%v\nwant:%v\nunexpected diff (-want, +got):\n%s", got, tc.want, diff)
			}
		})
	}

}

func TestRuneWrapNeeded(t *testing.T) {
	tests := []struct {
		desc  string
		r     rune
		posX  int
		width int
		want  bool
	}{
		{
			desc:  "half-width rune, falls within canvas",
			r:     'a',
			posX:  2,
			width: 3,
			want:  false,
		},
		{
			desc:  "full-width rune, falls within canvas",
			r:     '世',
			posX:  1,
			width: 3,
			want:  false,
		},
		{
			desc:  "half-width rune, falls outside of canvas, wrapping configured",
			r:     'a',
			posX:  3,
			width: 3,
			want:  true,
		},
		{
			desc:  "full-width rune, starts in and falls outside of canvas, wrapping configured",
			r:     '世',
			posX:  3,
			width: 3,
			want:  true,
		},
		{
			desc:  "full-width rune, starts outside of canvas, wrapping configured",
			r:     '世',
			posX:  3,
			width: 3,
			want:  true,
		},
		{
			desc:  "doesn't wrap for newline characters",
			r:     '\n',
			posX:  3,
			width: 3,
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := runeWrapNeeded(tc.r, tc.posX, tc.width)
			if got != tc.want {
				t.Errorf("runeWrapNeeded => got %v, want %v", got, tc.want)
			}
		})
	}
}
