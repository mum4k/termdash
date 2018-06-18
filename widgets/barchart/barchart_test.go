package barchart

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/widgetapi"
)

func TestGauge(t *testing.T) {
	tests := []struct {
		desc          string
		bc            *BarChart
		update        func(*BarChart) error // update gets called before drawing of the widget.
		canvas        image.Rectangle
		want          func(size image.Point) *faketerm.Terminal
		wantUpdateErr bool // whether to expect an error on a call to the update function
		wantDrawErr   bool
	}{
		{
			desc: "displays bars",
			bc: New(
				Char('o'),
			),
			update: func(bc *BarChart) error {
				return bc.Values([]int{1, 2}, 10)
			},
			canvas: image.Rect(0, 0, 3, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				//testdraw.MustRectangle(c, image.Rect(0, 0, 3, 3),
				//	draw.RectChar('o'),
				//	draw.RectCellOpts(cell.BgColor(cell.ColorRed)),
				//)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			err = tc.update(tc.bc)
			if (err != nil) != tc.wantUpdateErr {
				t.Errorf("update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)

			}
			if err != nil {
				return
			}

			err = tc.bc.Draw(c)
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
				t.Errorf("Rectangle => %v", diff)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc   string
		create func() (*BarChart, error)
		want   widgetapi.Options
	}{
		{
			desc: "minimum size for no bars",
			create: func() (*BarChart, error) {
				return New(), nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for no bars, but have labels",
			create: func() (*BarChart, error) {
				return New(
					Labels([]string{"foo"}),
				), nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for one bar, default width, gap and no labels",
			create: func() (*BarChart, error) {
				bc := New()
				if err := bc.Values([]int{1}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for two bars, default width, gap and no labels",
			create: func() (*BarChart, error) {
				bc := New()
				if err := bc.Values([]int{1, 2}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{3, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for two bars, custom width, gap and no labels",
			create: func() (*BarChart, error) {
				bc := New(
					BarWidth(3),
					BarGap(2),
				)
				if err := bc.Values([]int{1, 2}, 3); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
		{
			desc: "minimum size for two bars, custom width, gap and labels",
			create: func() (*BarChart, error) {
				bc := New(
					BarWidth(3),
					BarGap(2),
				)
				if err := bc.Values([]int{1, 2}, 3, Labels([]string{"foo", "bar"})); err != nil {
					return nil, err
				}
				return bc, nil
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{8, 2},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bc, err := tc.create()
			if err != nil {
				t.Fatalf("create => unexpected error: %v", err)
			}

			got := bc.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}

		})
	}
}
