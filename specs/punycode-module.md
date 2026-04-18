# punycode-module

## Purpose

This spec documents the top-level module structure of `punycode.js`: the constants, error-message table, regex patterns, convenience shortcuts, and the shape of the exported public API object. It covers everything declared at module scope that a Go porter must replicate in package-level declarations before implementing any function.

## Overflow Sentinel: `maxInt`

**Declaration:** `punycode.js:4`

```
const maxInt = 2147483647;  // 0x7FFFFFFF, 2^31 - 1
```

`maxInt` is the largest positive value of a signed 32-bit integer. The Bootstring algorithm (RFC 3492) performs all intermediate arithmetic inside the range of a 32-bit signed integer; any computation that would exceed this value triggers an overflow error. Every overflow guard in `decode` (punycode.js:243–245, 255–257, 268–270) and `encode` (punycode.js:337–338, 345–346) compares against `maxInt`.

In a Go port, use `const maxInt = math.MaxInt32` (2147483647) or the literal. The important constraint is that `maxInt` must equal exactly 2^31 − 1; using a wider integer type as the sentinel would silently change overflow behavior.

## Bootstring Parameter Constants

All declared at `punycode.js:7–14`. These are the fixed parameters of the Bootstring encoding scheme as defined in RFC 3492 §5.

| Name | Value | Line | Meaning |
|---|---|---|---|
| `base` | 36 | 7 | The radix of the generalized variable-length integer encoding. Digits range 0–35. |
| `tMin` | 1 | 8 | Minimum threshold value for a digit position in variable-length integer encoding. |
| `tMax` | 26 | 9 | Maximum threshold value for a digit position. |
| `skew` | 38 | 10 | Bias adaptation skew parameter used in `adapt`. |
| `damp` | 700 | 11 | Damping factor applied to delta on the very first call to `adapt`. |
| `initialBias` | 72 | 12 | Starting bias value for both encoder and decoder. |
| `initialN` | 128 | 13 | Starting code point threshold (0x80, the first non-ASCII code point). |
| `delimiter` | `'-'` | 14 | The ASCII hyphen-minus (U+002D, 0x2D) used to separate the basic-code-point prefix from the encoded suffix in a Punycode label. |

These constants must be reproducible verbatim in any port. Changing any value produces an incompatible codec.

## Regex Patterns

Declared at `punycode.js:17–19`. Three compiled patterns are used throughout the module.

### `regexPunycode` (punycode.js:17)

**Pattern:** `/^xn--/`

**Purpose:** Tests whether a domain label begins with the ACE prefix `xn--`. Used in `toUnicode` (punycode.js:391) to decide whether a label should be decoded. The match is case-sensitive; labels beginning with `XN--` or `Xn--` do not match and are passed through unchanged.

### `regexNonASCII` (punycode.js:18)

**Pattern:** `/[^\0-\x7F]/`

**Purpose:** Tests whether a string contains any character with a code unit value greater than U+007F. Used in `toASCII` (punycode.js:410) to decide whether a label needs Punycode encoding. The exclusion range is `\0` through `\x7F`, which means U+007F (DEL) is treated as ASCII and does NOT trigger encoding. This is a deliberate quirk documented in an inline comment. A label composed entirely of code points 0x00–0x7F (including control characters and DEL) passes through `toASCII` unchanged.

**Note:** The comment on punycode.js:18 reads: `// Note: U+007F DEL is excluded too.` — "excluded" here means excluded from the non-ASCII set, i.e., it is treated as ASCII.

### `regexSeparators` (punycode.js:19)

**Pattern:** `/[\x2E\u3002\uFF0E\uFF61]/g`

**Purpose:** Matches any of the four IDNA2003 label-separator characters and replaces them with U+002E (full stop) in `mapDomain` (punycode.js:82). The `g` flag causes all occurrences to be replaced, not just the first. The four separators are:

| Code Point | Name |
|---|---|
| U+002E | Full stop (period) |
| U+3002 | Ideographic full stop |
| U+FF0E | Fullwidth full stop |
| U+FF61 | Halfwidth ideographic full stop |

Used by: `mapDomain` (see `punycode-internal-helpers.md`), which is in turn used by `toASCII` and `toUnicode`.

## Error-Message Table

Declared at `punycode.js:22–26`. A plain key-to-string map called `errors` with three entries:

| Key | Message string |
|---|---|
| `'overflow'` | `'Overflow: input needs wider integers to process'` |
| `'not-basic'` | `'Illegal input >= 0x80 (not a basic code point)'` |
| `'invalid-input'` | `'Invalid input'` |

These strings are the exact text of `RangeError` messages thrown by the `error()` helper (punycode.js:41–43). A Go port should use these exact strings so that callers comparing error messages against known values continue to work. See `punycode-internal-helpers.md` for the `error()` function specification.

## Convenience Shortcuts

Declared at `punycode.js:29–31`. Three module-scope aliases:

| Name | Value | Line | Purpose |
|---|---|---|---|
| `baseMinusTMin` | `base - tMin` = 35 | 29 | Precomputed constant used in `adapt` (see `punycode-bias-adapt.md`) and in overflow guards during decode. Avoids recomputing `36 - 1` repeatedly. |
| `floor` | `Math.floor` | 30 | Alias for integer floor division; used throughout `adapt`, `decode`, and `encode`. In a Go port, integer division with truncation toward zero is equivalent for non-negative operands; all intermediate values in these algorithms are non-negative. |
| `stringFromCharCode` | `String.fromCharCode` | 31 | Converts a single numeric code point to a single-character string. Used in `encode` (punycode.js:307, 358, 364) when appending basic code points and encoded digits to the output. In Go, the equivalent is `string(rune(codePoint))`. |

## Public API Object Shape

Declared at `punycode.js:418–441`.

A single object `punycode` is assembled from all top-level function declarations and then exported. Its shape is:

```
punycode
  .version          string    "2.3.1"   (punycode.js:425)
  .ucs2             object
    .decode         function  ucs2decode (punycode.js:101-123; see punycode-ucs2-decode.md)
    .encode         function  ucs2encode (punycode.js:133; see punycode-ucs2-encode.md)
  .decode           function  decode     (punycode.js:196-281; see punycode-decode.md)
  .encode           function  encode     (punycode.js:290-376; see punycode-encode.md)
  .toASCII          function  toASCII    (punycode.js:408-414; see punycode-to-ascii.md)
  .toUnicode        function  toUnicode  (punycode.js:389-395; see punycode-to-unicode.md)
```

The `version` field value is the string `'2.3.1'` (punycode.js:425). This is a purely informational field; no algorithm depends on it.

## Module Export

`punycode.js:443`:

```
module.exports = punycode;
```

The entire API is exported as a single CommonJS module export. The exported object is exactly the `punycode` object described above. In a Go port, the equivalent is a package that exports these six symbols as exported functions/values.

## Cross-References

- [punycode-internal-helpers.md](./punycode-internal-helpers.md) — `error()`, `map()`, and `mapDomain()` helper functions
- [punycode-bootstring-digit.md](./punycode-bootstring-digit.md) — `basicToDigit()` and `digitToBasic()` digit-conversion helpers
- [punycode-bias-adapt.md](./punycode-bias-adapt.md) — `adapt()` bias adaptation function using `baseMinusTMin` and `floor`
- [punycode-decode.md](./punycode-decode.md) — `punycode.decode` implementation
- [punycode-encode.md](./punycode-encode.md) — `punycode.encode` implementation
- [punycode-to-ascii.md](./punycode-to-ascii.md) — `punycode.toASCII` implementation
- [punycode-to-unicode.md](./punycode-to-unicode.md) — `punycode.toUnicode` implementation
- [punycode-ucs2-decode.md](./punycode-ucs2-decode.md) — `punycode.ucs2.decode` implementation
- [punycode-ucs2-encode.md](./punycode-ucs2-encode.md) — `punycode.ucs2.encode` implementation
