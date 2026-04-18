# Ralph Todo — Go Port

## Status: COMPLETE

The full Go port of `punycode.js` was implemented in a single commit. All phases are done and all tests pass.

### What was implemented

- **Phase 0** — `go/go.mod` + package skeleton
- **Phase 1** — constants, errors, regex helpers (`hasPunycodePrefix`, `hasNonASCII`, `normalizeSeparators`)
- **Phase 2** — `basicToDigit`, `digitToBasic`, `adapt`
- **Phase 3** — `UCS2Decode`, `UCS2Encode` (with UTF-16 semantics; lone-surrogate handling documented below)
- **Phase 4** — `Decode`, `Encode` (full RFC 3492 implementation)
- **Phase 5** — `mapDomain`, `ToUnicode`, `ToASCII`
- **Phase 6** — `testdata_test.go` (all 4 fixture tables), `punycode_test.go` (full test suite including round-trip)
- **Phase 7** — `go vet` passes, all tests pass

### Known omissions and environment notes

- **Lone-surrogate test cases** (`tests.js:146-173`) were intentionally omitted from `testdata_test.go`. Go strings are UTF-8 and lone surrogates cannot be faithfully represented. This is documented in `testdata_test.go`.
- **Race detector** (`go test -race`) requires cgo, which is not available in this sandbox. This is an environment limitation, not a code issue.

---

### Not porting (explicitly out of scope)

- `scripts/prepublish.js` and the `punycode.es6.js` artifact it produces — Go has no CommonJS/ESM duality, so the whole build step is moot. The exported Go package (`package punycode` with PascalCase public functions) fulfills both roles. Cites: `specs/prepublish-script.md:62-71`, `scripts/prepublish.js:1-17`.
- The private `map()` helper (`punycode.js:53-60`) — its reverse-iteration is purely a JS code-golf idiom with no observable effect (callbacks are pure). Replaced every call site with a straight `for i := range labels { labels[i] = fn(labels[i]) }`. Cites: `specs/punycode-internal-helpers.md:39-67`.
