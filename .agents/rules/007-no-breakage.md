---
trigger: always_on
alwaysApply: true
description: Policy for public API stability and backward compatibility
globs: *
---

# No-Breakage Policy

Termdash is a published Go library used by downstream projects. Before suggesting or making changes that affect the public API:

- **Analyze the codebase** (e.g. search for usages of the type, function, or interface) to assess impact on both internal code and downstream consumers.
- **Public API types:** If changing exported types, function signatures, or interface contracts in non-`/private/` packages, ensure backward compatibility. Prefer additive changes (new optional fields via functional options, new methods) over breaking modifications.
- **Widget API:** If modifying `widgetapi.Widget` or related interfaces, check all widget implementations for compatibility. Every widget must continue to compile and pass tests.
- **Private packages:** Changes to packages under `/private/` have more flexibility since they are explicitly not part of the public API. However, still check all internal consumers.
- **Breaking changes:** If a breaking change is truly necessary, call it out explicitly to the user and document it for the `CHANGELOG.md`. The project has not yet reached v1.0.0, so breaking changes are permitted but should be minimized.

When in doubt, prefer smaller, backward-compatible changes and call out any breaking changes explicitly.
