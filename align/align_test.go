package align

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestRectangle(t *testing.T) {
	tests := []struct {
		desc    string
		rect    image.Rectangle
		area    image.Rectangle
		hAlign  Horizontal
		vAlign  Vertical
		want    image.Rectangle
		wantErr bool
	}{
		{
			desc:    "area falls outside of the rectangle",
			rect:    image.Rect(0, 0, 1, 1),
			area:    image.Rect(1, 1, 2, 2),
			hAlign:  HorizontalLeft,
			vAlign:  VerticalTop,
			wantErr: true,
		},
		{
			desc:    "unsupported horizontal alignment",
			rect:    image.Rect(0, 0, 2, 2),
			area:    image.Rect(0, 0, 1, 1),
			hAlign:  Horizontal(-1),
			vAlign:  VerticalTop,
			wantErr: true,
		},
		{
			desc:    "unsupported vertical alignment",
			rect:    image.Rect(0, 0, 2, 2),
			area:    image.Rect(0, 0, 1, 1),
			hAlign:  HorizontalLeft,
			vAlign:  Vertical(-1),
			wantErr: true,
		},
		{
			desc:   "nothing to align if the rectangles are equal",
			rect:   image.Rect(0, 0, 2, 2),
			area:   image.Rect(0, 0, 2, 2),
			hAlign: HorizontalLeft,
			vAlign: VerticalTop,
			want:   image.Rect(0, 0, 2, 2),
		},
		{
			desc:   "aligns top and left, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalLeft,
			vAlign: VerticalTop,
			want:   image.Rect(0, 0, 1, 1),
		},
		{
			desc:   "aligns top and center, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalCenter,
			vAlign: VerticalTop,
			want:   image.Rect(1, 0, 2, 1),
		},
		{
			desc:   "aligns top and right, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalRight,
			vAlign: VerticalTop,
			want:   image.Rect(2, 0, 3, 1),
		},
		{
			desc:   "aligns middle and left, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalLeft,
			vAlign: VerticalMiddle,
			want:   image.Rect(0, 1, 1, 2),
		},
		{
			desc:   "aligns middle and center, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalCenter,
			vAlign: VerticalMiddle,
			want:   image.Rect(1, 1, 2, 2),
		},
		{
			desc:   "aligns middle and right, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalRight,
			vAlign: VerticalMiddle,
			want:   image.Rect(2, 1, 3, 2),
		},
		{
			desc:   "aligns bottom and left, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalLeft,
			vAlign: VerticalBottom,
			want:   image.Rect(0, 2, 1, 3),
		},
		{
			desc:   "aligns bottom and center, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalCenter,
			vAlign: VerticalBottom,
			want:   image.Rect(1, 2, 2, 3),
		},
		{
			desc:   "aligns bottom and right, area is zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalRight,
			vAlign: VerticalBottom,
			want:   image.Rect(2, 2, 3, 3),
		},
		{
			desc:   "aligns top and left, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(0, 0, 1, 1),
			hAlign: HorizontalLeft,
			vAlign: VerticalTop,
			want:   image.Rect(0, 0, 1, 1),
		},
		{
			desc:   "aligns top and center, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(1, 1, 2, 2),
			hAlign: HorizontalCenter,
			vAlign: VerticalTop,
			want:   image.Rect(1, 0, 2, 1),
		},
		{
			desc:   "aligns top and right, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(1, 1, 2, 2),
			hAlign: HorizontalRight,
			vAlign: VerticalTop,
			want:   image.Rect(2, 0, 3, 1),
		},
		{
			desc:   "aligns middle and left, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(1, 1, 2, 2),
			hAlign: HorizontalLeft,
			vAlign: VerticalMiddle,
			want:   image.Rect(0, 1, 1, 2),
		},
		{
			desc:   "aligns middle and center, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(1, 1, 2, 2),
			hAlign: HorizontalCenter,
			vAlign: VerticalMiddle,
			want:   image.Rect(1, 1, 2, 2),
		},
		{
			desc:   "aligns middle and right, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(1, 1, 2, 2),
			hAlign: HorizontalRight,
			vAlign: VerticalMiddle,
			want:   image.Rect(2, 1, 3, 2),
		},
		{
			desc:   "aligns bottom and left, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(1, 1, 2, 2),
			hAlign: HorizontalLeft,
			vAlign: VerticalBottom,
			want:   image.Rect(0, 2, 1, 3),
		},
		{
			desc:   "aligns bottom and center, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(1, 1, 2, 2),
			hAlign: HorizontalCenter,
			vAlign: VerticalBottom,
			want:   image.Rect(1, 2, 2, 3),
		},
		{
			desc:   "aligns bottom and right, area isn't zero based",
			rect:   image.Rect(0, 0, 3, 3),
			area:   image.Rect(1, 1, 2, 2),
			hAlign: HorizontalRight,
			vAlign: VerticalBottom,
			want:   image.Rect(2, 2, 3, 3),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := Rectangle(tc.rect, tc.area, tc.hAlign, tc.vAlign)
			if (err != nil) != tc.wantErr {
				t.Errorf("Rectangle => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Rectangle => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
