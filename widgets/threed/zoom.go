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

// threed/zoom.go

package threed

import "math"

const (
	// minZoomScale is the lower bound for ZoomHandler.Scale.
	minZoomScale = 5.0
	// maxZoomScale is the upper bound for ZoomHandler.Scale.
	maxZoomScale = 100.0
)

// ZoomHandler manages the zoom scale for the ThreeD widget.
//
// All methods are called with ThreeD.mu already held; no additional
// locking is needed here.
type ZoomHandler struct {
	Scale float64 // Current scale factor; default 20.0.
}

// NewZoomHandler creates a ZoomHandler with the default scale.
func NewZoomHandler() *ZoomHandler {
	return &ZoomHandler{Scale: 20.0}
}

// ZoomIn multiplies Scale by 1.1 and clamps to [minZoomScale, maxZoomScale].
func (z *ZoomHandler) ZoomIn() {
	z.Scale *= 1.1
	z.clampScale()
}

// ZoomOut multiplies Scale by 0.9 and clamps to [minZoomScale, maxZoomScale].
func (z *ZoomHandler) ZoomOut() {
	z.Scale *= 0.9
	z.clampScale()
}

// clampScale keeps Scale within the allowed range.
func (z *ZoomHandler) clampScale() {
	z.Scale = math.Max(minZoomScale, math.Min(maxZoomScale, z.Scale))
}
