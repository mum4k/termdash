package cell

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestNewOptions(t *testing.T) {
	tests := []struct {
		desc string
		opts []Option
		want *Options
	}{
		{
			desc: "no provided options",
			want: &Options{},
		},
		{
			desc: "setting foreground color",
			opts: []Option{
				FgColor(ColorBlack),
			},
			want: &Options{
				FgColor: ColorBlack,
			},
		},
		{
			desc: "setting background color",
			opts: []Option{
				BgColor(ColorRed),
			},
			want: &Options{
				BgColor: ColorRed,
			},
		},
		{
			desc: "setting multiple options",
			opts: []Option{
				FgColor(ColorCyan),
				BgColor(ColorMagenta),
			},
			want: &Options{
				FgColor: ColorCyan,
				BgColor: ColorMagenta,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := NewOptions(tc.opts...)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewOptions => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		desc string
		r    rune
		opts []Option
		want Cell
	}{
		{
			desc: "creates empty cell with default options",
			want: Cell{
				Opts: &Options{},
			},
		},
		{
			desc: "cell with the specified rune",
			r:    'X',
			want: Cell{
				Rune: 'X',
				Opts: &Options{},
			},
		},
		{
			desc: "cell with options",
			r:    'X',
			opts: []Option{
				FgColor(ColorCyan),
				BgColor(ColorMagenta),
			},
			want: Cell{
				Rune: 'X',
				Opts: &Options{
					FgColor: ColorCyan,
					BgColor: ColorMagenta,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := New(tc.r, tc.opts...)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("New => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestCellApply(t *testing.T) {
	tests := []struct {
		desc string
		cell *Cell
		opts []Option
		want *Cell
	}{
		{
			desc: "no options provided",
			cell: New(0),
			want: New(0),
		},
		{
			desc: "no change in options",
			cell: New(0, FgColor(ColorCyan)),
			opts: []Option{
				FgColor(ColorCyan),
			},
			want: New(0, FgColor(ColorCyan)),
		},
		{
			desc: "retains previous values",
			cell: New(0, FgColor(ColorCyan)),
			opts: []Option{
				BgColor(ColorBlack),
			},
			want: New(
				0,
				FgColor(ColorCyan),
				BgColor(ColorBlack),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.cell
			got.Apply(tc.opts...)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Apply => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNewBuffer(t *testing.T) {
	tests := []struct {
		desc    string
		size    image.Point
		want    Buffer
		wantErr bool
	}{
		{
			desc:    "zero buffer is invalid",
			wantErr: true,
		},
		{
			desc:    "width cannot be negative",
			size:    image.Point{-1, 1},
			wantErr: true,
		},
		{
			desc:    "height cannot be negative",
			size:    image.Point{1, -1},
			wantErr: true,
		},
		{
			desc: "creates single cell buffer",
			size: image.Point{1, 1},
			want: Buffer{
				{
					New(0),
				},
			},
		},
		{
			desc: "creates the buffer",
			size: image.Point{2, 3},
			want: Buffer{
				{
					New(0),
					New(0),
					New(0),
				},
				{
					New(0),
					New(0),
					New(0),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := NewBuffer(tc.size)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewBuffer => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewBuffer => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestBufferSize(t *testing.T) {
	sizes := []image.Point{
		{1, 1},
		{2, 3},
	}

	for _, size := range sizes {
		t.Run("", func(t *testing.T) {
			b, err := NewBuffer(size)
			if err != nil {
				t.Fatalf("NewBuffer => unexpected error: %v", err)
			}

			got := b.Size()
			if diff := pretty.Compare(size, got); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestBufferArea(t *testing.T) {
	tests := []struct {
		desc string
		size image.Point
		want image.Rectangle
	}{
		{
			desc: "single cell buffer",
			size: image.Point{1, 1},
			want: image.Rectangle{
				Min: image.Point{0, 0},
				Max: image.Point{0, 0},
			},
		},
		{
			desc: "rectangular buffer",
			size: image.Point{3, 4},
			want: image.Rectangle{
				Min: image.Point{0, 0},
				Max: image.Point{2, 3},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			b, err := NewBuffer(tc.size)
			if err != nil {
				t.Fatalf("NewBuffer => unexpected error: %v", err)
			}

			got := b.Area()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Area => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
