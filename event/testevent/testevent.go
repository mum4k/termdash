// Package testevent provides utilities for tests that deal with concurrent
// events.
package testevent

import (
	"context"
	"fmt"
	"time"
)

// WaitFor waits until the provided function returns a nil error or the timeout.
// If the function doesn't return a nil error before the timeout expires,
// returns the last returned error.
func WaitFor(timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var err error
	for {
		tick := time.NewTimer(5 * time.Millisecond)
		select {
		case <-tick.C:
			if err = fn(); err != nil {
				continue
			}
			return nil

		case <-ctx.Done():
			return fmt.Errorf("timeout expired, error: %v", err)
		}
	}
}
