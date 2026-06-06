---
trigger: always_on
alwaysApply: true
description: Go best practices for termdash library code
globs: "**/*.go"
---

# Go Best Practices

- **Idiomatic Go:** Follow standard Go style (effective go, gofmt). Use clear names and small, focused functions.
- **Error handling:** Always handle errors explicitly; never ignore them. Return errors to callers or handle them meaningfully. Use Go's standard `errors` and `fmt.Errorf` with `%w` for wrapping.
- **Readability first:** Termdash's primary design goal is code readability. The functionality and efficiency come second. Write code that is easy to understand, enhance, and test.
- **Widget separation:** Widget implementations should contain high-level code only. Low-level drawing primitives belong in separate packages (e.g., `private/draw/`). See `doc/design_guidelines.md`.
- **Test helpers:** Provide test helpers (e.g., `MustRectangle()`) for all functions in the draw package to simplify widget unit tests.
- **Public vs Private:** Private packages live under `private/` and are identified by the presence of `/private/` in their import path. Stability of private packages isn't guaranteed. Public API surface is documented in the wiki.
- **Concurrency:** When writing concurrent code, prefer clear synchronization patterns (mutexes, channels). Ensure proper locking and avoid data races — the project runs tests with the race detector.
- **Performance:** Avoid unnecessary allocations in hot paths (e.g., drawing loops, event processing). Prefer bounded operations and avoid O(n²) patterns where linear alternatives exist.
