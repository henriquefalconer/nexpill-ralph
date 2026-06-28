# `punycode.decode` — Behavior Specification

## Overview

`punycode.decode(input)` converts a Punycode-encoded ASCII-only string into the corresponding Unicode string, implementing the decoding algorithm defined in RFC 3492 §6.2. The implementation lives at `punycode.js:196-281`.

The function explicitly avoids UCS-2 semantics (`punycode.js:197`): it builds an array of integer code points and reconstructs the final string via `String.fromCodePoint(...output)` (`punycode.js:280`), which correctly handles supplementary-plane characters (code points above U+FFFF) that UCS-2 cannot represent as single units.

---

## Bootstring Parameters

The decoder is governed by the Bootstring constants declared at `punycode.js:7-14`:

| Parameter      | Value | Meaning                                     |
|----------------|-------|---------------------------------------------|
| `base`         | 36    | Number of digit values (a–z + 0–9)         |
| `tMin`         | 1     | Minimum threshold                           |
| `tMax`         | 26    | Maximum threshold                           |
| `skew`         | 38    | Bias adaptation skew                        |
| `damp`         | 700   | Bias adaptation damping for first delta     |
| `initialBias`  | 72    | Starting bias value                         |
| `initialN`     | 128   | Starting value of the non-basic code point counter (`0x80`) |
| `delimiter`    | `'-'` | Separator between the basic and extended parts (`\x2D`) |

Additionally, `maxInt = 2147483647` (`punycode.js:4`) is used as an overflow sentinel.

---

## Error Definitions

Three `RangeError` messages are declared at `punycode.js:22-26` and thrown via the `error(type)` helper at `punycode.js:41-43`:

| Key              | Message                                           |
|------------------|---------------------------------------------------|
| `'overflow'`     | `'Overflow: input needs wider integers to process'` |
| `'not-basic'`    | `'Illegal input >= 0x80 (not a basic code point)'` |
| `'invalid-input'`| `'Invalid input'`                                 |

---

## Helper Functions

### `basicToDigit(codePoint)` — `punycode.js:144-155`

Maps an ASCII code point to its base-36 digit value:

- `0x30`–`0x39` (digits `'0'`–`'9'`) → `26`–`35` (`punycode.js:145-146`)
- `0x41`–`0x5A` (uppercase `'A'`–`'Z'`) → `0`–`25` (`punycode.js:148-149`)
- `0x61`–`0x7A` (lowercase `'a'`–`'z'`) → `0`–`25` (`punycode.js:151-152`)
- Any other code point → `base` (36), a sentinel for "not a valid digit" (`punycode.js:154`)

The ranges for `'A'`–`'Z'` and `'a'`–`'z'` both map to `0`–`25`, making digit recognition **case-insensitive** for alphabetic characters. This is the mechanism by which uppercase and lowercase ASCII letters are treated identically during decoding.

### `adapt(delta, numPoints, firstTime)` — `punycode.js:179-187`

Implements the bias adaptation function from RFC 3492 §3.4. After each inserted code point the bias is recalculated to keep subsequent threshold values well-calibrated:

1. If `firstTime`, scale `delta` down by `damp` (700); otherwise halve it (`punycode.js:181`).
2. Add `floor(delta / numPoints)` (`punycode.js:182`).
3. Repeatedly divide by `baseMinusTMin` (35) while `delta > baseMinusTMin * tMax >> 1`, accumulating multiples of `base` into `k` (`punycode.js:183-185`).
4. Return `floor(k + (baseMinusTMin + 1) * delta / (delta + skew))` (`punycode.js:186`).

---

## Decoding Algorithm

### State Initialisation — `punycode.js:198-202`

```
output  = []        // grows to hold code points in logical order
i       = 0         // running insertion index (generalized position)
n       = initialN  // 128 — first non-basic code point to consider
bias    = initialBias  // 72
```

### Phase 1: Basic Code Points — `punycode.js:204-219`

1. Find the last occurrence of `delimiter` (`'-'`) in `input` via `input.lastIndexOf(delimiter)` (`punycode.js:208`). Call that position `basic`.
2. If no delimiter is present, `basic = 0` (`punycode.js:209-211`), meaning there are no literal basic prefix characters.
3. Copy each character at positions `0` through `basic - 1` verbatim to `output` as its numeric code point (`punycode.js:213-219`).
4. Guard: if any of those characters has a code point `>= 0x80`, throw `RangeError('Illegal input >= 0x80 (not a basic code point)')` (`punycode.js:215-216`).

The characters in the basic prefix are plain ASCII and appear unchanged in the decoded output. For example, `'Bach-'` (`tests/tests.js:11`) has basic prefix `'Bach'`; the trailing `'-'` is the delimiter and is consumed but not emitted.

### Phase 2: Variable-Length Integer Decoding — `punycode.js:221-278`

The outer loop starts just after the delimiter (or at index 0 if none) and runs until all input is consumed (`punycode.js:224`).

Each iteration of the outer loop decodes one generalized variable-length integer (a "delta") that encodes the next code point insertion:

#### Inner digit-accumulation loop — `punycode.js:232-261`

Variables: `w = 1`, `k = base`.

For each digit position:

1. If `index >= inputLength`, throw `RangeError('Invalid input')` — input ended prematurely (`punycode.js:234-236`).
2. Read `digit = basicToDigit(input.charCodeAt(index++))` (`punycode.js:238`).
3. If `digit >= base` (i.e., `basicToDigit` returned 36), the character is not a valid base-36 digit; throw `RangeError('Invalid input')` (`punycode.js:240-242`).
4. Overflow guard: if `digit > floor((maxInt - i) / w)`, throw `RangeError('Overflow …')` (`punycode.js:243-245`).
5. Accumulate: `i += digit * w` (`punycode.js:247`).
6. Compute the threshold for this position (`punycode.js:248`):

   ```
   t = k <= bias         ? tMin
     : k >= bias + tMax  ? tMax
     :                     k - bias
   ```

7. If `digit < t`, this digit is the most significant digit of the current variable-length integer; break the inner loop (`punycode.js:250-252`).
8. Overflow guard on `w`: if `w > floor(maxInt / (base - t))`, throw `RangeError('Overflow …')` (`punycode.js:255-257`).
9. Advance: `w *= (base - t)`, `k += base`; continue (`punycode.js:259`).

#### Bias update and code point insertion — `punycode.js:263-276`

After the inner loop:

1. `out = output.length + 1` — the new output length after this insertion (`punycode.js:263`).
2. Update bias: `bias = adapt(i - oldi, out, oldi == 0)` (`punycode.js:264`).
3. Overflow guard: if `floor(i / out) > maxInt - n`, throw `RangeError('Overflow …')` (`punycode.js:268-270`).
4. Advance `n`: `n += floor(i / out)` (`punycode.js:272`).
5. Reduce `i` to a position within the current output: `i %= out` (`punycode.js:273`).
6. Insert code point `n` at position `i` of `output`, then increment `i` (`punycode.js:276`):

   ```js
   output.splice(i++, 0, n);
   ```

This splice-and-increment step is the canonical RFC 3492 insertion step: it places the new code point at the correct logical position within the string being reconstructed.

### Phase 3: Output — `punycode.js:280`

```js
return String.fromCodePoint(...output);
```

The accumulated array of integer code points is spread into `String.fromCodePoint`, producing the final Unicode string. Because `String.fromCodePoint` handles supplementary-plane code points natively, no UCS-2 surrogates are needed.

---

## Test Suite: `describe('punycode.decode', …)` — `tests/tests.js:290-310`

### Fixture-Driven Loop — `tests/tests.js:291-298`

```js
for (const object of testData.strings) {
    it(object.description || object.encoded, function() {
        assert.deepEqual(
            punycode.decode(object.encoded),
            object.decoded
        );
    });
}
```

Every entry in `testData.strings` (`tests/tests.js:7-136`) is iterated. The test name is `object.description` when present, otherwise `object.encoded` (`tests/tests.js:292`). The assertion is `assert.deepEqual` comparing `punycode.decode(object.encoded)` against `object.decoded` (`tests/tests.js:293-295`).

### Fixture Coverage

The `testData.strings` array (`tests/tests.js:7-136`) covers the following categories:

| Description | `encoded` | `decoded` (abbreviated) | Lines |
|---|---|---|---|
| Single basic code point | `'Bach-'` | `'Bach'` | 8–12 |
| Single non-ASCII character | `'tda'` | `'\xFC'` (ü) | 13–17 |
| Multiple non-ASCII characters | `'4can8av2009b'` | `'\xFC\xEB\xE4\xF6♥'` | 18–22 |
| Mix of ASCII and non-ASCII | `'bcher-kva'` | `'bücher'` | 23–27 |
| Long mixed string | `'Willst du … -x9e96lkal'` | German sentence with umlauts | 28–32 |
| **RFC 3492 §7.1 language samples:** | | | 33 |
| Arabic (Egyptian) | `'egbpdaj6bu4bxfgehfvwxn'` | Arabic phrase (17 chars) | 35–38 |
| Chinese (simplified) | `'ihqwcrb4cv8a8dqg056pqjye'` | 9 CJK characters | 39–43 |
| Chinese (traditional) | `'ihqwctvzc91f659drss3x8bo0yb'` | 9 CJK characters | 44–48 |
| Czech | `'Proprostnemluvesky-uyb24dma41a'` | Czech sentence | 49–53 |
| Hebrew | `'4dbcagdahymbxekheh6e0a7fei0b'` | Hebrew phrase | 54–58 |
| Hindi (Devanagari) | `'i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd'` | Devanagari phrase | 59–63 |
| Japanese (kanji and hiragana) | `'n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa'` | Mixed Japanese | 64–68 |
| Korean (Hangul syllables) | `'989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c'` | Korean phrase | 69–73 |
| Russian (Cyrillic) | `'b1abfaaepdrnnbgefbadotcwatmq2g4l'` | Cyrillic phrase | 83–87 |
| Spanish | `'PorqunopuedensimplementehablarenEspaol-fmd56a'` | Spanish phrase | 88–92 |
| Vietnamese | `'TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g'` | Vietnamese phrase | 93–97 |
| Mixed CJK with ASCII digits | `'3B-ww4c5e180e575a65lsy2b'` | `'3年B組金八先生'` | 98–101 |
| ASCII + CJK + hyphens | `'-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n'` | `'安室奈美恵-with-SUPER-MONKEYS'` | 102–105 |
| Long mixed with Japanese | `'Hello-Another-Way--fc4qua05auwb3674vfr0b'` | `'Hello-Another-Way-それぞれの場所'` | 106–109 |
| Digit prefix + CJK | `'2-u9tlzr9756bt3uc0v'` | `'ひとつ屋根の下2'` | 110–113 |
| Mixed alphanumeric + Japanese | `'MajiKoi5-783gue6qz075azm5e'` | `'Majiで Koiする5秒前'` | 114–117 |
| Japanese-only | `'de-jg4avhby1noc0d'` | `'パフィーdeルンバ'` | 118–121 |
| Japanese-only | `'d9juau41awczczp'` | `'そのスピードで'` | 122–125 |
| ASCII edge case (breaks host-name rules) | `'-> $1.00 <--'` | `'-> $1.00 <-'` | 131–135 |

The RFC 3492 note at `tests/tests.js:74-82` explains that the Russian sample deviates from the RFC's published encoding (`b1abfaaepdrnnbgefbaDotcwatmq2g4l` with mixed-case annotation) because JavaScript provides no way to implement mixed-case annotation; punycode.js uses the all-lowercase form `b1abfaaepdrnnbgefbadotcwatmq2g4l` (`tests/tests.js:86`).

---

## Special Tests — `tests/tests.js:299-309`

### "handles uppercase Z" — `tests/tests.js:299-301`

```js
it('handles uppercase Z', function() {
    assert.deepEqual(punycode.decode('ZZZ'), '箥');
});
```

The input `'ZZZ'` contains no delimiter, so `basic = 0` and the entire string is decoded as variable-length integer digits. Because `basicToDigit` maps both `'Z'` (`0x5A`, `punycode.js:148-149`) and `'z'` (`0x7A`, `punycode.js:151-152`) to digit value `25`, uppercase letters are treated identically to their lowercase equivalents. `'ZZZ'` thus decodes the same integer sequence as `'zzz'`, yielding the Unicode character U+7BA5 (`箥`).

This test directly validates the case-insensitivity of `basicToDigit` for the alphabetic range, which follows from the overlapping digit mappings at `punycode.js:148-153`.

### "throws RangeError: Invalid input" — `tests/tests.js:302-309`

```js
it('throws RangeError: Invalid input', function() {
    assert.throws(
        function() {
            punycode.decode('ls8h=');
        },
        RangeError
    );
});
```

The input `'ls8h='` contains the character `'='` (code point `0x3D`). This code point is not in any of the three digit ranges handled by `basicToDigit`, so the function returns `base` (36) (`punycode.js:154`). The inner decoding loop checks `if (digit >= base)` at `punycode.js:240-242` and immediately calls `error('invalid-input')`, which throws `new RangeError('Invalid input')` (`punycode.js:41-43`). The test asserts that the thrown error is an instance of `RangeError` using `assert.throws(..., RangeError)` (`tests/tests.js:303-308`).

---

## Assertion Mechanics

All fixture-loop tests use `assert.deepEqual` (`tests/tests.js:293`) from Node's built-in `assert` module (`tests/tests.js:3`). For primitive string values this is equivalent to strict equality (`===`), but `deepEqual` is used consistently throughout the suite for uniformity.

The invalid-input test uses `assert.throws(fn, RangeError)` (`tests/tests.js:303-308`), which verifies both that the function throws and that the thrown value is an instance of `RangeError`.

---

## Relationship to Other Operations

- **Inverse: `punycode.encode`** ([encode.md](encode.md)) — `punycode.encode(string)` is the exact inverse of `punycode.decode`. The same `testData.strings` fixture set is used for the `describe('punycode.encode', …)` suite at `tests/tests.js:312-321`, which calls `punycode.encode(object.decoded)` and asserts it deepEquals `object.encoded`.
- **Domain wrapper: `punycode.toUnicode`** ([to-unicode.md](to-unicode.md)) — `punycode.toUnicode(domain)` wraps `decode` for full domain names. It splits the input on RFC 3490 separators (`.`, `。`, `．`, `｡`, via `regexSeparators` at `punycode.js:19`), strips the `xn--` ACE prefix (matched by `regexPunycode` at `punycode.js:17`) from each label, calls `decode` on Punycode-encoded labels, then rejoins the labels. The `describe('punycode.toUnicode', …)` suite begins at `tests/tests.js:323`.
- **Two RangeError tests** for `decode` live (somewhat misleadingly) inside the `ucs2.decode` suite — see [ucs2-decode.md](ucs2-decode.md).
