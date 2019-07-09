// Copyright 2018 Google Inc.
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

package container

import (
	"fmt"
	"image"
	"sync"
	"testing"
	"time"

	"termdash/align"
	"termdash/cell"
	"termdash/internal/canvas/testcanvas"
	"termdash/internal/draw"
	"termdash/internal/draw/testdraw"
	"termdash/internal/event"
	"termdash/internal/event/testevent"
	"termdash/internal/faketerm"
	"termdash/internal/fakewidget"
	"termdash/keyboard"
	"termdash/linestyle"
	"termdash/mouse"
	"termdash/terminal/terminalapi"
	"termdash/widgetapi"
	"termdash/widgets/barchart"
)

// Example demonstrates how to use the Container API.
func Example() {
	bc, err := barchart.New()
	if err != nil {
		panic(err)
	}
	if _, err := New(
		/* terminal = */ nil,
		SplitVertical(
			Left(
				SplitHorizontal(
					Top(
						Border(linestyle.Light),
					),
					Bottom(
						SplitHorizontal(
							Top(
								Border(linestyle.Light),
							),
							Bottom(
								Border(linestyle.Light),
							),
						),
					),
					SplitPercent(30),
				),
			),
			Right(
				Border(linestyle.Light),
				PlaceWidget(bc),
			),
		),
	); err != nil {
		panic(err)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		desc             string
		termSize         image.Point
		container        func(ft *faketerm.Terminal) (*Container, error)
		wantContainerErr bool
		want             func(size image.Point) *faketerm.Terminal
	}{
		{
			desc:     "fails on MarginTop too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginTop(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on invalid option on the first vertical child container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							MarginTop(-1),
						),
						Right(),
					),
				)
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on invalid option on the second vertical child container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(),
						Right(
							MarginTop(-1),
						),
					),
				)
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on invalid option on the first horizontal child container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							MarginTop(-1),
						),
						Bottom(),
					),
				)
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on invalid option on the second horizontal child container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(),
						Bottom(
							MarginTop(-1),
						),
					),
				)
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginTopPercent too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginTopPercent(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginTopPercent too high",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginTopPercent(101))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both MarginTop and MarginTopPercent specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginTop(1), MarginTopPercent(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both MarginTopPercent and MarginTop specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginTopPercent(1), MarginTop(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginRight too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginRight(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginRightPercent too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginRightPercent(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginRightPercent too high",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginRightPercent(101))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both MarginRight and MarginRightPercent specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginRight(1), MarginRightPercent(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both MarginRightPercent and MarginRight specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginRightPercent(1), MarginRight(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginBottom too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginBottom(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginBottomPercent too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginBottomPercent(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginBottomPercent too high",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginBottomPercent(101))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both MarginBottom and MarginBottomPercent specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginBottom(1), MarginBottomPercent(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both MarginBottomPercent and MarginBottom specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginBottomPercent(1), MarginBottom(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginLeft too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginLeft(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginLeftPercent too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginLeftPercent(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on MarginLeftPercent too high",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginLeftPercent(101))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both MarginLeft and MarginLeftPercent specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginLeft(1), MarginLeftPercent(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both MarginLeftPercent and MarginLeft specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, MarginLeftPercent(1), MarginLeft(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingTop too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingTop(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingTopPercent too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingTopPercent(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingTopPercent too high",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingTopPercent(101))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both PaddingTop and PaddingTopPercent specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingTop(1), PaddingTopPercent(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both PaddingTopPercent and PaddingTop specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingTopPercent(1), PaddingTop(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingRight too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingRight(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingRightPercent too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingRightPercent(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingRightPercent too high",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingRightPercent(101))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both PaddingRight and PaddingRightPercent specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingRight(1), PaddingRightPercent(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both PaddingRightPercent and PaddingRight specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingRightPercent(1), PaddingRight(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingBottom too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingBottom(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingBottomPercent too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingBottomPercent(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingBottomPercent too high",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingBottomPercent(101))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both PaddingBottom and PaddingBottomPercent specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingBottom(1), PaddingBottomPercent(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both PaddingBottomPercent and PaddingBottom specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingBottomPercent(1), PaddingBottom(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingLeft too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingLeft(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingLeftPercent too low",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingLeftPercent(-1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on PaddingLeftPercent too high",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingLeftPercent(101))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both PaddingLeft and PaddingLeftPercent specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingLeft(1), PaddingLeftPercent(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both PaddingLeftPercent and PaddingLeft specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, PaddingLeftPercent(1), PaddingLeft(1))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on empty ID specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft, ID(""))
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on empty duplicate ID specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("0"),
					SplitHorizontal(
						Top(ID("1")),
						Bottom(ID("1")),
					),
				)
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails when both SplitFixed and SplitPercent are specified",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(linestyle.Light),
						),
						Bottom(
							Border(linestyle.Light),
						),
						SplitFixed(4),
						SplitPercent(20),
					),
				)
			},
			wantContainerErr: true,
		},
		{
			desc:     "fails on SplitFixed less than -1",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(linestyle.Light),
						),
						Bottom(
							Border(linestyle.Light),
						),
						SplitFixed(-2),
					),
				)
			},
			wantContainerErr: true,
		},
		{
			desc:     "empty container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "container with a border",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal split, children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(linestyle.Light),
						),
						Bottom(
							Border(linestyle.Light),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 5))
				testdraw.MustBorder(cvs, image.Rect(0, 5, 10, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "fails on horizontal split too small",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(linestyle.Light),
						),
						Bottom(
							Border(linestyle.Light),
						),
						SplitPercent(0),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "fails on horizontal split too large",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(linestyle.Light),
						),
						Bottom(
							Border(linestyle.Light),
						),
						SplitPercent(100),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "horizontal unequal split",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(linestyle.Light),
						),
						Bottom(
							Border(linestyle.Light),
						),
						SplitPercent(20),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 4))
				testdraw.MustBorder(cvs, image.Rect(0, 4, 10, 20))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal unequal split",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(linestyle.Light),
						),
						Bottom(
							Border(linestyle.Light),
						),
						SplitFixed(4),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 4))
				testdraw.MustBorder(cvs, image.Rect(0, 4, 10, 20))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal split, parent and children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					SplitHorizontal(
						Top(
							Border(linestyle.Light),
						),
						Bottom(
							Border(linestyle.Light),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 9, 5))
				testdraw.MustBorder(cvs, image.Rect(1, 5, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical split, children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 5, 10))
				testdraw.MustBorder(cvs, image.Rect(5, 0, 10, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "fails on vertical split too small",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
						SplitPercent(0),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "fails on vertical split too large",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
						SplitPercent(100),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "vertical unequal split",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
						SplitPercent(20),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 4, 10))
				testdraw.MustBorder(cvs, image.Rect(4, 0, 20, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical fixed splits",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
						SplitFixed(4),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 4, 10))
				testdraw.MustBorder(cvs, image.Rect(4, 0, 20, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical split, parent and children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 5, 9))
				testdraw.MustBorder(cvs, image.Rect(5, 1, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "multi level split",
			termSize: image.Point{10, 16},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							SplitHorizontal(
								Top(
									Border(linestyle.Light),
								),
								Bottom(
									SplitHorizontal(
										Top(
											Border(linestyle.Light),
										),
										Bottom(
											Border(linestyle.Light),
										),
									),
								),
							),
						),
						Right(
							Border(linestyle.Light),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 5, 8))
				testdraw.MustBorder(cvs, image.Rect(0, 8, 5, 12))
				testdraw.MustBorder(cvs, image.Rect(0, 12, 5, 16))
				testdraw.MustBorder(cvs, image.Rect(5, 0, 10, 16))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "inherits border and focused color",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					BorderColor(cell.ColorRed),
					FocusedColor(cell.ColorBlue),
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorBlue)),
				)
				testdraw.MustBorder(
					cvs,
					image.Rect(1, 1, 5, 9),
					draw.BorderCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testdraw.MustBorder(
					cvs,
					image.Rect(5, 1, 9, 9),
					draw.BorderCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "splitting a container removes the widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 5, 9))
				testdraw.MustBorder(cvs, image.Rect(5, 1, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "placing a widget removes container split",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(linestyle.Light),
						),
						Right(
							Border(linestyle.Light),
						),
					),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				fakewidget.MustDraw(
					ft,
					cvs,
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{},
				)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			cont, err := tc.container(got)
			if (err != nil) != tc.wantContainerErr {
				t.Errorf("tc.container => unexpected error:%v, wantErr:%v", err, tc.wantContainerErr)
			}
			if err != nil {
				return
			}
			contStr := cont.String()
			t.Logf("For container: %v", contStr)
			if err := cont.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			var want *faketerm.Terminal
			if tc.want != nil {
				want = tc.want(tc.termSize)
			} else {
				w, err := faketerm.New(tc.termSize)
				if err != nil {
					t.Fatalf("faketerm.New => unexpected error: %v", err)
				}
				want = w
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}

}

// errorHandler just stores the last error received.
type errorHandler struct {
	err error
	mu  sync.Mutex
}

func (eh *errorHandler) get() error {
	eh.mu.Lock()
	defer eh.mu.Unlock()
	return eh.err
}

func (eh *errorHandler) handle(err error) {
	eh.mu.Lock()
	defer eh.mu.Unlock()
	eh.err = err
}

func TestKeyboard(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) (*Container, error)
		events    []terminalapi.Event
		// If specified, waits for this number of events.
		// Otherwise waits for len(events).
		wantProcessed int
		want          func(size image.Point) *faketerm.Terminal
		wantErr       bool
	}{
		{
			desc:     "event not forwarded if container has no widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "event forwarded to focused container only",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
						),
						Right(
							SplitHorizontal(
								Top(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
								),
							),
						),
					),
				)
			},
			events: []terminalapi.Event{
				// Move focus to the target container.
				&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonRelease},
				// Send the keyboard event.
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				// Widgets that aren't focused don't get the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 20, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 0, 40, 10)),
					&widgetapi.Meta{},
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
				)

				// The focused widget receives the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 10, 40, 20)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				return ft
			},
		},
		{
			desc:     "event forwarded to all widgets that requested global key scope",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeGlobal})),
						),
						Right(
							SplitHorizontal(
								Top(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
								),
							),
						),
					),
				)
			},
			events: []terminalapi.Event{
				// Move focus to the target container.
				&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonRelease},
				// Send the keyboard event.
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				// Widget that isn't focused, but registered for global
				// keyboard events.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 20, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeGlobal},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)

				// Widget that isn't focused and only wants focused events.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 0, 40, 10)),
					&widgetapi.Meta{},
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
				)

				// The focused widget receives the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 10, 40, 20)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if the widget didn't request it",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeNone})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "widget returns an error when processing the event",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEsc},
			},
			wantProcessed: 2, // The error is also an event.
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{},
				)
				return ft
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			c, err := tc.container(got)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			eh := &errorHandler{}
			// Subscribe to receive errors.
			eds.Subscribe([]terminalapi.Event{terminalapi.NewError("")}, func(ev terminalapi.Event) {
				eh.handle(ev.(*terminalapi.Error).Error())
			})

			c.Subscribe(eds)
			// Initial draw to determine sizes of containers.
			if err := c.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}
			for _, ev := range tc.events {
				eds.Event(ev)
			}
			var wantEv int
			if tc.wantProcessed != 0 {
				wantEv = tc.wantProcessed
			} else {
				wantEv = len(tc.events)
			}

			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), wantEv; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			if err := c.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.termSize), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if err := eh.get(); (err != nil) != tc.wantErr {
				t.Errorf("errorHandler => unexpected error %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestMouse(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) (*Container, error)
		events    []terminalapi.Event
		// If specified, waits for this number of events.
		// Otherwise waits for len(events).
		wantProcessed int
		want          func(size image.Point) *faketerm.Terminal
		wantErr       bool
	}{
		{
			desc:     "mouse click outside of the terminal is ignored",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{10, 10}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if container has no widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "event forwarded to container at that point",
			termSize: image.Point{50, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
						),
						Right(
							SplitHorizontal(
								Top(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
								),
							),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{49, 9}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{49, 9}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				// Widgets that aren't targeted don't get the mouse clicks.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 25, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(25, 10, 50, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Keyboard{},
				)

				// The target widget receives the mouse event.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(25, 0, 50, 10)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{24, 9}, Button: mouse.ButtonLeft},
					&terminalapi.Mouse{Position: image.Point{24, 9}, Button: mouse.ButtonRelease},
				)
				return ft
			},
		},
		{
			desc:     "event focuses the target container after terminal resize (falls onto the new area), regression for #169",
			termSize: image.Point{50, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				// Decrease the terminal size, so when container is created, it
				// only sees width of 30.
				if err := ft.Resize(image.Point{30, 20}); err != nil {
					return nil, err
				}
				c, err := New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
						),
						Right(
							SplitHorizontal(
								Top(
									Border(linestyle.Light),
									PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
								),
							),
						),
					),
				)
				if err != nil {
					return nil, err
				}
				// Increase the width back to 50 so the mouse clicks land on the "new" area.
				if err := ft.Resize(image.Point{50, 20}); err != nil {
					return nil, err
				}
				// Draw once so the container has a chance to update the tracked area.
				if err := c.Draw(); err != nil {
					return nil, err
				}
				return c, nil
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{48, 8}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{48, 8}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				// The yellow border signifies that the container was focused.
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(25, 0, 50, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)

				// Widgets that aren't targeted don't get the mouse clicks.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 25, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(25, 10, 50, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Keyboard{},
				)

				// The target widget receives the mouse event.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(26, 1, 49, 9)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{22, 7}, Button: mouse.ButtonLeft},
					&terminalapi.Mouse{Position: image.Point{22, 7}, Button: mouse.ButtonRelease},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if the widget didn't request it",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeNone})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeWidget, event not forwarded if it falls on the container's border",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					PlaceWidget(
						fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget}),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(1, 1, 19, 19)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeContainer, event forwarded if it falls on the container's border",
			termSize: image.Point{21, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					PlaceWidget(
						fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeContainer}),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(1, 1, 20, 19)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeGlobal, event forwarded if it falls on the container's border",
			termSize: image.Point{21, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(linestyle.Light),
					PlaceWidget(
						fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeGlobal}),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(1, 1, 20, 19)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeWidget event not forwarded if it falls outside of widget's canvas",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeWidget,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeContainer event forwarded if it falls outside of widget's canvas",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeContainer,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeGlobal event forwarded if it falls outside of widget's canvas",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeGlobal,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeWidget event not forwarded if it falls to another container",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(),
						Bottom(
							PlaceWidget(
								fakewidget.New(widgetapi.Options{
									WantMouse: widgetapi.MouseScopeWidget,
									Ratio:     image.Point{2, 1},
								}),
							),
							AlignVertical(align.VerticalMiddle),
							AlignHorizontal(align.HorizontalCenter),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 10, 20, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeContainer event not forwarded if it falls to another container",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(),
						Bottom(
							PlaceWidget(
								fakewidget.New(widgetapi.Options{
									WantMouse: widgetapi.MouseScopeContainer,
									Ratio:     image.Point{2, 1},
								}),
							),
							AlignVertical(align.VerticalMiddle),
							AlignHorizontal(align.HorizontalCenter),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 10, 20, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "MouseScopeGlobal event forwarded if it falls to another container",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(),
						Bottom(
							PlaceWidget(
								fakewidget.New(widgetapi.Options{
									WantMouse: widgetapi.MouseScopeGlobal,
									Ratio:     image.Point{2, 1},
								}),
							),
							AlignVertical(align.VerticalMiddle),
							AlignHorizontal(align.HorizontalCenter),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 10, 20, 20)),
					&widgetapi.Meta{},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "mouse position adjusted relative to widget's canvas, vertical offset",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeWidget,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 5}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "mouse position adjusted relative to widget's canvas, horizontal offset",
			termSize: image.Point{30, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: widgetapi.MouseScopeWidget,
							Ratio:     image.Point{9, 10},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{6, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(6, 0, 24, 20)),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "widget returns an error when processing the event",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRight},
			},
			wantProcessed: 2, // The error is also an event.
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{},
				)
				return ft
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			c, err := tc.container(got)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			eh := &errorHandler{}
			// Subscribe to receive errors.
			eds.Subscribe([]terminalapi.Event{terminalapi.NewError("")}, func(ev terminalapi.Event) {
				eh.handle(ev.(*terminalapi.Error).Error())
			})
			c.Subscribe(eds)
			// Initial draw to determine sizes of containers.
			if err := c.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}
			for _, ev := range tc.events {
				eds.Event(ev)
			}
			var wantEv int
			if tc.wantProcessed != 0 {
				wantEv = tc.wantProcessed
			} else {
				wantEv = len(tc.events)
			}

			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), wantEv; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			if err := c.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.termSize), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if err := eh.get(); (err != nil) != tc.wantErr {
				t.Errorf("errorHandler => unexpected error %v, wantErr: %v", err, tc.wantErr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		desc       string
		termSize   image.Point
		container  func(ft *faketerm.Terminal) (*Container, error)
		updateID   string
		updateOpts []Option
		// events are events delivered before the update.
		beforeEvents []terminalapi.Event
		// events are events delivered after the update.
		afterEvents   []terminalapi.Event
		wantUpdateErr bool
		want          func(size image.Point) *faketerm.Terminal
	}{
		{
			desc:     "fails on empty updateID",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			wantUpdateErr: true,
		},
		{
			desc:     "fails when no container with the ID is found",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			updateID:      "myID",
			wantUpdateErr: true,
		},
		{
			desc:     "no changes when no options are provided",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myID"),
					Border(linestyle.Light),
				)
			},
			updateID: "myID",
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "fails on invalid options",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myID"),
					Border(linestyle.Light),
				)
			},
			updateID: "myID",
			updateOpts: []Option{
				MarginTop(-1),
			},
			wantUpdateErr: true,
		},
		{
			desc:     "fails when update introduces a duplicate ID",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myID"),
					Border(linestyle.Light),
				)
			},
			updateID: "myID",
			updateOpts: []Option{
				SplitVertical(
					Left(
						ID("left"),
					),
					Right(
						ID("myID"),
					),
				),
			},
			wantUpdateErr: true,
		},
		{
			desc:     "removes border from the container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myID"),
					Border(linestyle.Light),
				)
			},
			updateID: "myID",
			updateOpts: []Option{
				Border(linestyle.None),
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "places widget into a sub-container container",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myRoot"),
					SplitVertical(
						Left(
							ID("left"),
						),
						Right(
							ID("right"),
						),
					),
				)
			},
			updateID: "right",
			updateOpts: []Option{
				PlaceWidget(fakewidget.New(widgetapi.Options{})),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				wAr := image.Rect(10, 0, 20, 10)
				wCvs := testcanvas.MustNew(wAr)
				fakewidget.MustDraw(ft, wCvs, &widgetapi.Meta{}, widgetapi.Options{})
				testcanvas.MustCopyTo(wCvs, cvs)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "places widget into root which removes children",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myRoot"),
					SplitVertical(
						Left(
							ID("left"),
							Border(linestyle.Light),
						),
						Right(
							ID("right"),
							Border(linestyle.Light),
						),
					),
				)
			},
			updateID: "myRoot",
			updateOpts: []Option{
				PlaceWidget(fakewidget.New(widgetapi.Options{})),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				fakewidget.MustDraw(ft, cvs, &widgetapi.Meta{Focused: true}, widgetapi.Options{})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "changes container splits",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myRoot"),
					SplitVertical(
						Left(
							ID("left"),
							Border(linestyle.Light),
						),
						Right(
							ID("right"),
							Border(linestyle.Light),
						),
					),
				)
			},
			updateID: "myRoot",
			updateOpts: []Option{
				SplitHorizontal(
					Top(
						ID("left"),
						Border(linestyle.Light),
					),
					Bottom(
						ID("right"),
						Border(linestyle.Light),
					),
				),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 5))
				testdraw.MustBorder(cvs, image.Rect(0, 5, 10, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "update retains original focused container if it still exists",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myRoot"),
					SplitVertical(
						Left(
							ID("left"),
							Border(linestyle.Light),
						),
						Right(
							ID("right"),
							Border(linestyle.Light),
							SplitHorizontal(
								Top(
									ID("rightTop"),
									Border(linestyle.Light),
								),
								Bottom(
									ID("rightBottom"),
									Border(linestyle.Light),
								),
							),
						),
					),
				)
			},
			beforeEvents: []terminalapi.Event{
				// Move focus to container with ID "right".
				// It will continue to exist after the update.
				&terminalapi.Mouse{Position: image.Point{5, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{5, 0}, Button: mouse.ButtonRelease},
			},
			updateID: "right",
			updateOpts: []Option{
				Clear(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 5, 10))
				testdraw.MustBorder(cvs, image.Rect(5, 0, 10, 10), draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "update moves focus to the nearest parent when focused container is destroyed",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myRoot"),
					SplitVertical(
						Left(
							ID("left"),
							Border(linestyle.Light),
						),
						Right(
							ID("right"),
							Border(linestyle.Light),
							SplitHorizontal(
								Top(
									ID("rightTop"),
									Border(linestyle.Light),
								),
								Bottom(
									ID("rightBottom"),
									Border(linestyle.Light),
								),
							),
						),
					),
				)
			},
			beforeEvents: []terminalapi.Event{
				// Move focus to container with ID "rightTop".
				// It will be destroyed by calling update.
				&terminalapi.Mouse{Position: image.Point{6, 1}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{6, 1}, Button: mouse.ButtonRelease},
			},
			updateID: "right",
			updateOpts: []Option{
				Clear(),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 5, 10))
				testdraw.MustBorder(cvs, image.Rect(5, 0, 10, 10), draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "newly placed widget gets keyboard events",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myRoot"),
					PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
				)
			},
			beforeEvents: []terminalapi.Event{
				// Move focus to the target container.
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
			},
			afterEvents: []terminalapi.Event{
				// Send the keyboard event.
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			updateID: "myRoot",
			updateOpts: []Option{
				PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused})),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				fakewidget.MustDraw(
					ft,
					cvs,
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantKeyboard: widgetapi.KeyScopeFocused},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "newly placed widget gets mouse events",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					ID("myRoot"),
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
				)
			},
			afterEvents: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
			},
			updateID: "myRoot",
			updateOpts: []Option{
				PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget})),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				fakewidget.MustDraw(
					ft,
					cvs,
					&widgetapi.Meta{Focused: true},
					widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			cont, err := tc.container(got)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			eh := &errorHandler{}
			// Subscribe to receive errors.
			eds.Subscribe([]terminalapi.Event{terminalapi.NewError("")}, func(ev terminalapi.Event) {
				eh.handle(ev.(*terminalapi.Error).Error())
			})
			cont.Subscribe(eds)
			// Initial draw to determine sizes of containers.
			if err := cont.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			// Deliver the before events.
			for _, ev := range tc.beforeEvents {
				eds.Event(ev)
			}
			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), len(tc.beforeEvents); got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			{
				err := cont.Update(tc.updateID, tc.updateOpts...)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("Update => unexpected error:%v, wantErr:%v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}
			}

			// Deliver the after events.
			for _, ev := range tc.afterEvents {
				eds.Event(ev)
			}
			wantEv := len(tc.beforeEvents) + len(tc.afterEvents)
			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), wantEv; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
			}

			if err := cont.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			var want *faketerm.Terminal
			if tc.want != nil {
				want = tc.want(tc.termSize)
			} else {
				w, err := faketerm.New(tc.termSize)
				if err != nil {
					t.Fatalf("faketerm.New => unexpected error: %v", err)
				}
				want = w
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}

}
