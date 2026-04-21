PART 0
0.0. comprehensively study and compare current branch with [ref-branch] and also study deeply the content of the files present in current branch but not in [ref-branch]
PART 1
1.1. regarding the code implemented that was kept in current branch, point out all the fundamental problems with this implementation and all the assumptions that have any possibility of generating problems
1.2. out of the picked items, think about which can be considered a direct result of the code that was added in current branch. also, considering a deep study of the patterns already previously followed and generally adopted in [ref-branch], think of which fundamental problems listed could be appropriately disregarded
1.3. for every problem that cannot and should not be disregarded, think deeply of all the connected systems and uses of the item, all the variables and situations that in any shape or form influence this problem, and think of a more complete conclusion of all the consequences that deploying the current code would and could generate
1.4. create and overwrite the file in project root called SECURITY_REVIEW.md with what you learned
1.5. also create and overwrite a file in project root called SECURITY_ANALYSIS.md with the complete picture of the Fundamental Problems and Risky Assumptions and what was considered for each
1.6. also create and overwrite a file in project root called SECURITY_COMPARISON.md with the detailed picture of the code that was altered, and nothing but the code that was altered, in a way to not mistake with ANY code that was already there
PART 2
2.1. Append the final "Changes committed" progress block to `ralph/progress.txt` FIRST, then commit absolutely everything (`git add -A` — this MUST include `ralph/progress.txt`, `ralph/todo.md`, and `ralph/.last-branch` alongside the SECURITY_*.md files) with a message describing the changes. After the commit, `git push`.

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

After finishing item that was picked to be addressed, append the block BELOW to `ralph/progress.txt` FIRST, THEN run `git add -A` and `git commit` so the block is part of the same commit:
```

## YYYY-MM-DDTHH:mm:ss UTC - Changes committed.
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

- `<promise>NEXT</promise>` if more security review iterations are needed.
- `<promise>COMPLETE</promise>` if the review is fully complete.

Do NOT perform any additional work or verification after this signal.
