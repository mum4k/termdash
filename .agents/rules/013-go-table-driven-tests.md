---
trigger: always_on
description: Best practices for writing table-driven tests in Go
globs: *
---

# Go Table-Driven Tests Best Practices

When writing table-driven tests (TDTs) in Go within this repository, adhere strictly to the following guidelines:

1. **Keep Setup Logic Uniform:** The power of table-driven testing comes from running multiple nuanced inputs through the exact same logic execution and setup phases.
2. **Avoid Conditional Execution:** Do not introduce single-case conditional logic checks (e.g. `if tt.name == "SpecialCase" { ... }`) inside the standard test execution loop (`for _, tt := range tests { t.Run(...) }`).
3. **Extract Complex Cases:** If a test case requires fundamentally different environment bootstrapping, mocking, or contextual injection that cannot be cleanly modeled as parameters within the table struct, **it does not belong in the table**. 
4. **Standalone Functions:** Extract these highly custom scenarios into their own dedicated, standalone test functions (e.g., `func TestFeature_SpecialCase(t *testing.T)`).

Following constraints like this ensures tests remain readable, modular, and easy to maintain without turning the iterative execution block into unreadable spaghetti code.
