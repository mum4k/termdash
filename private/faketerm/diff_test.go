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
