// Copyright 2026 Google Inc.
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

// Package threed renders simple 3D models in a Termdash widget.
//
// The public API is intentionally small:
//
//	stage, _ := threed.New(threed.ShowAxes(false), threed.UprightOnly(true))
//	stage.SetModel(threed.Cube(threed.ModelSize(2), threed.ModelColor(threed.NeonCyan)))
//
// Models can come from primitives, charts, terminal boards, game maps, UTF-8
// glyphs, images, KML, or custom faces. Higher-level helpers keep caller code
// short while the package owns projection, shading, glyph masks, and board
// construction details.
//
// Primitive shapes:
//
//	model := threed.Pyramid(threed.ModelSize(1.8), threed.ModelRune('▲'))
//
// Logic boards and game maps:
//
//	model := threed.LogicBoard([]string{
//		"╭──── CPU ────╮",
//		"│ ◆──◆──◆  █ │",
//		"╰────────────╯",
//	}, threed.ModelCellSize(0.06, 0.16))
//
// UTF-8 glyphs and images:
//
//	glyph := threed.Glyph("✦", threed.ModelSize(1.2))
//	imageModel, err := threed.ModelFromImageFile("logo.png")
//	kmlModel, err := threed.ModelFromKMLURL(ctx, "https://example.com/map.kml")
package threed
