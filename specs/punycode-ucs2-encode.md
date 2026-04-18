# punycode.ucs2.encode — Specification

## Purpose
This function converts a sequence of integer Unicode code points back into a Unicode string. In the JavaScript reference implementation, the returned string is a UCS-2/UTF-16 string, which means code points above the Basic Multilingual Plane are represented as surrogate pairs of 16-bit code units. It is the inverse of `punycode.ucs2.decode` for well-formed inputs, and its implementation is a single delegation to the host's code-point-to-string conversion primitive (`punycode.js:125-133`).

## Contract
- Input: an ordered sequence (array) of integer Unicode code points. The implementation does not validate the range itself, but delegates to `String.fromCodePoint`, which accepts any integer in the range `0..0x10FFFF` (`punycode.js:133`).
- Output: a Unicode string whose code-unit content is determined by how each input code point is serialized into UTF-16 by the host (`punycode.js:133`).
- Invariant: the function MUST NOT mutate the input sequence. This is asserted explicitly by the test at `tests/tests.js:282-287`.
- Public API: exposed as `punycode.ucs2.encode` on the module's default export (`punycode.js:433-436`).

## Behavior rules

- Rule 1 — BMP code points pass through as single code units. Each input code point in the range `0x0000..0xFFFF` that is not a surrogate becomes exactly one 16-bit code unit in the output string. This is the baseline case exercised implicitly by every fixture that mixes ASCII letters with other values (`tests/tests.js:147-148`, `tests/tests.js:152-153`; implementation at `punycode.js:133`).

- Rule 2 — Astral (supplementary-plane) code points are split into UTF-16 surrogate pairs. A code point greater than `0xFFFF` (for example `0x1F355`, `0x1D400`, `0x1D306`, `0x1D356`) is emitted as a high-surrogate / low-surrogate pair of code units in the output string. Consecutive astral code points produce consecutive surrogate pairs with no separator and no merging across boundaries (fixture "Consecutive astral symbols" at `tests/tests.js:141-144`; implementation at `punycode.js:133`).

- Rule 3 — Lone surrogate code-unit values passed in as code points are emitted as-is. Although values in the range `0xD800..0xDFFF` are not valid Unicode scalar values, the underlying primitive accepts any integer up to `0x10FFFF` and emits each such value as a single 16-bit code unit. This lets the encoder faithfully round-trip malformed UTF-16 that a prior decode step may have produced. Covered by:
  - A lone high surrogate followed by non-surrogates (`tests/tests.js:145-149`).
  - A lone low surrogate followed by non-surrogates (`tests/tests.js:150-154`).
  - Two consecutive high surrogates (`tests/tests.js:155-159`).
  - An unmatched high surrogate, then a valid surrogate pair, then another unmatched high surrogate (`tests/tests.js:160-164`).
  - Two consecutive low surrogates (`tests/tests.js:165-169`).
  - An unmatched low surrogate, then a valid surrogate pair, then another unmatched low surrogate (`tests/tests.js:170-174`).
  Implementation reference: `punycode.js:133`.

- Rule 4 — Input immutability. Calling the function with a code-point sequence leaves that sequence equal to its pre-call value. The reference implementation achieves this because it only spreads the argument into `String.fromCodePoint` and never assigns to it. Asserted at `tests/tests.js:282-287`; implementation at `punycode.js:133`.

- Rule 5 — Empty input. Because the implementation simply spreads the sequence into `String.fromCodePoint`, an empty input sequence yields the empty string. This is not covered by a dedicated fixture but follows directly from `punycode.js:133` and is consistent with the round-trip relationship described below.

## Relationship to ucs2.decode
For any valid UTF-16 input string `s`, the round-trip identity `ucs2.encode(ucs2.decode(s)) == s` holds. The fixtures under `testData.ucs2` at `tests/tests.js:137-175` are deliberately designed so that each `decoded` integer array and each `encoded` string round-trip in both directions; the same table drives the `punycode.ucs2.decode` test block as well. Rule 3 above is what makes this round-trip exact even when `s` contains lone surrogates (`punycode.js:133`).

## Port notes (Go)
- The JavaScript implementation relies on `String.fromCodePoint`, which treats each numeric input as a Unicode code point (not a UTF-16 code unit) and auto-generates a surrogate pair for any value above `0xFFFF` (`punycode.js:133`).
- A faithful Go port must accept a code-point sequence (for example `[]rune` or `[]int32`) and emit UTF-16 code units, not UTF-8 bytes, in order to preserve the exact code-unit-level round-trip required by the lone-surrogate fixtures at `tests/tests.js:145-174`. A `[]uint16` output, or a string whose bytes encode those 16-bit code units, is appropriate; a straight `string` (UTF-8) will corrupt lone-surrogate fixtures because Go's UTF-8 encoder rejects or replaces surrogate values.
- For each code point `cp`: if `cp <= 0xFFFF`, append `cp` as a single `uint16`; if `cp > 0xFFFF`, append the two code units `0xD800 + ((cp - 0x10000) >> 10)` and `0xDC00 + ((cp - 0x10000) & 0x3FF)`. Values in `0xD800..0xDFFF` passed in directly are emitted as a single `uint16` unchanged (Rule 3).
- The port must not modify the caller's input slice (Rule 4 / `tests/tests.js:282-287`). In Go this is trivial because the implementation only reads the slice, but porters should avoid in-place reuse patterns.
- Overall porting guidance and the requirement that every test vector be preserved is recorded in `TARGET.md`.

## Test-vector source
All vectors listed in `specs/test-data-fixtures.md` under the `ucs2` section; originals at `tests/tests.js:137-175`.
