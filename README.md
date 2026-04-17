# Ralph Porting Workshop — `punycode.js` → Go

A hands-on workshop for porting a small-but-non-trivial library from one language to another using Ralph, Claude Code, and the [ghuntley porting method](https://ghuntley.com/porting/).

Everyone in the room runs the same four stages in parallel, each on their own branch, each producing their own Go port. We compare results at the end.

---

## What we're porting

**[mathiasbynens/punycode.js](https://github.com/mathiasbynens/punycode.js)** → Go

428 lines of pure JavaScript, zero dependencies, MIT licensed. Implements [RFC 3492](https://www.rfc-editor.org/rfc/rfc3492.txt) (the encoding behind Internationalized Domain Names). Six public functions, ~200 RFC-official test vectors. Small enough to finish, non-trivial enough to be interesting — variable-length integer encoding with bias adaptation.

## Why Go as the target?

- **Strict compiler.** Untyped JS → typed Go forces the agent to make every coercion explicit. Mistakes fail at `go build` time, not at runtime.
- **Complete stdlib.** `unicode/utf16`, `strings`, `math` — every primitive punycode.js reaches for already exists in Go's stdlib, so the port stays dependency-free.
- **Small language, few traps.** Go has almost no "surprise idioms" — the agent writes idiomatic Go reliably. Rust would be more dramatic but burn 2–3× the iterations on borrow-checker disputes.
- **Fast feedback loop.** `go test ./...` finishes in seconds, so each build iteration verifies itself before committing.
- **Real gap to fill.** Go's `golang.org/x/net/idna/punycode` is internal — the port produces a library you could actually publish.

## The ghuntley method, mapped to four Ralph invocations

1. **Tests → specs** (plan, Opus). Compress every test into a language-agnostic Markdown spec.
2. **Source → specs with citations** (plan, Opus). Document every source file with citations back to `path:line`.
3. **Specs → todo** (plan, Opus). Distill both spec sets into `ralph/todo.md`: a prioritized porting plan.
4. **Todo → port** (build, Sonnet). Classic Ralph loop: one item per iteration, one commit per iteration.

---

## The `./ds` wrapper

Every Ralph invocation runs inside a Docker sandbox via `./ds`, a tiny shell alias at the repo root. You never call `./ralph/ralph.sh` directly — you call `./ds ...` with the exact same arguments, and `ds` forwards them into the container.

```text
./ds plan 5 --goal "..."   # planning run inside sandbox
./ds 30                    # build loop inside sandbox
./ds shell                 # open a bash shell inside sandbox (for go test, debugging)
./ds login                 # run `claude login` inside sandbox (one-time per machine)
```

Why: the `--dangerously-skip-permissions` flag Ralph passes to Claude Code gives the agent full filesystem access. The sandbox confines that access to the repo directory. First run of `./ds` builds the image (~2 min); every run after that is instant.

The script is intentionally language-agnostic — no `npm run`, no `make`, no new tool to install. Just a bash script on your PATH via `./ds`.

---

## Setup (do this once per machine)

You only need **Docker** and **git** on the host. Everything else — Claude Code, Go, ralph tooling — lives inside the sandbox.

> **Anthropic subscription**: you need one. **A Max plan is strongly recommended** for Ralph — each plan iteration fans out up to 500 Sonnet subagents plus Opus synthesis. Pro will throttle before a stage finishes; Max gives roughly 5× the headroom.

### macOS

```bash
brew install --cask docker   # Docker Desktop
open -a Docker               # launch it; wait for the whale icon to stop animating
```

### Linux

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker "$USER"
newgrp docker                # or: log out and back in
```

### Windows (WSL2 already installed)

1. Install [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop/).
2. Docker Desktop → **Settings → Resources → WSL integration** → toggle on your WSL distro.
3. Restart Docker Desktop.
4. From here on, use your **WSL2 shell** (Ubuntu recommended) for everything.

### Verify

```bash
docker run --rm hello-world
```

---

## Stage 0 — Clone, vendor, log in

```bash
# 1. Fork this repo on GitHub, then clone your fork
git clone git@github.com:<your-handle>/nexpill-ralph.git
cd nexpill-ralph

# 2. Personal branch — everyone runs in parallel, nobody shares a branch
git checkout -b workshop/<your-name>

# 3. Log into Claude Code inside the sandbox (builds the image on first run)
./ds login
# → prints a URL + device code; open URL on any device, paste code, approve.
#   Credentials land in your host ~/.claude and persist across runs.

# 4. Vendor punycode.js into the repo root
git clone --depth 1 https://github.com/mathiasbynens/punycode.js /tmp/punycode-src
rm -rf /tmp/punycode-src/.git
mv /tmp/punycode-src/README.md /tmp/punycode-src/SOURCE-README.md   # avoid clobber
cp -R /tmp/punycode-src/. .

# 5. Tell Ralph where the port goes
cat > TARGET.md <<'EOF'
# Port Target

- **Language**: Go 1.22+
- **Output directory**: `port/`
- **Module path**: `github.com/<your-handle>/punycode-port`
- **Package**: `punycode`
- **Test command**: `cd port && go test ./...`
- **Style**: idiomatic Go — functions over classes, explicit error returns, table-driven tests.
- **Scope**: port every exported function from `punycode.js` (`decode`, `encode`, `toUnicode`, `toASCII`, `ucs2.decode`, `ucs2.encode`).
- **Fidelity**: every RFC test vector from the original test suite must pass in Go.
EOF

# 6. Commit and push the vendored baseline
git add -A
git commit -m "workshop: vendor punycode.js + target spec"
git push -u origin workshop/<your-name>

# 7. Branch off so Ralph's work lives on its own branch, keeping the vendored
#    baseline clean for later diffs
git checkout -b workshop/<your-name>/port
git push -u origin workshop/<your-name>/port
```

Ralph's auto-push runs against whatever branch you're on when you invoke `./ds`, so Stages 1–4 will land on `workshop/<your-name>/port`. At the end you can `git diff workshop/<your-name>..workshop/<your-name>/port` to see exactly what Ralph produced.

After Stage 0 your tree looks like:

```
nexpill-ralph/
├── ralph/               # Ralph tooling — don't touch
├── ds                   # sandboxed ralph wrapper (builds its own image)
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

Ralph reads every test file, fans out one subagent per file, writes one Markdown spec per test file. Output lands in `specs/tests/`.

```bash
./ds plan 5 --goal \
  "for every test file matching tests/**/*.js in this repo, use a separate \
   subagent to produce specs/tests/<basename>.md capturing EVERY behavior the \
   tests assert, in language-agnostic prose, with citations tests/<path>:<line>. \
   Describe what the tests observe and require, never how the implementation \
   works. Stop when every test file has a corresponding spec."
```

- `plan 5` caps at 5 iterations — a safety net so a stuck agent doesn't burn budget.
- Watch progress live from a second terminal: `tail -f ralph/progress.txt`.
- Success looks like `specs/tests/tests.md` with ~200 bullet-point behaviors, each citing a `tests/tests.js:<line>`.

**Checkpoint**: spot-check 5 random citations. If they're wrong, the agent was hallucinating — tighten the goal and re-run.

---

## Stage 2 — Source → specs with citations (≈ 10–30 min)

```bash
./ds plan 8 --goal \
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

## Stage 3 — Specs → `ralph/todo.md` (≈ 3–10 min)

Ralph turns the spec bundle into a prioritized, dependency-ordered porting plan. Every bullet is scoped to one Stage 4 build iteration.

```bash
./ds plan 3 --goal \
  "author ralph/todo.md as a prioritized porting plan from the specs under \
   specs/tests/** and specs/impl/** into Go per TARGET.md. Order items by \
   dependency: Go module scaffolding first (go.mod, package layout in port/), \
   then primitives (ucs2 codec, digit mapping, bias adaptation), then the \
   composite encoders/decoders (encode, decode, toASCII, toUnicode). Each \
   bullet must be scoped to one ralph build iteration (~one commit) and must \
   end with the test(s) from specs/tests/** that verify it. Finish when \
   ralph/todo.md is a clean ordered list covering every behavior in the specs."
```

**Checkpoint**: open `ralph/todo.md`. The top item should be scaffolding (`go.mod`, empty `port/punycode.go`, empty `port/punycode_test.go`). The last item should be the most composite function. Every item should reference a spec.

---

## Stage 4 — Port (≈ 30 min – 2 hours)

Classic Ralph loop. Each iteration picks the top item off `ralph/todo.md`, implements it into `port/`, writes Go tests, runs `go test`, commits, pushes, and moves on.

```bash
./ds 30
```

- `30` caps iterations — adjust for your budget and timebox.
- Sonnet handles build mode (faster and cheaper than Opus per iteration).
- Ralph auto-pushes after every commit, so your branch on GitHub grows in real time.

When Ralph emits `<promise>COMPLETE</promise>`, verify yourself inside the sandbox:

```bash
./ds shell
# inside the sandbox:
cd port && go test ./... -v
```

Every RFC test vector from `specs/tests/*.md` should have a Go counterpart that passes.

---

## What to watch for during the loop

| Signal | Meaning | What to do |
|---|---|---|
| `progress.txt` idle for minutes | Subagent hung or spinning | Wait — the stall watchdog (30 min default) will kill it |
| Same todo item keeps coming back | Agent is guessing, not reading the spec | Kill the loop, open the todo + spec, tighten the spec wording |
| `go test` fails but agent commits anyway | Build loop didn't verify | Add a "must run `go test` and show output" line to `TARGET.md`, re-run |
| `<promise>COMPLETE</promise>` before tests pass | Agent's definition of done is weak | Re-run `./ds 10` to keep iterating |

---

## Limitations & honest caveats

1. **Token cost scales with participants.** Ten people × four stages × Opus planning + Sonnet building = real money. Cap iterations aggressively; a Max plan subscription absorbs most of this.
2. **Rate limits on a shared account.** Plan prompts fan out up to 500 parallel subagents. Ten participants starting Stage 1 simultaneously can hit org-level request caps. Stagger stage starts by 30 seconds if needed.
3. **Non-determinism is the point.** Two participants with the same goal produce different specs and different ports. Plan a group diff-review at the end — it's the most interesting part of the workshop.
4. **Tests don't port 1:1.** Mocha's `describe`/`it` + array-driven vectors become Go's `testing` + table-driven subtests. The agent rewrites rather than translates; edge cases occasionally drop. Spot-check.
5. **"Complete" is agent-declared, not verified.** Ralph signals completion when `ralph/todo.md` is empty. That doesn't mean `go test` is green. Always re-run tests manually via `./ds shell`.
6. **Docker image size.** The sandbox image is ~800 MB (Ubuntu + Claude Code + Go). Budget disk accordingly on participant machines.
7. **Git push from inside the sandbox** needs your SSH key — `./ds` mounts `~/.ssh` read-only for this. Pushes to HTTPS remotes with cached credentials also work. If neither is set up, pushes fail silently and you can push manually from the host after each stage.
8. **Source protection is soft.** Nothing *forbids* editing `punycode.js`; plan mode says "markdown only" but build mode doesn't guard source paths. If the agent rewrites the original source, `git checkout -- punycode.js`.
9. **Workshop usually won't finish in-session.** A complete port takes 20–40 build iterations. In a 2-hour slot you'll see Stages 1–3 complete plus a partial port in Stage 4. That's the demo — participants continue at home.
10. **Go is pedagogically safe, not maximally dramatic.** For the full "compiler-as-teacher" effect ghuntley describes, target Rust instead — but expect 2–3× the iterations.

---

## Attribution

- Method: [Geoffrey Huntley, *Porting code bases using AI*](https://ghuntley.com/porting/).
- Source repo: [mathiasbynens/punycode.js](https://github.com/mathiasbynens/punycode.js), MIT licensed.
- Ralph runner: this repo's `ralph/` directory.
