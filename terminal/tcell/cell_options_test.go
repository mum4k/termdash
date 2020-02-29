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
