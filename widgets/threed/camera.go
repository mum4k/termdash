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

// Camera represents the viewer's perspective.
type Camera struct {
	Width     int         // Viewport width
	Height    int         // Viewport height
	Scale     float64     // Scale factor
	Direction Vector3D    // Forward direction the camera looks (world space)
	logger    *log.Logger // Logger for debugging
	Zoom      float64     // Zoom level (distance from camera to scene origin)
}

// NewCamera creates a new camera with default settings.
// The camera sits at z = -Zoom and looks in the +Z direction.
func NewCamera(logger *log.Logger) Camera {
	return Camera{
		Width:  80,
		Height: 24,
		Scale:  1.0,
		// Camera is at z = -Zoom, looking toward +Z.
		// Backface culling keeps faces where normal·Direction < 0,
		// i.e. normals that oppose the camera's forward vector.
		Direction: Vector3D{X: 0, Y: 0, Z: 1},
		logger:    logger,
		Zoom:      5.0,
	}
}

// Project projects a 3D point onto 2D screen space using perspective projection.
func (c *Camera) Project(v Vector3D) Vector2D {
	// Perspective projection parameters
	fov := 60.0 // Field of view in degrees
	aspectRatio := float64(c.Width) / float64(c.Height)
	fovRad := 1.0 / math.Tan(fov*0.5*math.Pi/180.0)

	// Camera is at z = -Zoom; adding Zoom gives depth from camera.
	z := v.Z + c.Zoom

	// Handle cases where z is too small or negative (behind camera)
	if z <= 0.1 {
		if c.logger != nil {
			c.logger.Printf("Warning: z=%.2f is too small or negative. Skipping projection.", z)
		}
		return Vector2D{X: math.NaN(), Y: math.NaN()}
	}

	x := (v.X * fovRad * aspectRatio) / z
	y := (v.Y * fovRad) / z

	// Scale, correct for tall terminal characters, and center
	x = (x * c.Scale) + float64(c.Width)/2
	y = (-y * c.Scale * terminalCellAspect) + float64(c.Height)/2 // invert Y; correct for char aspect

	return Vector2D{X: x, Y: y}
}

// AdjustScale adjusts the camera scale based on the model's dimensions.
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

	// Determine the required scale to fit the model within the canvas
	scaleX := float64(c.Width) / modelWidth
	scaleY := float64(c.Height) / modelHeight
	c.Scale = clampFloat(math.Min(scaleX, scaleY)*0.8, 5.0, 100.0) // Apply a padding factor and clamp

	if c.logger != nil {
		c.logger.Printf("Adjusted camera scale to %.2f", c.Scale)
	}
}
