package pie

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/canvas/braille/testbraille"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/draw/testdraw"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/widgetapi"
)

func TestPie(t *testing.T) {
	tests := []struct {
		desc          string
		values        []int
		colors        []cell.Color
		want          func(size image.Point) *faketerm.Terminal
		canvas        image.Rectangle
		wantDrawErr   bool
		wantNewErr    bool
		wantValuesErr bool
	}{
		{
			desc:       "New fails with no options",
			canvas:     image.Rect(0, 0, 5, 5),
			wantNewErr: false,
		},
		{
			desc:          "Values fails with empty values",
			values:        []int{},
			colors:        []cell.Color{cell.ColorRed},
			canvas:        image.Rect(0, 0, 5, 5),
			wantValuesErr: true,
		},
		{
			desc:          "Values fails with negative value",
			values:        []int{10, -5},
			colors:        []cell.Color{cell.ColorRed, cell.ColorBlue},
			canvas:        image.Rect(0, 0, 5, 5),
			wantValuesErr: true,
		},
		{
			desc:   "Draws pie chart with valid values and colors",
			values: []int{10, 20},
			colors: []cell.Color{cell.ColorRed, cell.ColorBlue},
			canvas: image.Rect(0, 0, 10, 10),
		},
		{
			desc:        "Fails to draw when canvas is too small",
			values:      []int{10, 20},
			colors:      []cell.Color{cell.ColorRed, cell.ColorBlue},
			canvas:      image.Rect(0, 0, 1, 1),
			wantDrawErr: true,
		},
		{
			desc:   "Draws a simple two-slice pie chart and verifies output",
			values: []int{1, 1},
			colors: []cell.Color{cell.ColorGreen, cell.ColorRed},
			canvas: image.Rect(0, 0, 10, 10),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				mid := image.Point{4, 4}

				testdraw.MustBrailleCircle(bc, mid, 4,
					draw.BrailleCircleArcOnly(90, 270),
					draw.BrailleCircleCellOpts(cell.FgColor(cell.ColorGreen)),
				)
				testdraw.MustBrailleCircle(bc, mid, 3,
					draw.BrailleCircleArcOnly(90, 270),
					draw.BrailleCircleCellOpts(cell.FgColor(cell.ColorGreen)),
				)

				// Semicerchio rosso a destra
				testdraw.MustBrailleCircle(bc, mid, 4,
					draw.BrailleCircleArcOnly(-90, 90),
					draw.BrailleCircleCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testdraw.MustBrailleCircle(bc, mid, 3,
					draw.BrailleCircleArcOnly(-90, 90),
					draw.BrailleCircleCellOpts(cell.FgColor(cell.ColorRed)),
				)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			p, err := New()
			if (err != nil) != tc.wantNewErr {
				t.Fatalf("New() error = %v, wantNewErr %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			if len(tc.values) > 0 {
				err = p.Values(tc.values)
				if (err != nil) != tc.wantValuesErr {
					t.Fatalf("Values() error = %v, wantValuesErr %v", err, tc.wantValuesErr)
				}
				if err != nil {
					return
				}
			}

			ft := faketerm.MustNew(tc.canvas.Size())
			cvs := testcanvas.MustNew(tc.canvas)
			meta := &widgetapi.Meta{}

			err = p.Draw(cvs, meta)
			if (err != nil) != tc.wantDrawErr {
				t.Fatalf("Draw() error = %v, wantDrawErr %v", err, tc.wantDrawErr)
			}

			if err == nil {
				testcanvas.MustApply(cvs, ft)
			}
		})
	}
}

func TestKeyboard(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = p.Keyboard(nil, nil)
	if err == nil {
		t.Fatalf("Keyboard() expected error, got nil")
	}
}

func TestMouse(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = p.Mouse(nil, nil)
	if err == nil {
		t.Fatalf("Mouse() expected error, got nil")
	}
}

func TestOptions(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	got := p.Options()
	want := widgetapi.Options{
		Ratio:        image.Point{braille.RowMult, braille.ColMult},
		MinimumSize:  image.Point{5, 5},
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}

	if got != want {
		t.Errorf("Options() = %v, want %v", got, want)
	}
}
