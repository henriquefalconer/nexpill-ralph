# Ralph todo — Port `punycode.js` to Go

One bullet = one build iteration = one commit. Items are ordered by dependency: land a bullet only when every bullet above it is green (`go build` + `go test ./...`). Each bullet cites (a) the governing spec sections in `specs/**` and (b) the source lines in `punycode.js` / `tests/tests.js` that the Go code must mirror.

The port lives at `port/` (new directory, a Go module — see bullet 1). There is intentionally no `src/` in the JS repo — the library sits at project root (`specs/README.md:5`, `specs/src-punycode.md:12`). Do NOT modify `punycode.js`, `tests/tests.js`, `scripts/prepublish.js`, `package.json`, `specs/**`, or any other existing file — only add files under `port/`.

---

## Ground rules (apply to every bullet)

- **Exact-match citations.** Every behavioural decision MUST be verifiable against a cited `punycode.js:line` or `tests/tests.js:line`. When spec prose and source disagree, source wins — flag the discrepancy in the commit message.
- **One-package layout.** Everything ships as `package punycode` under `port/`, one directory, flat — mirror the JS one-file shape (`specs/src-punycode.md:10-17`). Files may be split by topic (`constants.go`, `ucs2.go`, `bootstring.go`, `decode.go`, `encode.go`, `domain.go`, `punycode.go`) but no sub-packages.
- **Go signatures.**
  - `func Decode(input string) (string, error)` — mirror `punycode.js:196-281` (the `xn--` prefix is the caller's job; `specs/src-punycode.md:154-155`).
  - `func Encode(input string) (string, error)` — mirror `punycode.js:283-376`.
  - `func ToUnicode(input string) string` — mirror `punycode.js:389-395`. Returns a string (errors from `Decode` are swallowed per the JS behaviour at `punycode.js:391-393`; ugly labels pass through on failure-free paths only — match `tests/tests.js:323-344`).
  - `func ToASCII(input string) (string, error)` — mirror `punycode.js:408-414`. Returns an error because the inner `Encode` can overflow.
  - `func UCS2Decode(input string) []rune` — mirror `punycode.js:101-123`. Input is a Go string. See bullet 6 for the UTF-16 preservation strategy.
  - `func UCS2Encode(codePoints []rune) string` — mirror `punycode.js:125-133`.
- **Error values, not messages.** Export three sentinel errors mirroring `errors` at `punycode.js:22-26`: `ErrOverflow`, `ErrNotBasic`, `ErrInvalidInput`. Callers may `errors.Is(err, punycode.ErrOverflow)`. The messages must match the JS strings verbatim (`'Overflow: input needs wider integers to process'`, `'Illegal input >= 0x80 (not a basic code point)'`, `'Invalid input'`) so test output is grep-compatible.
- **Integer width.** All Bootstring arithmetic runs on `int32` (32-bit signed). The five overflow guards (`punycode.js:243-245, 255-257, 268-270, 337-339, 345-347`) MUST fire at the same boundaries they fire in JS. Widening to `int64` and deleting the guards changes observable error behaviour — do NOT do this (`specs/src-punycode.md:311-323`).
- **Lowercase-only emission.** `digitToBasic` is called with `flag=0` at both sites (`punycode.js:358, 364`; `specs/src-punycode.md:135-138`). The port may drop the flag parameter. The Russian fixture `b1abfaaepdrnnbgefbadotcwatmq2g4l` at `tests/tests.js:74-82` asserts lowercase output.
- **Case-insensitive input decode.** `basicToDigit` folds `A-Z` and `a-z` to the same digits (`punycode.js:148-153`). The hardcoded `ZZZ → U+7BA5` assertion at `tests/tests.js:299-301` verifies this — every iteration that touches decode must keep this green.
- **UTF-16 fidelity for `ucs2`.** Go strings are UTF-8; the port MUST expose a UTF-16 round-trip that preserves lone surrogates (`0xD800..0xDFFF`). The seven fixtures at `tests/tests.js:137-175` are the canonical stress test (`specs/punycode-ucs2-decode.md:13-21`, `specs/punycode-ucs2-encode.md:13-27`). See bullet 6 for the chosen strategy.
- **Commit discipline.** Every bullet is one commit. Commit message first line names the bullet number and feature. Body lists (i) the spec/source lines the commit implements, (ii) which Go files were added/changed, (iii) which tests pass.

---

## Bullets

### 1. Scaffold the Go module at `port/`

- Create `port/go.mod` with `module github.com/mathiasbynens/punycode` and `go 1.22`. No external dependencies (`specs/src-punycode.md:14-15` — the JS library is zero-dep and the Go port must stay zero-dep using only `unicode/utf16`, `strings`, `errors`, `fmt`, `testing`, `testing/fstest` as needed).
- Create `port/doc.go` with a package doc-comment summarising the public surface (`Decode`, `Encode`, `ToASCII`, `ToUnicode`, `UCS2Decode`, `UCS2Encode`, and the three sentinel errors). Reference RFC 3492 — the JS README does (`SOURCE-README.md:3`).
- Create `port/LICENSE` as a copy of `LICENSE-MIT.txt` so the port is distributable standalone (the JS module ships it per `specs/src-scripts-prepublish.md:80-83`).
- Add `port/README.md` with a one-paragraph pointer back to `specs/README.md` and the mapping "Go symbol → JS symbol → `punycode.js:line`".
- Acceptance: `cd port && go build ./...` succeeds; `go vet ./...` is clean. No tests yet.

### 2. Land Bootstring constants + sentinel errors

- Add `port/constants.go` translating `punycode.js:3-31`:
  - `maxInt int32 = 2147483647` — `punycode.js:4`.
  - `base int32 = 36` — `punycode.js:7`.
  - `tMin int32 = 1` — `punycode.js:8`.
  - `tMax int32 = 26` — `punycode.js:9`.
  - `skew int32 = 38` — `punycode.js:10`.
  - `damp int32 = 700` — `punycode.js:11`.
  - `initialBias int32 = 72` — `punycode.js:12`.
  - `initialN int32 = 128` — `punycode.js:13`.
  - `delimiter byte = '-'` — `punycode.js:14`.
  - `baseMinusTMin int32 = 35` — derived, `punycode.js:29`.
- Add `port/errors.go` with three exported sentinels (`ErrOverflow`, `ErrNotBasic`, `ErrInvalidInput`) via `errors.New` using the exact JS message strings (`punycode.js:23-25`).
- Do NOT translate the three regexes or `errors` object key-lookup indirection — the Go port inlines them where needed (see bullets 5, 10).
- Acceptance criteria: unit tests at `port/constants_test.go` assert each constant's numeric value matches RFC 3492 §5 (`specs/src-punycode.md:24-39`).
- Spec: `specs/src-punycode.md:22-61`.

### 3. Implement `basicToDigit` (private helper)

- Add `port/bootstring.go` — the private Bootstring arithmetic layer.
- Port `basicToDigit(codePoint int32) int32` per `punycode.js:135-155`, with the four-branch table at `specs/src-punycode.md:114-123`:
  - `0x30..0x39` → digit `26..35`.
  - `0x41..0x5A` → digit `0..25`.
  - `0x61..0x7A` → digit `0..25`.
  - otherwise → `base` (36, the sentinel).
- Unit tests in `port/bootstring_test.go`: cover each branch boundary (`'0'`, `'9'`, `'A'`, `'Z'`, `'a'`, `'z'`, `'/'`, `':'`, `'@'`, `'['`, `'`'`, `'{'`, `0x80`) — asserts case-insensitive folding (the precondition for the `ZZZ → U+7BA5` test at `tests/tests.js:299-301`).
- Keep the function unexported — only `Decode` calls it (`punycode.js:238`).
- Spec: `specs/src-punycode.md:112-123`.

### 4. Implement `digitToBasic` (private helper)

- In `port/bootstring.go`, port `digitToBasic(digit int32) byte` per `punycode.js:157-172`.
- Drop the `flag` parameter — both call sites in `punycode.js:358, 364` pass `0` and no test covers mixed-case output (`specs/src-punycode.md:135-138`).
- Formula: `digit + 22 + 75 * (digit < 26 ? 1 : 0)` — returns ASCII byte in `'a'..'z'` for digits `0..25` and `'0'..'9'` for digits `26..35`.
- Unit tests cover every digit `0..35` and assert the output byte matches the branchless formula and the spec table at `specs/src-punycode.md:126-138`.

### 5. Implement `adapt` (private helper)

- In `port/bootstring.go`, port `adapt(delta, numPoints int32, firstTime bool) int32` per `punycode.js:174-187`.
- Exact step sequence (`specs/src-punycode.md:142-150`):
  1. `k := 0`.
  2. If `firstTime`: `delta = delta / damp` else `delta = delta >> 1`.
  3. `delta += delta / numPoints`.
  4. While `delta > ((base - tMin) * tMax) >> 1` (== 455): `delta = delta / (base - tMin); k += base`.
  5. Return `k + (base - tMin + 1) * delta / (delta + skew)`.
- Use Go integer division — it truncates toward zero, matching `Math.floor` on non-negative ints (which is all we ever pass).
- Unit tests in `port/bootstring_test.go`: at minimum, assert `adapt(0, 1, true) == 0`, `adapt(700, 1, true) == 0`, and the known-good intermediate values from the RFC 3492 §7.1 encode traces.

### 6. Implement the UCS-2 / UTF-16 layer

- Add `port/ucs2.go` with:
  - `func UCS2Decode(input string) []rune` — mirrors `punycode.js:101-123` (`specs/punycode-ucs2-decode.md:13-21`).
  - `func UCS2Encode(codePoints []rune) string` — mirrors `punycode.js:125-133` (`specs/punycode-ucs2-encode.md:13-29`).
- **UTF-16 strategy.** Treat the Go string as UTF-16 in a `[]uint16` shadow buffer. Convert the input `string` → `[]uint16` by the fixture-preserving path described in `specs/punycode-ucs2-decode.md:31-34` and `specs/punycode-ucs2-encode.md:35-38`: round-trip through `utf16.Encode(utf16.Decode(...))` is NOT acceptable because it would collapse lone surrogates. Instead, for `UCS2Decode`:
  1. Iterate the input runes via `utf8.DecodeRuneInString` and push each rune whose value is in the full `0..0x10FFFF` space.
  2. For runes in `0x10000..0x10FFFF`, still push one `rune` (they are already supplementary scalars — Go's UTF-8 decoder emits them whole, so the fixture at `tests/tests.js:140-144` just works).
  3. Preserve the option for callers to pass a pre-encoded `[]uint16` via an alternate `UCS2DecodeUTF16(cu []uint16) []rune` constructor for the seven lone-surrogate fixtures at `tests/tests.js:145-174` — mirror the behaviour at `punycode.js:107-121`: high surrogate without a matching low surrogate is pushed as its raw code-unit integer and the cursor rewinds (`specs/punycode-ucs2-decode.md:14-18`).
- **`UCS2Encode` strategy.** Convert `[]rune` → `[]uint16` following `specs/punycode-ucs2-encode.md:37`: values `<= 0xFFFF` append as one `uint16`; values `> 0xFFFF` split into high/low surrogate pair; values in `0xD800..0xDFFF` pass through as one `uint16`. Return the `[]uint16` re-interpreted as a string via `string(utf16.Decode(cu))` **only when the caller is known-safe** — for the round-trip fixture path, expose a `UCS2EncodeUTF16(codePoints []rune) []uint16` that callers can ship as-is.
- Tests in `port/ucs2_test.go`: translate all 7 `testData.ucs2` records (`tests/tests.js:137-175`) into Go fixtures and assert round-trip both ways (`specs/punycode-ucs2-decode.md:13-21`, `specs/punycode-ucs2-encode.md:31-32`). Also verify input immutability of `UCS2Encode` (`tests/tests.js:282-287`, `specs/punycode-ucs2-encode.md:27`).
- Spec: `specs/src-punycode.md:82-108`, `specs/punycode-ucs2-decode.md`, `specs/punycode-ucs2-encode.md`.

### 7. Implement `Decode` (Bootstring decoder)

- Add `port/decode.go` with `func Decode(input string) (string, error)` mirroring `punycode.js:196-281` step-by-step.
- State: `output []int32` (decoded code points), cursor `i int32 = 0`, `n int32 = initialN`, `bias int32 = initialBias` — `specs/src-punycode.md:157-163`, `punycode.js:196-202`.
- **Basic-prefix copy** (`punycode.js:208-219`, `specs/src-punycode.md:164-170`):
  - `basic := strings.LastIndexByte(input, '-')`; if `< 0` set `basic = 0`.
  - For `j := 0; j < basic; j++`: if `input[j] >= 0x80` return `"", ErrNotBasic`. Otherwise append `int32(input[j])` to `output`. Use `input[j]` (byte), not rune iteration, because the ASCII guard makes UTF-8 multi-byte sequences impossible here.
- **Main loop** (`punycode.js:224-278`, `specs/src-punycode.md:173-198`):
  - `index := basic > 0 ? basic + 1 : 0`.
  - While `index < len(input)`: variable-length integer read (inner loop at `punycode.js:231-261`) — on empty input return `ErrInvalidInput` (`punycode.js:234-236`), on `digit >= base` return `ErrInvalidInput` (`punycode.js:240-242`), and enforce the two inner-loop overflow guards (`punycode.js:243-245, 255-257`).
  - After each variable-length integer: `out := int32(len(output)) + 1`; `bias = adapt(i - oldi, out, oldi == 0)`; enforce overflow guard C (`punycode.js:268-270`, `floor(i / out) > maxInt - n`); `n += i / out`; `i %= out`.
  - Insert `n` at position `i` into `output` (Go slice splice), then `i++`.
- **Finalise** (`punycode.js:280`): return `UCS2Encode(output)`.
- Tests in `port/decode_test.go`:
  - For every `testData.strings` fixture (tests/tests.js:7-136, per `specs/test-data-fixtures.md`), assert `Decode(encoded) == decoded` (`tests/tests.js:291-297`).
  - Assert `Decode("ZZZ") == "\u7BA5"` — the case-insensitive spot check at `tests/tests.js:299-301`.
  - Assert `Decode("\x81-")` returns `ErrNotBasic` and `Decode("\x81")` returns `ErrOverflow` (`tests/tests.js:255-270`, `specs/punycode-ucs2-decode.md:26-27` — two misplaced-throw tests).
  - Assert `Decode("ls8h=")` returns `ErrInvalidInput` (`tests/tests.js:302-309`).
- Spec: `specs/punycode-decode.md`, `specs/src-punycode.md:152-207`.

### 8. Implement `Encode` (Bootstring encoder)

- Add `port/encode.go` with `func Encode(input string) (string, error)` mirroring `punycode.js:283-376`.
- State: `output []byte`; `codePoints := UCS2Decode(input)`; `n int32 = initialN`; `delta int32 = 0`; `bias int32 = initialBias` — `punycode.js:290-302`, `specs/src-punycode.md:213-220`.
- **Basic-code-point copy** (`punycode.js:304-312`, `specs/src-punycode.md:222-224`): for each `cp` in `codePoints`: if `cp < 0x80` append `byte(cp)` to `output`. Record `basicLength := int32(len(output))` and `handledCPCount := basicLength`.
- **Delimiter emission** (`punycode.js:318-320`): if `basicLength > 0` append `delimiter`.
- **Main loop** (`punycode.js:323-374`, `specs/src-punycode.md:230-256`):
  - While `handledCPCount < int32(len(codePoints))`:
    - Find `m` = smallest `codePoint >= n` (init to `maxInt`) — `punycode.js:327-332`.
    - Enforce overflow guard D (`punycode.js:337-339`): if `m - n > (maxInt - delta) / (handledCPCount + 1)` return `ErrOverflow`.
    - `delta += (m - n) * (handledCPCount + 1)`; `n = m`.
    - For each `cp` in `codePoints`:
      - If `cp < n`: `delta++`; enforce overflow guard E (`punycode.js:345-347`): if `delta > maxInt` return `ErrOverflow` — note that post-increment on `int32` wraps silently, so guard BEFORE incrementing (check `delta == maxInt` pre-bump) or detect overflow via a widened temporary.
      - If `cp == n`: inner generalised variable-length integer emit loop (`punycode.js:350-362`); trailing `digitToBasic(q)` append (`punycode.js:364`); `bias = adapt(delta, handledCPCount+1, handledCPCount == basicLength)`; `delta = 0`; `handledCPCount++`.
    - `delta++; n++`.
- **Finalise** (`punycode.js:375`): return `string(output), nil`.
- Tests in `port/encode_test.go`:
  - For every `testData.strings` fixture, assert `Encode(decoded) == encoded` (`tests/tests.js:313-319`).
  - Add a round-trip property test: `Decode(Encode(decoded)) == decoded` for every `testData.strings` fixture — belts-and-braces.
- Spec: `specs/punycode-encode.md`, `specs/src-punycode.md:209-267`.

### 9. Implement `mapDomain` (private helper) with IDNA2003 separator normalisation + email splitting

- Add `port/domain.go` with `func mapDomain(domain string, callback func(label string) (string, error)) (string, error)` mirroring `punycode.js:62-86`.
- Algorithm (`specs/src-punycode.md:74-80`):
  1. `parts := strings.Split(domain, "@")`. If `len(parts) > 1`: `prefix := parts[0] + "@"`; `domain = parts[1]` — silently discards `parts[2:]` (untested edge case per `specs/src-punycode.md:331`).
  2. Normalise separators: replace U+3002, U+FF0E, U+FF61 with U+002E (`punycode.js:19, 82`). Use `strings.NewReplacer` with the four separator literals → `"."`.
  3. `labels := strings.Split(domain, ".")`.
  4. Apply `callback` to each label; collect results; return `prefix + strings.Join(results, ".")`.
  5. If any callback returns an error, short-circuit and return it — JS has no analogous path because `toUnicode` never throws and `toASCII` can throw only via `encode` (which does throw).
- Unit tests in `port/domain_test.go`: cover the four separators (`specs/test-data-fixtures.md` separators bucket, `tests/tests.js:221-242`); cover email splitting with `foo@bar.com` → prefix `"foo@"`, domain `"bar.com"`; cover plain domain with no `@`.
- Do NOT re-implement the private JS `map` (`punycode.js:45-60`) — use the Go `for range` idiom. The JS reverse-walk is non-observable (`specs/src-punycode.md:70`).

### 10. Implement `ToUnicode` + `ToASCII` wrappers

- In `port/punycode.go` add:
  - `func ToUnicode(input string) string` — mirror `punycode.js:389-395`.
    - Callback: if `strings.HasPrefix(label, "xn--")` (case-sensitive, matching `regexPunycode` at `punycode.js:17` — `specs/src-punycode.md:44`), then `Decode(strings.ToLower(label[4:]))`. If `Decode` returns an error, the JS behaviour at `punycode.js:391-393` would throw — but `toUnicode` is called by no test on malformed input, so `ToUnicode` may return the label unchanged on error (match JS semantics by returning `label` when `Decode` fails — belt-and-braces; document the choice in a one-line comment).
    - Else return the label unchanged (idempotence — `specs/punycode-to-unicode.md:8, 28`).
  - `func ToASCII(input string) (string, error)` — mirror `punycode.js:408-414`.
    - Callback: if the label contains any byte `>= 0x80` (mirror `regexNonASCII` at `punycode.js:18`; NOTE: DEL `0x7F` is ASCII per the inline comment and `tests/tests.js:216-219`), then `"xn--" + Encode(label)`. Else return the label unchanged.
- Use bullet-9's `mapDomain` for both.
- Tests in `port/wrappers_test.go`:
  - Domain fixtures (`tests/tests.js:176-220`, ten records): assert `ToUnicode(encoded) == decoded` and `ToASCII(decoded) == encoded` (`tests/tests.js:324-330, 347-353`).
  - Identity loop over all `testData.strings` (`tests/tests.js:332-343, 355-361`) — both wrappers must be the identity because the fixtures are pure ASCII and none starts with `xn--`.
  - Separator-normalisation loop (`tests/tests.js:363-370`): assert `ToASCII("mañana"+sep+"com") == "xn--maana-pta.com"` for all four separators (`tests/tests.js:221-242`).
  - DEL passthrough: `ToASCII("foo\x7f.example")` is identity (`tests/tests.js:216-219`).
  - Trailing-dot preservation: `ToASCII("example.com.") == "example.com."` and `ToUnicode("example.com.") == "example.com."` (`tests/tests.js:181-184`).
  - Email preservation: `ToUnicode("джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq") == "джумла@джpумлатест.bрфa"` and reverse (`tests/tests.js:211-215`).
- Spec: `specs/punycode-to-ascii.md`, `specs/punycode-to-unicode.md`, `specs/src-punycode.md:269-291`.

### 11. Port the `testData` fixture table to Go

- Add `port/testdata_test.go` with four `var` tables mirroring `testData` at `tests/tests.js:6-243`:
  - `stringsFixtures` — 25 records, schema `{Description, Decoded, Encoded string}` (`tests/tests.js:7-136`, `specs/test-data-fixtures.md` strings section).
  - `ucs2Fixtures` — 7 records, schema `{Description string; Decoded []int32; Encoded []uint16}` (the `Encoded` must be `[]uint16` to carry lone surrogates — JS strings conflate UTF-16 with strings, Go cannot). Source: `tests/tests.js:137-175`.
  - `domainFixtures` — 10 records, schema `{Description, Decoded, Encoded string}` (`tests/tests.js:176-220`).
  - `separatorFixtures` — 4 records, schema `{Description, Decoded, Encoded string}` (`tests/tests.js:221-242`).
- Keep the `Description` verbatim including any GitHub-issue references embedded in the JS (`tests/tests.js:181, 193, 197, 201, 211, 216` — the inline `// See …` comments must become Go `//` comments next to the record so grep-traceability is preserved).
- This is the single source of truth for tests in bullets 6, 7, 8, 10 — refactor those tests to consume these tables on the iteration that lands this bullet (if they were hardcoded in earlier bullets, collapse the duplication here).
- Spec: `specs/test-data-fixtures.md`, `specs/src-tests-tests.md`.

### 12. Mirror the six `describe` blocks as Go test functions

- In `port/punycode_test.go` (or one test file per unit — either works), define exactly six test functions, one per JS `describe` block (`tests/tests.js:245-371`):
  - `TestUCS2Decode` — iterates `ucs2Fixtures`, asserts `UCS2Decode(Encoded) == Decoded` (`tests/tests.js:245-271`, ignoring the two misplaced RangeError tests which belong with `TestDecode`).
  - `TestUCS2Encode` — iterates `ucs2Fixtures`, asserts `UCS2Encode(Decoded) == Encoded` AND that the input `Decoded` slice is unchanged after the call (`tests/tests.js:273-288`, `specs/punycode-ucs2-encode.md:27` / `tests/tests.js:282-287`).
  - `TestDecode` — iterates `stringsFixtures`; adds the `ZZZ` uppercase test; adds the three error-path subtests from `tests/tests.js:255-270, 302-309`.
  - `TestEncode` — iterates `stringsFixtures`.
  - `TestToUnicode` — iterates `domainFixtures`, then iterates `stringsFixtures` for the identity loop (`tests/tests.js:332-343`).
  - `TestToASCII` — iterates `domainFixtures`, then the `stringsFixtures` identity loop, then `separatorFixtures`.
- Use `t.Run(fixture.Description, …)` sub-tests so failures report by description — mirrors the JS `it(…, ...)` naming (`specs/src-tests-tests.md`).
- All six must pass on `go test ./port/...`.

### 13. Repository polish

- Add `port/go.sum` (empty or absent — there are no deps; `go mod tidy` will touch-and-leave).
- Add `.github/workflows/go.yml` (new workflow, does NOT modify any existing workflow file) that runs `go build ./port/... && go test ./port/...` on push/PR. Match Go version `1.22` (bullet 1).
- Add a row to `specs/README.md` ONLY IF mentioning a new port location would help a porter — but since `specs/README.md:3` already says "It is the source-of-truth for the Go port at `port/`", no edit is required. Verify by re-reading before the commit.
- Acceptance: whole repo `go test ./port/...` is green in CI.

---

## Out of scope

- `scripts/prepublish.js` and `punycode.es6.js` — JS-only build step with no Go analogue (`specs/src-scripts-prepublish.md:86-95`).
- `punycode.version` (`punycode.js:425`) — not under test and explicitly out of scope (`specs/src-punycode.md:309`). Skipping means we don't ship a matching `Version` constant.
- The five intentional scope gaps catalogued at `specs/src-punycode.md:325-335`:
  1. `regexPunycode` uppercase `XN--` — our port matches JS (case-sensitive) but no test pins it.
  2. DEL mixed with non-ASCII — untested; match JS behaviour (DEL is ASCII).
  3. Multi-`@` emails — untested; match JS behaviour (silently drop parts[2:]).
  4. `toASCII` on `xn--…` with non-ASCII content — untested; match JS behaviour (double-encode into `xn--xn--…`).
  5. Encode-side overflow guards D and E — untested but must be implemented (bullet 8) because dropping them changes observable behaviour.

## Done when

- `cd port && go build ./...` and `go test ./...` are both green.
- All 46 JS test fixtures (25 strings + 7 ucs2 + 10 domains + 4 separators — `specs/test-data-fixtures.md`) assert in the Go suite with the same expected values.
- All three error paths (`ErrNotBasic`, `ErrInvalidInput`, `ErrOverflow`) are covered by at least one test each.
- `go vet ./port/...` is clean; `gofmt -l port/` outputs nothing.
