---
trigger: always_on
description: Verification of compilation before completing work
globs: *
---

# Rule: Compile Check

Whenever the agent delivers a modification to the source code, the agent **MUST explicitly run `go build ./...`** via the command line to verify that the code compiles flawlessly **before** confirming the feature is complete and asking for the user's review.

Additionally, run `go vet ./...` to catch common issues that the compiler alone would not flag.

Never blindly present or finalize code that has not been confirmed to compile and pass vet checks locally via terminal output.
