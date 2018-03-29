package area

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestSize(t *testing.T) {
	tests := []struct {
		desc string
		area image.Rectangle
		want image.Point
	}{
		{
			desc: "zero area",
			area: image.Rect(0, 0, 0, 0),
			want: image.Point{0, 0},
		},
		{
			desc: "1-D on X axis",
			area: image.Rect(0, 0, 1, 0),
			want: image.Point{1, 0},
		},
		{
			desc: "1-D on Y axis",
			area: image.Rect(0, 0, 0, 1),
			want: image.Point{0, 1},
		},
		{
			desc: "area with a single cell",
			area: image.Rect(0, 0, 1, 1),
			want: image.Point{1, 1},
		},
		{
			desc: "a rectangle",
			area: image.Rect(0, 0, 2, 3),
			want: image.Point{2, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := Size(tc.area)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestFromSize(t *testing.T) {
	tests := []struct {
		desc    string
		size    image.Point
		want    image.Rectangle
		wantErr bool
	}{
		{
			desc:    "negative size on X axis",
			size:    image.Point{-1, 0},
			wantErr: true,
		},
		{
			desc:    "negative size on Y axis",
			size:    image.Point{0, -1},
			wantErr: true,
		},
		{
			desc: "zero size",
		},
		{
			desc: "1-D on X axis",
			size: image.Point{1, 0},
			want: image.Rect(0, 0, 1, 0),
		},
		{
			desc: "1-D on Y axis",
			size: image.Point{0, 1},
			want: image.Rect(0, 0, 0, 1),
		},
		{
			desc: "a rectangle",
			size: image.Point{2, 3},
			want: image.Rect(0, 0, 2, 3),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FromSize(tc.size)
			if (err != nil) != tc.wantErr {
				t.Fatalf("FromSize => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("FromSize => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
