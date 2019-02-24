package linestyle

import "testing"

func TestLineStyleName(t *testing.T) {
	tests := []struct {
		desc string
		ls   LineStyle
		want string
	}{
		{
			desc: "unknown",
			ls:   LineStyle(-1),
			want: "LineStyleUnknown",
		},
		{
			desc: "none",
			ls:   None,
			want: "LineStyleNone",
		},
		{
			desc: "light",
			ls:   Light,
			want: "LineStyleLight",
		},
		{
			desc: "double",
			ls:   Double,
			want: "LineStyleDouble",
		},
		{
			desc: "round",
			ls:   Round,
			want: "LineStyleRound",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.ls.String(); got != tc.want {
				t.Errorf("String => %q, want %q", got, tc.want)
			}

		})
	}
}
