package text

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/terminal/faketerm"
)

func TestLineTrim(t *testing.T) {
	cvsArea := image.Rect(0, 0, 10, 1)
	tests := []struct {
		desc     string
		cvs      *canvas.Canvas
		curPoint image.Point
		curRune  rune
		opts     *options
		wantRes  *trimResult
		want     func(size image.Point) *faketerm.Terminal
		wantErr  bool
	}{
		{
			desc:     "half-width rune, beginning of the canvas",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{0, 0},
			curRune:  'A',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  false,
				curPoint: image.Point{0, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "half-width rune, end of the canvas, fits",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{9, 0},
			curRune:  'A',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  false,
				curPoint: image.Point{9, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "full-width rune, end of the canvas, fits",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{8, 0},
			curRune:  '世',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  false,
				curPoint: image.Point{8, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "half-width rune, falls out of the canvas, not configured to trim",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{10, 0},
			curRune:  'A',
			opts: &options{
				wrapAtRunes: true,
			},
			wantRes: &trimResult{
				trimmed:  false,
				curPoint: image.Point{10, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "half-width rune, first that falls out of the canvas, trimmed and marked",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{10, 0},
			curRune:  'A',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  true,
				curPoint: image.Point{11, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				testdraw.MustText(c, "…", image.Point{9, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:     "full-width rune, starts in and falls out, trimmed and marked",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{9, 0},
			curRune:  '世',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  true,
				curPoint: image.Point{11, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				testdraw.MustText(c, "…", image.Point{9, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:     "full-width rune, starts out, trimmed and marked",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{10, 0},
			curRune:  '世',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  true,
				curPoint: image.Point{12, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				testdraw.MustText(c, "…", image.Point{9, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:     "newline rune, first that falls out of the canvas, not trimmed or marked",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{10, 0},
			curRune:  '\n',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  false,
				curPoint: image.Point{10, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "half-width rune, n-th that falls out of the canvas, trimmed and not marked",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{11, 0},
			curRune:  'A',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  true,
				curPoint: image.Point{12, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "full-width rune, n-th that falls out of the canvas, trimmed and not marked",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{11, 0},
			curRune:  '世',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  true,
				curPoint: image.Point{13, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "newline rune, n-th that falls out of the canvas, not trimmed or marked",
			cvs:      testcanvas.MustNew(cvsArea),
			curPoint: image.Point{11, 0},
			curRune:  '\n',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  false,
				curPoint: image.Point{11, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "full-width rune, starts out, previous is also full, trimmed and marked",
			cvs: func() *canvas.Canvas {
				cvs := testcanvas.MustNew(cvsArea)
				testcanvas.MustSetCell(cvs, image.Point{8, 0}, '世')
				return cvs
			}(),
			curPoint: image.Point{10, 0},
			curRune:  '世',
			opts: &options{
				wrapAtRunes: false,
			},
			wantRes: &trimResult{
				trimmed:  true,
				curPoint: image.Point{12, 0},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				testdraw.MustText(c, "…", image.Point{9, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotRes, err := lineTrim(tc.cvs, tc.curPoint, tc.curRune, tc.opts)
			if (err != nil) != tc.wantErr {
				t.Errorf("lineTrim => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.wantRes, gotRes); diff != "" {
				t.Errorf("lineTrim => unexpected result, diff (-want, +got):\n%s", diff)
			}

			got, err := faketerm.New(tc.cvs.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := tc.cvs.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.cvs.Size()), got); diff != "" {
				t.Errorf("lineTrim => %v", diff)
			}
		})
	}
}
