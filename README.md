# Ralph Porting Workshop — `punycode.js` → Go

A hands-on workshop for porting based on [ghuntley porting method](https://ghuntley.com/porting/) for language porting with Ralph.

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

### Linux

Linux uses the standalone `sbx` CLI. Here's one way to do it on **Ubuntu 26.04**:

```bash
curl -fsSL https://get.docker.com | sudo REPO_ONLY=1 sh
sudo apt-get install -y docker-sbx gh
sbx login            # one-time, opens a browser OAuth flow
gh auth login        # one-time; pick GitHub.com, then HTTPS or SSH
```

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

Now cache it inside the sandbox. The first `git pull` triggers the username/password prompt; paste the PAT as the password.

```bash
./ds git config --global --unset credential.helper
./ds git config --global credential.helper store
./ds git pull
# Username for 'https://github.com': <your-user>
# Password for 'https://<your-user>@github.com': <paste the PAT — input is hidden>
```

Credentials land in the sandbox's `~/.git-credentials` and persist across `./ds` calls.

### The Docker sandbox, and `./ds`

Everything happens inside a Docker sandbox — both the first `claude` to login and every ralph invocation.

`./ds` is a thin pass-through for `docker sandbox exec` — it runs whatever you hand it inside the project's sandbox, from the repo root:

```text
./ds                           ≡  docker sandbox run  claude .        # create/open sandbox + interactive claude
./ds ./ralph/ralph plan --goal …  ≡  docker sandbox exec … ./ralph/ralph plan --goal …
./ds ./ralph/ralph             ≡  docker sandbox exec … ./ralph/ralph
```

The first invocation of `./ds` builds the sandbox image (~2 min, one time). Every call after that reuses the cached image. Why sandbox at all: Ralph passes `--dangerously-skip-permissions` to Claude Code, giving the agent full filesystem access — the sandbox confines that access to the repo directory plus the mounts.

---

## Stage 1 — Running the Loop

And now, to do the conversion, run this in your project directory:

```bash
./ds ./ralph/ralph plan --goal "study every file in tests/* using separate subagents and document in /specs/*.md and link the implementation as citations in the specification (flat structure)" && ./ds ./ralph/ralph plan --goal "study every file in src/* using seperate subagents per file and link the implementation as citations in the specification (flat structure)" && ./ds ./ralph/ralph plan --goal "okay i want you to come up with a plan that implements the specs/*.md and porting it to Go, that we will use for a claude code sessions - separetely to this. it's important to cite line numbers of specifications and the source code that will be affected. search all the code as needed using up to 10 claude then write the ralph/todo.md as bullet points" && ./ds ./ralph/ralph
```

### How it works

1. **Tests → specs** (plan, Opus). Compresses every test into a language-agnostic Markdown spec.
2. **Source → specs with citations** (plan, Opus). Documents every source file with citations back to `path:line`.
3. **Specs → todo** (plan, Opus). Distills both spec sets into `ralph/todo.md`: a prioritized porting plan.
4. **Todo → port** (build, Sonnet). Classic Ralph loop: one item per iteration, one commit per iteration.

Whenever Ralph emits `<promise>COMPLETE</promise>` or `<promise>NEXT</promise>`, it means the current iteration has finished and the following is about to start.

After all the iterations have finished, the project should have been ported to Go, and it should include tests and functionality from original code, working and tested by Ralph. Check for all of those to validate functionality.

Happy Ralphing.

---

## Attribution

- Method: [Geoffrey Huntley, *Porting code bases using AI*](https://ghuntley.com/porting/).
- Source repo: [mathiasbynens/punycode.js](https://github.com/mathiasbynens/punycode.js), MIT licensed.
- Ralph runner: this repo's `ralph/` directory.
