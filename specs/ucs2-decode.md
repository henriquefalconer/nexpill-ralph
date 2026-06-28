# `punycode.ucs2.decode` — Behaviour Specification

## Overview

`punycode.ucs2.decode(string)` converts a JavaScript string (which the engine stores as a sequence of UTF-16 / UCS-2 code units) into a plain JavaScript `Array` of numeric Unicode code points. Its defining characteristic is that it collapses surrogate pairs — two adjacent 16-bit code units that together represent a single astral code point above U+FFFF — into one integer entry in the output array, matching the semantics of UTF-16 rather than the raw UCS-2 view. Lone (unmatched) surrogates are preserved as-is rather than discarded or errored on.

Source: `punycode.js:101-123`.

---

## Algorithm — Step-by-Step

The function is defined at `punycode.js:101-123`:

```
function ucs2decode(string) {           // punycode.js:101
    const output = [];                  // punycode.js:102
    let counter = 0;                    // punycode.js:103
    const length = string.length;       // punycode.js:104
    while (counter < length) {          // punycode.js:105
        const value = string.charCodeAt(counter++);  // punycode.js:106
        if (value >= 0xD800 && value <= 0xDBFF && counter < length) {
                                        // punycode.js:107
            const extra = string.charCodeAt(counter++);  // punycode.js:109
            if ((extra & 0xFC00) == 0xDC00) {            // punycode.js:110
                output.push(((value & 0x3FF) << 10) + (extra & 0x3FF) + 0x10000);
                                        // punycode.js:111
            } else {
                output.push(value);     // punycode.js:115
                counter--;              // punycode.js:116
            }
        } else {
            output.push(value);         // punycode.js:119
        }
    }
    return output;                      // punycode.js:122
}
```

### High-surrogate detection (`punycode.js:107`)

A code unit `value` is recognised as a high surrogate when it falls in the range `0xD800`–`0xDBFF` inclusive, **and** there is at least one more code unit remaining (`counter < length`). Both conditions must hold before the function speculatively reads the next unit.

### Low-surrogate detection (`punycode.js:110`)

The next code unit is read into `extra` (`punycode.js:109`). It is a low surrogate when `(extra & 0xFC00) == 0xDC00`, which tests that bits 15–10 equal `0b110111` — i.e. the value lies in `0xDC00`–`0xDFFF`.

### Surrogate-pair combination formula (`punycode.js:111`)

When both tests pass the two code units are combined into one astral code point:

```
codePoint = ((value & 0x3FF) << 10) + (extra & 0x3FF) + 0x10000
```

- `value & 0x3FF` extracts the 10 payload bits of the high surrogate.
- `extra & 0x3FF` extracts the 10 payload bits of the low surrogate.
- The high half is shifted left by 10 and added to the low half.
- Adding `0x10000` produces the final code point in the supplementary planes (U+10000–U+10FFFF).

### Unmatched-high-surrogate fallback (`punycode.js:112-117`)

If `extra` fails the low-surrogate test, the function:

1. Pushes the lone high surrogate value (`value`) onto `output` unchanged (`punycode.js:115`).
2. Decrements `counter` by one (`punycode.js:116`), rewinding back to `extra` so that it will be re-examined as the leading unit of the **next** iteration.

This design ensures that a sequence like `<high><high><low>` correctly preserves the first lone high surrogate and then pairs the second high surrogate with the following low surrogate.

If the high surrogate occurs at the very end of the string (`counter >= length` when the high-surrogate test fires), the outer `else` branch (`punycode.js:118-120`) is taken directly, pushing `value` unchanged.

### Non-surrogate and lone-low-surrogate pass-through (`punycode.js:118-120`)

Any code unit that is not a high surrogate (including lone low surrogates, BMP characters, and ASCII) reaches the `else` branch and is pushed directly to `output` as its raw 16-bit numeric value.

---

## Test Fixture (`testData.ucs2`)

The fixture is defined at `tests/tests.js:137-175`. The `describe` block iterates over every object in the array and asserts:

```js
assert.deepEqual(
    punycode.ucs2.decode(object.encoded),
    object.decoded,
    object.description
);
```
— `tests/tests.js:246-253`.

`assert.deepEqual` performs a recursive structural comparison, so every element of the returned array must equal the corresponding element of `decoded` in both value and type.

### Individual test cases

#### 1. Consecutive astral symbols (`tests/tests.js:140-144`)

| Property | Value |
|---|---|
| `encoded` | `'🍕𝐀𝌆𝍖'` |
| `decoded` | `[127829, 119808, 119558, 119638]` |

Four surrogate pairs are given back-to-back with no intervening BMP characters. Each adjacent `<high, low>` pair is combined by the formula at `punycode.js:111`. Verifies that the loop correctly advances `counter` past both code units of each pair and does not accidentally treat the low surrogate of one pair as the high surrogate of the next.

#### 2. Lone high surrogate followed by non-surrogates (`tests/tests.js:145-149`)

| Property | Value |
|---|---|
| `encoded` | `'\uD800ab'` |
| `decoded` | `[55296, 97, 98]` |

`\uD800` (0xD800) qualifies as a high surrogate (`punycode.js:107`). The next unit `'a'` (0x61) fails the low-surrogate test (`punycode.js:110`), so 0xD800 is pushed as a lone value (`punycode.js:115`) and `counter` is rewound (`punycode.js:116`). `'a'` and `'b'` then pass through the non-surrogate branch (`punycode.js:119`).

#### 3. Lone low surrogate followed by non-surrogates (`tests/tests.js:150-154`)

| Property | Value |
|---|---|
| `encoded` | `'\uDC00ab'` |
| `decoded` | `[56320, 97, 98]` |

`\uDC00` (0xDC00) does **not** satisfy `value >= 0xD800 && value <= 0xDBFF` (`punycode.js:107`) because 0xDC00 > 0xDBFF, so the non-surrogate branch is taken (`punycode.js:119`) and it is pushed as 56320 (0xDC00). `'a'` and `'b'` follow normally.

#### 4. High surrogate followed by another high surrogate (`tests/tests.js:155-159`)

| Property | Value |
|---|---|
| `encoded` | `'\uD800\uD800'` |
| `decoded` | `[0xD800, 0xD800]` |

The first 0xD800 triggers the high-surrogate branch (`punycode.js:107`); it reads the second 0xD800 as `extra`. The second value fails the low-surrogate test (`punycode.js:110`), so the first is pushed alone (`punycode.js:115`) and `counter` is rewound (`punycode.js:116`). The loop then processes the second 0xD800 in the same way, resulting in two separate entries.

#### 5. Unmatched high, surrogate pair, unmatched high (`tests/tests.js:160-164`)

| Property | Value |
|---|---|
| `encoded` | `'\uD800𝌆\uD800'` |
| `decoded` | `[0xD800, 0x1D306, 0xD800]` |

- Iteration 1: 0xD800 is a high surrogate; next unit 0xD834 is also a high surrogate (fails low-surrogate test). 0xD800 is pushed lone (`punycode.js:115`), counter rewinds (`punycode.js:116`).
- Iteration 2: 0xD834 is a high surrogate; next unit 0xDF06 passes the low-surrogate test (`punycode.js:110`). They combine to `((0x34 << 10) + 0x306 + 0x10000)` = 0x1D306 (`punycode.js:111`).
- Iteration 3: 0xD800 is at the end; the unmatched fallback in the non-surrogate `else` branch (`punycode.js:119`) pushes it as-is, because `counter` now equals `length` so the guard `counter < length` in `punycode.js:107` fails.

Result: `[0xD800, 0x1D306, 0xD800]` — three entries.

#### 6. Low surrogate followed by another low surrogate (`tests/tests.js:165-169`)

| Property | Value |
|---|---|
| `encoded` | `'\uDC00\uDC00'` |
| `decoded` | `[0xDC00, 0xDC00]` |

Neither 0xDC00 value satisfies the high-surrogate range check (`punycode.js:107`), so both pass through the non-surrogate branch (`punycode.js:119`) as plain 16-bit values.

#### 7. Unmatched low, surrogate pair, unmatched low (`tests/tests.js:170-174`)

| Property | Value |
|---|---|
| `encoded` | `'\uDC00𝌆\uDC00'` |
| `decoded` | `[0xDC00, 0x1D306, 0xDC00]` |

- Iteration 1: 0xDC00 is not a high surrogate; pushed directly (`punycode.js:119`).
- Iteration 2: 0xD834 is a high surrogate; 0xDF06 is a valid low surrogate (`punycode.js:110`); combined to 0x1D306 (`punycode.js:111`).
- Iteration 3: 0xDC00 is not a high surrogate; pushed directly (`punycode.js:119`).

Result: `[0xDC00, 0x1D306, 0xDC00]`.

---

## RangeError Tests

Both RangeError tests appear inside `describe('punycode.ucs2.decode', ...)` at `tests/tests.js:255-270` but they call `punycode.decode` (Punycode-to-Unicode decoding), not `punycode.ucs2.decode`. The `error` utility used by `decode` is defined at `punycode.js:41-43`; it always throws a `RangeError`.

Assertion mechanics: both tests use `assert.throws(fn, RangeError)` (`tests/tests.js:256`, `tests/tests.js:264`).

### Test 1 — "not-basic" error (`tests/tests.js:255-261`)

```js
punycode.decode('\x81-')
```

`\x81` has code point 0x81, which is >= 0x80. The `decode` function at `punycode.js:208-209` calls `input.lastIndexOf('-')` to find the Punycode delimiter. The `'-'` at index 1 is found, so `basic = 1`.

The loop at `punycode.js:213-218` copies the first `basic` (i.e. 1) code units to `output`, checking each one:

```js
if (input.charCodeAt(j) >= 0x80) {   // punycode.js:215
    error('not-basic');               // punycode.js:216
}
```

`\x81` (index 0) has `charCodeAt` value 0x81, which is >= 0x80, so `error('not-basic')` is called immediately, throwing `RangeError: Illegal input >= 0x80 (not a basic code point)` (`punycode.js:23-24`, `punycode.js:41-43`).

### Test 2 — "overflow" error (`tests/tests.js:263-270`)

```js
punycode.decode('\x81')
```

`\x81` has no `'-'` character, so `input.lastIndexOf(delimiter)` returns -1, and `basic` is set to 0 (`punycode.js:208-210`). The basic-copy loop is skipped entirely (runs 0 iterations).

The variable-length integer decoding loop begins at `punycode.js:224` with `index = 0` (since `basic` is 0, `punycode.js:224` starts at `index = 0`). On the first inner-loop iteration:

```js
const digit = basicToDigit(input.charCodeAt(index++));  // punycode.js:238
```

`input.charCodeAt(0)` is 0x81. `basicToDigit(0x81)` returns `base` (36) because 0x81 does not fall into any of the ASCII digit/letter ranges (`punycode.js:144-155`).

The check immediately following is:

```js
if (digit >= base) {                  // punycode.js:240
    error('invalid-input');           // punycode.js:241
}
```

`digit` (36) equals `base` (36), so `digit >= base` is true and `error('invalid-input')` is called, throwing `RangeError: Invalid input` (`punycode.js:25`, `punycode.js:41-43`).

> **Note on the test description vs. actual error:** The test at `tests/tests.js:263` is labelled *"throws RangeError: Overflow: input needs wider integers to process"*, but the actual error thrown by `punycode.decode('\x81')` is `RangeError: Invalid input` (the `'invalid-input'` branch at `punycode.js:240-241`). Because `assert.throws` only checks the constructor (`RangeError`) and not the message (`tests/tests.js:264`), the test still passes. The description in the test suite is therefore misleading — the code never reaches the overflow checks at `punycode.js:243-245` or `punycode.js:255-257` for this input.

---

## Summary of Assertion Mechanics

| Assertion type | Method | Location |
|---|---|---|
| Fixture round-trip equality | `assert.deepEqual(punycode.ucs2.decode(encoded), decoded)` | `tests/tests.js:248-251` |
| RangeError for not-basic | `assert.throws(fn, RangeError)` | `tests/tests.js:256-261` |
| RangeError for invalid/overflow | `assert.throws(fn, RangeError)` | `tests/tests.js:264-269` |

`assert.deepEqual` performs a recursive value-equality check: the returned array must have the same length and each element must equal the expected code-point integer. It does not require reference equality, only structural equality.

---

## Cross-References

- [ucs2-encode.md](ucs2-encode.md) — the inverse operation (`punycode.ucs2.encode`), tested over the same `testData.ucs2` fixtures.
- [decode.md](decode.md) — the `punycode.decode` function exercised by the two RangeError tests above.
