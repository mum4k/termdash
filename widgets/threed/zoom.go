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
