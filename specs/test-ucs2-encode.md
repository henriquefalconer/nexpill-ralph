# Spec: `punycode.ucs2.encode` test block

## 1. Subject

**Function:** `punycode.ucs2.encode(codePoints)`

**Signature:** Takes a single argument `codePoints`, an array of numeric Unicode code points, and returns a UCS-2 string constructed from those code points.

## 2. Contract

Source: `punycode.js:125-132` (JSDoc) and `punycode.js:133` (implementation).

- Accepts an array of numeric code points (`punycode.js:130`).
- Returns a new Unicode string described as UCS-2 (`punycode.js:131`).
- Implemented as `codePoints => String.fromCodePoint(...codePoints)` (`punycode.js:133`). The spread passes every element of the array as a separate argument to `String.fromCodePoint`, which converts each code point to its corresponding character, including surrogate-pair encoding for code points above `0xFFFF`.
- Does not mutate the input array. Because the implementation spreads the array into `String.fromCodePoint` without modifying it, the original array is guaranteed to remain unchanged after the call (`punycode.js:133`).
- Is exposed publicly as `punycode.ucs2.encode` (`punycode.js:435`).

## 3. Test cases

The describe block is at `tests/tests.js:273-288`.

### 3.1 Parameterized round-trip vectors

`tests/tests.js:274-281` loops over every object in `testData.ucs2` (`tests/tests.js:137-175`). For each object it calls `punycode.ucs2.encode(object.decoded)` and asserts deep equality with `object.encoded`. The vectors are defined below, with their line ranges in `tests/tests.js`.

| Description | `decoded` (input) | `encoded` (expected output) | Lines |
|---|---|---|---|
| Consecutive astral symbols | `[127829, 119808, 119558, 119638]` | `'\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56'` | `tests/tests.js:140-144` |
| U+D800 (high surrogate) followed by non-surrogates | `[55296, 97, 98]` | `'\uD800ab'` | `tests/tests.js:145-149` |
| U+DC00 (low surrogate) followed by non-surrogates | `[56320, 97, 98]` | `'\uDC00ab'` | `tests/tests.js:150-154` |
| High surrogate followed by another high surrogate | `[0xD800, 0xD800]` | `'\uD800\uD800'` | `tests/tests.js:155-159` |
| Unmatched high surrogate, followed by a surrogate pair, followed by an unmatched high surrogate | `[0xD800, 0x1D306, 0xD800]` | `'\uD800\uD834\uDF06\uD800'` | `tests/tests.js:160-164` |
| Low surrogate followed by another low surrogate | `[0xDC00, 0xDC00]` | `'\uDC00\uDC00'` | `tests/tests.js:165-169` |
| Unmatched low surrogate, followed by a surrogate pair, followed by an unmatched low surrogate | `[0xDC00, 0x1D306, 0xDC00]` | `'\uDC00\uD834\uDF06\uDC00'` | `tests/tests.js:170-174` |

### 3.2 Non-mutation test

`tests/tests.js:282-287`.

- `tests/tests.js:282`: `const codePoints = [0x61, 0x62, 0x63]` is defined outside the `it` callback.
- `tests/tests.js:283`: `punycode.ucs2.encode(codePoints)` is called immediately, storing the result in `result`.
- `tests/tests.js:284-287`: A single `it('does not mutate argument array', ...)` callback asserts two things:
  1. `result` deep-equals the string `'abc'` — confirming correct encoding of the three code points.
  2. `codePoints` deep-equals `[0x61, 0x62, 0x63]` — confirming that the original input array was not altered by the call.

The non-mutation guarantee is a direct consequence of the implementation at `punycode.js:133`: spreading the array into `String.fromCodePoint` reads elements without writing back to the array.

## 4. Implementation citations

- `punycode.js:133` — arrow function definition: `const ucs2encode = codePoints => String.fromCodePoint(...codePoints);`
- `punycode.js:435` — public binding: `'encode': ucs2encode` within the `punycode.ucs2` object literal.

## 5. Notes for porting

**Astral code points produce surrogate pairs.** Several input vectors contain code points above `0xFFFF` — for example `127829` and `119808` at `tests/tests.js:142`. The expected output for these contains surrogate pairs (e.g. `'\uD83C\uDF55'` for `127829`). Any port must produce the correct surrogate-pair encoding for such values.

**Surrogate code points are emitted as-is.** Several input vectors contain code points in the surrogate range (`0xD800`-`0xDFFF`) — for example `0xD800` and `0xDC00` appear directly in decoded arrays at `tests/tests.js:147-172`. The expected output faithfully includes these surrogate code units in the string unchanged. This is a deliberate deviation from strict UTF-16 validity: the function treats its input as an array of UCS-2 code units conceptually, and makes no attempt to validate or reject unpaired surrogates. A port must replicate this behavior and must not raise an error or substitute a replacement character when surrogate code points appear in the input.
