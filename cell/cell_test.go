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

package cell

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestNewOptions(t *testing.T) {
	tests := []struct {
		desc string
		opts []Option
		want *Options
	}{
		{

			desc: "no provided options",
			want: &Options{},
		},
		{
			desc: "setting foreground color",
			opts: []Option{
				FgColor(ColorBlack),
			},
			want: &Options{
				FgColor: ColorBlack,
			},
		},
		{
			desc: "setting background color",
			opts: []Option{
				BgColor(ColorRed),
			},
			want: &Options{
				BgColor: ColorRed,
			},
		},
		{
			desc: "setting multiple options",
			opts: []Option{
				FgColor(ColorCyan),
				BgColor(ColorMagenta),
			},
			want: &Options{
				FgColor: ColorCyan,
				BgColor: ColorMagenta,
			},
		},
		{
			desc: "setting options by passing the options struct",
			opts: []Option{
				&Options{
					FgColor: ColorCyan,
					BgColor: ColorMagenta,
				},
			},
			want: &Options{
				FgColor: ColorCyan,
				BgColor: ColorMagenta,
			},
		},
		{
			desc: "setting font attributes",
			opts: []Option{
				Bold(),
				Italic(),
				Underline(),
				Strikethrough(),
				Inverse(),
				Blink(),
			},
			want: &Options{
				Bold:          true,
				Italic:        true,
				Underline:     true,
				Strikethrough: true,
				Inverse:       true,
				Blink:         true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := NewOptions(tc.opts...)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("NewOptions => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
