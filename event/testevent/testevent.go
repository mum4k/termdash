// Copyright 2019 Google Inc.
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
