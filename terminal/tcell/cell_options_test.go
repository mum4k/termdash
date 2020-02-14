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
