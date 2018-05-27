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

package text

import (
	"testing"
)

func TestScrollTrackerNoContentRolling(t *testing.T) {
	tests := []struct {
		desc   string
		lines  int
		height int
		events func(*scrollTracker)
		want   int
	}{
		{
			desc:   "starts from the first line",
			lines:  2,
			height: 1,
			want:   0,
		},
		{
			desc:   "user can scroll down by a line",
			lines:  2,
			height: 1,
			events: func(st *scrollTracker) {
				st.downOneLine()
			},
			want: 1,
		},
		{
			desc:   "scroll down capped at the last line",
			lines:  2,
			height: 1,
			events: func(st *scrollTracker) {
				st.downOneLine()
				st.downOneLine()
			},
			want: 1,
		},
		{
			desc:   "larger terminal, scroll down capped at the last line",
			lines:  4,
			height: 2,
			events: func(st *scrollTracker) {
				st.downOneLine()
				st.downOneLine()
				st.downOneLine()
				st.downOneLine()
				st.downOneLine()
			},
			want: 2,
		},
		{
			desc:   "scroll up capped at the first line",
			lines:  2,
			height: 1,
			events: func(st *scrollTracker) {
				st.upOneLine()
				st.upOneLine()
			},
			want: 0,
		},
		{
			desc:   "processes multiple scroll events",
			lines:  4,
			height: 2,
			events: func(st *scrollTracker) {
				st.downOneLine()
				st.downOneLine()
				st.upOneLine()
			},
			want: 1,
		},
		{
			desc:   "scrolling down ignored when all content fits",
			lines:  2,
			height: 4,
			events: func(st *scrollTracker) {
				st.downOneLine()
				st.downOneLine()
			},
			want: 0,
		},
		{
			desc:   "scrolls down by a page",
			lines:  6,
			height: 2,
			events: func(st *scrollTracker) {
				st.downOnePage()
				st.downOnePage()
			},
			want: 4,
		},
		{
			desc:   "scrolling down by a page capped at the last line",
			lines:  6,
			height: 2,
			events: func(st *scrollTracker) {
				st.downOnePage()
				st.downOnePage()
				st.downOnePage()
				st.downOnePage()
			},
			want: 4,
		},
		{
			desc:   "scrolling up by a page capped at the first line",
			lines:  6,
			height: 2,
			events: func(st *scrollTracker) {
				st.downOnePage()
				st.upOnePage()
				st.upOnePage()
				st.upOnePage()
				st.upOnePage()
			},
			want: 0,
		},
		{
			desc:   "scrolling by lines and pages can be combined",
			lines:  8,
			height: 2,
			events: func(st *scrollTracker) {
				st.downOnePage() // first == 2
				st.upOneLine()   // first = 1
				st.downOneLine() // first = 2
				st.downOneLine() // first = 3
				st.downOneLine() // first = 4
				st.downOneLine() // first = 5
				st.upOnePage()   // first == 3
			},
			want: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			st := newScrollTracker(&options{})
			if tc.events != nil {
				tc.events(st)
			}
			got := st.firstLine(tc.lines, tc.height)
			if got != tc.want {
				t.Errorf("firstLine => got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestScrollTrackerContentRolling(t *testing.T) {
	st := newScrollTracker(&options{rollContent: true})
	// All of these test cases act on the same instance of the scroll tracker.
	tests := []struct {
		desc   string
		lines  int
		height int
		events func()
		want   int
	}{
		{
			desc:   "all content fits, draws from the first line",
			lines:  2,
			height: 2,
			want:   0,
		},
		{
			desc:   "content doesn't fit, draws up to the last line",
			lines:  4,
			height: 2,
			want:   2,
		},
		{
			desc:   "draws up to the last line when height decreases",
			lines:  4,
			height: 1,
			want:   3,
		},
		{
			desc:   "draws up to the last line when height increases",
			lines:  4,
			height: 2,
			want:   2,
		},
		{
			desc:   "user scrolling breaks away from the last line",
			lines:  4,
			height: 2,
			events: func() {
				st.upOneLine()
			},
			want: 1,
		},
		{
			desc:   "keeps scrolled to position when new content arrives",
			lines:  5,
			height: 2,
			want:   1,
		},
		{
			desc:   "scrolling down to the last line displays the latest line",
			lines:  5,
			height: 2,
			events: func() {
				st.downOneLine()
				st.downOneLine()
				st.downOneLine()
			},
			want: 3,
		},
		{
			desc:   "rolling of new content resumes",
			lines:  6,
			height: 2,
			want:   4,
		},
		{
			desc:   "scroll up breaks away from the last line again",
			lines:  6,
			height: 2,
			events: func() {
				st.upOneLine()
			},
			want: 3,
		},
		{
			desc:   "keeps scrolled to position when new content arrives again",
			lines:  7,
			height: 2,
			want:   3,
		},
		{
			desc:   "resize so that the last line becomes visible",
			lines:  7,
			height: 7,
			want:   0,
		},
		{
			desc:   "rolls content after the resize",
			lines:  8,
			height: 7,
			want:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.events != nil {
				tc.events()
			}
			got := st.firstLine(tc.lines, tc.height)
			if got != tc.want {
				t.Errorf("firstLine => got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestNormalizeScroll(t *testing.T) {
	tests := []struct {
		desc   string
		first  int
		lines  int
		height int
		want   int
	}{
		{
			desc:   "first line is negative",
			first:  -1,
			lines:  3,
			height: 1,
			want:   0,
		},
		{
			desc:   "no lines to be printed",
			first:  0,
			lines:  0,
			height: 1,
			want:   0,
		},
		{
			desc:   "zero height",
			first:  0,
			lines:  1,
			height: 0,
			want:   0,
		},
		{
			desc:   "first line is greater than the number of lines",
			first:  4,
			lines:  3,
			height: 2,
			want:   1,
		},
		{
			desc:   "first line reset to start if the full content fits",
			first:  1,
			lines:  3,
			height: 3,
			want:   0,
		},
		{
			desc:   "valid first line",
			first:  2,
			lines:  4,
			height: 2,
			want:   2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := normalizeScroll(tc.first, tc.lines, tc.height)
			if got != tc.want {
				t.Errorf("normalizeScroll => got %d, want %d", got, tc.want)
			}
		})
	}
}
