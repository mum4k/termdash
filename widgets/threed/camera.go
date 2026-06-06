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

// threed/camera.go

package threed

import (
	"log"
	"math"
)

// terminalCellAspect corrects for terminal characters being ~2x taller than
// wide in pixels. Without this the cube appears vertically stretched.
const terminalCellAspect = 0.5

// fovDegrees is the fixed field-of-view used by all cameras.
const fovDegrees = 60.0

// fovRadFactor is 1/tan(fov/2) computed once at package init — fov never
// changes, so there is no reason to recompute it on every vertex projection.
var fovRadFactor = 1.0 / math.Tan(fovDegrees*0.5*math.Pi/180.0)

// Camera represents the viewer's perspective.
type Camera struct {
	Width     int         // Viewport width in cells
	Height    int         // Viewport height in cells
	Scale     float64     // Scale factor applied after projection
	Direction Vector3D    // Forward direction the camera looks (world space)
	logger    *log.Logger // Logger for debugging
	Zoom      float64     // Distance from camera to scene origin along Z

	// Cached per-frame values — recomputed by UpdateProjection when dimensions change.
	aspectRatio float64 // Width/Height; avoids the division inside the hot Project path
	halfW       float64 // float64(Width)/2  — screen centre X
	halfH       float64 // float64(Height)/2 — screen centre Y
}

// NewCamera creates a new camera with default settings.
// The camera sits at z = -Zoom and looks in the +Z direction.
func NewCamera(logger *log.Logger) Camera {
	c := Camera{
		Width:  80,
		Height: 24,
		Scale:  1.0,
		// Camera looks toward +Z; backface culling keeps faces whose normal
		// opposes this direction (normal·Direction < 0).
		Direction: Vector3D{X: 0, Y: 0, Z: 1},
		logger:    logger,
		Zoom:      5.0,
	}
	c.UpdateProjection()
	return c
}

// UpdateProjection caches the aspect ratio and screen-centre values that
// Project uses for every vertex. Call this once after changing Width or Height
// rather than recomputing inside the hot per-vertex path.
func (c *Camera) UpdateProjection() {
	if c.Height > 0 {
		c.aspectRatio = float64(c.Width) / float64(c.Height)
	}
	c.halfW = float64(c.Width) / 2
	c.halfH = float64(c.Height) / 2
}

// Project maps a 3D point into 2D screen space using perspective projection.
// The per-frame constants (fovRadFactor, aspectRatio, halfW, halfH) are
// pre-cached by UpdateProjection; only a few multiplications and one division
// are needed here.
func (c *Camera) Project(v Vector3D) Vector2D {
	// Camera sits at z = -Zoom; shift so depth is measured from the camera.
	z := v.Z + c.Zoom

	// Clip vertices that are at or behind the camera plane.
	if z <= 0.1 {
		if c.logger != nil {
			c.logger.Printf("Warning: z=%.2f is behind camera, skipping projection.", z)
		}
		return Vector2D{X: math.NaN(), Y: math.NaN()}
	}

	x := (v.X * fovRadFactor * c.aspectRatio) / z
	y := (v.Y * fovRadFactor) / z

	// Apply scale, correct for tall terminal glyphs, and translate to screen centre.
	x = (x * c.Scale) + c.halfW
	y = (-y * c.Scale * terminalCellAspect) + c.halfH

	return Vector2D{X: x, Y: y}
}

// AdjustScale sets the camera scale so the model fits within the viewport.
func (c *Camera) AdjustScale(model *Model) {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	for _, face := range model.Faces {
		for _, vertex := range face.Vertices {
			if vertex.X < minX {
				minX = vertex.X
			}
			if vertex.X > maxX {
				maxX = vertex.X
			}
			if vertex.Y < minY {
				minY = vertex.Y
			}
			if vertex.Y > maxY {
				maxY = vertex.Y
			}
		}
	}

	modelWidth := maxX - minX
	modelHeight := maxY - minY

	// Choose the tighter of the two axes and leave a small margin.
	scaleX := float64(c.Width) / modelWidth
	scaleY := float64(c.Height) / modelHeight
	c.Scale = clampFloat(math.Min(scaleX, scaleY)*0.8, 5.0, 100.0)

	if c.logger != nil {
		c.logger.Printf("Adjusted camera scale to %.2f", c.Scale)
	}
}
