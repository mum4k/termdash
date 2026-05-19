---
trigger: always_on
description: Foundational AI agent operating rules for the termdash repository
globs: *
---

# Persistent Rules

## Rule Loading

Rules located in `.agents/rules/` are the authoritative set of operating boundaries for any AI agent working in this repository.

These rules establish the foundational operating boundaries for the repository. Under no circumstances should the agent ignore, bypass, or selectively apply these rules unless explicitly commanded by the user with a manual override.

### Rule File Format
All new rules must be created as standard `.md` files in `.agents/rules/`. 

To ensure Cursor loads these rules automatically, **every rule must have a corresponding `.mdc` symlink inside `.cursor/rules/`**. When creating or adding a new rule file `[name].md`, you MUST also create the symlink:
`ln -s ../../.agents/rules/[name].md .cursor/rules/[name].mdc`

All newly added rule files MUST contain a YAML frontmatter header at the very top of the file in the following format:

```yaml
---
trigger: always_on
description: [Short rule description]
globs: *
---
```
