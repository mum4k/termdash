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

package termbox

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/internal/terminalapi"
)

func TestNewTerminal(t *testing.T) {
	tests := []struct {
		desc string
		opts []Option
		want *Terminal
	}{
		{
			desc: "default options",
			want: &Terminal{
				colorMode: terminalapi.ColorMode256,
			},
		},
		{
			desc: "sets color mode",
			opts: []Option{
				ColorMode(terminalapi.ColorModeNormal),
			},
			want: &Terminal{
				colorMode: terminalapi.ColorModeNormal,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := newTerminal(tc.opts...)

			// Ignore these fields.
			got.events = nil
			got.done = nil

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("newTerminal => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
