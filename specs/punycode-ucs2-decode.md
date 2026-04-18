# punycode.ucs2.decode

## Purpose

Convert a UCS-2 string (as exposed by JavaScript and similar runtimes that represent Unicode characters outside the Basic Multilingual Plane as surrogate pairs) into an array of Unicode code points. Surrogate pairs are combined into single code points; unmatched surrogates are emitted as their raw code-unit values.

## Signature

**Input:** A string (sequence of code units in the UCS-2 range 0x0000–0xFFFF).

**Output:** An array of non-negative integers, each representing a Unicode code point (0x0000–0x10FFFF).

## Behavior

The function processes the input string left-to-right (punycode.js:105), maintaining a character counter:

- **Surrogate pair matching** (punycode.js:107–111): If the current code unit is in the high surrogate range (0xD800–0xDBFF) and another code unit exists, inspect the next code unit. If it is a low surrogate (0xDC00–0xDFFF, identified by the mask `(value & 0xFC00) == 0xDC00`), consume both and emit the combined code point: `((high & 0x3FF) << 10) + (low & 0x3FF) + 0x10000`.

- **Unmatched high surrogate** (punycode.js:113–117): If the current code unit is a high surrogate but the next is not a low surrogate (or no next unit exists), emit the high surrogate as-is and decrement the counter to reprocess the next code unit in the next iteration.

- **All other code units** (punycode.js:118–120): Emit the current code unit unchanged and advance.

## Examples

The test suite validates the following cases from `testData.ucs2` (tests.js:137–175):

- **Consecutive astral symbols** (tests.js:141–143): Input `\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56` (four surrogate pairs) decodes to `[127829, 119808, 119558, 119638]`.

- **High surrogate followed by non-surrogates** (tests.js:146–148): Input `\uD800ab` decodes to `[55296, 97, 98]`, treating the unmatched high surrogate as code point 0xD800.

- **Low surrogate followed by non-surrogates** (tests.js:151–153): Input `\uDC00ab` decodes to `[56320, 97, 98]`, treating the unmatched low surrogate as code point 0xDC00.

- **High surrogate followed by another high surrogate** (tests.js:156–158): Input `\uD800\uD800` decodes to `[0xD800, 0xD800]`.

- **Unmatched high surrogate, surrogate pair, unmatched high surrogate** (tests.js:161–163): Input `\uD800\uD834\uDF06\uD800` decodes to `[0xD800, 0x1D306, 0xD800]`.

- **Low surrogate followed by another low surrogate** (tests.js:166–168): Input `\uDC00\uDC00` decodes to `[0xDC00, 0xDC00]`.

- **Unmatched low surrogate, surrogate pair, unmatched low surrogate** (tests.js:171–173): Input `\uDC00\uD834\uDF06\uDC00` decodes to `[0xDC00, 0x1D306, 0xDC00]`.

## Error Conditions (Quirk)

The test suite at tests.js:255–270 includes two assertions labeled for `punycode.ucs2.decode`, but they actually invoke `punycode.decode` (the Punycode decoder, not the UCS-2 decoder). These tests are colocated by historical accident and are fully specified in `punycode-decode.md`:

- `punycode.decode('\x81-')` throws `RangeError: Illegal input >= 0x80 (not a basic code point)` (tests.js:255–262). Error message defined at punycode.js:24; thrown via `error('not-basic')` at punycode.js:41–43.

- `punycode.decode('\x81')` throws `RangeError: Overflow: input needs wider integers to process` (tests.js:263–270). Error message defined at punycode.js:23; thrown via `error('overflow')` at punycode.js:41–43.

## Cross-references

- `test-fixtures.md` — consolidated test data definitions.
- `punycode-ucs2-encode.md` — inverse operation (code points to UCS-2 string).
- `punycode-decode.md` — Punycode (domain-name) decoder, which processes the output of `ucs2.decode`.
