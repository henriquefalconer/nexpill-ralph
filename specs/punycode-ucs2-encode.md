# punycode.ucs2.encode

## Purpose

Converts an array of Unicode code points back into a UCS-2/UTF-16 string. This function is the inverse of `punycode.ucs2.decode` (see `punycode-ucs2-decode.md`).

## Signature

**Input:** An array of non-negative integers, each representing a Unicode code point.

**Output:** A string containing the encoded UCS-2/UTF-16 representation of the code points.

## Behavior

The function processes each code point in the input array and emits the corresponding character(s) according to these rules (punycode.js:133):

- Each code point >= 0x10000 MUST be emitted as a surrogate pair per UTF-16 encoding. For example, code point 0x1D306 is encoded as the high surrogate 0xD834 followed by the low surrogate 0xDF06.
- Each code point < 0x10000 MUST be emitted as a single code unit with that value.
- Code points in the surrogate range (0xD800–0xDFFF) are emitted as single code units at their literal values, producing lone surrogates in the output string (tests.js:146–173).
- The input array MUST NOT be mutated during processing (tests.js:282–287). The implementation uses the spread operator with `String.fromCodePoint`, which does not modify the original array.

## Examples

The following test cases are derived from `testData.ucs2` (tests.js:137–174):

| Input Code Points | Output String | Description |
|---|---|---|
| [127829, 119808, 119558, 119638] | `\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56` | Consecutive astral symbols (tests.js:141–143) |
| [55296, 97, 98] | `\uD800ab` | High surrogate (U+D800) followed by non-surrogates (tests.js:146–148) |
| [56320, 97, 98] | `\uDC00ab` | Low surrogate (U+DC00) followed by non-surrogates (tests.js:150–153) |
| [0xD800, 0xD800] | `\uD800\uD800` | High surrogate followed by another high surrogate (tests.js:156–158) |
| [0xD800, 0x1D306, 0xD800] | `\uD800\uD834\uDF06\uD800` | Unmatched high surrogate, surrogate pair, unmatched high surrogate (tests.js:161–163) |
| [0xDC00, 0xDC00] | `\uDC00\uDC00` | Low surrogate followed by another low surrogate (tests.js:166–168) |
| [0xDC00, 0x1D306, 0xDC00] | `\uDC00\uD834\uDF06\uDC00` | Unmatched low surrogate, surrogate pair, unmatched low surrogate (tests.js:171–173) |
| [0x61, 0x62, 0x63] | `abc` | ASCII characters; input array must not be mutated (tests.js:282–287) |

## Cross-references

- [test-fixtures.md](test-fixtures.md) — full fixture descriptions
- [punycode-ucs2-decode.md](punycode-ucs2-decode.md) — inverse operation
