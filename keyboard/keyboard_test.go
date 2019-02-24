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

package keyboard

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		desc string
		key  Key
		want string
	}{
		{
			desc: "unknown",
			key:  Key(-1000),
			want: "KeyUnknown",
		},
		{
			desc: "defined value",
			key:  KeyEnter,
			want: "KeyEnter",
		},
		{
			desc: "standard key",
			key:  'a',
			want: "a",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.key.String(); got != tc.want {
				t.Errorf("String => %q, want %q", got, tc.want)
			}
		})
	}
}
