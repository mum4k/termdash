---
trigger: always_on
description: Read only question mode
globs: *
---

# Q: Prefix Rule

If the user's prompt begins exactly with the prefix `q: ` (case-insensitive), you MUST operate in a **read-only / answer-only mode**. 

Under this condition:
1. You may use read-oriented tools (like reading files, searching code, or checking git status) to gather context to answer the user's question.
2. You are **STRICTLY FORBIDDEN** from making any modifications to the codebase, creating new files, running destructive commands, or initiating any deployments.
3. Your final output should only be an explanation, analysis, or direct answer to the question asked. 
