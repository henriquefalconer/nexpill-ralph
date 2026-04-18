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

## Setup (do this once per machine)

You only need **Docker** and **git** on the host. Everything else — Claude Code, Go, ralph tooling — lives inside the sandbox.

**Prerequisites** (both installed in the per-OS subsection below):

- A sandbox driver for `./ds` (`docker sandbox` or `sbx`).
- The GitHub CLI (`gh`) for forking and cloning in one command.

> **Anthropic subscription**: you need one. **A Max plan is strongly recommended** for Ralph — the loop is token-heavy (Opus planning plus Sonnet building, iteration after iteration).

### macOS

Install Docker Desktop 4.50+ (bundles the `docker sandbox` command that `./ds` drives) and the GitHub CLI:

```bash
# Install
brew install --cask docker
brew install gh
open -a Docker       # launch; wait for the whale icon to stop animating
gh auth login        # one-time; pick GitHub.com, then HTTPS or SSH

# Update later
brew upgrade --cask docker && brew upgrade gh
```

### Linux (Ubuntu 26.04)

Docker Desktop for Linux doesn't ship `docker sandbox` yet, so Linux uses the standalone `sbx` CLI — and the clean apt path only works on **Ubuntu 26.04** (or Rocky Linux 8). Earlier Ubuntu releases aren't in Docker's apt repo; upgrade to 26.04 or use a different machine.

```bash
curl -fsSL https://get.docker.com | sudo REPO_ONLY=1 sh
sudo apt-get install -y docker-sbx gh
sbx login            # one-time, opens a browser OAuth flow
gh auth login        # one-time; pick GitHub.com, then HTTPS or SSH
```

`./ds` auto-detects `sbx` and uses it in place of `docker sandbox`.

### Windows

The `./ds` script is bash, so you'll want a **bash-compatible terminal** to run it from — Git Bash (bundled with Git for Windows) is the easiest option; MSYS2, Cygwin, or similar also work.

1. **Docker Desktop for Windows 4.50+** (native ARM64 and amd64 builds) — <https://www.docker.com/products/docker-desktop/>. Bundles `docker sandbox`, runs the sandbox VM via Hyper-V.
2. **A bash-compatible terminal** — Git for Windows (Git Bash recommended) from <https://git-scm.com/download/win>; accept the installer defaults.
3. **GitHub CLI (`gh`)** — `winget install --id GitHub.cli` (or download from <https://cli.github.com>).

Open your bash-compatible terminal and authenticate:

```bash
gh auth login        # one-time; pick GitHub.com, then HTTPS or SSH
docker sandbox --help  # sanity check — should list subcommands
```

---

## Stage 0 — Clone, branch, log in

The vendored source (`punycode.js`, `tests/tests.js`, `LICENSE-MIT.txt`) and `TARGET.md` are already checked in on `main`. You just need your own branch and an authenticated sandbox.

### 1. Fork and clone

`gh repo fork` creates `<your-user>/nexpill-ralph` on GitHub, clones it locally, sets `origin` to your fork, and adds `upstream` pointing at `henriquefalconer/nexpill-ralph`.

```bash
gh repo fork henriquefalconer/nexpill-ralph --clone --remote
cd nexpill-ralph
```


### 2. Create the sandbox and log into Claude Code

Running `./ds` with no arguments invokes `docker sandbox run claude .` on the repo, which builds the sandbox image (~2 min, one time) and drops you into an interactive claude session.

```bash
./ds
```

Claude starts interactively. Pick **"Log in with your Anthropic account"** (the subscription flow), complete OAuth in the browser, then exit claude with Ctrl+C twice. Credentials land in the sandbox and persist for every subsequent `./ds <cmd>` call.

### 3. Authenticate git inside the sandbox

Ralph's auto-push needs cached credentials inside the sandbox. First, create a fine-grained Personal Access Token scoped to **only your fork**:

- Go to <https://github.com/settings/tokens?type=beta>
- **Token name:** anything (e.g. `ds-nexpill-ralph`)
- **Expiration:** whatever fits your workshop timebox
- **Repository access:** "Only select repositories" → pick `<your-user>/nexpill-ralph`
- **Repository permissions:** Contents → Read and write (Metadata → Read-only auto-selects; required)
- Generate the token and copy it once — you won't see it again.

Now cache it inside the sandbox. The first `git push` triggers the username/password prompt; paste the PAT as the password.

```bash
./ds git pull
./ds git push
# Username for 'https://github.com': <your-user>
# Password for 'https://<your-user>@github.com': <paste the PAT — input is hidden>
```

Credentials land in the sandbox's `~/.git-credentials` and persist across `./ds` calls.

### The Docker sandbox, and `./ds`

Everything happens inside a Docker sandbox — both the first `claude` to login and every ralph invocation.

`./ds` is a thin pass-through for `docker sandbox exec` — it runs whatever you hand it inside the project's sandbox, from the repo root. No built-in verbs, no ralph-specific knowledge:

```text
./ds                           ≡  docker sandbox run  claude .        # create/open sandbox + interactive claude
./ds ./ralph/ralph.sh plan --goal …  ≡  docker sandbox exec … ./ralph/ralph.sh plan --goal …
./ds ./ralph/ralph.sh             ≡  docker sandbox exec … ./ralph/ralph.sh
./ds go test ./port/...        ≡  docker sandbox exec … go test ./port/...   # any command works
```

The first invocation of `./ds` builds the sandbox image (~2 min, one time). Every call after that reuses the cached image. Why sandbox at all: Ralph passes `--dangerously-skip-permissions` to Claude Code, giving the agent full filesystem access — the sandbox confines that access to the repo directory plus the mounts.

---

## Stage 1 — Tests → specs (≈ 5–15 min)

Ralph reads every test file, fans out one subagent per file, writes one Markdown spec per test file. Output lands in `specs/tests/`.

```bash
./ds ./ralph/ralph.sh plan --goal "for every test file matching tests/**/*.js in this repo, use a separate subagent to produce specs/tests/<basename>.md capturing EVERY behavior the tests assert, in language-agnostic prose, with citations tests/<path>:<line>. Describe what the tests observe and require, never how the implementation works. Stop when every test file has a corresponding spec."
```

- Plan mode runs a single iteration by default — enough to produce the spec bundle in one pass.
- Watch progress live from a second terminal: `tail -f ralph/progress.txt`.
- Success looks like `specs/tests/tests.md` with ~200 bullet-point behaviors, each citing a `tests/tests.js:<line>`.

**Checkpoint**: spot-check 5 random citations. If they're wrong, the agent was hallucinating — tighten the goal and re-run.

---

## Stage 2 — Source → specs with citations (≈ 10–30 min)

```bash
./ds ./ralph/ralph.sh plan --goal "use a subagent to read punycode.js in full and produce specs/impl/punycode.md documenting public and internal behavior, invariants, data flow, and edge cases, with citations to punycode.js:<line>. The spec must be sufficient for a from-scratch reimplementation in any language — do not use JavaScript syntax in the prose. Follow the RFC 3492 structure where it maps naturally."
```

**Checkpoint**: open `specs/impl/punycode.md`. It should read like the RFC 3492 algorithm description, with line citations — not like JavaScript with comments.

---

## Stage 3 — Specs → `ralph/todo.md` (≈ 3–10 min)

Ralph turns the spec bundle into a prioritized, dependency-ordered porting plan. Every bullet is scoped to one Stage 4 build iteration.

```bash
./ds ./ralph/ralph.sh plan --goal "author ralph/todo.md as a prioritized porting plan from the specs under specs/tests/** and specs/impl/** into Go per TARGET.md. Order items by dependency: Go module scaffolding first (go.mod, package layout in port/), then primitives (ucs2 codec, digit mapping, bias adaptation), then the composite encoders/decoders (encode, decode, toASCII, toUnicode). Each bullet must be scoped to one ralph build iteration (~one commit) and must end with the test(s) from specs/tests/** that verify it. Finish when ralph/todo.md is a clean ordered list covering every behavior in the specs."
```

**Checkpoint**: open `ralph/todo.md`. The top item should be scaffolding (`go.mod`, empty `port/punycode.go`, empty `port/punycode_test.go`). The last item should be the most composite function. Every item should reference a spec.

---

## Stage 4 — Port (≈ 30 min – 2 hours)

Classic Ralph loop. Each iteration picks the top item off `ralph/todo.md`, implements it into `port/`, writes Go tests, runs `go test`, commits, pushes, and moves on.

```bash
./ds ./ralph/ralph.sh
```

- Build mode runs until Ralph signals `<promise>COMPLETE</promise>` — interrupt with Ctrl+C whenever your budget or timebox is up.
- Sonnet handles build mode (faster and cheaper than Opus per iteration).
- Ralph auto-pushes after every commit, so your branch on GitHub grows in real time.

When Ralph emits `<promise>COMPLETE</promise>`, run the Go test suite yourself from the host:

```bash
./ds go test ./port/... -v
```

Every RFC test vector from `specs/tests/*.md` should have a Go counterpart that passes.

---

## What to watch for during the loop

| Signal | Meaning | What to do |
|---|---|---|
| `progress.txt` idle for minutes | Subagent hung or spinning | Wait — the stall watchdog (30 min default) will kill it |
| Same todo item keeps coming back | Agent is guessing, not reading the spec | Kill the loop, open the todo + spec, tighten the spec wording |
| `go test` fails but agent commits anyway | Build loop didn't verify | Add a "must run `go test` and show output" line to `TARGET.md`, re-run |
| `<promise>COMPLETE</promise>` before tests pass | Agent's definition of done is weak | Re-run `./ds ./ralph/ralph.sh` to keep iterating |

---

## Limitations & honest caveats

1. **Token cost scales with participants.** Ten people × four stages × Opus planning + Sonnet building = real money. Interrupt the build loop with Ctrl+C once you've seen enough; a Max plan subscription absorbs most of this.
2. **Rate limits on a shared account.** Plan prompts fan out many parallel subagents. Ten participants starting Stage 1 simultaneously can hit org-level request caps. Stagger stage starts by 30 seconds if needed.
3. **Non-determinism is the point.** Two participants with the same goal produce different specs and different ports. Plan a group diff-review at the end — it's the most interesting part of the workshop.
4. **Tests don't port 1:1.** Mocha's `describe`/`it` + array-driven vectors become Go's `testing` + table-driven subtests. The agent rewrites rather than translates; edge cases occasionally drop. Spot-check.
5. **"Complete" is agent-declared, not verified.** Ralph signals completion when `ralph/todo.md` is empty. That doesn't mean `go test` is green. Always re-run tests manually via `./ds go test ./port/...`.
6. **Docker image size.** The sandbox image is ~800 MB (Ubuntu + Claude Code + Go). Budget disk accordingly on participant machines.
7. **Git push from inside the sandbox** needs credentials. The Stage 0 PAT step caches HTTPS creds in the sandbox's `~/.git-credentials`; alternatively, `./ds` mounts `~/.ssh` read-only so an SSH remote works too. If neither is set up, pushes fail silently and you can push manually from the host after each stage.
8. **Source protection is soft.** Nothing *forbids* editing `punycode.js`; plan mode says "markdown only" but build mode doesn't guard source paths. If the agent rewrites the original source, `git checkout -- punycode.js`.
9. **Workshop usually won't finish in-session.** A complete port takes 20–40 build iterations. In a 2-hour slot you'll see Stages 1–3 complete plus a partial port in Stage 4. That's the demo — participants continue at home.
10. **Go is pedagogically safe, not maximally dramatic.** For the full "compiler-as-teacher" effect ghuntley describes, target Rust instead — but expect 2–3× the iterations.

---

## Attribution

- Method: [Geoffrey Huntley, *Porting code bases using AI*](https://ghuntley.com/porting/).
- Source repo: [mathiasbynens/punycode.js](https://github.com/mathiasbynens/punycode.js), MIT licensed.
- Ralph runner: this repo's `ralph/` directory.
