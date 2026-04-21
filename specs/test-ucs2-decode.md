# Spec: `punycode.ucs2.decode` test block

Source block: `tests/tests.js:245-271`

---

## 1. Subject

**Function under test:** `punycode.ucs2.decode`

**Binding:** `punycode.ucs2.decode` is bound to the internal `ucs2decode` function at `punycode.js:434`, inside the `punycode.ucs2` object declared at `punycode.js:433-436`.

**Signature:**

- Input: a UCS-2 string (JavaScript's native string type, which internally uses UTF-16 / UCS-2 code units).
- Output: an array of Unicode code points (non-negative integers), one entry per logical Unicode character.

---

## 2. Contract

Derived from the JSDoc at `punycode.js:88-100` and the implementation at `punycode.js:101-123`.

The function iterates over the code units of a UCS-2 string one at a time. For each code unit it encounters:

- If the code unit is a high surrogate (0xD800–0xDBFF) and is immediately followed by a low surrogate (0xDC00–0xDFFF), it combines the pair into a single astral code point and advances the cursor past both code units.
- If the code unit is a high surrogate but is not followed by a low surrogate (i.e., it is an unmatched high surrogate), it emits the raw high-surrogate code unit value as-is and does not advance the cursor past the following code unit, so that the next iteration can re-examine it as a potential high surrogate.
- In all other cases (non-surrogates, lone low surrogates), the code unit value is emitted directly.

The result is an array of code points that correctly represents the full Unicode scalar values of the string, treating surrogate pairs as single astral characters while preserving unmatched surrogates as raw code unit values.

---

## 3. Test cases

### 3.1 Parameterized loop — `tests/tests.js:246-254`

The loop iterates over `testData.ucs2` (defined at `tests/tests.js:137-175`). For each entry it calls `punycode.ucs2.decode(object.encoded)` and asserts deep equality with `object.decoded` (`tests/tests.js:248-252`).

The direction of the test is: `object.encoded` is the UCS-2 input string; `object.decoded` is the expected array of code points. This matches the function semantics: decode a UCS-2 string into an array of Unicode code points.

Each vector (`tests/tests.js:137-175`):

#### Vector 1 — Consecutive astral symbols (`tests/tests.js:141-144`)

- Description: `'Consecutive astral symbols'`
- Input string: `'\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56'` (four surrogate pairs back-to-back)
- Expected output: `[127829, 119808, 119558, 119638]`
- Purpose: verifies that multiple consecutive valid surrogate pairs are each decoded into their respective astral code points without interference between pairs.

#### Vector 2 — U+D800 followed by non-surrogates (`tests/tests.js:145-149`)

- Description: `'U+D800 (high surrogate) followed by non-surrogates'`
- Input string: `'\uD800ab'`
- Expected output: `[55296, 97, 98]`
- Purpose: verifies that a high surrogate not followed by a low surrogate is emitted as its raw code unit value (55296 = 0xD800), and the following ASCII characters are emitted normally.

#### Vector 3 — U+DC00 followed by non-surrogates (`tests/tests.js:150-154`)

- Description: `'U+DC00 (low surrogate) followed by non-surrogates'`
- Input string: `'\uDC00ab'`
- Expected output: `[56320, 97, 98]`
- Purpose: verifies that a lone low surrogate at the start of the string is emitted as its raw code unit value (56320 = 0xDC00), and subsequent ASCII characters are emitted normally.

#### Vector 4 — High surrogate followed by another high surrogate (`tests/tests.js:155-159`)

- Description: `'High surrogate followed by another high surrogate'`
- Input string: `'\uD800\uD800'`
- Expected output: `[0xD800, 0xD800]`
- Purpose: verifies that when a high surrogate is followed by another high surrogate (not a low surrogate), both are emitted individually as unmatched surrogates.

#### Vector 5 — Unmatched high surrogate, surrogate pair, unmatched high surrogate (`tests/tests.js:160-164`)

- Description: `'Unmatched high surrogate, followed by a surrogate pair, followed by an unmatched high surrogate'`
- Input string: `'\uD800\uD834\uDF06\uD800'`
- Expected output: `[0xD800, 0x1D306, 0xD800]`
- Purpose: verifies correct re-examination behavior. The first `\uD800` is a high surrogate followed by another high surrogate (`\uD834`), so it is emitted as unmatched and the cursor backs up. The pair `\uD834\uDF06` is then decoded as astral code point 0x1D306. The final `\uD800` is a high surrogate at the end of the string (no following character), so it is emitted as unmatched.

#### Vector 6 — Low surrogate followed by another low surrogate (`tests/tests.js:165-169`)

- Description: `'Low surrogate followed by another low surrogate'`
- Input string: `'\uDC00\uDC00'`
- Expected output: `[0xDC00, 0xDC00]`
- Purpose: verifies that two consecutive low surrogates are each emitted individually as raw code unit values, since neither can form a valid pair (no preceding high surrogate at the point of emission).

#### Vector 7 — Unmatched low surrogate, surrogate pair, unmatched low surrogate (`tests/tests.js:170-174`)

- Description: `'Unmatched low surrogate, followed by a surrogate pair, followed by an unmatched low surrogate'`
- Input string: `'\uDC00\uD834\uDF06\uDC00'`
- Expected output: `[0xDC00, 0x1D306, 0xDC00]`
- Purpose: verifies that a leading lone low surrogate is emitted as unmatched, the subsequent valid surrogate pair `\uD834\uDF06` is decoded as 0x1D306, and the trailing lone low surrogate is emitted as unmatched.

---

### 3.2 Error tests — `tests/tests.js:255-270`

**Important quirk:** These two `it` blocks are filed inside the `describe('punycode.ucs2.decode', ...)` block (`tests/tests.js:245`) but they do NOT call `punycode.ucs2.decode`. They call `punycode.decode` (the Punycode label decoder, not the UCS-2 string decoder). This is a mis-filing in the existing test suite. Implementations porting these tests should be aware that the error behavior being specified belongs to `punycode.decode`, not `punycode.ucs2.decode`.

#### Error test 1 — `tests/tests.js:255-262`

- It label: `'throws RangeError: Illegal input >= 0x80 (not a basic code point)'`
- Call: `punycode.decode('\x81-')`
- Expected: throws `RangeError`
- Explanation: The input `'\x81-'` contains a hyphen-minus (`-`), which acts as the Punycode delimiter. This means `basic` is set to the index of the last delimiter, and the pre-delimiter portion contains `\x81` (code point 0x81, a non-basic character >= 0x80). The check at `punycode.js:215-216` calls `error('not-basic')` when any pre-delimiter character has code point >= 0x80. The `error` helper throws a `RangeError` with the message `'Illegal input >= 0x80 (not a basic code point)'` (`punycode.js:24`).

#### Error test 2 — `tests/tests.js:263-270`

- It label: `'throws RangeError: Overflow: input needs wider integers to process'`
- Call: `punycode.decode('\x81')`
- Expected: throws `RangeError`
- Explanation: The input `'\x81'` has no delimiter, so `basic` is 0 and the main decoding loop begins at index 0. The character `\x81` (code point 0x81) is processed by `basicToDigit`, which returns `base` (36) for any character outside the valid digit/letter ranges. The check at `punycode.js:240-242` compares the returned digit against `base` and calls `error('invalid-input')` when the digit equals or exceeds `base`, which also throws a `RangeError`. Alternatively, depending on execution path, the overflow guard at `punycode.js:243-245` may trigger `error('overflow')` with message `'Overflow: input needs wider integers to process'` (`punycode.js:23`). Both errors throw `RangeError`; the test asserts only the error type, not the message.

---

## 4. Implementation citations

| Symbol | Location |
|---|---|
| `ucs2decode` function definition | `punycode.js:101-123` |
| `punycode.ucs2.decode` binding | `punycode.js:434` |
| Surrogate pair combination formula | `punycode.js:111` |
| Unmatched high surrogate emission and cursor rollback | `punycode.js:113-116` |
| Non-surrogate / lone low surrogate emission | `punycode.js:118-120` |
| `errors['overflow']` message string | `punycode.js:23` |
| `errors['not-basic']` message string | `punycode.js:24` |
| `error('not-basic')` call site in `decode` | `punycode.js:215-216` |
| `error('invalid-input')` call site in `decode` | `punycode.js:240-242` |
| JSDoc for `ucs2decode` | `punycode.js:88-100` |

---

## 5. Surrogate handling details

The `ucs2decode` function (`punycode.js:101-123`) processes code units in a single forward pass with a mutable cursor.

**Valid surrogate pair (high then low):**

When the current code unit is in the high-surrogate range 0xD800–0xDBFF (`punycode.js:107`) and the next code unit is in the low-surrogate range 0xDC00–0xDFFF (tested via `(extra & 0xFC00) == 0xDC00` at `punycode.js:110`), the pair is combined into a single astral code point using the formula at `punycode.js:111`:

```
astral = ((high & 0x3FF) << 10) + (low & 0x3FF) + 0x10000
```

Both code units are consumed (the cursor advances twice).

**Unmatched high surrogate:**

When the current code unit is a high surrogate but the immediately following code unit is not a low surrogate (`punycode.js:112-117`), only the high-surrogate code unit value is appended to the output (`punycode.js:115`). The cursor is decremented by one (`punycode.js:116`) so that the next iteration re-examines the following code unit. This prevents the following code unit from being silently consumed as the low half of a non-pair.

**Lone low surrogate or non-surrogate:**

Any code unit that is not in the high-surrogate range falls through to `punycode.js:118-120`, where its raw code unit value is appended directly. This covers lone low surrogates (0xDC00–0xDFFF) as well as all BMP characters outside the surrogate range.
