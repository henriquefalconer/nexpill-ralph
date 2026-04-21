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

### Setup

- [x] **1. Bootstrap Go module — constants, errors, Version**
  - Source: `punycode.js:1-32` (params + regexes + errors), `:425` (`version: '2.3.1'`)
  - Specs: `specs/src-constants.md:1-117`, `specs/src-error.md:1-116`, `specs/src-publicAPI.md:1-133`
  - Create: `go.mod` (`module github.com/<fork>/nexpill-ralph/punycode` or similar), `punycode.go` containing `package punycode`, `const Version = "2.3.1"`, bootstring parameters (`maxInt`, `base`, `tMin`, `tMax`, `skew`, `damp`, `initialBias`, `initialN`, `delimiter = '-'`), the precomputed `baseMinusTMin`, and the three sentinel `var Err* = errors.New(...)` with messages from `punycode.js:23-25`.
  - No tests yet. `go build ./...` must succeed.

### Pure helpers (leaf nodes — no inter-dependencies)

- [x] **2. Port `basicToDigit` + table-driven test**
  - Source: `punycode.js:144-155`
  - Spec: `specs/src-basicToDigit.md:1-155`
  - Add to `helpers.go`: `func basicToDigit(cp rune) int` returning `0..35` or `base` sentinel. Branches: digit `0x30-0x39` → `26 + (cp - 0x30)`; upper `0x41-0x5A` → `cp - 0x41`; lower `0x61-0x7A` → `cp - 0x61`; else `base`.
  - `helpers_test.go`: cover all four branches plus boundary code points (`0x2F`, `0x3A`, `0x40`, `0x5B`, `0x60`, `0x7B`, `0x80`).

- [x] **3. Port `digitToBasic` + test**
  - Source: `punycode.js:168-172`
  - Spec: `specs/src-digitToBasic.md:1-172`
  - Add to `helpers.go`: `func digitToBasic(digit int, flag bool) rune`. Go has no implicit bool→int; translate `75 * (digit < 26)` and `((flag != 0) << 5)` explicitly (e.g. `if digit < 26 { r += 75 }`; `if flag { r -= 32 }`).
  - Test: cover digits 0..35 with `flag=false` (the only flag used by callers at `punycode.js:359, 364`). Add one smoke case with `flag=true` to document the uppercase path even though callers never use it.

- [x] **4. Port `adapt` (RFC 3492 §3.4) + test**
  - Source: `punycode.js:179-187`
  - Spec: `specs/src-adapt.md:1-187`
  - Add to `helpers.go`: `func adapt(delta, numPoints int, firstTime bool) int`.
  - Preserve the exact structure of the JS loop; `delta >> 1` stays a bitwise shift. Convergence threshold is the precomputed `baseMinusTMin * tMax / 2 = 455` (avoid recomputing inside the loop).
  - Test: sanity vectors from RFC 3492 Appendix A worked examples (cite `specs/src-adapt.md` if it lists any); otherwise table of `(delta, numPoints, firstTime) → expected` crafted from running the JS reference.

### UCS-2 codec

- [ ] **5. Port `UCS2Decode` + tests**
  - Source: `punycode.js:101-123`
  - Specs: `specs/src-ucs2decode.md:1-60`, `specs/test-ucs2-decode.md:41-92` (7 vectors)
  - Create `ucs2.go`: `func UCS2Decode(s string) []rune`. Iterate over UTF-16 code units (not Go runes — we need the raw UCS-2 view). The simplest faithful port: range by `uint16` units via `utf16.Encode([]rune(s))` is **wrong** for inputs containing lone surrogates (Go's rune iteration replaces them with `\uFFFD`). Instead walk `s` decoding as UTF-8 to `rune`, then re-encode surrogate-bearing runes back to code units, OR iterate over `s` as `[]byte` and decode UTF-8 manually preserving surrogate semantics.
  - Expected decision: accept Go `string` as UTF-8; if input contains astral chars they arrive as single runes already. Simulate the JS surrogate-pair recombination only when the caller hands us a `[]uint16` — so expose an additional `UCS2DecodeUnits(units []uint16) []rune` that mirrors the JS algorithm exactly. Document this divergence in `doc.go`. **Confirm with user before committing** — test vectors at `specs/test-ucs2-decode.md:41-92` use raw UTF-16 sequences that Go `string` cannot represent natively.
  - `ucs2_test.go`: table-driven test exercising each of the 7 vectors in `specs/test-ucs2-decode.md`.

- [ ] **6. Port `UCS2Encode` + tests**
  - Source: `punycode.js:133`
  - Specs: `specs/src-ucs2encode.md:1-103`, `specs/test-ucs2-encode.md:19-48` (7 vectors + non-mutation)
  - `ucs2.go`: `func UCS2Encode(codePoints []rune) string` — build via `strings.Builder` + `WriteRune`. Go `[]rune` is a copy-on-pass slice header; non-mutation (the assertion at `tests/tests.js:282-287`) is automatic but add a test that verifies the input slice is unchanged after the call.
  - `ucs2_test.go`: reuse the 7 vectors + non-mutation assertion.

### Core codec (depends on helpers + UCS-2)

- [ ] **7. Port `Decode` + tests (includes all three error cases)**
  - Source: `punycode.js:196-281`
  - Specs: `specs/src-decode.md:1-328`, `specs/test-decode.md:34-316` (23 pass vectors + 1 invalid-input case at `tests/tests.js:302-309`), plus `specs/test-ucs2-decode.md:95-112` (two misfiled error tests actually targeting `punycode.decode` at `tests/tests.js:255-270`).
  - Create `decode.go`: `func Decode(input string) (string, error)`. Follow the structure at `punycode.js:196-281`:
    - Basic prefix detection via `strings.LastIndex(input, "-")`.
    - Main loop: variable-length integer decode with three overflow guards (`punycode.js:243-245`, `:255-257`, `:268-270`) — each returns `ErrOverflow`. Invalid base-36 digit returns `ErrInvalidInput` (`:235, 241`). Non-basic code point in prefix returns `ErrNotBasic` (`:216`).
    - Insertion at `:276` (`output.splice(i, 0, n)`) → in Go: `output = append(output[:i], append([]rune{n}, output[i:]...)...)`; pre-grow capacity once outside the loop.
    - Final assembly via `string(output)`.
  - `decode_test.go`: table with all 23 passing vectors from `tests/tests.js:7-136` (the `testData.strings` block), plus `decode("ZZZ") == "\u7BA5"` at `tests/tests.js:299-301`, plus three error-case subtests: `"\x81-"` → `ErrNotBasic`, `"\x81"` → `ErrOverflow`, `"ls8h="` → `ErrInvalidInput`.

- [ ] **8. Port `Encode` + tests**
  - Source: `punycode.js:290-376`
  - Specs: `specs/src-encode.md:1-286`, `specs/test-encode.md:18-136` (23 vectors — same fixture as Decode)
  - Create `encode.go`: `func Encode(input string) (string, error)`.
    - Step 1: run `UCS2Decode` on input (`punycode.js:294`).
    - Step 2: copy basic code points (`< 0x80`) to output, append `-` if any (`:305-320`).
    - Step 3: main loop — scan for next `m >= n` (`:328-332`); overflow guard at `:337-339` → `ErrOverflow`; inner overflow guard at `:345-347`.
    - Step 4: emit variable-length digits via `digitToBasic(..., false)` (lowercase-only; flag is always 0 in the JS source at `:359, 364`).
  - `encode_test.go`: same 23 vectors; no JS error-case tests exist for encode (the `ErrOverflow` paths are unreachable with realistic strings), so add one synthetic overflow test if feasible (extremely long string of distinct astral code points); otherwise comment that the error path is exercised transitively.

### Domain-aware wrappers (depend on Encode/Decode)

- [ ] **9. Port private `mapDomain` + separator handling**
  - Source: `punycode.js:62-86` (`mapDomain`), `:19` (`regexSeparators`), `:45-60` (`map` helper)
  - Specs: `specs/src-mapDomain.md:1-140`, `specs/src-map.md:1-98`
  - Add to `domain.go`: private `func mapDomain(domain string, fn func(string) (string, error)) (string, error)`.
    - Email split: `strings.SplitN(domain, "@", 2)` — preserve first part verbatim; only pass second part through label mapping. (Matches `punycode.js:73-80`.)
    - Separator normalization: `strings.NewReplacer("\u3002", ".", "\uFF0E", ".", "\uFF61", ".").Replace(rest)` — U+002E is already `.`, do not include it in the replacer. (Matches regex at `punycode.js:19`.)
    - Label split: `strings.Split(normalized, ".")`; apply `fn` per label; rejoin with `.`.
    - Drop the `map` helper from `punycode.js:53-60` — it existed only for IE8 compat. Use a plain range loop.
  - No public test for `mapDomain` directly; it's covered transitively by `ToASCII`/`ToUnicode` tests in the next two iterations.

- [ ] **10. Port `ToASCII` + tests**
  - Source: `punycode.js:408-414`, `regexNonASCII` at `:18` (`/[^\0-\x7F]/` — note this **excludes** U+007F DEL).
  - Specs: `specs/src-toASCII.md:1-126`, `specs/test-toASCII.md:27-91` (10 domain vectors + 23 ASCII-only pass-through + 4 separator vectors)
  - Add to `domain.go`: `func ToASCII(input string) (string, error)` delegating to `mapDomain`. Per-label callback: scan runes — if any rune is `> 0x7F`, return `"xn--" + encoded`; else return the label unchanged.
  - The DEL exclusion matters for the `tests/tests.js:216-219` vector (`"foo\x7F.example"` → unchanged). The check is strict `> 0x7F`, not `>= 0x7F`.
  - `domain_test.go`: the three test fixtures from `testData.domains`, `testData.strings` (pass-through assertion), `testData.separators` — all 10+23+4 cases.

- [ ] **11. Port `ToUnicode` + tests**
  - Source: `punycode.js:389-395`, `regexPunycode` at `:17` (`/^xn--/`, **case-sensitive**, lowercase only).
  - Specs: `specs/src-toUnicode.md:1-113`, `specs/test-toUnicode.md:27-84` (10 domain vectors + 23 pass-through)
  - Add to `domain.go`: `func ToUnicode(input string) (string, error)` delegating to `mapDomain`. Per-label callback: if `strings.HasPrefix(label, "xn--")`, decode `strings.ToLower(label[4:])`; else return the label unchanged. Case sensitivity intentional — `XN--foo` does not decode, matching JS.
  - `domain_test.go`: `testData.domains` (decoded ↔ encoded) plus `testData.strings` pass-through assertions.

### Polish

- [ ] **12. Centralize test fixture (`testdata_test.go`)**
  - Specs: `specs/test-vectors.md:10-200` (master index of all four fixture blocks)
  - Mirror the JS `testData` object at `tests/tests.js:6-243` as a single package-internal Go value: `var testVectors = struct{ Strings, UCS2, Domains, Separators []vector }{ ... }`. Each subtest (`TestEncode`, `TestDecode`, `TestToASCII`, `TestToUnicode`, `TestUCS2Decode`, `TestUCS2Encode`) iterates this one fixture instead of inlining vectors. Prevents drift between test files.
  - Only after iterations 5–11 have landed their inline vectors. This iteration is a pure refactor; no new behavior.

- [ ] **13. Add package documentation (`doc.go`)**
  - Specs: `specs/src-publicAPI.md:1-133`
  - Create `doc.go` with a package-level comment citing RFC 3492 and RFC 5891, and summarizing the public surface. Each exported symbol (`Version`, `Encode`, `Decode`, `ToASCII`, `ToUnicode`, `UCS2Encode`, `UCS2Decode`, the three `Err*` sentinels) gets a Go-doc comment starting with its own name. Link to the upstream `mathiasbynens/punycode.js` and the commit hash that corresponds to version 2.3.1.

- [ ] **14. Final verification + CI hook**
  - Run `go build ./...`, `go test -race ./...`, `go vet ./...`. All three must pass.
  - Add a minimal `.github/workflows/go.yml` (or extend existing) running the same on push. Use `go 1.22` (matches Go's stable release at spec time; no specific version pinned in the source repo).
  - Cross-check every vector count against the digest numbers: `testData.strings` = 24 entries (`tests/tests.js:7-136`), `testData.ucs2` = 7 (`:137-175`), `testData.domains` = 10 (`:176-220`), `testData.separators` = 4 (`:221-242`). If any table drifts from these counts, stop and reconcile.

## Out of scope (intentionally skipped)

- **`scripts/prepublish.js`** (`specs/src-scripts-prepublish.md:1-250`) — generates an ES6 module from CommonJS source. Irrelevant to a Go target.
- **`ds` shell script** — project tooling (Docker sandbox launcher), not library source.
- **`tests/tests.js`** as a standalone port — its content is absorbed into the per-iteration `_test.go` files + the consolidated `testdata_test.go` in iteration 12. No 1:1 file translation.
- **`punycode.es6.js`** (generated artifact) — don't regenerate in Go; the Go package *is* the artifact.
