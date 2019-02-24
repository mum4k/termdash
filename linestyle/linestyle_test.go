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

package linestyle

import "testing"

func TestLineStyleName(t *testing.T) {
	tests := []struct {
		desc string
		ls   LineStyle
		want string
	}{
		{
			desc: "unknown",
			ls:   LineStyle(-1),
			want: "LineStyleUnknown",
		},
		{
			desc: "none",
			ls:   None,
			want: "LineStyleNone",
		},
		{
			desc: "light",
			ls:   Light,
			want: "LineStyleLight",
		},
		{
			desc: "double",
			ls:   Double,
			want: "LineStyleDouble",
		},
		{
			desc: "round",
			ls:   Round,
			want: "LineStyleRound",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.ls.String(); got != tc.want {
				t.Errorf("String => %q, want %q", got, tc.want)
			}

		})
	}
}
