// Copyright 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package segmentdisplay

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/private/segdisp"
	"github.com/mum4k/termdash/private/segdisp/dotseg"
	"github.com/mum4k/termdash/private/segdisp/dotseg/testdotseg"
	"github.com/mum4k/termdash/private/segdisp/sixteen"
	"github.com/mum4k/termdash/private/segdisp/sixteen/testsixteen"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// mustDrawChar draws the provided character in the area of the canvas or panics.
func mustDrawChar(cvs *canvas.Canvas, char rune, ar image.Rectangle, cOpts ...cell.Option) {
	c := testcanvas.MustNew(ar)
	switch {
	case char == '.' || char == ':':
		d := dotseg.New()
		testdotseg.MustSetCharacter(d, char)
		testdotseg.MustDraw(d, c, dotseg.CellOpts(cOpts...))

	default:
		d := sixteen.New()
		testsixteen.MustSetCharacter(d, char)
		testsixteen.MustDraw(d, c, sixteen.CellOpts(cOpts...))
	}

	testcanvas.MustCopyTo(c, cvs)
}

func TestSegmentDisplay(t *testing.T) {
	tests := []struct {
		desc          string
		opts          []Option
		update        func(*SegmentDisplay) error // update gets called before drawing of the widget.
		canvas        image.Rectangle
		meta          *widgetapi.Meta
		want          func(size image.Point) *faketerm.Terminal
		wantCapacity  int
		wantNewErr    bool
		wantUpdateErr bool // whether to expect an error on a call to the update function
		wantDrawErr   bool
	}{
		{
			desc: "New fails on invalid GapPercent (too low)",
			opts: []Option{
				GapPercent(-1),
			},
			canvas:     image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			wantNewErr: true,
		},
		{
			desc: "New fails on invalid GapPercent (too high)",
			opts: []Option{
				GapPercent(101),
			},
			canvas:     image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			wantNewErr: true,
		},
		{
			desc:   "write fails on invalid GapPercent (too low)",
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write(
					[]*TextChunk{NewChunk("1")},
					GapPercent(-1),
				)
			},
			wantUpdateErr: true,
		},
		{
			desc:   "write fails on invalid GapPercent (too high)",
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write(
					[]*TextChunk{NewChunk("1")},
					GapPercent(101),
				)
			},
			wantUpdateErr: true,
		},
		{
			desc:   "fails on area too small for a segment",
			canvas: image.Rect(0, 0, segdisp.MinCols-1, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("1")})
			},
			wantDrawErr: true,
		},
		{
			desc:   "write fails without chunks",
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write(nil)
			},
			wantUpdateErr: true,
		},
		{
			desc:   "write fails with an empty chunk",
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("")})
			},
			wantUpdateErr: true,
		},
		{
			desc:   "write fails on unsupported characters when requested",
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("←", WriteErrOnUnsupported())})
			},
			wantUpdateErr: true,
		},
		{
			desc:         "draws empty without text",
			canvas:       image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			wantCapacity: 1,
		},
		{
			desc: "draws multiple segments, all fit exactly",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("123")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)},
					{'2', image.Rect(segdisp.MinCols, 0, segdisp.MinCols*2, segdisp.MinRows)},
					{'3', image.Rect(segdisp.MinCols*2, 0, segdisp.MinCols*3, segdisp.MinRows)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "uses the dot segment for a colon",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("1:3")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)},
					{':', image.Rect(segdisp.MinCols, 0, segdisp.MinCols*2, segdisp.MinRows)},
					{'3', image.Rect(segdisp.MinCols*2, 0, segdisp.MinCols*3, segdisp.MinRows)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "uses the dot segment for a dot",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("1.3")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)},
					{'.', image.Rect(segdisp.MinCols, 0, segdisp.MinCols*2, segdisp.MinRows)},
					{'3', image.Rect(segdisp.MinCols*2, 0, segdisp.MinCols*3, segdisp.MinRows)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "write sanitizes text by default",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("←1")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '1', image.Rect(segdisp.MinCols, 0, segdisp.MinCols*2, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 2,
		},
		{
			desc: "write sanitizes text with option",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("←1", WriteSanitize())})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '1', image.Rect(segdisp.MinCols, 0, segdisp.MinCols*2, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 2,
		},
		{
			desc:   "aligns segment vertical middle by default",
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("1")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '1', image.Rect(0, 1, segdisp.MinCols, segdisp.MinRows+1))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc:   "subsequent calls to write overwrite previous text",
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				if err := sd.Write([]*TextChunk{NewChunk("123")}); err != nil {
					return err
				}
				return sd.Write([]*TextChunk{NewChunk("4")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '4', image.Rect(0, 1, segdisp.MinCols, segdisp.MinRows+1))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "sets cell options per text chunk",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write(
					[]*TextChunk{
						NewChunk("1", WriteCellOpts(
							cell.FgColor(cell.ColorRed),
							cell.BgColor(cell.ColorBlue),
						)),
						NewChunk("2", WriteCellOpts(
							cell.FgColor(cell.ColorGreen),
							cell.BgColor(cell.ColorYellow),
						)),
					})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(
					cvs, '1',
					image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorBlue),
				)
				mustDrawChar(
					cvs, '2',
					image.Rect(segdisp.MinCols, 0, segdisp.MinCols*2, segdisp.MinRows),
					cell.FgColor(cell.ColorGreen),
					cell.BgColor(cell.ColorYellow),
				)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 2,
		},
		{
			desc: "reset resets the text content and reports capacity when maximizing fit and with gaps",
			opts: []Option{
				MaximizeDisplayedText(),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				if err := sd.Write([]*TextChunk{NewChunk("123")}); err != nil {
					return err
				}
				sd.Reset()
				return nil
			},
			wantCapacity: 2,
		},
		{
			desc: "reset resets the text content and reports capacity when maximizing fit and without gaps",
			opts: []Option{
				GapPercent(0),
				MaximizeDisplayedText(),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				if err := sd.Write([]*TextChunk{NewChunk("123")}); err != nil {
					return err
				}
				sd.Reset()
				return nil
			},
			wantCapacity: 3,
		},
		{
			desc: "reset resets the text content and reports capacity when maximizing segment height and gaps",
			opts: []Option{
				MaximizeSegmentHeight(),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				if err := sd.Write([]*TextChunk{NewChunk("123")}); err != nil {
					return err
				}
				sd.Reset()
				return nil
			},
			wantCapacity: 2,
		},
		{
			desc: "reset resets the text content and reports capacity when maximizing segment height and no gaps",
			opts: []Option{
				GapPercent(0),
				MaximizeSegmentHeight(),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				if err := sd.Write([]*TextChunk{NewChunk("123")}); err != nil {
					return err
				}
				sd.Reset()
				return nil
			},
			wantCapacity: 2,
		},
		{
			desc:   "reset resets provided cell options",
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				if err := sd.Write(
					[]*TextChunk{
						NewChunk("1", WriteCellOpts(
							cell.FgColor(cell.ColorRed),
							cell.BgColor(cell.ColorBlue),
						)),
					}); err != nil {
					return err
				}
				sd.Reset()
				return sd.Write([]*TextChunk{NewChunk("1")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "aligns segment vertical middle with option",
			opts: []Option{
				AlignVertical(align.VerticalMiddle),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("1")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '1', image.Rect(0, 1, segdisp.MinCols, segdisp.MinRows+1))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "aligns segment vertical top with option",
			opts: []Option{
				AlignVertical(align.VerticalTop),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("1")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "options given to Write override those given to New so aligns top",
			opts: []Option{
				AlignVertical(align.VerticalBottom),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				return sd.Write(
					[]*TextChunk{NewChunk("1")},
					AlignVertical(align.VerticalTop),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "aligns segment vertical bottom with option",
			opts: []Option{
				AlignVertical(align.VerticalBottom),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows+2),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("1")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '1', image.Rect(0, 2, segdisp.MinCols, segdisp.MinRows+2))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc:   "aligns segment horizontal center by default",
			canvas: image.Rect(0, 0, segdisp.MinCols+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("8")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '8', image.Rect(1, 0, segdisp.MinCols+1, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "aligns segment horizontal center with option",
			opts: []Option{
				AlignHorizontal(align.HorizontalCenter),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("8")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '8', image.Rect(1, 0, segdisp.MinCols+1, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "aligns segment horizontal left with option",
			opts: []Option{
				AlignHorizontal(align.HorizontalLeft),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("8")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '8', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "aligns segment horizontal right with option",
			opts: []Option{
				AlignHorizontal(align.HorizontalRight),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("8")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				mustDrawChar(cvs, '8', image.Rect(2, 0, segdisp.MinCols+2, segdisp.MinRows))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 1,
		},
		{
			desc: "draws multiple segments, not enough space, maximizes segment height with option",
			opts: []Option{
				MaximizeSegmentHeight(),
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("123")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)},
					{'2', image.Rect(segdisp.MinCols, 0, segdisp.MinCols*2, segdisp.MinRows)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 2,
		},
		{
			desc: "draws multiple segments, not enough space, maximizes displayed text by default and fits all",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows*4),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("123")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 7, 6, 12)},
					{'2', image.Rect(6, 7, 12, 12)},
					{'3', image.Rect(12, 7, 18, 12)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "draws multiple segments, not enough space, maximizes displayed text but cannot fit all",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows*4),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("1234")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 7, 6, 12)},
					{'2', image.Rect(6, 7, 12, 12)},
					{'3', image.Rect(12, 7, 18, 12)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "draws multiple segments, not enough space, maximizes displayed text with option",
			opts: []Option{
				MaximizeDisplayedText(),
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows*4),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("123")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 7, 6, 12)},
					{'2', image.Rect(6, 7, 12, 12)},
					{'3', image.Rect(12, 7, 18, 12)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc:   "draws multiple segments with a gap by default",
			canvas: image.Rect(0, 0, segdisp.MinCols*3+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("123")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)},
					{'2', image.Rect(segdisp.MinCols+1, 0, segdisp.MinCols*2+1, segdisp.MinRows)},
					{'3', image.Rect(segdisp.MinCols*2+2, 0, segdisp.MinCols*3+2, segdisp.MinRows)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "draws multiple segments with a gap, exact fit",
			opts: []Option{
				GapPercent(20),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("123")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)},
					{'2', image.Rect(segdisp.MinCols+1, 0, segdisp.MinCols*2+1, segdisp.MinRows)},
					{'3', image.Rect(segdisp.MinCols*2+2, 0, segdisp.MinCols*3+2, segdisp.MinRows)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "draws multiple segments with a larger gap",
			opts: []Option{
				GapPercent(40),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("123")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(3, 0, 9, 5)},
					{'2', image.Rect(11, 0, 17, 5)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 2,
		},
		{
			desc: "draws multiple segments with a gap, not all fit, maximizes displayed text",
			opts: []Option{
				GapPercent(20),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("8888")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'8', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)},
					{'8', image.Rect(segdisp.MinCols+1, 0, segdisp.MinCols*2+1, segdisp.MinRows)},
					{'8', image.Rect(segdisp.MinCols*2+2, 0, segdisp.MinCols*3+2, segdisp.MinRows)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "draws multiple segments with a gap, not all fit, last segment would fit without a gap",
			opts: []Option{
				GapPercent(20),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*4+2, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("8888")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'8', image.Rect(3, 0, 9, 5)},
					{'8', image.Rect(10, 0, 16, 5)},
					{'8', image.Rect(17, 0, 23, 5)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
		{
			desc: "draws multiple segments with a gap, not enough space, maximizes segment height with option",
			opts: []Option{
				MaximizeSegmentHeight(),
				GapPercent(20),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*5, segdisp.MinRows*2),
			update: func(sd *SegmentDisplay) error {
				return sd.Write([]*TextChunk{NewChunk("123")})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(2, 0, 14, 10)},
					{'2', image.Rect(16, 0, 28, 10)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 2,
		},
		{
			desc: "regression for #174, protects against external data mutation",
			opts: []Option{
				GapPercent(0),
			},
			canvas: image.Rect(0, 0, segdisp.MinCols*3, segdisp.MinRows),
			update: func(sd *SegmentDisplay) error {
				chunks := []*TextChunk{NewChunk("123")}
				if err := sd.Write(chunks); err != nil {
					return err
				}
				// Mutates and adds unsupported characters.
				chunks[0] = NewChunk("\r\t")
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, tc := range []struct {
					char rune
					area image.Rectangle
				}{
					{'1', image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)},
					{'2', image.Rect(segdisp.MinCols, 0, segdisp.MinCols*2, segdisp.MinRows)},
					{'3', image.Rect(segdisp.MinCols*2, 0, segdisp.MinCols*3, segdisp.MinRows)},
				} {
					mustDrawChar(cvs, tc.char, tc.area)
				}

				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCapacity: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			sd, err := New(tc.opts...)
			if (err != nil) != tc.wantNewErr {
				t.Errorf("New => unexpected error: %v, wantNewErr: %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			if tc.update != nil {
				err = tc.update(sd)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}
			}

			err = sd.Draw(c, tc.meta)
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

			var want *faketerm.Terminal
			if tc.want != nil {
				want = tc.want(c.Size())
			} else {
				want = faketerm.MustNew(c.Size())
			}

			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if gotCapacity := sd.Capacity(); gotCapacity != tc.wantCapacity {
				t.Errorf("Capacity => %d, want %d", gotCapacity, tc.wantCapacity)
			}
		})
	}
}

func TestKeyboard(t *testing.T) {
	sd, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := sd.Keyboard(&terminalapi.Keyboard{}); err == nil {
		t.Errorf("Keyboard => got nil err, wanted one")
	}
}

func TestMouse(t *testing.T) {
	sd, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := sd.Mouse(&terminalapi.Mouse{}); err == nil {
		t.Errorf("Mouse => got nil err, wanted one")
	}
}

func TestOptions(t *testing.T) {
	sd, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	got := sd.Options()
	want := widgetapi.Options{
		MinimumSize:  image.Point{segdisp.MinCols, segdisp.MinRows},
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
	}

}
