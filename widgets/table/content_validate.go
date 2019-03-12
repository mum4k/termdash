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

package table

// contant_validate.go contains code that validates the user provided Content.

import "fmt"

// validateContent validates the content instance.
func validateContent(c *Content) error {
	if min := 1; int(c.cols) < min {
		return fmt.Errorf("invalid number of columns %d, must be a value in range %d <= v", c.cols, min)
	}

	for _, r := range c.rows {
		if got, want := r.effectiveColumns(), int(c.cols); got != want {
			return fmt.Errorf("content has %d columns, but row %v has %d, all rows must occupy the same amount of columns", want, r, got)
		}
	}
	return nil
}
