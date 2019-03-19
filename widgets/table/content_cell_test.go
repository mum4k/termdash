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
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestWrapToWidth(t *testing.T) {
	tests := []struct {
		desc    string
		data    string
		opts    []CellOption
		width   columnWidth
		want    []*Data
		wantErr bool
	}{
		{
			desc:  "no wrapping without content",
			width: 1,
		},
		{
			desc:  "single line when wrapping disabled",
			data:  "hello world",
			width: 5,
			want: []*Data{
				NewData("hello world"),
			},
		},
		{
			desc: "wraps at words",
			data: "hello world",
			opts: []CellOption{
				CellWrapAtWords(),
			},
			width: 5,
			want: []*Data{
				NewData("hello"),
				NewData("world"),
			},
		},
		{
			desc: "wrapping fails on unsupported control runes",
			data: "hello\tworld",
			opts: []CellOption{
				CellWrapAtWords(),
			},
			width:   5,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c := NewCellWithOpts([]*Data{NewData(tc.data)}, tc.opts...)
			err := c.wrapToWidth(tc.width)
			if (err != nil) != tc.wantErr {
				t.Errorf("c.wrapToWidth => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			got := c.wrapped
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("c.wrapToWidth =>\n  got:%v\n want:%v\n diff (-want, +got):\n%s", got, tc.want, diff)
			}
		})
	}
}
