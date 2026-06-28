# Lean Port — Architecture & Porting Plan

This spec defines **how** the punycode.js codec is ported to **Lean 4 (Lean lang)**.
It is the durable design baseline that every build iteration in
[`ralph/todo.md`](../ralph/todo.md) references. The behaviour being ported is
fully documented in the existing flat specs — this file only adds the
Lean-specific decisions (type mapping, monad strategy, termination, file
layout, test strategy). Source citations use `punycode.js:<line>`; behaviour
citations point at the relevant `specs/impl-*.md` / `specs/*.md`.

## Goal

Produce a Lean 4 library (`Punycode`) that is behaviourally equivalent to
`punycode.js` (`punycode.js:1-444`) — same outputs and same error conditions —
and a Lean test suite that ports `tests/tests.js` (`tests/tests.js:1-372`,
documented in [overview.md](overview.md) and [test-fixtures.md](test-fixtures.md)).

`scripts/prepublish.js` ([impl-prepublish.md](impl-prepublish.md)) is build
tooling for the ES6 JS variant and has **no Lean analogue** — it is explicitly
out of scope.

## The four cross-cutting porting decisions

### D1 — Errors: `Except PunyError`, not exceptions

JS `error(type)` throws `new RangeError(errors[type])` and never returns
(`punycode.js:41-43`; [impl-helpers.md](impl-helpers.md)). Lean functions are
pure, so every function that can call `error` returns `Except PunyError α`:

- `inductive PunyError | overflow | notBasic | invalidInput` mirrors the three
  keys of the `errors` table (`punycode.js:21-26`; [impl-constants.md](impl-constants.md)).
- Affected functions: `decode` (`punycode.js:196-281`), `encode`
  (`punycode.js:290-376`), `toASCII`/`toUnicode` (`punycode.js:389-414`), and
  `mapDomain`'s callback type (`punycode.js:72-86`) all become `Except`-valued.
- Error *messages* (the exact strings at `punycode.js:22-25`) are recovered by a
  `PunyError.message : String` function so the test suite can assert on them, and
  so a future `RangeError`-style wrapper can reproduce JS message text.
- The known discrepancy — `decode('\x81')` yields `invalidInput`, not `overflow`
  ([impl-decode.md](impl-decode.md):69-76, [overview.md](overview.md):56-62) — is
  preserved exactly (the Lean port reproduces JS behaviour, not the test label).

### D2 — UTF-16 modelling: `Array UInt16` for the UCS-2 layer

JS strings are UTF-16 internally and `charCodeAt` exposes **lone surrogate code
units** (`punycode.js:106`, `punycode.js:238`). Lean's `String` is UTF-8 and
`Char` is a valid Unicode scalar value (`0..0x10FFFF` minus `0xD800..0xDFFF`), so
a Lean `String` **cannot** hold a lone surrogate. The `ucs2` fixtures depend on
lone surrogates (`'\uD800ab'`, `[0xD800, 0xD800]`, etc. —
[test-fixtures.md](test-fixtures.md):45-86, `tests/tests.js:137-175`). Therefore:

- Model a "JS UCS-2 string" as `Array UInt16` (a sequence of UTF-16 code units).
- `ucs2decode : Array UInt16 → Array UInt32` ports `punycode.js:101-123`
  ([impl-ucs2.md](impl-ucs2.md)) faithfully, including the high-surrogate /
  low-surrogate / unmatched-surrogate branches (`punycode.js:107-120`) and the
  `counter--` reprocessing step (`punycode.js:116`).
- `ucs2encode : Array UInt32 → Array UInt16` ports
  `codePoints => String.fromCodePoint(...codePoints)` (`punycode.js:133`) by
  splitting code points ≥ `0x10000` back into surrogate pairs (the inverse of
  `punycode.js:111`).
- Bridge helpers convert at the API boundary:
  `String → Array UInt16` (encode the Lean string to UTF-16 code units) and
  `Array UInt16 → String` (decode UTF-16 back, used for `decode`'s output
  `String.fromCodePoint(...output)` at `punycode.js:280`). Lone surrogates that
  cannot be represented in a Lean `String` are an documented boundary limitation
  that affects only the `ucs2.*` unit tests, which therefore assert over
  `Array UInt16` / `Array UInt32` directly, never over `String`.

`decode`/`encode`/`toASCII`/`toUnicode` operate on real Unicode and so use Lean
`String` at their public boundary; internally `encode` still routes through
`ucs2decode` over `Array UInt16` (`punycode.js:294`).

### D3 — Arithmetic: `Nat` with explicit `maxInt` guards

Every running value (`n`, `i`, `bias`, `w`, `k`, `q`, `delta`, `m`) is
non-negative throughout (`punycode.js:200-202`, `300-302`, etc.), and the
differences taken (`i - oldi` at `punycode.js:264`, `m - n` at `punycode.js:341`)
are non-negative by construction. So:

- Use `Nat` (arbitrary precision) for all codec arithmetic — this removes
  silent machine-int wraparound and lets the overflow *guards* be the only place
  `maxInt` matters.
- `maxInt = 2147483647` (`punycode.js:4`; [impl-constants.md](impl-constants.md):13-21)
  is a plain `Nat`; the guards at `punycode.js:243`, `255`, `268`, `327`, `337`,
  `345` are ported verbatim as comparisons returning `PunyError.overflow`.
- `floor(a / b)` is Lean `Nat` division (already floored); `delta >> 1`
  (`punycode.js:181`) is `delta / 2`; the `digitToBasic` bit tricks
  (`punycode.js:171`) port to explicit arithmetic (`+ 75 * (if digit < 26 then 1
  else 0) - (if flag != 0 then 32 else 0)`); the `ucs2decode` masks/shifts
  (`& 0x3FF`, `<< 10`, `& 0xFC00`) use `UInt16`/`Nat` `land`/`shiftLeft`.
- `Nat` subtraction is truncating; only use it where the JS value is provably
  ≥ 0 (it always is at the subtraction sites above), otherwise compute on the
  guard side.

### D4 — Control flow & termination: `do`/`Id`/`Except` with `for`, bounded loops

Lean 4 `do`-notation supports `let mut` + `for x in xs do` + early `return`/
`break`, so the imperative JS loops port structurally:

- Bounded `for` loops (basic-code-point copy `punycode.js:213-219`, the input
  scans `punycode.js:305-309`/`punycode.js:328-332`/`punycode.js:344-369`) map
  directly to `for ... in ...` over arrays/ranges.
- The **unbounded** generalized-variable-length-integer loops
  (`for(;;)` at `punycode.js:232-261` in decode, `punycode.js:351-362` in encode)
  and the bias scaling loop (`punycode.js:183-185` in `adapt`) need an explicit
  termination argument for Lean. Strategy: bound them by a `fuel : Nat` parameter
  (e.g. derived from `input.length`/`maxInt`) and treat fuel exhaustion as
  unreachable (`PunyError.overflow` or `panic`/`unreachable`), **or** prove
  termination via a decreasing measure (`adapt`'s `delta` strictly decreases
  while `delta > 455` since `baseMinusTMin = 35 > 1`, `punycode.js:184`). Prefer a
  decreasing-measure proof where cheap (`adapt`), fuel elsewhere. The chosen
  approach per loop is recorded in the relevant build iteration.
- `output.splice(i, 0, n)` (`punycode.js:276`) ports to `Array.insertIdx`.
- JS string ops port to Lean equivalents (no regex engine needed):
  `lastIndexOf('-')` (`punycode.js:208`), `slice(4)` + ASCII `toLowerCase`
  (`punycode.js:392`), `split('@')`/`split('.')`/`join` (`punycode.js:73-84`),
  `regexPunycode.test` → `String.startsWith "xn--"` (`punycode.js:391`),
  `regexNonASCII.test` → "any code point ≥ 0x80" (`punycode.js:410`),
  `regexSeparators` replace → replace the four separator chars U+002E/U+3002/
  U+FF0E/U+FF61 with `'.'` (`punycode.js:82`, [impl-constants.md](impl-constants.md):47-51,
  [test-fixtures.md](test-fixtures.md):123-163).

## Target file layout

A Lake package `Punycode` (library) plus a test executable. One Lean module per
implementation spec, so the mapping is 1:1 and citations stay tight:

| Lean file | Ports | Source | Behaviour spec |
|---|---|---|---|
| `lakefile.toml` + `lean-toolchain` | build config | — | this file |
| `Punycode.lean` (root) | re-exports public API | `punycode.js:419-443` | [impl-domain-api.md](impl-domain-api.md) |
| `Punycode/Constants.lean` | `maxInt`, Bootstring params, separators, `PunyError` | `punycode.js:1-31`, `21-26` | [impl-constants.md](impl-constants.md) |
| `Punycode/Helpers.lean` | `error`→`Except`, `map`, `mapDomain` | `punycode.js:41-86` | [impl-helpers.md](impl-helpers.md) |
| `Punycode/UCS2.lean` | `ucs2decode`, `ucs2encode`, UTF-16 bridges | `punycode.js:101-133` | [impl-ucs2.md](impl-ucs2.md) |
| `Punycode/Bootstring.lean` | `basicToDigit`, `digitToBasic`, `adapt` | `punycode.js:144-187` | [impl-bootstring-helpers.md](impl-bootstring-helpers.md) |
| `Punycode/Decode.lean` | `decode` | `punycode.js:196-281` | [impl-decode.md](impl-decode.md) |
| `Punycode/Encode.lean` | `encode` | `punycode.js:290-376` | [impl-encode.md](impl-encode.md) |
| `Punycode/Domain.lean` | `toUnicode`, `toASCII`, `version` | `punycode.js:389-441` | [impl-domain-api.md](impl-domain-api.md) |
| `Tests/Fixtures.lean` | `testData` (4 arrays) | `tests/tests.js:6-243` | [test-fixtures.md](test-fixtures.md) |
| `Tests/Main.lean` | the six `describe` suites | `tests/tests.js:245-371` | [overview.md](overview.md) + per-method specs |

## Public API parity

The Lean `Punycode` namespace must expose the members of the JS `punycode`
object (`punycode.js:419-441`; [impl-domain-api.md](impl-domain-api.md):30-43):
`version = "2.3.1"` (`punycode.js:425`), `ucs2.decode`/`ucs2.encode`
(`punycode.js:434-435`), `decode`, `encode`, `toASCII`, `toUnicode`
(`punycode.js:437-440`). Names and signatures are Lean-idiomatic but the
behaviour is identical.

## Test strategy

- Port `testData` (`tests/tests.js:6-243`) into typed Lean fixtures: `strings`
  (`tests/tests.js:7-136`), `ucs2` (`tests/tests.js:137-175`, modelled as
  `Array UInt16` ↔ `Array UInt32`), `domains` (`tests/tests.js:176-220`),
  `separators` (`tests/tests.js:221-242`). See [test-fixtures.md](test-fixtures.md).
- Port the six suites and their assertions (decode↔encode, ucs2 decode↔encode,
  toUnicode↔toASCII round trips — [overview.md](overview.md):46-53), plus the
  passthrough loops ([test-fixtures.md](test-fixtures.md):26-29) and the
  separator-normalisation loop ([to-ascii.md](to-ascii.md), `tests/tests.js:363-370`).
- Error-condition tests assert on `PunyError`/`message`, reproducing the JS
  `assert.throws(..., RangeError)` cases (e.g. `tests/tests.js:255-270`,
  `tests/tests.js:296-309`), including the mislabeled-but-correct `decode('\x81')`
  → `invalidInput` case (D1).
- Any framework is acceptable (e.g. `LSpec`, or a plain `IO` runner with a pass
  counter); the runner must exit non-zero on failure so `lake` / CI catches it.

## Prerequisites (for the build iterations)

- **Toolchain not yet installed**: there is no `elan`/`lake`/`lean` on PATH and
  no `lean-toolchain`/`lakefile` in the tree (verified). The first build
  iteration must install `elan` (which provides `lake`/`lean`) and pin a
  `lean-toolchain`. This needs outbound network access to the Lean release
  hosts (`elan`/`leanprover` GitHub releases, `release.lean-lang.org`) — request
  it if the firewall blocks the installer.
- No Mathlib dependency is required; the port uses only Lean core + `Std`/`Init`.
