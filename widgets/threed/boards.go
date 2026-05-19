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

package threed

import "strings"

type glyphColorFunc func(rune) Color

// TextBoard converts terminal text rows into centered 3D glyph billboards.
//
// Use it for labels, lightweight diagrams, game maps, or any other terminal
// native composition where the source is already readable UTF-8 text.
func TextBoard(rows []string, opts ...ModelOption) *Model {
	cfg := newModelOptions(opts...)
	return rowsToModel(rows, cfg, terminalBoardColor)
}

// LogicBoard converts circuit/logic-board text into a colored 3D model.
func LogicBoard(rows []string, opts ...ModelOption) *Model {
	cfg := newModelOptions(opts...)
	return rowsToModel(rows, cfg, logicBoardColor)
}

// GameBoard converts an ASCII/UTF-8 game map into a colored 3D model.
func GameBoard(rows []string, opts ...ModelOption) *Model {
	cfg := newModelOptions(opts...)
	return rowsToModel(rows, cfg, gameBoardColor)
}

func rowsToModel(rows []string, cfg modelOptions, colorFor glyphColorFunc) *Model {
	if cfg.centered {
		return createCenteredGlyphRows(cfg.position, rows, cfg.cellWidth, cfg.cellHeight, colorFor)
	}
	return createGlyphRows(cfg.position, rows, cfg.cellWidth, cfg.cellHeight, colorFor)
}

func createGlyphRows(origin Vector3D, rows []string, cellWidth, cellHeight float64, colorFor glyphColorFunc) *Model {
	model := NewModel()
	for row, line := range rows {
		col := 0
		for _, r := range line {
			if r != ' ' {
				addGlyphBillboard(model, Vector3D{
					X: origin.X + float64(col)*cellWidth,
					Y: origin.Y - float64(row)*cellHeight,
					Z: origin.Z,
				}, cellWidth*1.65, r, colorFor(r))
			}
			col++
		}
	}
	return model
}

func createCenteredGlyphRows(center Vector3D, rows []string, cellWidth, cellHeight float64, colorFor glyphColorFunc) *Model {
	maxCols := 0
	for _, row := range rows {
		if cols := len([]rune(row)); cols > maxCols {
			maxCols = cols
		}
	}
	origin := Vector3D{
		X: center.X - float64(maxCols-1)*cellWidth/2,
		Y: center.Y,
		Z: center.Z,
	}
	return createGlyphRows(origin, rows, cellWidth, cellHeight, colorFor)
}

func terminalBoardColor(r rune) Color {
	switch {
	case strings.ContainsRune("┌┐└┘─│├┤┬┴┼╭╮╰╯╦╩═║╔╗╚╝", r):
		return Color{R: 0.56, G: 0.92, B: 0.96}
	case strings.ContainsRune("█▇▆▅▄▃▂▁▣■▪", r):
		return Color{R: 0.78, G: 0.94, B: 1.00}
	case strings.ContainsRune("◆◇●○", r):
		return Color{R: 0.54, G: 0.96, B: 0.38}
	case strings.ContainsRune("▲△╱╲", r):
		return Color{R: 0.98, G: 0.78, B: 0.28}
	default:
		return Color{R: 0.86, G: 0.88, B: 0.90}
	}
}

func logicBoardColor(r rune) Color {
	switch {
	case strings.ContainsRune("┌┐└┘─│├┤┬┴┼╭╮╰╯╦╩═║╔╗╚╝", r):
		return Color{R: 0.48, G: 0.90, B: 0.94}
	case strings.ContainsRune("◆◇", r):
		return Color{R: 0.52, G: 0.98, B: 0.38}
	case strings.ContainsRune("○●", r):
		return Color{R: 0.96, G: 0.80, B: 0.28}
	case strings.ContainsRune("█▣■▪", r):
		return Color{R: 0.96, G: 0.36, B: 0.52}
	default:
		return Color{R: 0.84, G: 0.88, B: 0.90}
	}
}

func gameBoardColor(r rune) Color {
	switch r {
	case '#', '█', '▓', '▒', '░':
		return Color{R: 0.42, G: 0.72, B: 0.95}
	case '@', '☺', '☻', '●':
		return Color{R: 0.98, G: 0.88, B: 0.34}
	case '*', '✦', '✧', '+':
		return Color{R: 0.62, G: 0.98, B: 0.46}
	case '!', '▲', '▼', '◆':
		return Color{R: 0.98, G: 0.42, B: 0.58}
	default:
		return terminalBoardColor(r)
	}
}
