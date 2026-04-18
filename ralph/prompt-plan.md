You are in PLAN MODE. The ULTIMATE GOAL at the bottom of this prompt tells you *what* to produce this iteration. Everything above is *how* to behave regardless of goal. Read the goal first; let it drive every decision below.

0a. Plan mode writes ONLY markdown — typically under `specs/**` and `ralph/todo.md`. Never modify source, tests, configuration, or any non-markdown file. Read anything; write only markdown.

0b. Read `ralph/todo.md` and any existing `specs/**` to see what prior iterations have produced. If the ULTIMATE GOAL is already satisfied by what exists, emit `<promise>COMPLETE</promise>` immediately without further work.

1. Execute the ULTIMATE GOAL. Ultrathink about:
   - Which inputs to read (source tree, tests, existing specs, notes) given the goal.
   - Which outputs satisfy the goal and where they belong on disk.
   - The right granularity — one spec per logical unit (test file, source module, feature) is usually correct.
   - Dependencies between outputs — write primitives before composites.
   Use up to 1 Sonnet subagent for independent reads. Use Opus subagents for synthesis when the task requires reconciling findings across files.

2. Default spec-writing rules (apply unless the goal overrides them):
   - One markdown file per logical unit.
   - Every claim cites its source with `path:line` so future readers can `file_read` the original without guessing.
   - Prose is language-agnostic — describe behavior, not the source language's syntax. This matters when specs will be re-implemented in another language.
   - Describe *what* the system does and *must* do. Describe *how* only when the goal asks for it.

3. `ralph/todo.md` format (when the goal is to produce or update it):
   - Prioritized bullet list; top item = most important thing yet to do.
   - Each item is scoped to ~one ralph build iteration (one commit).
   - Order respects dependencies.
   - Keep it lean — prune completed items and historical details that don't inform future work. Specs are the durable baseline.

IMPORTANT: Plan only — do NOT implement. Before claiming something is missing, search for it first. Commit everything when finished writing plan.

ULTIMATE GOAL: [project-specific goal]

## Progress Logging — Mandatory

You MUST append progress updates to ralph/progress.txt at the exact moments described below. This file is tailed live in the terminal — it is the ONLY way the user sees what is happening. Use `echo "..." >> ralph/progress.txt` (append, not overwrite).

Most importantly, the first thing you should do is append:
```
═══════════════════════════════════════════════════════
  Ralph Iteration [ralph-iteration]
═══════════════════════════════════════════════════════

Brief explanation of what you will do (starting with a verb like "Finding most important item to address...", ending in ...)

```
The first line appended should be "═══════════════════════════════════════════════════════". If it's empty, make sure the first line is exactly "═══════════════════════════════════════════════════════".

If still doing a specific task and it is taking too long to finish, append (try to always append once per minute):
```

Found/did/finished X, still doing/finding/etc Y...
```
The first line appended should be an empty line.

After picking item to be addressed, append:
```

Chose X, it's the Y of Z.
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

After finishing all steps and committing, reply with ONLY ONE of:

- `<promise>NEXT</promise>` if more planning iterations are needed.
- `<promise>COMPLETE</promise>` if the plan is fully written and no further iterations are required.

Do NOT perform any additional work or verification after this signal.
