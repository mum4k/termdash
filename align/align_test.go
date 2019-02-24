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

package align

import "testing"

func TestHorizontal(t *testing.T) {
	tests := []struct {
		desc  string
		align Horizontal
		want  string
	}{
		{
			desc:  "unknown",
			align: Horizontal(-1),
			want:  "HorizontalUnknown",
		},
		{
			desc:  "left",
			align: HorizontalLeft,
			want:  "HorizontalLeft",
		},
		{
			desc:  "center",
			align: HorizontalCenter,
			want:  "HorizontalCenter",
		},
		{
			desc:  "right",
			align: HorizontalRight,
			want:  "HorizontalRight",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.align.String(); got != tc.want {
				t.Errorf("String => %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVertical(t *testing.T) {
	tests := []struct {
		desc  string
		align Vertical
		want  string
	}{
		{
			desc:  "unknown",
			align: Vertical(-1),
			want:  "VerticalUnknown",
		},
		{
			desc:  "top",
			align: VerticalTop,
			want:  "VerticalTop",
		},
		{
			desc:  "middle",
			align: VerticalMiddle,
			want:  "VerticalMiddle",
		},
		{
			desc:  "bottom",
			align: VerticalBottom,
			want:  "VerticalBottom",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.align.String(); got != tc.want {
				t.Errorf("String => %q, want %q", got, tc.want)
			}
		})
	}
}
