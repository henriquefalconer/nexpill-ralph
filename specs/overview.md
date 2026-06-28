# Test Suite Overview — punycode.js

This directory documents the test suite of the **punycode.js** library (v2.3.1,
`package.json:3`), a Punycode / IDNA converter that fully complies with RFC 3492
and RFC 5891 (`package.json:4`).

## Source layout

- Implementation under test: `punycode.js` (the entire codec, `punycode.js:1-444`).
- Public API export: `punycode.js:419-441` exposes `version`, `ucs2.decode`,
  `ucs2.encode`, `decode`, `encode`, `toASCII`, and `toUnicode`.
- Tests: a single file `tests/tests.js` (`tests/tests.js:1-372`).
- Test runner: Mocha — `npm test` runs `mocha tests` (`package.json:43`).
  The suite uses Node's built-in `assert` module (`tests/tests.js:3`) and
  requires the implementation via `require('../punycode.js')` (`tests/tests.js:4`).

## Structure of `tests/tests.js`

The file is a shared fixture object followed by six `describe` blocks — one per
public method. Each spec in this directory documents one logical unit:

| Spec file | Covers | `describe` block | Implementation |
|---|---|---|---|
| [test-fixtures.md](test-fixtures.md) | The shared `testData` object (4 arrays) | n/a (`tests/tests.js:6-243`) | — |
| [ucs2-decode.md](ucs2-decode.md) | `punycode.ucs2.decode` | `tests/tests.js:245-271` | `punycode.js:101-123` |
| [ucs2-encode.md](ucs2-encode.md) | `punycode.ucs2.encode` | `tests/tests.js:273-288` | `punycode.js:133` |
| [decode.md](decode.md) | `punycode.decode` | `tests/tests.js:290-310` | `punycode.js:196-281` |
| [encode.md](encode.md) | `punycode.encode` | `tests/tests.js:312-321` | `punycode.js:290-376` |
| [to-unicode.md](to-unicode.md) | `punycode.toUnicode` | `tests/tests.js:323-344` | `punycode.js:389-395` |
| [to-ascii.md](to-ascii.md) | `punycode.toASCII` | `tests/tests.js:346-371` | `punycode.js:408-414` |

## How the fixtures feed the suites

The single `testData` object (`tests/tests.js:6-243`) holds four arrays that are
shared across the suites:

- `testData.strings` (`tests/tests.js:7-136`) → `decode`, `encode`, and the
  passthrough loops of `toUnicode` / `toASCII`.
- `testData.ucs2` (`tests/tests.js:137-175`) → `ucs2.decode`, `ucs2.encode`.
- `testData.domains` (`tests/tests.js:176-220`) → `toUnicode`, `toASCII`.
- `testData.separators` (`tests/tests.js:221-242`) → `toASCII` (IDNA2003
  separator backwards compatibility).

See [test-fixtures.md](test-fixtures.md) for the full field semantics.

## Inverse relationships

The suites form three inverse pairs over shared fixtures, which together assert
round-trip correctness:

- `ucs2.decode` ↔ `ucs2.encode` over `testData.ucs2`.
- `decode` ↔ `encode` over `testData.strings`.
- `toUnicode` ↔ `toASCII` over `testData.domains`.

## Known discrepancy

The test named *"throws RangeError: Overflow…"* at `tests/tests.js:263-270`
calls `punycode.decode('\x81')`, which actually throws `RangeError: Invalid
input` (`punycode.js:240-242`), not the overflow error. The test still passes
because `assert.throws(fn, RangeError)` (`tests/tests.js:264`) checks only the
error constructor, not the message. See [ucs2-decode.md](ucs2-decode.md) for
details.
