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

package axes

import (
	"image"
	"reflect"
	"testing"
)

func TestLongestString(t *testing.T) {
	tests := []struct {
		desc    string
		strings []string
		want    int
	}{
		{
			desc: "all strings' length are different",
			strings: []string{
				"1",
				"12",
				"123",
				"1234",
				"12345",
			},
			want: 5,
		},
		{
			desc: "same strings",
			strings: []string{
				"123456789",
				"123456789",
				"123456789",
				"123456789",
				"123456789",
			},
			want: 9,
		},
		{
			desc: "different strings with same length",
			strings: []string{
				"123456789",
				"987654321",
				"aaaaaaaaa",
				"bbbbbbbbb",
				"ccccccccc",
			},
			want: 9,
		},
		{
			desc: "all strings are empty",
			strings: []string{
				"",
				"",
				"",
			},
			want: 0,
		},
		{
			desc:    "empty array",
			strings: []string{},
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := LongestString(tt.strings); got != tt.want {
				t.Errorf("LongestString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewXDetails(t *testing.T) {
	type args struct {
		yEnd      image.Point
		labels    []string
		cellWidth int
	}
	tests := []struct {
		desc    string
		args    args
		want    *XDetails
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := NewXDetails(tt.args.yEnd, tt.args.labels, tt.args.cellWidth)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewXDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewXDetails() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewYDetails(t *testing.T) {
	tests := []struct {
		desc    string
		labels  []string
		want    *YDetails
		wantErr bool
	}{
		// TODO：add test cases
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := NewYDetails(tt.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewYDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewYDetails() got = %v, want %v", got, tt.want)
			}
		})
	}
}
