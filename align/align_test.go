package align

import "testing"

func TestHorizontal(t *testing.T) {
	tests := []struct {
		desc  string
		align Horizontal
		want  string
	}{
		{
			desc:  "unknown",
			align: Horizontal(-1),
			want:  "HorizontalUnknown",
		},
		{
			desc:  "left",
			align: HorizontalLeft,
			want:  "HorizontalLeft",
		},
		{
			desc:  "center",
			align: HorizontalCenter,
			want:  "HorizontalCenter",
		},
		{
			desc:  "right",
			align: HorizontalRight,
			want:  "HorizontalRight",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.align.String(); got != tc.want {
				t.Errorf("String => %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVertical(t *testing.T) {
	tests := []struct {
		desc  string
		align Vertical
		want  string
	}{
		{
			desc:  "unknown",
			align: Vertical(-1),
			want:  "VerticalUnknown",
		},
		{
			desc:  "top",
			align: VerticalTop,
			want:  "VerticalTop",
		},
		{
			desc:  "middle",
			align: VerticalMiddle,
			want:  "VerticalMiddle",
		},
		{
			desc:  "bottom",
			align: VerticalBottom,
			want:  "VerticalBottom",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.align.String(); got != tc.want {
				t.Errorf("String => %q, want %q", got, tc.want)
			}
		})
	}
}
