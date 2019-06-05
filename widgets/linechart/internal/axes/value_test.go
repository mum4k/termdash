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

package axes

import (
	"fmt"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestValue(t *testing.T) {
	formatter := func(float64) string { return "test" }

	tests := []struct {
		desc            string
		float           float64
		nonZeroDecimals int
		formatter       func(float64) string
		want            *Value
	}{
		{
			desc:            "handles zeroes",
			float:           0,
			nonZeroDecimals: 0,
			want: &Value{
				Value:           0,
				Rounded:         0,
				ZeroDecimals:    0,
				NonZeroDecimals: 0,
			},
		},
		{
			desc:            "rounds to requested precision",
			float:           1.01234,
			nonZeroDecimals: 2,
			want: &Value{
				Value:           1.01234,
				Rounded:         1.013,
				ZeroDecimals:    1,
				NonZeroDecimals: 2,
			},
		},
		{
			desc:            "no rounding when not requested",
			float:           1.01234,
			nonZeroDecimals: 0,
			want: &Value{
				Value:           1.01234,
				Rounded:         1.01234,
				ZeroDecimals:    1,
				NonZeroDecimals: 0,
			},
		},
		{
			desc:            "formatter value when value formatter as option",
			float:           1.01234,
			nonZeroDecimals: 0,
			formatter:       formatter,
			want: &Value{
				Value:           1.01234,
				Rounded:         1.01234,
				ZeroDecimals:    1,
				NonZeroDecimals: 0,
				formatter:       formatter,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := NewValue(tc.float, tc.nonZeroDecimals, ValueFormatter(tc.formatter))
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewValue => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestText(t *testing.T) {
	tests := []struct {
		value           float64
		nonZeroDecimals int
		wantRounded     float64
		wantText        string
	}{
		{0, 2, 0, "0"},
		{10, 2, 10, "10"},
		{-10, 2, -10, "-10"},
		{0.5, 2, 0.5, "0.50"},
		{-0.5, 2, -0.5, "-0.50"},
		{100.5, 2, 100.5, "100.50"},
		{-100.5, 2, -100.5, "-100.50"},
		{0.12345, 1, 0.2, "0.2"},
		{0.12345, 2, 0.13, "0.13"},
		{0.123, 4, 0.123, "0.1230"},
		{-0.12345, 2, -0.12, "-0.12"},
		{999.12345, 2, 999.13, "999.13"},
		{-999.12345, 2, -999.12, "-999.12"},
		{999.00012345, 2, 999.00013, "999.00013"},
		{-999.00012345, 2, -999.00012, "-999.00012"},
		{100000.1, 2, 100000.1, "100000.10"},
		{1000000.1, 2, 1000000.1, "1.00e+06"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v_%v", tc.value, tc.nonZeroDecimals), func(t *testing.T) {
			v := NewValue(tc.value, tc.nonZeroDecimals)
			gotRounded := v.Rounded
			if gotRounded != tc.wantRounded {
				t.Errorf("newValue(%v, %v).Rounded => got %v, want %v", tc.value, tc.nonZeroDecimals, gotRounded, tc.wantRounded)
			}

			gotText := v.Text()
			if gotText != tc.wantText {
				t.Errorf("newValue(%v, %v).Text => got %q, want %q", tc.value, tc.nonZeroDecimals, gotText, tc.wantText)
			}

		})
	}
}

func TestNewTextValue(t *testing.T) {
	const want = "foo"
	v := NewTextValue(want)
	got := v.Text()
	if got != want {
		t.Errorf("v.Text => got %q, want %q", got, want)
	}
}
