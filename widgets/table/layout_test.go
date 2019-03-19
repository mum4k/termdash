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

package table

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestContentLayout(t *testing.T) {
	tests := []struct {
		desc    string
		columns Columns
		rows    []*Row
		opts    []ContentOption
		cvsAr   image.Rectangle
		want    *contentLayout
		wantErr bool
	}{
		{
			desc:    "fails when canvas not wide enough for the columns",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("a"),
					NewCell("b"),
				),
			},
			cvsAr:   image.Rect(0, 0, 1, 1),
			wantErr: true,
		},
		{
			desc:    "success",
			columns: Columns(2),
			rows: []*Row{
				NewRow(
					NewCell("a"),
					NewCell("b"),
				),
			},
			cvsAr: image.Rect(0, 0, 2, 1),
			want: &contentLayout{
				lastCvsAr:    image.Rect(0, 0, 2, 1),
				columnWidths: []columnWidth{1, 1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			content, err := NewContent(tc.columns, tc.rows, tc.opts...)
			if err != nil {
				t.Fatalf("NewContent => unexpected error: %v", err)
			}

			got, err := newContentLayout(content, tc.cvsAr)
			if (err != nil) != tc.wantErr {
				t.Errorf("newContentLayout => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("newContentLayout => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
