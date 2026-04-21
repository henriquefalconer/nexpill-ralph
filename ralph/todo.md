# Ralph TODO — punycode.js → Go port

Dependency-ordered build queue. One bullet ≈ one Ralph build iteration ≈ one commit. Each item depends only on items above it. Every bullet cites the specs it implements and the `punycode.js` lines it mirrors, so Claude Code sessions can work in isolation from this file + the `specs/` directory.

## Design decisions (apply to every item below)

- **Package layout**: single Go module, single `package punycode`. Files split by logical unit for readability, not encapsulation.
- **Integer type**: use `int` (≥ 32-bit on every Go target) and bounds-check against `maxInt = 2147483647` explicitly at the same JS guard points. Do **not** use `int32` — intermediate multiplications would silently wrap before the guards run. Right-shift `>> 1` on non-negative `int` matches JS semantics.
- **Floor division**: JS `Math.floor(a/b)` with non-negative operands == Go `a / b`. All `floor(...)` calls in `punycode.js` are on non-negative values — plain `/` is correct. Do not introduce a `floor` helper.
- **Errors**: three sentinel values mirroring `errors` object at `punycode.js:22-26`: `ErrOverflow`, `ErrNotBasic`, `ErrInvalidInput`. Public functions return `(string, error)`; `UCS2Decode` / `UCS2Encode` have no error path.
- **Public API casing**: idiomatic Go — `Encode`, `Decode`, `ToASCII`, `ToUnicode`, `UCS2Decode`, `UCS2Encode`, `Version`. The JS nested namespace `punycode.ucs2.{encode,decode}` flattens to `UCS2Encode`/`UCS2Decode`; document the mapping in the package doc.
- **Regex replacements**: prefer `strings` primitives over `regexp`. `regexSeparators` (`punycode.js:19`) → `strings.NewReplacer`; `regexNonASCII` (`punycode.js:18`) → `for _, r := range s { if r > 0x7F … }`; `regexPunycode` (`punycode.js:17`) → `strings.HasPrefix(label, "xn--")` (case-sensitive, matching JS).
- **Commit discipline**: each iteration leaves `go build ./...` and `go test ./...` green. A new unit lands with its own `_test.go` in the same commit.

## Environment notes

- `go test -race ./...` requires cgo + gcc, which are absent in the sandbox. Use `go test ./...` for verification instead. `go build ./...` and `go vet ./...` are always available.

## Iterations

All 14 items complete. Port is finished.

- [x] **14. Final verification + CI hook**
  - `go build ./...`, `go test ./...`, `go vet ./...` all pass green (`go test -race` skipped — cgo/gcc absent in sandbox).
  - Vector counts verified: Strings=24, UCS2=7, Domains=10, Separators=4 — all match spec.
  - `.github/workflows/go.yml` created locally (runs `go build ./...`, `go vet ./...`, `go test ./...` on push/PR using go 1.22) but **push to GitHub was blocked** — the GitHub PAT lacks `workflow` scope, which is required to push files under `.github/workflows/`.
  - **Action required to complete CI setup (choose one) — STILL BLOCKED as of iteration 10:**
    - **(a)** Add `workflow` scope to the GitHub PAT and re-push the branch, or
    - **(b)** Push `.github/workflows/go.yml` manually (e.g. via the GitHub web UI or a PAT with the correct scope).
  - `.github/workflows/go.yml` exists locally and is correct; every push attempt is rejected by GitHub with "refusing to allow a Personal Access Token to create or update workflow … without `workflow` scope".

## Out of scope (intentionally skipped)

- **`scripts/prepublish.js`** (`specs/src-scripts-prepublish.md:1-250`) — generates an ES6 module from CommonJS source. Irrelevant to a Go target.
- **`ds` shell script** — project tooling (Docker sandbox launcher), not library source.
- **`tests/tests.js`** as a standalone port — its content is absorbed into the per-iteration `_test.go` files + the consolidated `testdata_test.go` in iteration 12. No 1:1 file translation.
- **`punycode.es6.js`** (generated artifact) — don't regenerate in Go; the Go package *is* the artifact.
