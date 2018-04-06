package container

// focus.go contains code that tracks the focused container.

import (
	"image"
	"sync"

	"github.com/mum4k/termdash/terminalapi"
)

// pointCont finds the top-most (on the screen) container whose area contains
// the given point. Returns nil if none of the containers in the tree contain
// this point.
func pointCont(c *Container, p image.Point) *Container {
	var (
		errStr string
		cont   *Container
	)
	postOrder(rootCont(c), &errStr, visitFunc(func(c *Container) error {
		if p.In(c.area) && cont == nil {
			cont = c
		}
		return nil
	}))
	return cont
}

// focusTracker tracks the active (focused) container.
// This object is thread-safe and must not be copied.
type focusTracker struct {
	// container is the currently focused container.
	container *Container

	// candidate is the container that might become focused next. I.e. we got
	// a mouse click and now waiting for a release or a timeout.
	candidate *Container

	// mouseFSM is a state machine tracking mouse clicks in containers and
	// moving focus from one container to the next.
	mouseFSM mouseStateFn

	mu sync.RWMutex
}

// newFocusTracker returns a new focus tracker with focus set at the provided
// container.
func newFocusTracker(c *Container) *focusTracker {
	return &focusTracker{
		container: c,
		mouseFSM:  mouseWantLeftButton,
	}
}

func (ft *focusTracker) isActive(c *Container) bool {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	return ft.container == c
}

// mouse identifies mouse events that change the focused container and track
// the focused container in the tree.
func (ft *focusTracker) mouse(m *terminalapi.Mouse) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	ft.mouseFSM = ft.mouseFSM(ft, m)
}
