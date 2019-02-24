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

package draw

import (
	"image"
	"sort"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/internal/canvas"
)

func TestMultiEdgeNodes(t *testing.T) {
	tests := []struct {
		desc  string
		lines []HVLine
		want  []*hVLineNode
	}{
		{
			desc: "no lines added",
		},
		{
			desc: "single-edge nodes only",
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 1},
				},
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 1},
				},
			},
		},
		{
			desc: "lines don't cross",
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 2},
				},
			},
			want: []*hVLineNode{
				{
					p: image.Point{0, 1},
					edges: map[hVLineEdge]bool{
						newHVLineEdge(image.Point{0, 0}, image.Point{0, 1}): true,
						newHVLineEdge(image.Point{0, 1}, image.Point{0, 2}): true,
					},
				},
				{
					p: image.Point{1, 1},
					edges: map[hVLineEdge]bool{
						newHVLineEdge(image.Point{1, 0}, image.Point{1, 1}): true,
						newHVLineEdge(image.Point{1, 1}, image.Point{1, 2}): true,
					},
				},
			},
		},
		{
			desc: "lines cross, node has two edges",
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 1},
				},
				{
					Start: image.Point{0, 0},
					End:   image.Point{1, 0},
				},
			},
			want: []*hVLineNode{
				{
					p: image.Point{0, 0},
					edges: map[hVLineEdge]bool{
						newHVLineEdge(image.Point{0, 0}, image.Point{0, 1}): true,
						newHVLineEdge(image.Point{0, 0}, image.Point{1, 0}): true,
					},
				},
			},
		},
		{
			desc: "lines cross, node has three edges",
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
				{
					Start: image.Point{0, 1},
					End:   image.Point{1, 1},
				},
			},
			want: []*hVLineNode{
				{
					p: image.Point{0, 1},
					edges: map[hVLineEdge]bool{
						newHVLineEdge(image.Point{0, 0}, image.Point{0, 1}): true,
						newHVLineEdge(image.Point{0, 1}, image.Point{1, 1}): true,
						newHVLineEdge(image.Point{0, 1}, image.Point{0, 2}): true,
					},
				},
			},
		},
		{
			desc: "lines cross, node has four edges",
			lines: []HVLine{
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 2},
				},
				{
					Start: image.Point{0, 1},
					End:   image.Point{2, 1},
				},
			},
			want: []*hVLineNode{
				{
					p: image.Point{1, 1},
					edges: map[hVLineEdge]bool{
						newHVLineEdge(image.Point{1, 0}, image.Point{1, 1}): true,
						newHVLineEdge(image.Point{0, 1}, image.Point{1, 1}): true,
						newHVLineEdge(image.Point{1, 1}, image.Point{2, 1}): true,
						newHVLineEdge(image.Point{1, 1}, image.Point{1, 2}): true,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := canvas.New(image.Rect(0, 0, 3, 3))
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			g := newHVLineGraph()
			for i, l := range tc.lines {
				line, err := newHVLine(c, l.Start, l.End, newHVLineOptions())
				if err != nil {
					t.Fatalf("newHVLine[%d] => unexpected error: %v", i, err)
				}
				g.addLine(line)
			}

			got := g.multiEdgeNodes()

			lessFn := func(i, j int) bool {
				return got[i].p.X < got[j].p.X || got[i].p.Y < got[j].p.Y
			}
			sort.Slice(got, lessFn)
			sort.Slice(tc.want, lessFn)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("multiEdgeNodes => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}

}

func TestNodeRune(t *testing.T) {
	tests := []struct {
		desc    string
		node    *hVLineNode
		ls      LineStyle
		want    rune
		wantErr bool
	}{
		{
			desc:    "fails on node with no edges",
			node:    &hVLineNode{},
			wantErr: true,
		},
		{
			desc: "fails on unsupported two edge combination",
			node: &hVLineNode{
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{0, 0}, image.Point{1, 1}): true,
					newHVLineEdge(image.Point{1, 1}, image.Point{2, 2}): true,
				},
			},
			ls:      LineStyleLight,
			wantErr: true,
		},
		{
			desc: "fails on unsupported three edge combination",
			node: &hVLineNode{
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{0, 0}, image.Point{1, 1}): true,
					newHVLineEdge(image.Point{0, 0}, image.Point{0, 1}): true,
					newHVLineEdge(image.Point{1, 1}, image.Point{2, 2}): true,
				},
			},
			ls:      LineStyleLight,
			wantErr: true,
		},
		{
			desc:    "fails on unsupported line style",
			node:    &hVLineNode{},
			ls:      LineStyle(-1),
			wantErr: true,
		},
		{
			desc: "horizontal line",
			node: &hVLineNode{
				p: image.Point{1, 1},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{0, 1}, image.Point{1, 1}): true,
					newHVLineEdge(image.Point{1, 1}, image.Point{2, 1}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][hLine],
		},
		{
			desc: "vertical line",
			node: &hVLineNode{
				p: image.Point{1, 1},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{1, 0}, image.Point{1, 1}): true,
					newHVLineEdge(image.Point{1, 1}, image.Point{1, 2}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][vLine],
		},
		{
			desc: "top left corner",
			node: &hVLineNode{
				p: image.Point{0, 0},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{0, 0}, image.Point{1, 0}): true,
					newHVLineEdge(image.Point{0, 0}, image.Point{0, 1}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][topLeftCorner],
		},
		{
			desc: "top right corner",
			node: &hVLineNode{
				p: image.Point{2, 0},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{1, 0}, image.Point{2, 0}): true,
					newHVLineEdge(image.Point{2, 0}, image.Point{2, 1}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][topRightCorner],
		},
		{
			desc: "bottom left corner",
			node: &hVLineNode{
				p: image.Point{0, 2},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{0, 1}, image.Point{0, 2}): true,
					newHVLineEdge(image.Point{0, 2}, image.Point{1, 2}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][bottomLeftCorner],
		},
		{
			desc: "bottom right corner",
			node: &hVLineNode{
				p: image.Point{2, 2},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{1, 2}, image.Point{2, 2}): true,
					newHVLineEdge(image.Point{2, 1}, image.Point{2, 2}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][bottomRightCorner],
		},
		{
			desc: "T horizontal and up",
			node: &hVLineNode{
				p: image.Point{1, 2},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{1, 1}, image.Point{1, 2}): true,
					newHVLineEdge(image.Point{0, 2}, image.Point{1, 2}): true,
					newHVLineEdge(image.Point{1, 2}, image.Point{2, 2}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][hAndUp],
		},
		{
			desc: "T horizontal and down",
			node: &hVLineNode{
				p: image.Point{1, 0},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{0, 0}, image.Point{1, 0}): true,
					newHVLineEdge(image.Point{1, 0}, image.Point{2, 0}): true,
					newHVLineEdge(image.Point{1, 0}, image.Point{1, 1}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][hAndDown],
		},
		{
			desc: "T vertical and right",
			node: &hVLineNode{
				p: image.Point{0, 1},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{0, 0}, image.Point{0, 1}): true,
					newHVLineEdge(image.Point{0, 1}, image.Point{1, 1}): true,
					newHVLineEdge(image.Point{0, 1}, image.Point{0, 2}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][vAndRight],
		},
		{
			desc: "T vertical and left",
			node: &hVLineNode{
				p: image.Point{2, 1},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{2, 0}, image.Point{2, 1}): true,
					newHVLineEdge(image.Point{1, 1}, image.Point{2, 1}): true,
					newHVLineEdge(image.Point{2, 1}, image.Point{2, 2}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][vAndLeft],
		},
		{
			desc: "cross",
			node: &hVLineNode{
				p: image.Point{1, 1},
				edges: map[hVLineEdge]bool{
					newHVLineEdge(image.Point{1, 0}, image.Point{1, 1}): true,
					newHVLineEdge(image.Point{0, 1}, image.Point{1, 1}): true,
					newHVLineEdge(image.Point{1, 1}, image.Point{2, 1}): true,
					newHVLineEdge(image.Point{1, 1}, image.Point{1, 2}): true,
				},
			},
			ls:   LineStyleLight,
			want: lineStyleChars[LineStyleLight][vAndH],
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.node.rune(tc.ls)
			if (err != nil) != tc.wantErr {
				t.Errorf("rune => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if got != tc.want {
				t.Errorf("rune => got %c, want %c", got, tc.want)
			}
		})
	}
}
