# Ralph Todo — Go Port Plan

Porting `punycode.js` to Go. Each bullet below is **one build iteration = one commit**. Items are listed **in dependency order** — do them top-to-bottom. Each item cites both the spec file and the JS source it derives from, and names the Go file(s) it should produce or modify.

Target layout (established by the first iteration, assumed by every later one):

- `go/go.mod` — module declaration
- `go/punycode.go` — library implementation (package `punycode`)
- `go/punycode_test.go` — full test suite
- `go/testdata.go` (internal to the test package) — ported test fixture tables
- Existing JS files (`punycode.js`, `tests/tests.js`, `scripts/prepublish.js`) remain untouched as reference.

---

## Priority queue

### Phase 0 — Scaffolding

- **Initialize Go module and package skeleton.** Create `go/go.mod` (`module github.com/mathiasbynens/punycode` or similar; `go 1.22`). Create empty `go/punycode.go` with `package punycode` and a file-level doc comment mirroring `SOURCE-README.md` intent. Create empty `go/punycode_test.go` with `package punycode_test` (black-box) or `package punycode` depending on whether internal helpers will be unit-tested. Verify `go build ./...` and `go test ./...` succeed with zero files in the package. Cites: `specs/punycode-module.md:1-30`, `package.json` (for the `"version": "2.3.1"` string).

### Phase 1 — Module-level constants, errors, regex equivalents

- **Port Bootstring constants and the `maxInt` sentinel.** In `go/punycode.go`, declare as `const`: `base = 36`, `tMin = 1`, `tMax = 26`, `skew = 38`, `damp = 700`, `initialBias = 72`, `initialN = 128`, `delimiter = '-'`, `maxInt = 2147483647`, `baseMinusTMin = base - tMin`. Also declare the library version constant `version = "2.3.1"`. Cites: `punycode.js:4-14,29`, `specs/punycode-module.md:19-34,81-89`.

- **Port error sentinels / `error(type)` helper.** Declare three exported error values (`ErrOverflow`, `ErrNotBasic`, `ErrInvalidInput`) wrapping the exact message strings from the JS `errors` table. Decide between `panic(ErrXxx)` (mirrors JS `throw`, simpler for pure-function ports) and returning `(string, error)` from every public function (idiomatic Go). Recommendation: return `error` from public API, and have internal helpers use `panic` caught by a top-level `recover` in `decode`/`encode` to keep the inner loops clean. Cites: `punycode.js:22-26,41-43`, `specs/punycode-internal-helpers.md:7-35`, `specs/punycode-module.md:69-79`.

- **Port separator-set and ASCII-prefix helpers (regex replacements).** Add three helper functions instead of compiling `regexp.Regexp` objects: `hasPunycodePrefix(s string) bool` (replaces `/^xn--/`, must be case-sensitive — use `strings.HasPrefix(s, "xn--")`), `hasNonASCII(s string) bool` (replaces `/[^\0-\x7F]/`, iterate runes and return true if any rune `> 0x7F`), and `normalizeSeparators(s string) string` (replaces `regexSeparators = /[\x2E\u3002\uFF0E\uFF61]/g`, use `strings.Map` or a small switch to collapse U+002E, U+3002, U+FF0E, U+FF61 → `'.'`). Cites: `punycode.js:17-19`, `specs/punycode-module.md:36-67`.

### Phase 2 — Pure Bootstring helpers

- **Port `basicToDigit(cp int) int`.** Pure function. Map ASCII code point → Bootstring digit (0–35) or `base` (=36) for invalid. Ranges from `punycode.js:144-155`: `0x30-0x39` (`'0'-'9'`) → `26-35`; `0x41-0x5A` (`'A'-'Z'`) → `0-25`; `0x61-0x7A` (`'a'-'z'`) → `0-25`; else → `base`. Cites: `punycode.js:144-155`, `specs/punycode-bootstring-digit.md:7-60`.

- **Port `digitToBasic(digit int, flag int) int`.** Pure function. Formula `digit + 22 + 75*(digit<26) - ((flag!=0)<<5)`. In Go: `result := digit + 22; if digit < 26 { result += 75 }; if flag != 0 { result -= 32 }`. Note: `encode()` only calls with `flag = 0` (`punycode.js:359,364`), so the uppercase branch is dead code from the public API but must remain faithful. Cites: `punycode.js:168-172`, `specs/punycode-bootstring-digit.md:64-114`.

- **Port `adapt(delta, numPoints int, firstTime bool) int`.** RFC 3492 §3.4 bias adaptation. Logic: (1) if `firstTime`, `delta /= damp` (=700); else `delta >>= 1` (= `delta/2`); (2) `delta += delta / numPoints`; (3) while `delta > floor((base-tMin)*tMax/2) == 455`, do `delta /= base - tMin` (=35) and `k += base` (=36); (4) return `k + (base-tMin+1)*delta / (delta+skew)` i.e. `k + 36*delta/(delta+38)`. Cites: `punycode.js:179-187`, `specs/punycode-bias-adapt.md:7-145`. All arithmetic is non-negative integer division — trivial in Go.

### Phase 3 — UCS-2 code-point ↔ string conversion

- **Port `ucs2Decode(s string) []int32` (faithful UTF-16 semantics).** **Critical design decision.** JS strings are UTF-16 and the test fixtures `testData.ucs2` (`tests/tests.js:137-175`) encode lone surrogates and astral code points as UTF-16 surrogate pairs. A naive port that takes Go's `[]rune` of the UTF-8 string will FAIL tests at `tests/tests.js:142` (astral emoji `[127829, 119808, 119558, 119638]`) and `tests/tests.js:147-173` (orphan-surrogate round-trips). Therefore the port must: (a) interpret the input string's UTF-8 bytes, convert each code point to one-or-two UTF-16 code units via `utf16.Encode` (stdlib), then walk those code units applying the surrogate-pairing logic from `punycode.js:102-122`. Result type is `[]int32` (or `[]rune`) of code points, potentially including lone surrogate values `0xD800-0xDFFF`. Cites: `punycode.js:102-122`, `specs/punycode-ucs2-decode.md:7-53`, `tests/tests.js:137-175`.

- **Port `ucs2Encode(codePoints []int32) string` (faithful, non-mutating).** For each code point: if `>= 0x10000`, split into high+low surrogates `0xD800 + ((cp-0x10000)>>10)` / `0xDC00 + ((cp-0x10000)&0x3FF)` (mirrors `punycode.js:89-99`); else emit as a single UTF-16 code unit. Then convert the resulting `[]uint16` to a Go string via `string(utf16.Decode(units))`. **Must not mutate the input slice** (`tests/tests.js:282-287`). Because Go slices are pass-by-reference-like, either take `[]int32` and never assign to it, or copy defensively. Cites: `punycode.js:128-134`, `specs/punycode-ucs2-encode.md:7-40`.

### Phase 4 — Bootstring `decode` / `encode`

- **Port `Decode(input string) (string, error)`.** Implements RFC 3492 decode (`punycode.js:195-285`). Algorithm: (1) find last `'-'` with `strings.LastIndex(input, "-")`; (2) copy pre-delimiter chars to an `[]int32` output, erroring `ErrNotBasic` if any `>= 0x80`; (3) loop reading digits, accumulating `i`, guarding overflow with `maxInt`, calling `adapt` and inserting code points. All three error messages must match exactly (`ErrNotBasic`, `ErrInvalidInput`, `ErrOverflow`). Finish by converting the `[]int32` back to a string via the Phase 3 `ucs2Encode`. Integer type: use `int` (64-bit on modern platforms) with explicit `> maxInt` pre-checks before every addition/multiplication that could exceed the RFC's 32-bit bound. Cites: `punycode.js:195-285`, `specs/punycode-decode.md:1-228`.

- **Port `Encode(input string) (string, error)`.** Implements RFC 3492 encode (`punycode.js:295-377`). Steps: (1) `ucs2Decode(input)` → code-point slice; (2) extract basic (`< 0x80`) code points into output, emit delimiter if any; (3) outer loop finds smallest unmapped code point `m > n`, inner loop emits variable-length integer per matching code point using `digitToBasic(_, 0)`; (4) bias adaptation after each integer. Both overflow checks (`punycode.js:337-338,345-346`) must be translated literally. Use a `strings.Builder` for the output rather than repeated string concatenation. Cites: `punycode.js:295-377`, `specs/punycode-encode.md:1-144`.

### Phase 5 — Domain / email wrappers

- **Port the internal `mapDomain(s string, fn func(string) string) string` helper.** Steps (`punycode.js:72-85`): (1) split on `'@'` into `(local, domain)` preserving the `'@'` in the result if present; (2) replace U+002E / U+3002 / U+FF0E / U+FF61 in `domain` with `'.'` (reuse `normalizeSeparators` from Phase 1); (3) `strings.Split(domain, ".")`; (4) apply `fn` to every label; (5) rejoin with `'.'`; (6) prepend `local + "@"` if present. Cites: `punycode.js:62-85`, `specs/punycode-internal-helpers.md:71-119`.

- **Port `ToUnicode(input string) (string, error)`.** For each label via `mapDomain`, if `strings.HasPrefix(label, "xn--")`, strip the prefix, `strings.ToLower` the remainder, pass to `Decode`; propagate any error. Otherwise return the label unchanged. Cites: `punycode.js:389-395`, `specs/punycode-to-unicode.md:1-132`, `tests/tests.js:175-219`.

- **Port `ToASCII(input string) (string, error)`.** For each label via `mapDomain`, if `hasNonASCII(label)`, run `Encode(label)` and prepend `"xn--"`; else return unchanged. Cites: `punycode.js:408-414`, `specs/punycode-to-ascii.md:1-88`, `tests/tests.js:175-219,221-241`.

### Phase 6 — Test suite

- **Port the four test-data tables into `go/testdata.go`.** Translate the JS fixtures literally — Go source supports UTF-8, so Unicode characters in string literals port directly. Four slice-of-struct variables: `stringsData` (25 entries from `tests/tests.js:7-136` — fields `description, decoded, encoded`); `ucs2Data` (7 entries from `tests/tests.js:137-175` — `description string; decoded []int32; encoded string`); `domainsData` (9 entries from `tests/tests.js:176-220`); `separatorsData` (4 entries from `tests/tests.js:221-242`). The `\uXXXX` JS escapes translate to `\uXXXX` Go escapes (same syntax) for BMP; astral code points use Go's `\U000XXXXX` form. Cites: `specs/test-fixtures.md:1-93`, `tests/tests.js:7-242`.

- **Port `punycode.ucs2.decode` + `ucs2.encode` tests.** Table-driven `TestUCS2Decode` / `TestUCS2Encode` over `ucs2Data`. Add standalone subtests: decode must panic/return-error on `"\x81-"` (`tests/tests.js:255-262`), decode must return `ErrOverflow` on the overflow string (`tests/tests.js:263-270`), encode must not mutate its input slice (`tests/tests.js:282-287`). Cites: `specs/punycode-ucs2-decode.md`, `specs/punycode-ucs2-encode.md`, `tests/tests.js:245-288`.

- **Port `punycode.decode` + `punycode.encode` tests.** Table-driven `TestDecode` / `TestEncode` over `stringsData` (`tests/tests.js:290-321`). Standalone subtest: `Decode("ZZZ")` handles uppercase (`tests/tests.js:299-301`). Standalone subtest: invalid inputs produce `ErrInvalidInput` / `ErrNotBasic` (`tests/tests.js:302-309`). Cites: `specs/punycode-decode.md`, `specs/punycode-encode.md`.

- **Port `toUnicode` + `toASCII` tests.** Table-driven `TestToUnicode` / `TestToASCII` over `domainsData` and `separatorsData`. Subtests for the "skips non-`xn--`" cases (`tests/tests.js:332-343`) and "skips ASCII-only" cases (`tests/tests.js:355-362`). Cites: `specs/punycode-to-unicode.md`, `specs/punycode-to-ascii.md`, `tests/tests.js:323-370`.

- **Add a RFC-3492-round-trip integration subtest.** For every entry in `stringsData`, assert `Decode(Encode(x.decoded)) == x.decoded` and `Encode(Decode(x.encoded)) == x.encoded`. This is the belt-and-suspenders check that catches any silent asymmetry between encode and decode. Cites: `specs/punycode-encode.md` (round-trip section).

### Phase 7 — Polish

- **Add Go doc comments on every exported symbol.** Each `Decode`, `Encode`, `ToASCII`, `ToUnicode`, `UCS2Decode`, `UCS2Encode`, and the three error vars must have a `// Name ...` doc comment so `go doc` output matches the JSDoc in `punycode.js:136-142,190-194,289-294,380-388,403-407`. Cites: the JSDoc blocks above.

- **Run `go vet ./...` and `gofmt -s -w .`; verify `go test ./... -race` passes with zero failures.** Final polish iteration — no new functionality.

### Not porting (explicitly out of scope)

- `scripts/prepublish.js` and the `punycode.es6.js` artifact it produces — Go has no CommonJS/ESM duality, so the whole build step is moot. The exported Go package (`package punycode` with PascalCase public functions) fulfills both roles. Cites: `specs/prepublish-script.md:62-71`, `scripts/prepublish.js:1-17`.
- The private `map()` helper (`punycode.js:53-60`) — its reverse-iteration is purely a JS code-golf idiom with no observable effect (callbacks are pure). Replace every call site with a straight `for i := range labels { labels[i] = fn(labels[i]) }`. Cites: `specs/punycode-internal-helpers.md:39-67`.

---

## Cross-reference: spec file → Go artifact

| Spec file | Primary Go target | Phase |
|---|---|---|
| `specs/punycode-module.md` | constants block at top of `go/punycode.go` | 0, 1 |
| `specs/punycode-internal-helpers.md` | error values + `normalizeSeparators` + `mapDomain` | 1, 5 |
| `specs/punycode-bootstring-digit.md` | `basicToDigit`, `digitToBasic` | 2 |
| `specs/punycode-bias-adapt.md` | `adapt` | 2 |
| `specs/punycode-ucs2-decode.md` | `UCS2Decode` | 3 |
| `specs/punycode-ucs2-encode.md` | `UCS2Encode` | 3 |
| `specs/punycode-decode.md` | `Decode` | 4 |
| `specs/punycode-encode.md` | `Encode` | 4 |
| `specs/punycode-to-unicode.md` | `ToUnicode` | 5 |
| `specs/punycode-to-ascii.md` | `ToASCII` | 5 |
| `specs/test-fixtures.md` | `go/testdata.go` | 6 |
| `specs/prepublish-script.md` | (not ported — Go package replaces it) | — |

## Cross-reference: JS source → Go artifact

| `punycode.js` range | Go target |
|---|---|
| `:1-31` (use-strict, constants, regex, error table, shortcuts) | top-of-file constants block |
| `:41-43` (error helper) | `ErrOverflow` / `ErrNotBasic` / `ErrInvalidInput` + panic-recover glue |
| `:53-60` (`map`) | inline `for` loops (see "Not porting") |
| `:62-85` (`mapDomain`) | unexported `mapDomain` |
| `:87-99` (ucs2 decode tail) / `:102-122` (main decode) | `UCS2Decode` |
| `:128-134` (ucs2 encode) | `UCS2Encode` |
| `:144-155` (`basicToDigit`) | `basicToDigit` |
| `:168-172` (`digitToBasic`) | `digitToBasic` |
| `:179-187` (`adapt`) | `adapt` |
| `:195-285` (`decode`) | `Decode` |
| `:295-377` (`encode`) | `Encode` |
| `:389-395` (`toUnicode`) | `ToUnicode` |
| `:408-414` (`toASCII`) | `ToASCII` |
| `:418-441` (public API object) | exported-symbol surface of the Go package |
| `scripts/prepublish.js:1-17` | (not ported) |

## Key risks to flag in the build loop

1. **UTF-16 vs UTF-8 semantics** (Phase 3) — the single place where "just use `[]rune`" would silently break tests. Build-mode Claude must read the ucs2 spec and use `unicode/utf16` from the stdlib, not naive rune iteration.
2. **Integer overflow** (Phase 4) — RFC 3492 requires checks against `maxInt = 2^31-1` even on 64-bit platforms. Every `w *= base - t`, `i += digit*w`, and `delta += (m-n)*(h+1)` needs its pre-check translated verbatim from `punycode.js:243,255,268,337,345`.
3. **Error model choice** (Phase 1) — picking panic-vs-error once, up front, and holding to it. Mixing the two across `Decode` / `ToASCII` would create an ugly API.
4. **Exact error-message strings** — tests assert on them (`tests/tests.js:257-270,302-309`). Must match `punycode.js:22-26` byte-for-byte.
