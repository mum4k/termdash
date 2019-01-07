package axes

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

type updateY struct {
	minVal float64
	maxVal float64
}

func TestY(t *testing.T) {
	tests := []struct {
		desc           string
		minVal         float64
		maxVal         float64
		update         *updateY
		cvsHeight      int
		wantWidth      int
		want           *YDetails
		wantNewErr     bool
		wantUpdateErr  bool
		wantWidthErr   bool
		wantDetailsErr bool
	}{
		{
			desc:      "zero based positive ints",
			minVal:    0,
			maxVal:    3,
			cvsHeight: 4,
			wantWidth: 2,
			want: &YDetails{
				Width: 2,
				Labels: []*Label{
					{"0", image.Point{0, 3}},
					{"3", image.Point{0, 0}},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			y, err := NewY(tc.minVal, tc.maxVal)
			if (err != nil) != tc.wantNewErr {
				t.Errorf("NewY => unexpected error: %v, wantErr: %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			if tc.update != nil {
				err := y.Update(tc.update.minVal, tc.update.maxVal)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("Update => unexpected error: %v, wantErr: %v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}
			}

			gotWidth, err := y.RequiredWidth()
			if (err != nil) != tc.wantWidthErr {
				t.Errorf("RequiredWidth => unexpected error: %v, wantErr: %v", err, tc.wantWidthErr)
			}
			if err != nil {
				return
			}
			if gotWidth != tc.wantWidth {
				t.Errorf("RequiredWidth => got %v, want %v", gotWidth, tc.wantWidth)
			}

			got, err := y.Details(tc.cvsHeight)
			if (err != nil) != tc.wantDetailsErr {
				t.Errorf("Details => unexpected error: %v, wantErr: %v", err, tc.wantDetailsErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Details => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
