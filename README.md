# Ralph Porting Workshop — `punycode.js` → Go

A hands-on workshop for porting a small-but-non-trivial library from one language to another using Ralph, Claude Code, and the [ghuntley porting method](https://ghuntley.com/porting/).

Everyone in the room runs the same three stages in parallel, each on their own branch, each producing their own Go port. We compare results at the end.

---

## What we're porting

**[mathiasbynens/punycode.js](https://github.com/mathiasbynens/punycode.js)** → Go

- **Why this repo**: 428 lines of pure JavaScript, zero dependencies, MIT licensed. Implements [RFC 3492](https://www.rfc-editor.org/rfc/rfc3492.txt) (the encoding behind Internationalized Domain Names). Six public functions, ~200 RFC-official test vectors. Small enough to finish, non-trivial enough to be interesting (variable-length integer encoding with bias adaptation).
- **Why Go as target**: strict compiler catches porting mistakes early, stdlib has everything we need, Go doesn't ship a public `punycode` package (the one in `golang.org/x/net/idna/punycode` is internal) — so the port fills a real gap.

---

## The ghuntley method, in three stages

1. **Tests → specs.** Compress every test into a language-agnostic Markdown spec. Tests become the source of truth for *what the system does*.
2. **Source → specs with citations.** Document every source file as a spec that cites the original `path:line`. These specs become the PRD for the re-implementation.
3. **Specs → port.** Classic Ralph loop: one prioritized item per iteration, one commit per iteration, driven by a todo file derived from the specs.

Ralph runs stages 1 and 2 in `plan` mode (Opus, goal-driven, markdown-only outputs). Stage 3 runs in `build` mode (Sonnet, one commit per loop).

---

## Prerequisites

- A working [Claude Code](https://docs.claude.com/en/docs/claude-code/overview) CLI authenticated to an Anthropic account with Opus + Sonnet access.
- Git, bash, `go` 1.22+ (for running the port's tests at the end).
- A personal fork of this repo, so your pushes don't collide with anyone else's.
- Rough budget: **$20–$50 of Anthropic credit per participant** for a full run. Use `max_iterations` caps to stay in budget.

---

## Stage 0 — Setup (do this before the workshop starts)

```bash
# 1. Fork this repo on GitHub, then clone your fork
git clone git@github.com:<your-handle>/nexpill-ralph.git
cd nexpill-ralph

# 2. Make your own branch — everyone runs in parallel, nobody shares a branch
git checkout -b workshop/<your-name>

# 3. Vendor punycode.js into the repo root
git clone --depth 1 https://github.com/mathiasbynens/punycode.js /tmp/punycode-src
rm -rf /tmp/punycode-src/.git
# Rename the source README so it doesn't clobber this one
mv /tmp/punycode-src/README.md /tmp/punycode-src/SOURCE-README.md
cp -R /tmp/punycode-src/. .

# 4. Tell Ralph where the port goes
cat > TARGET.md <<'EOF'
# Port Target

- **Language**: Go 1.22+
- **Output directory**: `port/`
- **Module path**: `github.com/<your-handle>/punycode-port`
- **Package**: `punycode`
- **Test command**: `cd port && go test ./...`
- **Style**: idiomatic Go — functions over classes, explicit error returns, table-driven tests.
- **Scope**: port every exported function from `punycode.js` (`decode`, `encode`, `toUnicode`, `toASCII`, `ucs2.decode`, `ucs2.encode`).
- **Fidelity**: every RFC test vector in `tests/tests.js` must pass in Go. Add Go-idiomatic additions (e.g. fuzz tests) where valuable.
EOF

# 5. Commit the vendored source + TARGET.md on your branch
git add -A
git commit -m "workshop: vendor punycode.js + target spec"
git push -u origin workshop/<your-name>
```

At this point your working tree has:

```
nexpill-ralph/
├── ralph/               # Ralph tooling — don't touch
├── punycode.js          # vendored source
├── punycode.es6.js      # vendored source (ES6 twin)
├── tests/               # vendored tests
├── package.json         # vendored
├── LICENSE-MIT.txt      # vendored
├── SOURCE-README.md     # punycode.js's original README, renamed
├── TARGET.md            # port target spec — you just wrote this
└── README.md            # this file
```

---

## Stage 1 — Tests → specs (≈ 5–15 min)

Ralph reads every test file, fans out one subagent per file, writes one Markdown spec per test file capturing every asserted behavior. Output lands in `specs/tests/`.

```bash
./ralph/ralph.sh plan 5 --goal \
  "for every test file matching tests/**/*.js in this repo, use a separate \
   subagent to produce specs/tests/<basename>.md capturing EVERY behavior the \
   tests assert, in language-agnostic prose, with citations tests/<path>:<line>. \
   Describe what the tests observe and require, never how the implementation \
   works. Stop when every test file has a corresponding spec."
```

- `plan 5` caps at 5 iterations — a safety net so a stuck agent doesn't burn budget.
- Watch `ralph/progress.txt` in another terminal: `tail -f ralph/progress.txt`.
- Success looks like: `specs/tests/tests.md` (or similar) with ~200 bullet-point behaviors, each citing a `tests/tests.js:<line>`.

**Checkpoint before Stage 2**: open a spec, spot-check 5 random claims against the cited line. If citations are wrong, Ralph was hallucinating — re-run with tighter goal wording before continuing.

---

## Stage 2 — Source → specs with citations (≈ 10–30 min)

Ralph reads every source file and writes one spec per source module, documenting the algorithm in prose with citations back to `punycode.js:<line>`.

```bash
./ralph/ralph.sh plan 8 --goal \
  "for every non-test JavaScript source file at the repo root (punycode.js \
   and punycode.es6.js), use a separate subagent to produce \
   specs/impl/<module>.md documenting public and internal behavior, invariants, \
   data flow, and edge cases, with citations to <path>:<line>. These specs \
   must be sufficient for a from-scratch reimplementation in any language — \
   do not use JavaScript syntax in the prose. If punycode.js and punycode.es6.js \
   are semantic duplicates, produce one spec and note the equivalence."
```

**Checkpoint**: open `specs/impl/punycode.md`. It should read like the RFC 3492 algorithm description, with line citations — not like JavaScript with comments.

---

## Stage 3 — Port (≈ 30 min – 2 hours)

Now the classic Ralph loop. Build mode on the first iteration will notice `ralph/todo.md` is empty, generate it from your specs + `TARGET.md`, commit, and stop. From iteration 2 onward, each iteration picks one item off the todo, implements it into `port/`, writes Go tests, runs `go test`, commits, and pushes.

```bash
./ralph/ralph.sh 30
```

- `30` caps iterations — adjust for your budget and timebox.
- Sonnet handles build mode (faster and cheaper than Opus per iteration).
- Ralph auto-pushes after every commit, so your branch on GitHub grows in real time — great for demoing to the room.

When Ralph emits `<promise>COMPLETE</promise>`, it thinks it's done. Verify:

```bash
cd port && go test ./... -v
```

Every RFC test vector from `specs/tests/*.md` should have a Go counterpart that passes.

---

## What to watch for during the loop

| Signal | Meaning | What to do |
|---|---|---|
| `progress.txt` idle for minutes | Subagent hung or spinning | Wait — the stall watchdog (30 min default) will kill it |
| Same todo item keeps coming back | Agent is guessing, not reading the spec | Kill the loop, open the todo + spec, tighten the spec |
| `go test` fails but agent commits anyway | Build loop didn't verify | Add a "must run `go test` and show output" line to `TARGET.md`, re-run |
| `<promise>COMPLETE</promise>` before tests pass | Agent's definition of done is weak | Add a completion gate to `prompt-build.md` for a future workshop |

---

## Limitations & honest caveats

1. **Token cost scales with participants.** Ten people × three stages × Opus planning + Sonnet building = real money. Cap iterations aggressively.
2. **Anthropic rate limits on a shared org key.** The plan prompt asks for up to 500 parallel subagents. Ten participants starting stage 1 simultaneously can hit org-level request caps. Stagger stage starts by 30 seconds, or split participants across multiple API keys.
3. **Non-determinism is the point.** Two participants with the same goal produce different specs and different ports. Plan a group diff-review at the end — it's the most interesting part of the workshop.
4. **Tests don't port 1:1.** Mocha's `describe`/`it` + array-driven vectors become Go's `testing` + table-driven subtests. The agent rewrites rather than translates; edge cases occasionally drop. Spot-check.
5. **"Complete" is agent-declared, not verified.** Ralph signals completion when `ralph/todo.md` is empty. That doesn't mean `go test` is green. Always re-run tests manually.
6. **Platform quirks.** `ralph.sh` uses GNU `stat -c %Y`; macOS participants need to run inside the sandbox or adapt the script. The `CLAUDE.md` in this repo's parent documents the persistent-env setup.
7. **Source protection is soft.** Nothing in the prompts *forbids* editing `punycode.js`; the plan-mode prompt says "markdown only" but build mode doesn't guard source paths. Don't let the agent rewrite the original source — if it happens, `git checkout -- punycode.js`.
8. **Workshop will usually not finish in-session.** A complete port takes 20–40 build iterations. In a 2-hour slot you'll see stages 1 and 2 complete, plus a partial port in stage 3. That's the demo — participants continue at home.
9. **Go target is chosen for pedagogy, not difficulty.** Want the full "compiler-as-teacher" effect ghuntley describes? Target Rust instead — but expect 2–3× the iterations.
10. **The TARGET.md → build-bootstrap flow is opinionated.** The current `prompt-build.md` requires `TARGET.md` at repo root to generate a useful first todo. For non-porting workflows, ignore or customize this file.

---

## Attribution

- Method: [Geoffrey Huntley, *Porting code bases using AI*](https://ghuntley.com/porting/).
- Source repo: [mathiasbynens/punycode.js](https://github.com/mathiasbynens/punycode.js), MIT licensed.
- Ralph runner: this repo's `ralph/` directory.
