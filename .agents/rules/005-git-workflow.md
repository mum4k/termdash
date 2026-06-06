---
trigger: always_on
description: Git workflow, branching policy, and safety rules
globs: *
alwaysApply: true
---

# Rule: git-workflow

## Git Version Control, Safety & Review Process

To protect the integrity of the project code and ensure effective review of logic changes, the AI agent must adhere strictly to the following Git workflow and safety procedures.

### Branching Policy

The agent is explicitly forbidden from committing work directly to the `master` or `devel` branches.

**Termdash uses the [nvie branching model](https://nvie.com/posts/a-successful-git-branching-model/).** All development happens on the `devel` branch. Feature branches must be based off `devel` and PRs must target `devel`. The `master` branch is reserved for releases, major bug fixes, and documentation updates.

Whenever a new feature, iteration, or rule modification is requested by the user, the agent must:
1. Check if there is already an active feature branch we are iterating on for the current session or logical feature. If so, **you MUST checkout, reuse it, and push to it.** Do NOT create a new branch/PR for every single follow-up prompt or UI tweak.
2. If there truly isn't an active feature branch, create and checkout a new feature branch using `git checkout -b feature/[short-description] devel`. **NEVER create multiple feature branches for a single logical set of updates or consecutive user requests. Consolidate them into one branch.**
3. Create a GitHub Pull Request (PR) for the branch targeting `devel` if it doesn't already exist.
4. Perform all research, modifications, rule generations, and script edits on that single unified feature branch.

The user mandates a strict **One Branch / One PR Policy** for all ongoing work.

Whenever taking action, the agent MUST:
1. Keep exactly **one** unified feature branch open (e.g., `feature/current-updates`) for the entirety of the conversational session or block of work.
2. **Never create multiple feature branches or Pull Requests.** Push all incremental modifications, completely distinct files, and entirely different features directly to that single, unified active branch.
3. Create one GitHub Pull Request (PR) for everything, and just keep adding commits to it.
4. DO NOT create "per-task" branches. The user explicitly only wants one feature branch and one PR to review at a time.

### Git Safety and Failure Handling
1. **No Force Pushing**: The agent is STRICTLY FORBIDDEN from ever using `--force` or `-f` flags when pushing to any remote branch (`git push --force`, `git push -f`).
2. **No Force Committing**: The agent is STRICTLY FORBIDDEN from bypassing git hooks using the `--no-verify` flag during commits (`git commit --no-verify`), unless explicitly instructed by the user.
3. **Failure Handling**: If a `git commit` or `git push` command fails for any reason (e.g., pre-push hook fails, branch conflicts, test script failures), the agent MUST STOP and immediately ask the user how to proceed.
4. **Resolution**: The agent should present the error output to the user and await explicit instructions rather than attempting to force the git operation through.

### Review via Pull Requests
When the agent has completed the iteration on the feature branch:
1. Stage and commit the changes (`git add . && git commit -m "..."`).
2. Push the feature branch to the remote repository.
3. Notify the user that the branch is ready for review.
4. **CRITICAL: NEVER automatically merge the PR (`gh pr merge`) into `devel` or `master` unless the user explicitly gives you the command to do so.** You must stop and wait for their review feedback.
5. **Post-Merge Cleanup**: When the user explicitly gives you permission to merge the PR, you MUST clean up both local and remote feature branches immediately after merging (e.g., `git push origin --delete branch-name` and `git branch -D branch-name`). Do not leave orphaned feature branches in the local or remote environments.

### Processing Review Feedback
If the user requests changes via GitHub PR comments:
1. Methodically address each comment directly in the codebase or rules.
2. Once pushed, explicitly acknowledge every addressed comment so the user can smoothly track your resolution status.

**Only if the user explicitly inputs a manual override command** (e.g., "you may commit directly to devel this one time") is the agent allowed to bypass this safety protocol. Otherwise, feature branches must be isolated and submitted for PR review.
