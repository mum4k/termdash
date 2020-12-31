// Copyright 2021 Google Inc.
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

package heatmap

import (
	"reflect"
	"testing"
)

func Test_initLabels(t *testing.T) {
	type args struct {
		l int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "when the number of labels exceeds 10",
			args: args{
				l: 11,
			},
			want: []string{
				"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initLabels(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
