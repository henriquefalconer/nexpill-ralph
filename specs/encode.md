# `punycode.encode` Behavior Specification

## Overview

`punycode.encode(input)` converts a string of Unicode symbols into a bare Punycode ASCII-only string per RFC 3492 (Bootstring). The function is defined at `punycode.js:290-376`. It returns only the encoded label, **without** the `xn--` ACE prefix; that prefix is prepended exclusively by `toASCII` (`punycode.js:411`) when processing domain-name labels.

---

## Bootstring Parameters

The algorithm is governed by six constants declared at `punycode.js:7-14`:

| Parameter      | Value | Constant name   |
|----------------|-------|-----------------|
| `base`         | 36    | `base`          |
| `tMin`         | 1     | `tMin`          |
| `tMax`         | 26    | `tMax`          |
| `skew`         | 38    | `skew`          |
| `damp`         | 700   | `damp`          |
| `initialBias`  | 72    | `initialBias`   |
| `initialN`     | 128   | `initialN`      |
| `delimiter`    | `'-'` | `delimiter`     |

The derived constant `baseMinusTMin` (`base - tMin = 35`) is computed once at `punycode.js:29`.

Overflow is bounded by `maxInt = 2147483647` (`punycode.js:4`).

---

## Algorithm (Step by Step)

### Step 1 — UCS-2 to Code Points (`ucs2decode`)

`punycode.js:294`:

```js
input = ucs2decode(input);
```

`ucs2decode` (`punycode.js:101-123`) walks the input string one UTF-16 code unit at a time. When it encounters a high surrogate (`0xD800–0xDBFF`) immediately followed by a low surrogate (`0xDC00–0xDFFF`), it combines them into a single supplementary code point (`((high & 0x3FF) << 10) + (low & 0x3FF) + 0x10000`, `punycode.js:111`). Unmatched surrogates are emitted as-is (`punycode.js:115`). The result is a plain JavaScript array of integer code points.

### Step 2 — State Initialization

`punycode.js:300-302`:

- `n = initialN` (128 / `0x80`) — the current code-point threshold
- `delta = 0` — running offset for the generalized integer encoder
- `bias = initialBias` (72) — current bias for threshold calculation

### Step 3 — Copy Basic Code Points

`punycode.js:305-309`: Every code point `< 0x80` (i.e., ASCII) is appended directly to `output` as its character.

`punycode.js:311-312`: `basicLength` is recorded (number of basic characters emitted), and `handledCPCount` is initialised to the same value.

`punycode.js:318-320`: If at least one basic code point was emitted, a literal `'-'` delimiter is appended to `output`.

- **Pure-ASCII input** (e.g., `'-> $1.00 <-'`): all characters go through unchanged, followed by a single extra `'-'`, producing `'-> $1.00 <--'` (`tests/tests.js:131-135`).
- **Pure-non-ASCII input** (e.g., `'\xFC'`): `basicLength` is 0, no delimiter is appended, and the output begins directly with the variable-length integer encoding (`tests/tests.js:14-17`).

### Step 4 — Main Encoding Loop

`punycode.js:323`: The outer `while` loop continues until every code point has been handled (`handledCPCount < inputLength`).

#### 4a. Find the Next Smallest Unhandled Code Point

`punycode.js:327-332`: `m` is set to `maxInt`, then every code point in `input` is scanned; `m` is narrowed to the smallest value that is `>= n` and `< m`. This finds the next non-basic code point to be encoded in this pass.

#### 4b. Advance Delta with Overflow Guard

`punycode.js:336-341`:

```
delta += (m - n) * (handledCPCount + 1)
```

Before the addition, an overflow check is performed (`punycode.js:337`):

```js
if (m - n > floor((maxInt - delta) / handledCPCountPlusOne)) {
    error('overflow');
}
```

On overflow, a `RangeError` with message `'Overflow: input needs wider integers to process'` is thrown (`punycode.js:22-23`, `punycode.js:41-43`).

`n` is then set to `m` (`punycode.js:342`).

#### 4c. Emit Variable-Length Integers for Each Occurrence of n

`punycode.js:344-369`: The inner `for` loop iterates over all code points again:

- For any code point strictly less than the current `n`: `delta` is incremented by 1 (`punycode.js:345`), with overflow guarded by `> maxInt`.
- When a code point equals `n` (`punycode.js:348`), a generalized variable-length base-36 integer is emitted for the current `delta`:

  `punycode.js:350-363`: Starting with `q = delta`, an inner loop emits digits one at a time. The threshold `t` for each position `k` is:

  ```
  t = k <= bias  ? tMin
    : k >= bias + tMax ? tMax
    : k - bias
  ```

  (`punycode.js:352`). This is the RFC 3492 threshold formula (section 5). While `q >= t`, the digit emitted is `t + (q - t) % (base - t)` and `q` is floored-divided by `(base - t)` (`punycode.js:358-361`). When `q < t`, the final digit `q` is emitted (`punycode.js:364`).

  Each digit value is converted to its ASCII character via `digitToBasic` (`punycode.js:168-172`):

  ```js
  const digitToBasic = function(digit, flag) {
      return digit + 22 + 75 * (digit < 26) - ((flag != 0) << 5);
  };
  ```

  Digits 0–25 map to lowercase `a–z`; digits 26–35 map to `0–9`. The `flag` parameter controls case: `encode` always passes `0` (`punycode.js:359`, `punycode.js:364`), so **output is always lowercase**.

  After emitting a complete integer, bias is updated via `adapt` (`punycode.js:365`), `delta` is reset to 0 (`punycode.js:366`), and `handledCPCount` is incremented (`punycode.js:367`).

#### 4d. Advance n and delta

`punycode.js:371-372`: After the inner loop, `delta` is incremented by 1 and `n` is incremented by 1, preparing for the next pass.

### Step 5 — Return

`punycode.js:375`: `output.join('')` concatenates all emitted characters (basic code points, delimiter if any, and encoded non-basic digits) into the final ASCII string.

---

## Bias Adaptation (`adapt`)

`punycode.js:179-187`, implementing RFC 3492 section 3.4:

```js
const adapt = function(delta, numPoints, firstTime) {
    let k = 0;
    delta = firstTime ? floor(delta / damp) : delta >> 1;
    delta += floor(delta / numPoints);
    for (; delta > baseMinusTMin * tMax >> 1; k += base) {
        delta = floor(delta / baseMinusTMin);
    }
    return floor(k + (baseMinusTMin + 1) * delta / (delta + skew));
};
```

- On the first call (`firstTime = true`, i.e., when `handledCPCount === basicLength`, `punycode.js:365`), delta is divided by `damp` (700).
- On subsequent calls, delta is halved.
- The loop and final formula bring the bias into the range where the next insertion point's digit threshold is accurate.

---

## Test Suite: `describe('punycode.encode', ...)`

### Structure

`tests/tests.js:312-321`:

```js
describe('punycode.encode', function() {
    for (const object of testData.strings) {
        it(object.description || object.decoded, function() {
            assert.deepEqual(
                punycode.encode(object.decoded),
                object.encoded
            );
        });
    }
});
```

- **Iteration source**: `testData.strings` (`tests/tests.js:7-136`), an array of fixture objects, each with `decoded` (Unicode string) and `encoded` (Punycode string) fields, plus an optional `description`.
- **Test name**: `object.description` if present, otherwise `object.decoded` (`tests/tests.js:314`).
- **Assertion**: `assert.deepEqual(punycode.encode(object.decoded), object.encoded)` (`tests/tests.js:315-317`). `deepEqual` is Node's built-in `assert.deepEqual` (`tests/tests.js:3`).
- **Inverse relationship**: This suite is the exact inverse of `describe('punycode.decode', ...)` (`tests/tests.js:290-310`), which iterates the same fixture set and calls `punycode.decode(object.encoded)` expecting `object.decoded`. Every fixture must round-trip correctly.

### Fixture Coverage

The fixtures in `testData.strings` (`tests/tests.js:7-136`) cover the following categories:

**Pure ASCII with trailing delimiter**
- `tests/tests.js:8-12`: `'Bach'` → `'Bach-'`. A single basic code point. All four characters are ASCII, so they are copied verbatim and the delimiter is appended.

**Single non-ASCII character**
- `tests/tests.js:13-17`: `'\xFC'` (U+00FC, ü) → `'tda'`. No basic code points, no delimiter, pure encoded integer.

**Multiple non-ASCII characters only**
- `tests/tests.js:18-22`: `'\xFC\xEB\xE4\xF6♥'` → `'4can8av2009b'`.

**Mixed ASCII and non-ASCII**
- `tests/tests.js:23-27`: `'bücher'` → `'bcher-kva'`. Basic characters `bcher` are copied, delimiter appended, then non-ASCII `ü` is encoded.
- `tests/tests.js:28-32`: Long German string with both ASCII words and non-ASCII characters.

**RFC 3492 Section 7.1 Language Samples** (`tests/tests.js:33-125`):
- Arabic (Egyptian), `tests/tests.js:34-38`
- Chinese (Simplified), `tests/tests.js:39-43`
- Chinese (Traditional), `tests/tests.js:44-48`
- Czech, `tests/tests.js:49-53`
- Hebrew, `tests/tests.js:54-58`
- Hindi (Devanagari), `tests/tests.js:59-63`
- Japanese (kanji and hiragana), `tests/tests.js:64-68`
- Korean (Hangul syllables), `tests/tests.js:69-73`
- Russian (Cyrillic), `tests/tests.js:83-87` (see mixed-case annotation note below)
- Spanish, `tests/tests.js:88-92`
- Vietnamese, `tests/tests.js:93-97`

**Additional RFC 3492 examples** (`tests/tests.js:98-125`): Seven fixtures without `description` fields, exercising strings with numerals, hyphens, ASCII uppercase letters, and Japanese katakana mixed with ASCII.

**ASCII edge case**
- `tests/tests.js:131-135`:
  ```js
  {
      'description': 'ASCII string that breaks the existing rules for host-name labels',
      'decoded': '-> $1.00 <-',
      'encoded': '-> $1.00 <--'
  }
  ```
  This is a pure-ASCII string (all code points `< 0x80`). Every character is copied to output, then one `'-'` delimiter is appended, producing an output that ends with `'<--'` (the original trailing `'-'` plus the appended delimiter).

---

## Mixed-Case Annotation Limitation

`tests/tests.js:74-87` documents a known limitation:

```
As there's no way to do it in JavaScript, Punycode.js doesn't support
mixed-case annotation (which is entirely optional as per the RFC).
So, while the RFC sample string encodes to:
`b1abfaaepdrnnbgefbaDotcwatmq2g4l`
Without mixed-case annotation it has to encode to:
`b1abfaaepdrnnbgefbadotcwatmq2g4l`
https://github.com/mathiasbynens/punycode.js/issues/3
```

Because `digitToBasic` is always called with `flag = 0` (`punycode.js:359`, `punycode.js:364`), the subtraction `- ((flag != 0) << 5)` in `punycode.js:171` never fires, and digit values 0–25 always yield lowercase `a–z`. RFC 3492 allows (but does not require) mixed-case annotation to encode the case of basic code points in the original label; `punycode.encode` does not implement it. The Russian fixture (`tests/tests.js:83-87`) therefore expects the fully lowercase `'b1abfaaepdrnnbgefbadotcwatmq2g4l'` rather than the mixed-case `'b1abfaaepdrnnbgefbaDotcwatmq2g4l'` that a conforming RFC implementation with optional mixed-case support would produce.

---

## Relation to `toASCII` and `decode`

**`toASCII`** ([to-ascii.md](to-ascii.md), `punycode.js:408-414`) is the domain-level wrapper. It calls `mapDomain` which splits a full domain on label separators, then for each label containing non-ASCII characters (`regexNonASCII.test(string)`, `punycode.js:410`) it prepends `'xn--'` to the result of `encode(string)` (`punycode.js:411`). Labels that are already ASCII pass through unchanged. The `punycode.encode` test suite therefore tests the inner encoding step only, without the ACE prefix or label splitting.

**`punycode.decode`** ([decode.md](decode.md), `punycode.js:196-281`) is the exact inverse of `punycode.encode`. The test suite at `tests/tests.js:290-310` iterates the same `testData.strings` fixtures and asserts `punycode.decode(object.encoded) === object.decoded`. The `punycode.encode` suite at `tests/tests.js:312-321` asserts the reverse direction over the same fixtures, confirming round-trip correctness for every language sample and edge case.

---

## Error Conditions

Although the `encode` test suite does not include explicit error-throwing tests, the function guards against the overflow `RangeError` condition defined at `punycode.js:22-26`:

| Key              | Message                                          | Trigger location        |
|------------------|--------------------------------------------------|-------------------------|
| `'overflow'`     | `'Overflow: input needs wider integers to process'` | `punycode.js:338`, `punycode.js:346` |

All errors are thrown via the shared `error(type)` utility (`punycode.js:41-43`), which throws a `RangeError` with the corresponding message string. The `'not-basic'` and `'invalid-input'` errors are decode-side only.
