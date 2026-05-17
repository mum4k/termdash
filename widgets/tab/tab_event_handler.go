// Package tab provides functionality for managing tabbed interfaces.
package tab

import (
	"context"
	"image"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// EventHandler handles keyboard and mouse events for tab navigation.
type EventHandler struct {
	term           terminalapi.Terminal // Terminal interface.
	tm             *Manager             // Reference to the Tab Manager.
	th             *Header              // Reference to the Header.
	tc             *Content             // Reference to the Content.
	container      *container.Container // Container for updating content.
	ctx            context.Context      // Context for cancellation.
	cancel         context.CancelFunc   // Function to cancel the context.
	opts           *Options             // Configuration options.
	logger         *log.Logger          // Logger for event handling.
	logFile        io.Closer            // File backing the logger when logging is enabled.
	lastSwitchTime time.Time            // Tracks when the last tab switch due to notification occurred.
	refreshMu      sync.Mutex           // Serializes redraw-triggering updates.
}

// NewEventHandler initializes a new EventHandler.
func NewEventHandler(ctx context.Context, term terminalapi.Terminal, tm *Manager, th *Header, tc *Content, cont *container.Container, cancel context.CancelFunc, opts *Options) *EventHandler {
	var logger *log.Logger
	var logFile io.Closer

	// Only create the log file if EnableLogging is true.
	if opts.EnableLogging {
		file, err := os.OpenFile("tab_demo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("failed to open log file: %v", err)
			logger = log.New(io.Discard, "", 0) // Discard logging if file creation fails.
		} else {
			logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
			logFile = file
		}
	} else {
		// Discard logs if logging is disabled, and do not create the file.
		logger = log.New(io.Discard, "", 0)
	}

	eh := &EventHandler{
		term:      term,
		tm:        tm,
		th:        th,
		tc:        tc,
		container: cont,
		ctx:       ctx,
		cancel:    cancel,
		opts:      opts,
		logger:    logger,
		logFile:   logFile,
	}

	// Start the notification watcher if FollowNotifications is enabled.
	if opts.FollowNotifications {
		go eh.startNotificationWatcher()
	}
	if eh.logFile != nil {
		go func() {
			<-ctx.Done()
			_ = eh.logFile.Close()
		}()
	}

	return eh
}

// startNotificationWatcher monitors for tab notifications and switches tabs.
func (eh *EventHandler) startNotificationWatcher() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	queued := make(map[int]struct{})
	var order []int
	currentIndex := -1
	var tabViewStartTime time.Time

	for {
		select {
		case <-eh.ctx.Done():
			return
		case now := <-ticker.C:
			for _, index := range eh.tm.GetNotifiedIndexes() {
				if _, ok := queued[index]; ok {
					continue
				}
				queued[index] = struct{}{}
				order = append(order, index)
			}

			for i := 0; i < len(order); {
				index := order[i]
				if !eh.indexHasNotification(index) {
					delete(queued, index)
					order = append(order[:i], order[i+1:]...)
					continue
				}
				i++
			}

			if currentIndex >= 0 && now.Sub(tabViewStartTime) >= time.Second {
				if eh.tm.SetNotification(currentIndex, false, 0) {
					eh.refresh()
				}
				delete(queued, currentIndex)
				order = removeIndex(order, currentIndex)
				currentIndex = -1
			}

			if currentIndex == -1 && len(order) > 0 && now.Sub(eh.lastSwitchTime) >= time.Second {
				currentIndex = order[0]
				order = order[1:]
				delete(queued, currentIndex)
				if eh.tm.SetActiveTab(currentIndex) {
					eh.refresh()
				}
				eh.lastSwitchTime = now
				tabViewStartTime = now
			}
		}
	}
}

// HandleKeyboard processes keyboard events for tab switching.
func (eh *EventHandler) HandleKeyboard(k *terminalapi.Keyboard) {
	eh.logger.Printf("Keyboard event: key=%v", k.Key)
	// Handle Ctrl+C and 'q' to exit.
	if k.Key == keyboard.KeyCtrlC || k.Key == keyboard.KeyEsc || k.Key == 'q' || k.Key == 'Q' {
		eh.cancel()
		return
	}

	// Switch to the next tab with Tab key or Right arrow key.
	if k.Key == keyboard.KeyTab || k.Key == keyboard.KeyArrowRight {
		eh.tm.NextTab()
		eh.refresh()
		return
	}

	// Switch to the previous tab with Left arrow key.
	if k.Key == keyboard.KeyArrowLeft {
		eh.tm.PreviousTab()
		eh.refresh()
		return
	}
}

// HandleMouse processes mouse events for tab switching.
func (eh *EventHandler) HandleMouse(m *terminalapi.Mouse) {
	eh.logger.Printf("Mouse event: button=%v, position=%v", m.Button, m.Position)

	if m.Button != mouse.ButtonLeft {
		return // Only handle left-click presses.
	}

	// Get terminal size.
	size := eh.term.Size()
	height := size.Y

	// Calculate the height of the Header.
	headerHeight := eh.th.Height()
	if headerHeight == 0 {
		headerHeight = int(float64(height) * 0.1) // Fallback to 10% of terminal height.
		if headerHeight == 0 {
			headerHeight = 1
		}
	}

	// Check if the click is within the Header area.
	if m.Position.Y >= headerHeight {
		// Click is outside the Header.
		return
	}

	// Adjust the mouse position to be relative to the Header.
	adjustedPosition := image.Point{
		X: m.Position.X,
		Y: m.Position.Y, // Since Header starts at Y=0, this remains the same.
	}

	// Determine which tab was clicked.
	clickedTabIndex := eh.th.GetClickedTab(adjustedPosition)
	if clickedTabIndex != -1 {
		eh.tm.SetActiveTab(clickedTabIndex)
		eh.refresh()
	}
}

// refresh redraws the header and active tab content after a state change.
func (eh *EventHandler) refresh() {
	eh.refreshMu.Lock()
	defer eh.refreshMu.Unlock()

	if err := eh.th.Update(); err != nil {
		eh.logger.Printf("failed to update tab header: %v", err)
	}
	if err := eh.tc.Update(eh.container); err != nil {
		eh.logger.Printf("failed to update tab content: %v", err)
	}
}

// Refresh redraws the tab header and content after an external state change.
func (eh *EventHandler) Refresh() {
	eh.refresh()
}

// indexHasNotification reports whether the tab at index still has a notification.
func (eh *EventHandler) indexHasNotification(index int) bool {
	return eh.tm.HasNotificationAt(index)
}

// removeIndex removes the first matching value from a queue of indexes.
func removeIndex(indexes []int, target int) []int {
	for i, index := range indexes {
		if index == target {
			return append(indexes[:i], indexes[i+1:]...)
		}
	}
	return indexes
}
