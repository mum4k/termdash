// threed/threed.go

package threed

import (
	"image"
	"io"
	"log"
	"math"
	"os"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// ThreeD is a custom Termdash widget that renders enhanced 3D objects.
type ThreeD struct {
	mu           sync.Mutex      // Protects concurrent access
	model        *Model          // The 3D model to render
	rotation     Vector3D        // Rotation angles around X, Y, Z axes
	options      *Options        // Widget options
	camera       Camera          // Camera parameters
	light        Vector3D        // Light direction for shading
	projected    []ProjectedFace // Projected faces for rendering
	doubleBuffer *canvas.Canvas  // Off-screen canvas for double buffering
	zoomHandler  *ZoomHandler    // Handles zooming
	logger       *log.Logger     // Logger for debugging

	// Per-frame reuse buffers — zero allocations after warmup.
	fillScene    *subcellScene // Reused braille fill scene; re-cleared each frame.
	transformBuf []Vector3D    // Scratch space for per-face rotated vertices.
	pointsBuf    []Vector2D    // Backing store for all ProjectedFace.Points slices.
	pointsIdx    int           // Write cursor into pointsBuf for the current frame.
}

// New creates a new ThreeD widget with enhanced rendering.
func New(opts ...Option) (*ThreeD, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt.set(options)
	}

	// Initialize the logger.
	var logger *log.Logger
	if options.EnableLogging {
		file, err := os.OpenFile("threed_demo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("failed to open log file: %v", err)
			logger = log.New(os.Stdout, "", log.LstdFlags)
		} else {
			logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	} else {
		// Discard logs if logging is disabled.
		logger = log.New(io.Discard, "", 0)
	}

	zoomHandler := NewZoomHandler()
	if options.ZoomScale > 0 {
		zoomHandler.Scale = options.ZoomScale
		zoomHandler.clampScale()
	}

	widget := &ThreeD{
		model:    NewModel(),
		rotation: Vector3D{},
		options:  options,
		camera:   NewCamera(logger),
		// Light at (1, 2, -3): positioned above, slightly right, and toward the
		// viewer.  This gives each visible cube face a distinct brightness:
		//   front  (~0.80 diffuse) > top (~0.53) > right (~0.27)
		// compared to the old (1,1,-1) which gave every face the same 0.577.
		light:       Vector3D{X: 1, Y: 2, Z: -3}.Normalize(),
		zoomHandler: zoomHandler, // Initialize the zoom handler.
		logger:      logger,      // Set the logger.
	}
	return widget, nil
}

// Draw renders the widget onto the canvas with advanced shading and colors.
func (t *ThreeD) Draw(cvs *canvas.Canvas, _ *widgetapi.Meta) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	width := cvs.Area().Dx()
	height := cvs.Area().Dy()

	// Initialize double buffering.
	if t.doubleBuffer == nil || t.doubleBuffer.Area().Dx() != width || t.doubleBuffer.Area().Dy() != height {
		var err error
		t.doubleBuffer, err = canvas.New(image.Rect(0, 0, width, height))
		if err != nil {
			return err
		}
	}

	// Clear the off-screen canvas.
	if err := t.doubleBuffer.Clear(); err != nil {
		return err
	}

	// Update camera dimensions and base scale.
	t.camera.Width = width
	t.camera.Height = height

	// Base scale from canvas size.
	baseScale := float64(height) / 2.5

	// Zoom factor from the ZoomHandler.
	// 20.0 is the "neutral" zoom (see zoom.go).
	zoomFactor := t.zoomHandler.Scale / 20.0

	t.camera.Scale = baseScale * zoomFactor

	// Build a rotation matrix once per frame: 6 trig calls instead of 6 per vertex.
	rotMat := BuildRotationMatrix(t.rotation)

	// Pre-size the points backing buffer so the face loop needs no allocations.
	totalVerts := 0
	for i := range t.model.Faces {
		totalVerts += len(t.model.Faces[i].Vertices)
	}
	if cap(t.pointsBuf) < totalVerts {
		t.pointsBuf = make([]Vector2D, totalVerts)
	} else {
		t.pointsBuf = t.pointsBuf[:totalVerts]
	}
	t.pointsIdx = 0

	// Apply transformations and project vertices.
	t.projected = t.projected[:0]
	for _, face := range t.model.Faces {
		n := len(face.Vertices)

		// Grow the transform scratch buffer only when needed; reslice otherwise.
		if cap(t.transformBuf) < n {
			t.transformBuf = make([]Vector3D, n)
		} else {
			t.transformBuf = t.transformBuf[:n]
		}

		// Transform vertices using the pre-built rotation matrix.
		for i, vertex := range face.Vertices {
			t.transformBuf[i] = rotMat.Apply(vertex)
		}

		// Rotate the pre-computed model-space normal into world space.
		// Rotation matrices are orthogonal so length is preserved — no sqrt needed.
		// If face.Normal is zero (e.g. faces added via direct slice assignment that
		// bypassed AddFace), fall back to computing the normal from transformed vertices.
		var normal Vector3D
		if face.Normal != (Vector3D{}) {
			normal = rotMat.Apply(face.Normal)
		} else if n >= 3 {
			normal = computeFaceNormal(t.transformBuf[:n])
		}

		// Backface culling only for polygonal faces.
		// With outward-pointing normals, a face is visible when its normal
		// opposes the camera direction (dot < 0). Cull when dot >= 0.
		if t.options.BackfaceCulling && n >= 3 {
			dot := normal.Dot(t.camera.Direction)
			if dot >= 0 {
				continue
			}
		}

		// Carve out a slice of the pre-allocated points buffer for this face.
		start := t.pointsIdx
		t.pointsIdx += n

		// Initialize ProjectedFace with the rotated normal.
		pf := ProjectedFace{
			Points:     t.pointsBuf[start:t.pointsIdx:t.pointsIdx],
			Normal:     normal,
			Depth:      0,
			Char:       face.Char,
			RenderMode: face.RenderMode,
			Color:      face.Color,
			HasColor:   face.HasColor,
		}

		// Calculate average depth for sorting and project each vertex.
		depthSum := 0.0
		validProjection := true
		for i, vertex := range t.transformBuf {
			projected := t.camera.Project(vertex)
			if math.IsNaN(projected.X) || math.IsNaN(projected.Y) {
				validProjection = false
				break
			}
			pf.Points[i] = projected
			depthSum += vertex.Z
		}

		if !validProjection {
			// Roll back the cursor so the slice is reused next iteration.
			t.pointsIdx = start
			continue
		}

		pf.Depth = depthSum / float64(n)
		t.projected = append(t.projected, pf)
	}

	// Sort faces by depth (Painter's Algorithm).
	sortFacesByDepth(t.projected)

	// Resolve shading once so both the fill pass and the overlay pass use the
	// same lighting.
	for i := range t.projected {
		pf := &t.projected[i]
		baseColor := t.options.DiffuseColor
		if pf.HasColor {
			baseColor = pf.Color
		}
		pf.Brightness, pf.ShadeColor, pf.Char = calculatePhongShading(pf.Normal, t.light, baseColor, t.options, pf.Char)
	}

	// Reuse the subcell fill scene across frames; only allocate when the canvas
	// area changes (avoids ~800 KB of allocations per frame).
	if t.fillScene == nil || t.fillScene.cellArea != t.doubleBuffer.Area() {
		t.fillScene = newSubcellScene(t.doubleBuffer.Area())
	} else {
		t.fillScene.Clear()
	}

	// Render filled polygons into a higher-density braille scene first so
	// extruded glyph masks retain much more shape detail at normal terminal
	// sizes.
	for _, pf := range t.projected {
		if len(pf.Points) > 2 && pf.RenderMode == FaceRenderFill {
			t.fillScene.FillPolygon(pf.Points, pf.ShadeColor)
		}
	}
	if err := t.fillScene.CopyTo(t.doubleBuffer); err != nil {
		return err
	}

	// Render glyph overlays and explicit line faces on top of the filled scene.
	for _, pf := range t.projected {
		if len(pf.Points) == 2 {
			char := pf.Char
			if char == 0 {
				char = '─'
			}
			drawLine(t.doubleBuffer, pf.Points[0], pf.Points[1], char, cell.FgColor(pf.ShadeColor.ToCellColor()))
			continue
		}
		if len(pf.Points) > 2 && pf.RenderMode != FaceRenderFill {
			drawFillFace(t.doubleBuffer, pf.Points, pf.ShadeColor, pf.Char, pf.RenderMode)
		}
	}

	// Optionally, draw axes.
	if t.options.ShowAxes {
		drawAxes(t.doubleBuffer)
	}

	// Copy the off-screen buffer to the main canvas.
	return t.doubleBuffer.CopyTo(cvs)
}

// SetModel sets the 3D model to render.
func (t *ThreeD) SetModel(model *Model) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.model = model
	// Log the number of faces in the model.
	t.logger.Printf("Set new model with %d faces", len(model.Faces))
}

// normalizeRotation ensures rotation angles are within 0 to 2π.
func (t *ThreeD) normalizeRotation() {
	if t.options.UprightOnly {
		t.rotation.X = 0
		t.rotation.Z = 0
	}
	t.rotation.X = math.Mod(t.rotation.X, 2*math.Pi)
	t.rotation.Y = math.Mod(t.rotation.Y, 2*math.Pi)
	t.rotation.Z = math.Mod(t.rotation.Z, 2*math.Pi)

	if t.rotation.X < 0 {
		t.rotation.X += 2 * math.Pi
	}
	if t.rotation.Y < 0 {
		t.rotation.Y += 2 * math.Pi
	}
	if t.rotation.Z < 0 {
		t.rotation.Z += 2 * math.Pi
	}
}

// Rotate applies the provided delta rotation around the X, Y, and Z axes.
func (t *ThreeD) Rotate(delta Vector3D) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rotation.X += delta.X
	t.rotation.Y += delta.Y
	t.rotation.Z += delta.Z
	t.normalizeRotation()
	t.logger.Printf("Rotated to X: %.2f, Y: %.2f, Z: %.2f", t.rotation.X, t.rotation.Y, t.rotation.Z)
}

// Keyboard handles keyboard events.
func (t *ThreeD) Keyboard(k *terminalapi.Keyboard, _ *widgetapi.EventMeta) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch k.Key {
	case keyboard.KeyArrowUp:
		t.rotation.X -= t.options.RotationStep
	case keyboard.KeyArrowDown:
		t.rotation.X += t.options.RotationStep
	case keyboard.KeyArrowLeft:
		t.rotation.Y -= t.options.RotationStep
	case keyboard.KeyArrowRight:
		t.rotation.Y += t.options.RotationStep
	}

	t.normalizeRotation()
	return nil
}

// Mouse handles mouse events.
func (t *ThreeD) Mouse(m *terminalapi.Mouse, _ *widgetapi.EventMeta) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Handle zoom using the scroll wheel.
	switch m.Button {
	case mouse.ButtonWheelUp:
		t.zoomHandler.ZoomIn()
	case mouse.ButtonWheelDown:
		t.zoomHandler.ZoomOut()
	}
	return nil
}

// Options returns the options for this widget.
func (t *ThreeD) Options() widgetapi.Options {
	return widgetapi.Options{
		// Indicate that the widget wants keyboard and mouse events.
		WantKeyboard: widgetapi.KeyScopeGlobal,
		WantMouse:    widgetapi.MouseScopeWidget,
	}
}

// Logger returns the logger associated with the widget.
func (t *ThreeD) Logger() *log.Logger {
	return t.logger
}

// Ensure ThreeD implements widgetapi.Widget interface.
var _ widgetapi.Widget = (*ThreeD)(nil)
