HARD STOP: Implement exactly ONE commit per invocation. When the commit lands, append the final progress block, output a single promise (`<promise>NEXT</promise>` if @ralph/todo.md still has pending items, `<promise>COMPLETE</promise>` if it has ZERO pending items), and stop. Do NOT continue past the first commit. The next loop iteration will pick up remaining work.

0a. Study `specs/*` with multiple Sonnet subagents to learn the application specifications.
0b. Study @ralph/todo.md.

1. Your task is to implement functionality per the specifications and @ralph/todo.md file using parallel subagents. Follow @ralph/todo.md and choose the most important item to address. Before making changes, search the codebase (don't assume not implemented) using Sonnet subagents. You may use multiple Sonnet subagents for searches/reads and only 1 Sonnet subagent for build/tests. Use Opus subagents when complex reasoning is needed (debugging, architectural decisions).
2. Author property based tests or unit tests (which ever is best).
3. After implementing functionality or resolving problems, run the tests for that unit of code that was improved. If functionality is missing then it's your job to add it as per the application specifications. Ultrathink.
4. When you discover issues, immediately update @ralph/todo.md with your findings using a subagent. When resolved, update and remove the item.
5. When the tests pass, update @ralph/todo.md, then commit absolutely everything with a message describing the changes. After the commit, `git push`.
6. Do not write "Iteration NN" anywhere in @ralph/todo.md.

9999. Important: You can study the specifications and follow the citations to reference source code.
99999. Important: When authoring documentation, capture the why — tests and implementation importance.
999999. Important: Single sources of truth, no migrations/adapters. If tests unrelated to your work fail, resolve them as part of the increment.
9999999. You may add extra logging if required to debug issues.
99999999. Keep @ralph/todo.md current with learnings using a subagent — future work depends on this to avoid duplicating efforts. Update especially after finishing your turn.
999999999. When you learn something new about how to run the application, update @CLAUDE.md using a subagent but keep it brief. For example if you run commands multiple times before learning the correct command then that file should be updated.
9999999999. For any bugs you notice, resolve them or document them in @ralph/todo.md using a subagent even if it is unrelated to the current piece of work.
99999999999. Implement functionality completely. Placeholders and stubs waste efforts and time redoing the same work.
999999999999. When @ralph/todo.md becomes large periodically clean out the items that are completed from the file using a subagent.
9999999999999. If you find inconsistencies in the specs/* then use an Opus 4.6 subagent with 'ultrathink' requested to update the specs.
99999999999999. IMPORTANT: Keep @CLAUDE.md operational only — status updates and progress notes belong in @ralph/todo.md. A bloated CLAUDE.md pollutes every future loop's context.
999999999999999. NEVER emit a single Write over 400 lines. For larger files, create a ≤400-line skeleton with Write, then grow it with Edits. Placeholders are forbidden.
9999999999999999. DONE: emit `<promise>COMPLETE</promise>` ONLY when @ralph/todo.md has ZERO pending items. Re-read @ralph/todo.md to verify before choosing between `<promise>NEXT</promise>` and `<promise>COMPLETE</promise>`. Finishing this iteration's task is NOT agent complete; it is a NEXT.

## Progress Logging — Mandatory

You MUST append progress updates to ralph/progress.txt at the exact moments described below. This file is tailed live in the terminal — it is the ONLY way the user sees what is happening. Use `echo "..." >> ralph/progress.txt` (append, not overwrite).

Most importantly, the first thing you should do is append (iteration number should be exactly "[ralph-iteration]"):
```
═══════════════════════════════════════════════════════
  Ralph Iteration [ralph-iteration]
═══════════════════════════════════════════════════════

Brief explanation of what you will do (starting with a verb like "Finding most important item to address...", ending in ...)

```
The first line appended should be "═══════════════════════════════════════════════════════". If it's empty, make sure the first line is exactly "═══════════════════════════════════════════════════════".

After picking item to be addressed, append:
```

Chose X, it's the Y of Z.
```
The first line appended should be an empty line.

Whenever something meaningful happens, append a short note. Lean toward narrating more rather than less; silence looks like a stall.
```

Found/did/finished X. Now doing/investigating Y...
```
The first line appended should be an empty line.

After important finding, append:
```

Brief explanation of what was done/found. [Then "Continuing task..." or something like that]
```
The first line appended should be an empty line.

After finishing item that was picked to be addressed and committing, append:
```

## [Date] [Time] UTC - Changes committed.
- What was implemented
- Files changed
- **Brief description of changes:**
  - [change 1]
  - [change 2]
  - ...
---
```
The first line appended should be an empty line.

## Stop Condition

After your ONE commit lands and the final progress block is appended, reply with ONLY ONE of:

- `<promise>NEXT</promise>` if @ralph/todo.md still has pending items. The outer loop will start a fresh iteration.
- `<promise>COMPLETE</promise>` if @ralph/todo.md has ZERO pending items (verify by re-reading). The outer loop will exit.

Do NOT perform any additional work or verification after this signal.
