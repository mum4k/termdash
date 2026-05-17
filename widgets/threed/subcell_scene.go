package threed

import (
	"image"
	"math"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
)

// subcellScene stores filled polygon coverage at braille-pixel resolution so
// the final terminal output can preserve more detail than a one-cell-per-sample
// raster.
type subcellScene struct {
	cellArea image.Rectangle
	width    int
	height   int
	filled   []bool
	colors   []Color
}

// newSubcellScene allocates a new high-resolution scene buffer for the given
// terminal cell area.
func newSubcellScene(cellArea image.Rectangle) *subcellScene {
	width := cellArea.Dx() * braille.ColMult
	height := cellArea.Dy() * braille.RowMult
	return &subcellScene{
		cellArea: cellArea,
		width:    width,
		height:   height,
		filled:   make([]bool, width*height),
		colors:   make([]Color, width*height),
	}
}

// Clear resets the scene contents while retaining the allocated storage.
func (s *subcellScene) Clear() {
	for i := range s.filled {
		s.filled[i] = false
		s.colors[i] = Color{}
	}
}

// FillPolygon rasterizes the projected polygon into braille subcells.
func (s *subcellScene) FillPolygon(points []Vector2D, clr Color) {
	if len(points) < 3 || s.width == 0 || s.height == 0 {
		return
	}

	scaled := make([]Vector2D, len(points))
	for i, p := range points {
		scaled[i] = Vector2D{
			X: p.X * braille.ColMult,
			Y: p.Y * braille.RowMult,
		}
	}

	minX, maxX := scaled[0].X, scaled[0].X
	minY, maxY := scaled[0].Y, scaled[0].Y
	for _, p := range scaled[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	startX := maxInt(0, int(math.Floor(minX)))
	endX := minInt(s.width-1, int(math.Ceil(maxX)))
	startY := maxInt(0, int(math.Floor(minY)))
	endY := minInt(s.height-1, int(math.Ceil(maxY)))

	for x := startX; x <= endX; x++ {
		for y := startY; y <= endY; y++ {
			if !pointInPolygon(float64(x)+0.5, float64(y)+0.5, scaled) {
				continue
			}
			idx := y*s.width + x
			s.filled[idx] = true
			s.colors[idx] = clr
		}
	}
}

// CopyTo converts the subcell scene into braille characters and copies them to
// the destination canvas.
func (s *subcellScene) CopyTo(dst *canvas.Canvas) error {
	if s.width == 0 || s.height == 0 {
		return nil
	}

	bc, err := braille.New(s.cellArea)
	if err != nil {
		return err
	}

	for cellY := 0; cellY < s.cellArea.Dy(); cellY++ {
		for cellX := 0; cellX < s.cellArea.Dx(); cellX++ {
			var colorSum Color
			var lit int
			for py := 0; py < braille.RowMult; py++ {
				for px := 0; px < braille.ColMult; px++ {
					x := cellX*braille.ColMult + px
					y := cellY*braille.RowMult + py
					if x < 0 || x >= s.width || y < 0 || y >= s.height {
						continue
					}
					idx := y*s.width + x
					if !s.filled[idx] {
						continue
					}
					colorSum = colorSum.Add(s.colors[idx])
					lit++
				}
			}
			if lit == 0 {
				continue
			}

			avgColor := colorSum.Multiply(1 / float64(lit))
			opts := []cell.Option{cell.FgColor(avgColor.ToCellColor())}
			for py := 0; py < braille.RowMult; py++ {
				for px := 0; px < braille.ColMult; px++ {
					x := cellX*braille.ColMult + px
					y := cellY*braille.RowMult + py
					idx := y*s.width + x
					if !s.filled[idx] {
						continue
					}
					if err := bc.SetPixel(image.Point{X: x, Y: y}, opts...); err != nil {
						return err
					}
				}
			}
		}
	}

	return bc.CopyTo(dst)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
