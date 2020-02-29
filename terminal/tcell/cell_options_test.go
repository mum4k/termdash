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
	"testing"

	"github.com/gdamore/tcell"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

func TestCellColor(t *testing.T) {
	tests := []struct {
		color cell.Color
		want  tcell.Color
	}{
		{cell.ColorDefault, tcell.ColorDefault},
		{cell.ColorBlack, tcell.ColorBlack},
		{cell.ColorRed, tcell.ColorMaroon},
		{cell.ColorGreen, tcell.ColorGreen},
		{cell.ColorYellow, tcell.ColorOlive},
		{cell.ColorBlue, tcell.ColorNavy},
		{cell.ColorMagenta, tcell.ColorPurple},
		{cell.ColorCyan, tcell.ColorTeal},
		{cell.ColorWhite, tcell.ColorSilver},
		{cell.ColorNumber(42), tcell.Color(42)},
	}

	for _, tc := range tests {
		t.Run(tc.color.String(), func(t *testing.T) {
			got := cellColor(tc.color)
			if got != tc.want {
				t.Errorf("cellColor(%v) => got %v, want %v", tc.color, got, tc.want)
			}
		})
	}
}

func TestFixColor(t *testing.T) {
	tests := []struct {
		colorMode terminalapi.ColorMode
		color     cell.Color
		want      tcell.Color
	}{
		// See https://jonasjacek.github.io/colors/ for a good reference of all 256 xterm colors
		// All 256 colors
		{terminalapi.ColorMode256, cell.ColorDefault, tcell.ColorDefault},
		{terminalapi.ColorMode256, cell.ColorBlack, tcell.ColorBlack},
		{terminalapi.ColorMode256, cell.ColorRed, tcell.ColorMaroon},
		{terminalapi.ColorMode256, cell.ColorGreen, tcell.ColorGreen},
		{terminalapi.ColorMode256, cell.ColorYellow, tcell.ColorOlive},
		{terminalapi.ColorMode256, cell.ColorBlue, tcell.ColorNavy},
		{terminalapi.ColorMode256, cell.ColorMagenta, tcell.ColorPurple},
		{terminalapi.ColorMode256, cell.ColorCyan, tcell.ColorTeal},
		{terminalapi.ColorMode256, cell.ColorWhite, tcell.ColorSilver},
		{terminalapi.ColorMode256, cell.ColorNumber(42), tcell.Color(42)},
		// 8 system colors
		{terminalapi.ColorModeNormal, cell.ColorDefault, tcell.ColorDefault},
		{terminalapi.ColorModeNormal, cell.ColorBlack, tcell.ColorBlack},
		{terminalapi.ColorModeNormal, cell.ColorRed, tcell.ColorMaroon},
		{terminalapi.ColorModeNormal, cell.ColorGreen, tcell.ColorGreen},
		{terminalapi.ColorModeNormal, cell.ColorYellow, tcell.ColorOlive},
		{terminalapi.ColorModeNormal, cell.ColorBlue, tcell.ColorNavy},
		{terminalapi.ColorModeNormal, cell.ColorMagenta, tcell.ColorPurple},
		{terminalapi.ColorModeNormal, cell.ColorCyan, tcell.ColorTeal},
		{terminalapi.ColorModeNormal, cell.ColorWhite, tcell.ColorSilver},
		{terminalapi.ColorModeNormal, cell.ColorNumber(42), tcell.Color(10)},
		// Grayscale colors (all the grey colours from 231 to 255)
		{terminalapi.ColorModeGrayscale, cell.ColorDefault, tcell.Color231},
		{terminalapi.ColorModeGrayscale, cell.ColorBlack, tcell.Color232},
		{terminalapi.ColorModeGrayscale, cell.ColorRed, tcell.Color233},
		{terminalapi.ColorModeGrayscale, cell.ColorGreen, tcell.Color234},
		{terminalapi.ColorModeGrayscale, cell.ColorYellow, tcell.Color235},
		{terminalapi.ColorModeGrayscale, cell.ColorBlue, tcell.Color236},
		{terminalapi.ColorModeGrayscale, cell.ColorMagenta, tcell.Color237},
		{terminalapi.ColorModeGrayscale, cell.ColorCyan, tcell.Color238},
		{terminalapi.ColorModeGrayscale, cell.ColorWhite, tcell.Color239},
		{terminalapi.ColorModeGrayscale, cell.ColorNumber(42), tcell.Color(250)},
		// 216 colors (16 to 231)
		{terminalapi.ColorMode216, cell.ColorDefault, tcell.ColorWhite},
		{terminalapi.ColorMode216, cell.ColorBlack, tcell.Color16},
		{terminalapi.ColorMode216, cell.ColorRed, tcell.Color17},
		{terminalapi.ColorMode216, cell.ColorGreen, tcell.Color18},
		{terminalapi.ColorMode216, cell.ColorYellow, tcell.Color19},
		{terminalapi.ColorMode216, cell.ColorBlue, tcell.Color20},
		{terminalapi.ColorMode216, cell.ColorMagenta, tcell.Color21},
		{terminalapi.ColorMode216, cell.ColorCyan, tcell.Color22},
		{terminalapi.ColorMode216, cell.ColorWhite, tcell.Color23},
		{terminalapi.ColorMode216, cell.ColorNumber(42), tcell.Color(58)},
	}

	for _, tc := range tests {
		t.Run(tc.colorMode.String()+"_"+tc.color.String(), func(t *testing.T) {
			color := cellColor(tc.color)
			got := fixColor(color, tc.colorMode)
			if got != tc.want {
				t.Errorf("fixColor(%v_%v), => got %v, want %v", tc.colorMode, tc.color, got, tc.want)
			}
		})
	}
}

func TestCellOptsToStyle(t *testing.T) {
	tests := []struct {
		colorMode terminalapi.ColorMode
		opts      cell.Options
		want      tcell.Style
	}{
		{
			colorMode: terminalapi.ColorMode256,
			opts:      cell.Options{FgColor: cell.ColorWhite, BgColor: cell.ColorBlack},
			want:      tcell.StyleDefault.Foreground(tcell.ColorSilver).Background(tcell.ColorBlack),
		},
		{
			colorMode: terminalapi.ColorModeNormal,
			opts:      cell.Options{FgColor: cell.ColorWhite, BgColor: cell.ColorBlack},
			want:      tcell.StyleDefault.Foreground(tcell.ColorSilver).Background(tcell.ColorBlack),
		},
		{
			colorMode: terminalapi.ColorModeGrayscale,
			opts:      cell.Options{FgColor: cell.ColorWhite, BgColor: cell.ColorBlack},
			want:      tcell.StyleDefault.Foreground(tcell.Color239).Background(tcell.Color232),
		},
		{
			colorMode: terminalapi.ColorMode216,
			opts:      cell.Options{FgColor: cell.ColorWhite, BgColor: cell.ColorBlack},
			want:      tcell.StyleDefault.Foreground(tcell.Color23).Background(tcell.Color16),
		},
	}

	for _, tc := range tests {
		t.Run(tc.opts.FgColor.String()+"+"+tc.opts.BgColor.String(), func(t *testing.T) {
			got := cellOptsToStyle(&tc.opts, tc.colorMode)
			if got != tc.want {
				fg, bg, _ := got.Decompose()
				wantFg, wantBg, _ := tc.want.Decompose()
				t.Errorf("cellOptsToStyle(%v, fg=%v, bg=%v) => got (fg=%X, bg=%X), want (fg=%X, bg=%X)",
					tc.colorMode, tc.opts.FgColor, tc.opts.BgColor, fg, bg, wantFg, wantBg)
			}
		})
	}
}
