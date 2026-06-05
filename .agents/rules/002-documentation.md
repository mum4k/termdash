---
trigger: always_on
alwaysApply: true
description: When and how to update project documentation
globs: *
---

# Documentation

- **Update docs on each major change.** When you:
  - Add or remove a dependency or build requirement: update `README.md` and `CONTRIBUTING.md` as relevant.
  - Add or change a widget, public API surface, or drawing primitive: update the relevant package-level doc comments and `README.md` widget listing.
  - Remove a feature or deprecate a public API: update all references in README, CHANGELOG, doc/ files, and code comments.
  - Introduce a new concept or pattern (e.g. new widget category, new drawing primitive): add documentation in `doc/` so future contributors (and AI) know the intended design.
- **Docs to maintain:** `README.md` (quick start, widget listing, badges), `CONTRIBUTING.md` (fork/merge workflow, CLA), `CHANGELOG.md` (version history), `doc/` directory (design goals, guidelines, HLD, widget development guide).
- **Agent rules:** When making an important decision that should be followed in future work (e.g. conventions, patterns, "always do X" or "never do Y"), update rules in `.agents/rules/` so that AI and contributors consistently apply it. Add or adjust a rule in the appropriate file rather than relying on one-off comments or memory.
