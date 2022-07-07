// Copyright 2020 Google Inc.
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

package tcell

import (
	"reflect"
	"testing"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

func TestCellOptsToStyle(t *testing.T) {
	tests := []struct {
		desc      string
		colorMode terminalapi.ColorMode
		opts      cell.Options
		want      tcell.Style
	}{
		{
			desc:      "ColorMode256: ColorDefault and ColorBlack",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorDefault,
				BgColor: cell.ColorBlack,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorDefault).
				Background(tcell.ColorBlack),
		},
		{
			desc:      "ColorMode256: ColorMaroon and ColorGreen",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorMaroon,
				BgColor: cell.ColorGreen,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorMaroon).
				Background(tcell.ColorGreen),
		},
		{
			desc:      "ColorMode256: ColorOlive and ColorNavy",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorOlive,
				BgColor: cell.ColorNavy,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorOlive).
				Background(tcell.ColorNavy),
		},
		{
			desc:      "ColorMode256: ColorPurple and ColorTeal",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorPurple,
				BgColor: cell.ColorTeal,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorPurple).
				Background(tcell.ColorTeal),
		},
		{
			desc:      "ColorMode256: ColorSilver and ColorGray",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorSilver,
				BgColor: cell.ColorGray,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorSilver).
				Background(tcell.ColorGray),
		},
		{
			desc:      "ColorMode256: ColorRed and ColorLime",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorRed,
				BgColor: cell.ColorLime,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorRed).
				Background(tcell.ColorLime),
		},
		{
			desc:      "ColorMode256: ColorYellow and ColorBlue",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorYellow,
				BgColor: cell.ColorBlue,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorYellow).
				Background(tcell.ColorBlue),
		},
		{
			desc:      "ColorMode256: ColorFuchsia and ColorAqua",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorFuchsia,
				BgColor: cell.ColorAqua,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorFuchsia).
				Background(tcell.ColorAqua),
		},
		{
			desc:      "ColorMode256: ColorWhite and ColorDefault",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorWhite,
				BgColor: cell.ColorDefault,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorWhite).
				Background(tcell.ColorDefault),
		},
		{
			desc:      "ColorMode256: termbox compatibility colors ColorMagenta and ColorCyan",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorMagenta,
				BgColor: cell.ColorCyan,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorPurple).
				Background(tcell.ColorTeal),
		},
		{
			desc:      "ColorMode256: first(0) and last(255) numbered color",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorNumber(0),
				BgColor: cell.ColorNumber(255),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorBlack).
				Background(tcell.Color255),
		},
		{
			desc:      "ColorMode256: two numbered colors",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorNumber(33),
				BgColor: cell.ColorNumber(200),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color33).
				Background(tcell.Color200),
		},
		{
			desc:      "ColorMode256: first and last RGB6 color",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorRGB6(0, 0, 0),
				BgColor: cell.ColorRGB6(5, 5, 5),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color16).
				Background(tcell.Color231),
		},
		{
			desc:      "ColorMode256: first and last RGB24 color",
			colorMode: terminalapi.ColorMode256,
			opts: cell.Options{
				FgColor: cell.ColorRGB24(0, 0, 0),
				BgColor: cell.ColorRGB24(255, 255, 255),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color16).
				Background(tcell.Color231),
		},
		{
			desc:      "ColorModeNormal: first and last color",
			colorMode: terminalapi.ColorModeNormal,
			opts: cell.Options{
				FgColor: cell.ColorBlack,
				BgColor: cell.ColorWhite,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorBlack).
				Background(tcell.ColorWhite),
		},
		{
			desc:      "ColorModeNormal: colors in the middle",
			colorMode: terminalapi.ColorModeNormal,
			opts: cell.Options{
				FgColor: cell.ColorGreen,
				BgColor: cell.ColorOlive,
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorGreen).
				Background(tcell.ColorOlive),
		},
		{
			desc:      "ColorModeNormal: colors above the range rotate back",
			colorMode: terminalapi.ColorModeNormal,
			opts: cell.Options{
				FgColor: cell.ColorNumber(17),
				BgColor: cell.ColorNumber(18),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorBlack).
				Background(tcell.ColorMaroon),
		},
		{
			desc:      "ColorMode216: first and last color",
			colorMode: terminalapi.ColorMode216,
			opts: cell.Options{
				FgColor: cell.ColorNumber(0),
				BgColor: cell.ColorNumber(215),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color16).
				Background(tcell.Color231),
		},
		{
			desc:      "ColorMode216: colors in the middle",
			colorMode: terminalapi.ColorMode216,
			opts: cell.Options{
				FgColor: cell.ColorNumber(1),
				BgColor: cell.ColorNumber(2),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color17).
				Background(tcell.Color18),
		},
		{
			desc:      "ColorMode216: colors above the range rotate back",
			colorMode: terminalapi.ColorMode216,
			opts: cell.Options{
				FgColor: cell.ColorNumber(216),
				BgColor: cell.ColorNumber(217),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color16).
				Background(tcell.Color17),
		},
		{
			desc:      "ColorModeGrayscale: first and last color",
			colorMode: terminalapi.ColorModeGrayscale,
			opts: cell.Options{
				FgColor: cell.ColorNumber(0),
				BgColor: cell.ColorNumber(23),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color232).
				Background(tcell.Color255),
		},
		{
			desc:      "ColorModeGrayscale: colors in the middle",
			colorMode: terminalapi.ColorModeGrayscale,
			opts: cell.Options{
				FgColor: cell.ColorNumber(1),
				BgColor: cell.ColorNumber(2),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color233).
				Background(tcell.Color234),
		},
		{
			desc:      "ColorModeGrayscale: colors above the range rotate back",
			colorMode: terminalapi.ColorModeGrayscale,
			opts: cell.Options{
				FgColor: cell.ColorNumber(24),
				BgColor: cell.ColorNumber(25),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.Color232).
				Background(tcell.Color233),
		},
		{
			desc:      "Unknown color mode converts to default color",
			colorMode: terminalapi.ColorMode(-1),
			opts: cell.Options{
				FgColor: cell.ColorNumber(24),
				BgColor: cell.ColorNumber(25),
			},
			want: tcell.StyleDefault.
				Foreground(tcell.ColorDefault).
				Background(tcell.ColorDefault),
		},
		{
			colorMode: terminalapi.ColorModeNormal,
			opts:      cell.Options{Bold: true},
			want:      tcell.StyleDefault.Bold(true),
		},
		{
			colorMode: terminalapi.ColorModeNormal,
			opts:      cell.Options{Italic: true},
			want:      tcell.StyleDefault.Italic(true),
		},
		{
			colorMode: terminalapi.ColorModeNormal,
			opts:      cell.Options{Underline: true},
			want:      tcell.StyleDefault.Underline(true),
		},
		{
			colorMode: terminalapi.ColorModeNormal,
			opts:      cell.Options{Strikethrough: true},
			want:      tcell.StyleDefault.StrikeThrough(true),
		},
		{
			colorMode: terminalapi.ColorModeNormal,
			opts:      cell.Options{Inverse: true},
			want:      tcell.StyleDefault.Reverse(true),
		},
		{
			colorMode: terminalapi.ColorModeNormal,
			opts:      cell.Options{Blink: true},
			want:      tcell.StyleDefault.Blink(true),
		},
		{
			colorMode: terminalapi.ColorModeNormal,
			opts:      cell.Options{Dim: true},
			want:      tcell.StyleDefault.Dim(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := cellOptsToStyle(&tc.opts, tc.colorMode)
			if !reflect.DeepEqual(got, tc.want) {
				diff := pretty.Compare(tc.want, got)
				t.Logf("opts: %+v\nstyle:%+v", tc.opts, got)
				t.Errorf("cellOptsToStyle => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
