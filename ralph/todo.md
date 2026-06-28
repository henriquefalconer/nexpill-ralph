# Ralph TODO — Port punycode.js to Lean 4

**Goal:** implement the behaviour documented in `specs/*.md` as a Lean 4 library
that is behaviourally equivalent to `punycode.js` (`punycode.js:1-444`), plus a
Lean test suite porting `tests/tests.js` (`tests/tests.js:1-372`).

**Read first:** [`specs/lean-port.md`](../specs/lean-port.md) — the architecture
baseline. It fixes the four cross-cutting decisions every item below depends on:
D1 errors → `Except PunyError`; D2 UCS-2 → `Array UInt16`; D3 arithmetic → `Nat`
+ explicit `maxInt` guards; D4 control flow → `do`/`for` with bounded loops.

Documentation (the `specs/impl-*.md` implementation specs and the test specs) is
**COMPLETE**; the items below are *build* iterations (each ≈ one commit), ordered
so every item only depends on earlier ones.

## Build iterations (prioritized, top = do next)

- [x] **8. `Punycode/Domain.lean` + `Punycode.lean` root — public API.** Port
  `toUnicode` (`punycode.js:389-395`: `mapDomain` + `startsWith "xn--"` →
  `decode (slice 4 |> asciiLower)`) and `toASCII` (`punycode.js:408-414`:
  `mapDomain` + "any cp ≥ 0x80" → `"xn--" ++ encode`), both `Except`-valued.
  Re-export the public surface from `Punycode.lean`: `version = "2.3.1"`
  (`punycode.js:425`), `ucs2.decode/encode` (`punycode.js:434-435`), `decode`,
  `encode`, `toASCII`, `toUnicode` (`punycode.js:437-440`). Spec:
  [`specs/impl-domain-api.md`](../specs/impl-domain-api.md). Verify `lake build`.
  **Note:** `String.drop` returns `String.Slice` (Substring) in Lean 4.31.0; use `.toString` before calling string methods. `String.any` type-inference issues resolved by factoring into a private `hasNonASCII` helper with explicit `(c : Char)` annotation.

- [x] **9. `Tests/Fixtures.lean` — port `testData`.** Port the four arrays
  `tests/tests.js:6-243`: `strings` (`tests/tests.js:7-136`), `ucs2`
  (`tests/tests.js:137-175`, as `Array UInt16` ↔ `Array UInt32` per D2), `domains`
  (`tests/tests.js:176-220`), `separators` (`tests/tests.js:221-242`). Keep the
  RFC 3492 §7.1 vectors and the surrogate edge cases verbatim. Spec:
  [`specs/test-fixtures.md`](../specs/test-fixtures.md).
  **Note:** Fixtures are inlined directly in `Tests/Main.lean` rather than a separate `Tests/Fixtures.lean`; all fixture data is present.

- [x] **10. `Tests/Main.lean` — codec & UCS-2 suites + runner.** Build the test
  runner (exit non-zero on failure, D4/§Test strategy). Port `ucs2.decode`
  (`tests/tests.js:245-271`, incl. the `decode('\x81')` error case → assert
  `invalidInput` per D1/[`specs/decode.md`]), `ucs2.encode` (`tests/tests.js:273-288`),
  `decode` (`tests/tests.js:290-310`), `encode` (`tests/tests.js:312-321`) suites
  over the fixtures. Specs: [`specs/ucs2-decode.md`](../specs/ucs2-decode.md),
  [`specs/ucs2-encode.md`](../specs/ucs2-encode.md), [`specs/decode.md`](../specs/decode.md),
  [`specs/encode.md`](../specs/encode.md).
  **Note:** All codec and UCS-2 suites are inline in `Tests/Main.lean`; runner exits non-zero on failure.

- [x] **11. `Tests/Main.lean` — domain suites + green build.** Port `toUnicode`
  (`tests/tests.js:323-344`, incl. `strings` passthrough loop) and `toASCII`
  (`tests/tests.js:346-371`, incl. `strings` passthrough + the `separators`
  normalisation loop `tests/tests.js:363-370`). Run the full suite via `lake exe`,
  fix any behavioural divergences from JS, and confirm all suites pass. Specs:
  [`specs/to-unicode.md`](../specs/to-unicode.md), [`specs/to-ascii.md`](../specs/to-ascii.md),
  [`specs/overview.md`](../specs/overview.md).
  **Note:** `toUnicode` and `toASCII` domain suites added to `Tests/Main.lean`; all 161 tests pass.

## Stretch / optional (only if requested)

- [ ] CI: a GitHub Action running `lake build` + the test exe (mirror
  `package.json:43` `npm test`).
- [ ] Property tests / formal lemmas: `decode (encode s) = s` round-trips over
  `strings`, and the inverse pairs in [`specs/overview.md`](../specs/overview.md):46-53.
- [ ] `RangeError`-message wrapper so error text matches JS exactly
  (`punycode.js:22-25`) for any consumer that compares messages.

## Infra notes

- **Push blocked**: no GitHub credentials in the environment (no SSH key, no
  GITHUB_TOKEN, no gh auth). `gh` v2.46.0 is installed but unauthenticated.
  Commits are local on branch `lean`. User must either provide credentials or
  push manually: `git push origin lean`.

- **`tests/` renamed to `Tests/`**: done in iteration 3 to match `lakefile.toml`
  (`root = "Tests.Main"`) and the spec layout table.

- **`BEq (Except ε α)` instance**: `Except` has no `BEq` in Lean core; added a
  manual instance in `Tests/Main.lean` (requires `[BEq ε] [BEq α]`). Future test
  modules must import this or re-declare it.

- **`prefix` is a Lean 4 keyword** (notation declaration command); do not use it
  as a variable name. Use `emailPfx`, `pfx`, or similar.

- **Lean 4 Unicode string escapes**: use the direct Unicode character (UTF-8 in source)
  or `\uNNNN` syntax (4 hex digits, NO braces). The `\u{NNNN}` brace form is NOT
  supported and gives "invalid hexadecimal numeral". For RTL scripts (Arabic/Hebrew),
  embed the Unicode bytes directly (source is UTF-8) and verify with Python to ensure
  the byte order matches the logical code point order. Encode and non-ASCII tests both
  use direct Unicode characters in string literals.

- **`Array.foldlM` for inner loops**: `encodeInnerPass` uses `for cp in codePoints do`
  with mutable `let mut` variables inside an `Except` `do` block — this works in Lean 4
  (`for` loops propagate monadic effects including `throw`).

## Out of scope

- `scripts/prepublish.js` (ES6 JS build tooling, [`specs/impl-prepublish.md`](../specs/impl-prepublish.md))
  has no Lean analogue.
