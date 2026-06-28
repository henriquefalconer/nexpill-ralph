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

- [x] **1. Lean project scaffolding.** Installed `elan` v4.2.3 + Lean 4 v4.31.0
  (pinned in `lean-toolchain`). Created `lakefile.toml` (lib `Punycode` + exe
  `PunycodeTests`), empty `Punycode.lean` root, stub `Tests/Main.lean`,
  `.gitignore` entries for `.lake/`/`build/`. `lake build` passes.

- [x] **2. `Punycode/Constants.lean` + `PunyError`.** Port the config block
  `punycode.js:1-31`: `maxInt` (`punycode.js:4`), Bootstring params `base/tMin/
  tMax/skew/damp/initialBias/initialN/delimiter` (`punycode.js:6-14`),
  `baseMinusTMin` (`punycode.js:29`), and the four separator code points for
  later use (`punycode.js:19`). Define `inductive PunyError | overflow | notBasic
  | invalidInput` from the `errors` table (`punycode.js:21-26`) plus
  `PunyError.message : String` returning the exact strings `punycode.js:22-25`.
  All as `Nat`/`UInt32` per D3. Spec: [`specs/impl-constants.md`](../specs/impl-constants.md).

- [x] **3. `Punycode/Helpers.lean` — `error`, `map`, `mapDomain`.** Port
  `punycode.js:41-86`. `error` → `Except.error` over `PunyError` (D1). `map`
  (`punycode.js:53-60`) → `Array.map`/`List.map` (reverse-traversal detail is
  irrelevant once order-preserving). `mapDomain` (`punycode.js:72-86`): split on
  `'@'` keeping the email local part intact (`punycode.js:73-80`), replace the
  four `regexSeparators` chars with `'.'` (`punycode.js:82`, no regex engine —
  see D4), split on `'.'`, map an `Except`-valued callback `(String → Except
  PunyError String)` over labels, `join "."`, reassemble (`punycode.js:84-85`).
  Spec: [`specs/impl-helpers.md`](../specs/impl-helpers.md).

- [x] **4. `Punycode/UCS2.lean` — `ucs2decode`/`ucs2encode` + UTF-16 bridges.**
  Port `punycode.js:101-133` over `Array UInt16` per D2. `ucs2decode : Array
  UInt16 → Array UInt32` reproduces the high/low/unmatched-surrogate branches
  (`punycode.js:107-120`) and the `counter--` reprocess step (`punycode.js:116`).
  `ucs2encode : Array UInt32 → Array UInt16` inverts the pair-combine
  (`punycode.js:111`, `punycode.js:133`). Add `String ↔ Array UInt16` bridge
  helpers for the API boundary (document the lone-surrogate `String` limitation,
  D2). Spec: [`specs/impl-ucs2.md`](../specs/impl-ucs2.md).
  **Note:** `ucs2decodeAux` uses `partial def` because Lean's `omega` tactic cannot
  close `Nat.sub` termination goals involving `Array.size` automatically. The
  function trivially terminates since the index `i` strictly increases every branch.

- [x] **5. `Punycode/Bootstring.lean` — `basicToDigit`, `digitToBasic`,
  `adapt`.** Port `punycode.js:144-187`. `basicToDigit` (`punycode.js:144-155`)
  → branch table returning `0..35` or `base` sentinel. `digitToBasic`
  (`punycode.js:168-172`) → explicit `Nat` arithmetic replacing the branchless
  bit trick (D3). `adapt` (`punycode.js:179-187`): scaling loop
  `punycode.js:183-185` needs termination — prefer the decreasing-`delta` measure
  proof (`baseMinusTMin = 35 > 1`), fall back to fuel (D4). Spec:
  [`specs/impl-bootstring-helpers.md`](../specs/impl-bootstring-helpers.md).
  **Note:** `adaptLoop` uses `partial def` (same pattern as `ucs2decodeAux`):
  Lean cannot prove `delta / baseMinusTMin < delta` from the `delta > 455` guard
  automatically; the loop trivially terminates since `35 > 1` ensures strict
  decrease.

- [x] **6. `Punycode/Decode.lean` — `decode`.** Port `punycode.js:196-281` →
  `decode : String → Except PunyError String` (output via `Array UInt32` then the
  D2 bridge, mirroring `String.fromCodePoint(...output)` `punycode.js:280`). State
  init `punycode.js:198-202`; basic-points copy + `not-basic` guard
  `punycode.js:208-219`; main loop `punycode.js:224-278`; the unbounded
  variable-length-integer inner loop `punycode.js:232-261` (fuel/termination per
  D4); all six overflow/invalid guards (`punycode.js:235,241,244,256,269` +
  `not-basic` 216); `adapt` call `punycode.js:264`; `Array.insertIdx` for
  `splice` `punycode.js:276`. Preserve the `decode('\x81')`→`invalidInput`
  behaviour (D1). Spec: [`specs/impl-decode.md`](../specs/impl-decode.md).
  **Note:** `decodeVLI` and `decodeLoop` use `partial def` (same pattern): Lean cannot prove termination of the VLI digit-accumulation loop automatically; the loops trivially terminate since `index` strictly increases on every iteration.

- [x] **7. `Punycode/Encode.lean` — `encode`.** Port `punycode.js:290-376` →
  `encode : String → Except PunyError String`. `ucs2decode` the input
  (`punycode.js:294`, via UCS2 module); state `punycode.js:300-302`; basic-points
  + delimiter `punycode.js:305-320`; main loop `punycode.js:323-374` — next-min
  scan `punycode.js:327-332`, delta-advance + overflow guard `punycode.js:336-342`,
  per-code-point pass `punycode.js:344-369` with the unbounded digit loop
  `punycode.js:351-362` (D4) calling `digitToBasic` (`punycode.js:359,364`) and
  `adapt` (`punycode.js:365`); `join ""` (`punycode.js:375`). Spec:
  [`specs/impl-encode.md`](../specs/impl-encode.md).

- [ ] **8. `Punycode/Domain.lean` + `Punycode.lean` root — public API.** Port
  `toUnicode` (`punycode.js:389-395`: `mapDomain` + `startsWith "xn--"` →
  `decode (slice 4 |> asciiLower)`) and `toASCII` (`punycode.js:408-414`:
  `mapDomain` + "any cp ≥ 0x80" → `"xn--" ++ encode`), both `Except`-valued.
  Re-export the public surface from `Punycode.lean`: `version = "2.3.1"`
  (`punycode.js:425`), `ucs2.decode/encode` (`punycode.js:434-435`), `decode`,
  `encode`, `toASCII`, `toUnicode` (`punycode.js:437-440`). Spec:
  [`specs/impl-domain-api.md`](../specs/impl-domain-api.md). Verify `lake build`.

- [ ] **9. `Tests/Fixtures.lean` — port `testData`.** Port the four arrays
  `tests/tests.js:6-243`: `strings` (`tests/tests.js:7-136`), `ucs2`
  (`tests/tests.js:137-175`, as `Array UInt16` ↔ `Array UInt32` per D2), `domains`
  (`tests/tests.js:176-220`), `separators` (`tests/tests.js:221-242`). Keep the
  RFC 3492 §7.1 vectors and the surrogate edge cases verbatim. Spec:
  [`specs/test-fixtures.md`](../specs/test-fixtures.md).

- [ ] **10. `Tests/Main.lean` — codec & UCS-2 suites + runner.** Build the test
  runner (exit non-zero on failure, D4/§Test strategy). Port `ucs2.decode`
  (`tests/tests.js:245-271`, incl. the `decode('\x81')` error case → assert
  `invalidInput` per D1/[`specs/decode.md`]), `ucs2.encode` (`tests/tests.js:273-288`),
  `decode` (`tests/tests.js:290-310`), `encode` (`tests/tests.js:312-321`) suites
  over the fixtures. Specs: [`specs/ucs2-decode.md`](../specs/ucs2-decode.md),
  [`specs/ucs2-encode.md`](../specs/ucs2-encode.md), [`specs/decode.md`](../specs/decode.md),
  [`specs/encode.md`](../specs/encode.md).

- [ ] **11. `Tests/Main.lean` — domain suites + green build.** Port `toUnicode`
  (`tests/tests.js:323-344`, incl. `strings` passthrough loop) and `toASCII`
  (`tests/tests.js:346-371`, incl. `strings` passthrough + the `separators`
  normalisation loop `tests/tests.js:363-370`). Run the full suite via `lake exe`,
  fix any behavioural divergences from JS, and confirm all suites pass. Specs:
  [`specs/to-unicode.md`](../specs/to-unicode.md), [`specs/to-ascii.md`](../specs/to-ascii.md),
  [`specs/overview.md`](../specs/overview.md).

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
