package canvas

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestNew(t *testing.T) {
	tests := []struct {
		desc     string
		area     image.Rectangle
		wantSize image.Point
		wantErr  bool
	}{
		{
			desc:    "area min has negative X",
			area:    image.Rect(-1, 0, 0, 0),
			wantErr: true,
		},
		{
			desc:    "area min has negative Y",
			area:    image.Rect(0, -1, 0, 0),
			wantErr: true,
		},
		{
			desc:    "area max has negative X",
			area:    image.Rect(0, 0, -1, 0),
			wantErr: true,
		},
		{
			desc:    "area max has negative Y",
			area:    image.Rect(0, 0, 0, -1),
			wantErr: true,
		},
		{
			desc:     "smallest valid size",
			area:     image.Rect(0, 0, 0, 0),
			wantSize: image.Point{1, 1},
		},
		{
			desc:     "rectangular canvas 3 by 4",
			area:     image.Rect(0, 0, 2, 3),
			wantSize: image.Point{3, 4},
		},
		{
			desc:     "non-zero based area",
			area:     image.Rect(1, 1, 2, 3),
			wantSize: image.Point{2, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := New(tc.area)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			got := c.Size()
			if diff := pretty.Compare(tc.wantSize, got); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
