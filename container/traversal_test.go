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
	"errors"
	"image"
	"reflect"
	"testing"

	"github.com/mum4k/termdash/internal/cell"
	"github.com/mum4k/termdash/internal/faketerm"
)

func TestRoot(t *testing.T) {
	size := image.Point{4, 4}
	ft, err := faketerm.New(size)
	if err != nil {
		t.Fatalf("faketerm.New => unexpected error: %v", err)
	}
	want, err := New(
		ft,
		SplitHorizontal(
			Top(
				SplitHorizontal(
					Top(),
					Bottom(),
				),
			),
			Bottom(),
		),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	if got := rootCont(want); got != want {
		t.Errorf("rootCont(root) => got %p, want %p", got, want)
	}

	if got := rootCont(want.first.first); got != want {
		t.Errorf("rootCont(root.first.first) => got %p, want %p", got, want)
	}
}

func TestTraversal(t *testing.T) {
	size := image.Point{4, 4}
	ft, err := faketerm.New(size)
	if err != nil {
		t.Fatalf("faketerm.New => unexpected error: %v", err)
	}
	cont, err := New(
		ft,
		BorderColor(cell.ColorBlack),
		SplitVertical(
			Left(
				BorderColor(cell.ColorRed),
				SplitVertical(
					Left(
						BorderColor(cell.ColorYellow),
					),
					Right(
						BorderColor(cell.ColorBlue),
					),
				),
			),
			Right(
				BorderColor(cell.ColorGreen),
				SplitVertical(
					Left(
						BorderColor(cell.ColorMagenta),
					),
					Right(
						BorderColor(cell.ColorCyan),
					),
				),
			),
		),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	tests := []struct {
		desc       string
		travFunc   func(*Container, *string, visitFunc)
		visitErr   error
		wantColors []cell.Color
		wantErr    bool
	}{
		{
			desc:     "preOrder success",
			travFunc: preOrder,
			wantColors: []cell.Color{
				cell.ColorBlack,
				cell.ColorRed,
				cell.ColorYellow,
				cell.ColorBlue,
				cell.ColorGreen,
				cell.ColorMagenta,
				cell.ColorCyan,
			},
		},
		{
			desc:     "preOrder error",
			travFunc: preOrder,
			visitErr: errors.New("visit error"),
			wantErr:  true,
		},
		{
			desc:     "postOrder success",
			travFunc: postOrder,
			wantColors: []cell.Color{
				cell.ColorYellow,
				cell.ColorBlue,
				cell.ColorRed,
				cell.ColorMagenta,
				cell.ColorCyan,
				cell.ColorGreen,
				cell.ColorBlack,
			},
		},
		{
			desc:     "postOrder error",
			travFunc: postOrder,
			visitErr: errors.New("visit error"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			var (
				errStr    string
				gotColors []cell.Color
			)

			tc.travFunc(cont, &errStr, visitFunc(func(c *Container) error {
				gotColors = append(gotColors, c.opts.inherited.borderColor)
				return tc.visitErr
			}))

			if (errStr != "") != tc.wantErr {
				t.Fatalf("traversal => unexpected error: %v, wantErr: %v", errStr, tc.wantErr)
			}
			if errStr != "" {
				return
			}

			if !reflect.DeepEqual(gotColors, tc.wantColors) {
				t.Fatalf("traversal => unexpected order\n  got:%v\n want:%v", gotColors, tc.wantColors)
			}
		})
	}
}
