---
trigger: always_on
alwaysApply: true
description: Unit test requirements and coverage standards for Go code
globs: *
---

# Tests and Coverage

- **Unit tests for all logic:** Every new or meaningfully changed behavior (new widget feature, new drawing primitive, bug fix in core packages) must have **corresponding unit tests** in the same change. Place tests in the same package (e.g., `widget_test.go`, `draw_test.go`).
- **Table-driven tests:** Use table-driven tests where multiple cases fit one pattern. See rule `013-go-table-driven-tests` for detailed guidelines.
- **Test helpers:** Leverage existing test helpers in `private/faketerm/`, `private/canvas/testcanvas/`, and similar packages. When adding new drawing primitives, provide corresponding `Must*` test helpers.
- **Race detector:** All tests must pass with the race detector enabled (`go test -race ./...`). Never introduce data races.
- **Run tests before completing:** Before confirming a feature is complete, the agent MUST run `go test ./...` to verify all tests pass. Do not present code as finished without confirming it compiles and tests pass.
- **Bug fixes:** For bug fixes, add a **regression test** that would have failed before the fix and passes after.
- **Coverage direction:** When adding features, ensure test coverage does not regress. Aim for thorough coverage of state mutations, boundary conditions, and error paths.
