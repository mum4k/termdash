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

package dotseg

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestAttributes(t *testing.T) {
	tests := []struct {
		desc      string
		brailleAr image.Rectangle
		seg       Segment
		want      image.Rectangle
		wantErr   bool
	}{
		{
			desc:      "fails on unsupported segment",
			brailleAr: image.Rect(0, 0, 1, 1),
			seg:       Segment(-1),
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			attr := newAttributes(tc.brailleAr)
			got, err := attr.segArea(tc.seg)
			if (err != nil) != tc.wantErr {
				t.Errorf("segArea => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("segArea => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
