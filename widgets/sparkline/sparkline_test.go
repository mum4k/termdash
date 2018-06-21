package sparkline

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/widgetapi"
)

func TestSparkLine(t *testing.T) {
	tests := []struct {
		desc          string
		sparkLine     *SparkLine
		update        func(*SparkLine) error // update gets called before drawing of the widget.
		canvas        image.Rectangle
		want          func(size image.Point) *faketerm.Terminal
		wantUpdateErr bool // whether to expect an error on a call to the update function
		wantDrawErr   bool
	}{
		{
			desc:      "draws empty for no values",
			sparkLine: New(),
			update: func(bc *SparkLine) error {
				return nil
			},
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			err = tc.update(tc.sparkLine)
			if (err != nil) != tc.wantUpdateErr {
				t.Errorf("update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)

			}
			if err != nil {
				return
			}

			err = tc.sparkLine.Draw(c)
			if (err != nil) != tc.wantDrawErr {
				t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
			}
			if err != nil {
				return
			}

			got, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(c.Size()), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc      string
		sparkLine *SparkLine
		want      widgetapi.Options
	}{
		{
			desc:      "no label and no fixed height",
			sparkLine: New(),
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "label and no fixed height",
			sparkLine: New(
				Label("foo"),
			),
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 2},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "no label and fixed height",
			sparkLine: New(
				Height(3),
			),
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 3},
				MaximumSize:  image.Point{1, 3},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "label and fixed height",
			sparkLine: New(
				Label("foo"),
				Height(3),
			),
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 4},
				MaximumSize:  image.Point{1, 4},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.sparkLine.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}

		})
	}
}
