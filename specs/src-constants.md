# Spec: Top-level constants, regex, error messages, and convenience shortcuts

Source block: `punycode.js:1-31`

---

## 1. Subject

**Block purpose:** Declares Bootstring (RFC 3492) algorithm parameters, regular expressions for domain label inspection, error messages, and utility function shortcuts that are used throughout the Punycode encoder and decoder.

**External reference:** [RFC 3492 – Bootstring: a generalized ASCII-compatible encoding syntax](https://tools.ietf.org/html/rfc3492)

This block establishes the foundational numeric, symbolic, and messaging constants that configure how the Punycode algorithm encodes and decodes Unicode domain labels. Every parameter is specified in RFC 3492 Section 5.

---

## 2. Bootstring Parameters

Bootstring is a generic algorithm for encoding Unicode using only ASCII characters. The following parameters tune its behavior for Punycode domain labels. All are RFC 3492 defaults (ibid., Section 5.1).

| Constant | Value | Meaning | RFC 3492 | Used at |
|---|---|---|---|---|
| `maxInt` (`punycode.js:4`) | `2147483647` (0x7FFFFFFF, 2^31−1) | Highest positive signed 32-bit integer; used as an overflow guard to detect when accumulated deltas or indices exceed the safe range for further computation. | Section 5.1 | :243, :255, :268, :327, :337, :345 |
| `base` (`punycode.js:7`) | `36` | Alphabet size. Valid digits are 0–9 (values 0–9) and a–z (values 10–35), forming a 36-character set. Used to determine how many code points can be represented per iteration. | Section 5.1 | :183, :184, :186, :232, :240, :254, :255, :259, :351, :357, :359, :361 |
| `tMin` (`punycode.js:8`) | `1` | Minimum threshold. Part of a step-function determining which digits are active in each bias window. | Section 5.1 | :248, :352 (within bias threshold computations) |
| `tMax` (`punycode.js:9`) | `26` | Maximum threshold. Used with `tMin` to define the range of active digits. | Section 5.1 | :183, :248, :352 |
| `skew` (`punycode.js:10`) | `38` | Skew parameter in bias adaptation. Adjusts how aggressively the bias shifts after each output digit. | Section 5.1 | :186 (bias computation) |
| `damp` (`punycode.js:11`) | `700` | Damping parameter for bias adaptation; used in the first iteration to slow the rate at which the algorithm adapts its bias. | Section 5.1 | :181 (bias adaptation on first iteration) |
| `initialBias` (`punycode.js:12`) | `72` | Starting bias value before any input is processed. Controls which digits are active on the first iteration. | Section 5.1 | :202 (encode), :302 (decode) |
| `initialN` (`punycode.js:13`) | `128` (0x80) | Starting code point value. Bootstring always begins by assuming all ASCII (code points 0–127) have been output; subsequent iterations encode code points >= 128. | Section 5.1 | :201 (encode), :300 (decode) |
| `delimiter` (`punycode.js:14`) | `'-'` (U+002D, HYPHEN-MINUS) | Separator between the basic (unencoded ASCII) and encoded portions of a label. | Section 5.1 | :208 (find last), :319 (append) |

---

## 3. Regular Expressions

**Match patterns for domain label classification.**

| Constant | Pattern | Purpose | Unicode codepoints matched / notes |
|---|---|---|---|
| `regexPunycode` (`punycode.js:17`) | `/^xn--/` | Tests whether a string is a Punycode label (i.e., starts with the "xn--" ASCII-compatible encoding prefix). | Literal ASCII "xn--" (U+0078, U+006E, U+002D, U+002D) |
| `regexNonASCII` (`punycode.js:18`) | `/[^\0-\x7F]/` | Tests whether a string contains any non-ASCII character. Matches any code unit outside the range U+0000 to U+007F (ASCII printable + control). Note: U+007F (DEL) is excluded. | Any Unicode codepoint >= U+0080 |
| `regexSeparators` (`punycode.js:19`) | `/[\x2E\u3002\uFF0E\uFF61]/g` | Matches domain label separators per RFC 3490. Global flag allows replacement of all occurrences. Separators recognized: U+002E (FULL STOP, the ASCII dot), U+3002 (IDEOGRAPHIC FULL STOP, Chinese/Japanese period), U+FF0E (FULLWIDTH FULL STOP, full-width dot for CJK), U+FF61 (HALFWIDTH IDEOGRAPHIC FULL STOP, half-width form). | U+002E, U+3002, U+FF0E, U+FF61 |

**Usage:** `regexPunycode` is used at :391 to check if a label needs decoding. `regexNonASCII` is used at :410 to check if a label needs encoding. `regexSeparators` is used at :82 to normalize domain separators to ASCII dots before splitting.

---

## 4. Error Messages

**Named error conditions and their user-facing messages.**

The `errors` object (`:22-26`) maps error type keys to `RangeError` messages. Error handling is centralized in the `error(type)` function (`:41-43`), which throws `RangeError(errors[type])`.

| Key | Message | Trigger condition |
|---|---|---|
| `'overflow'` (`:23`) | `'Overflow: input needs wider integers to process'` | A numeric accumulation (`delta`, `i`, `w`) exceeds `maxInt`, indicating the input requires integer arithmetic wider than 32 bits. Checked at :243, :255, :268, :337, :345. |
| `'not-basic'` (`:24`) | `'Illegal input >= 0x80 (not a basic code point)'` | A character in the basic (pre-delimiter) portion of a Punycode label has code point >= 0x80. Checked at :215-216 during decode. |
| `'invalid-input'` (`:25`) | `'Invalid input'` | Catch-all for malformed input (e.g., invalid digit value). Checked at :240 during decode. |

---

## 5. Convenience Shortcuts

**Pre-computed and aliased values to reduce code size and improve runtime performance.**

| Constant | Definition | Purpose | Used at |
|---|---|---|---|
| `baseMinusTMin` (`punycode.js:29`) | `base - tMin` = 36 − 1 = `35` | Pre-computed difference; used in bias calculations and threshold tests rather than recomputing `base - tMin` repeatedly. | :183, :184, :186, :255, :259, :359, :361 |
| `floor` (`punycode.js:30`) | `Math.floor` | Alias to the `Math.floor` function. Saves bytes by using a shorter identifier and avoids repeated object property lookups. | :181, :183, :184, :186, :243, :255, :259, :268, :337, :351, :361 |
| `stringFromCharCode` (`punycode.js:31`) | `String.fromCharCode` | Alias to `String.fromCharCode` for the same reason. Creates a string from an array of Unicode code points. | :359 (during encoding digit-to-char conversion) |

---

## 6. Cross-references and data flow

### Flow: Encoding a Unicode label

1. `encode()` function (`:273-365`) initializes `n = initialN` (`:300`), `bias = initialBias` (`:302`).
2. Loops over increasing code point values; for each code point, computes a `delta` (distance from previous code point).
3. For each delta, applies bias adaptation (`:340-344`) using `damp`, `baseMinusTMin`, `skew` to adjust the bias.
4. Outputs digits in base 36 (`:351-361`); checks against `tMin`, `tMax`, and `base` to compute the digit value.
5. Appends `delimiter` (`:319`) if output is non-empty.

### Flow: Decoding a Punycode label

1. `decode()` function (`:197-268`) initializes `n = initialN` (`:201`), `bias = initialBias` (`:202`).
2. Splits the input at `delimiter` (`:208`), taking the part after the last occurrence as the encoded digits.
3. Loops over the encoded digits (`:232-268`); each digit is converted via `basicToDigit()` (`:137-155`) and validated against `base`.
4. Accumulates `delta` and `i` values; checks against `maxInt` to prevent overflow (`:243, :255, :268`).
5. Applies bias adaptation to compute the next active code point value and updates `n` and `bias`.

### Regex workflow

- `regexPunycode` (`:391`) determines whether a label is already encoded and needs decoding.
- `regexNonASCII` (`:410`) determines whether a label needs encoding.
- `regexSeparators` (`:82`) normalizes domain separator characters before label splitting.

---

## 7. Implementation citations

| Item | Location |
|---|---|
| Bootstring parameter block | `punycode.js:3-14` |
| Regex block | `punycode.js:16-19` |
| Error messages object | `punycode.js:21-26` |
| Convenience shortcuts block | `punycode.js:28-31` |
| Error helper function | `punycode.js:41-43` |
| `basicToDigit` (uses `base`, regex handling) | `punycode.js:137-155` |
| `digitToBasic` (uses `tMin`, `tMax`) | `punycode.js:159-167` |
| `bias` function (uses `baseMinusTMin`, `tMax`, `skew`) | `punycode.js:178-187` |
| `mapDomain` (uses `regexSeparators`) | `punycode.js:72-86` |
| `encode` function (uses all Bootstring params, error checks) | `punycode.js:273-365` |
| `decode` function (uses all Bootstring params, error checks) | `punycode.js:197-268` |
| Label-inspection public APIs (use regex) | `punycode.js:389-411` |
