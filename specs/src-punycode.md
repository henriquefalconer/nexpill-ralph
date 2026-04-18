# `punycode.js` — source-file specification

This spec is the source-driven view of the library. Every section cites the exact line range in `punycode.js` (the project's single library file) so a Go porter can verify each claim against the reference implementation. The test-driven specs in this directory (one per Mocha `describe` block) describe *what the tests require*; this spec describes *what the file contains and how its pieces cooperate*.

Read this alongside:

- `specs/test-data-fixtures.md` — the vector tables.
- `specs/punycode-decode.md`, `specs/punycode-encode.md`, `specs/punycode-to-unicode.md`, `specs/punycode-to-ascii.md`, `specs/punycode-ucs2-decode.md`, `specs/punycode-ucs2-encode.md` — per-function behaviour.

## File overview

- **Path:** `punycode.js` (project root — there is no `src/` directory).
- **Length:** 443 lines (`punycode.js:1-443`).
- **Language level:** Strict-mode CommonJS (`punycode.js:1`). Single `module.exports` at `punycode.js:443`.
- **Runtime dependencies:** none. All primitives (`Math.floor`, `String.fromCharCode`, `String.fromCodePoint`, `Array#splice`, `Array#lastIndexOf`, `Array#map`, regex, `String#slice`/`#toLowerCase`) are standard ECMAScript.
- **Published artefacts:** `punycode.js` itself and the generated `punycode.es6.js` produced by `scripts/prepublish.js` (see `specs/src-scripts-prepublish.md`).

## 1. Module preamble (`punycode.js:1`)

`'use strict';` is declared at the top of the file. Under strict mode: undeclared-variable assignment throws, `this` is undefined in top-level function calls, duplicate parameter names are errors, and octal literals are forbidden. A port to a statically-typed language needs no analogue — strict mode is a JS-only artefact — but the port must match the observable behaviour that strict mode enforces here (in practice, no hidden global leaks).

## 2. Module-level constants (`punycode.js:3-31`)

### 2.1 Bootstring parameters (`punycode.js:3-14`)

All values are taken from RFC 3492 §5 and MUST be matched exactly.

| Name | Value | Line | Role |
|---|---|---|---|
| `maxInt` | 2 147 483 647 (`2^31 - 1`) | `punycode.js:4` | 32-bit signed overflow bound used by every overflow guard |
| `base` | 36 | `punycode.js:7` | Radix of the generalised variable-length integer |
| `tMin` | 1 | `punycode.js:8` | Lower threshold bound |
| `tMax` | 26 | `punycode.js:9` | Upper threshold bound |
| `skew` | 38 | `punycode.js:10` | Denominator term in `adapt()` |
| `damp` | 700 | `punycode.js:11` | First-iteration delta divisor in `adapt()` |
| `initialBias` | 72 | `punycode.js:12` | Starting bias for encode/decode |
| `initialN` | 128 (`0x80`) | `punycode.js:13` | First non-basic code point; starting `n` |
| `delimiter` | `-` (U+002D) | `punycode.js:14` | Separator between the basic prefix and the extended section of a Punycode label |

### 2.2 Regular expressions (`punycode.js:16-19`)

| Name | Pattern | Line | Behaviour |
|---|---|---|---|
| `regexPunycode` | `/^xn--/` | `punycode.js:17` | Case-**sensitive**, anchored match for the ACE prefix. A label starting with `XN--` would NOT match — see §11. |
| `regexNonASCII` | `/[^\0-\x7F]/` | `punycode.js:18` | Any single code unit outside U+0000..U+007F. The inline comment "U+007F DEL is excluded too" is load-bearing: DEL counts as ASCII and does not trigger encoding. |
| `regexSeparators` | `/[\x2E\u3002\uFF0E\uFF61]/g` | `punycode.js:19` | Global match for the four IDNA2003 label separators: U+002E (FULL STOP), U+3002 (IDEOGRAPHIC FULL STOP), U+FF0E (FULLWIDTH FULL STOP), U+FF61 (HALFWIDTH IDEOGRAPHIC FULL STOP). Used to normalise to U+002E before splitting. |

### 2.3 Error messages (`punycode.js:21-26`)

The `errors` object maps three type keys to human-readable messages:

- `'overflow'` → `'Overflow: input needs wider integers to process'` (`punycode.js:23`). Thrown by the five overflow guards — see §10.
- `'not-basic'` → `'Illegal input >= 0x80 (not a basic code point)'` (`punycode.js:24`). Thrown by `decode` when a non-ASCII character appears in the basic prefix.
- `'invalid-input'` → `'Invalid input'` (`punycode.js:25`). Thrown by `decode` for malformed digit input.

### 2.4 Convenience shortcuts (`punycode.js:28-31`)

- `baseMinusTMin = base - tMin` = 35 (`punycode.js:29`). Appears in `adapt()` and in threshold arithmetic.
- `floor = Math.floor` (`punycode.js:30`). Alias used wherever integer division is required.
- `stringFromCharCode = String.fromCharCode` (`punycode.js:31`). Alias used to emit ASCII code units from the encoder.

## 3. Private helpers (`punycode.js:35-86`)

### 3.1 `error(type)` (`punycode.js:35-43`)

Throws `RangeError(errors[type])`. Never returns. Called from `decode` at `punycode.js:216, 235, 241, 244, 256, 269` and from `encode` at `punycode.js:338, 346`.

### 3.2 `map(array, callback)` (`punycode.js:45-60`)

Private re-implementation of `Array#map`. Iterates in reverse (via `while (length--)` at `punycode.js:56`) and writes results into `result[length]` at `punycode.js:57`, preserving input order in the output. The reverse walk is a JS micro-optimisation and is observable only in that the callback is invoked in reverse index order. A port may use any ordered map; no caller depends on the invocation order. Called once, at `punycode.js:84`.

### 3.3 `mapDomain(domain, callback)` (`punycode.js:62-86`)

The only place that understands "domain" and "email" shape. Algorithm:

1. **Split on `@`** (`punycode.js:73`). If the input contains one or more `@`, take `parts[0] + '@'` as a preserved prefix (`punycode.js:78`) and use `parts[1]` as the domain (`punycode.js:79`). Everything after a hypothetical second `@` is silently discarded — a multi-`@` input like `a@b@c` becomes `"a@" + callback("b")`. Not exercised by any test; see §11.
2. **Normalise separators** (`punycode.js:82`). Apply `regexSeparators` with a global replace, mapping U+3002 / U+FF0E / U+FF61 to U+002E.
3. **Split on `.`** (`punycode.js:83`) to obtain labels.
4. **Apply the callback** to each label via `map()` and join with `.` (`punycode.js:84`).
5. **Return** the preserved prefix concatenated with the encoded domain (`punycode.js:85`).

## 4. UCS-2 layer (`punycode.js:88-133`)

### 4.1 `ucs2decode(string)` (`punycode.js:88-123`)

Converts a JavaScript string (UCS-2 / UTF-16 code-unit sequence) into an array of Unicode scalar code points, combining surrogate pairs. Never throws.

State (`punycode.js:102-104`): an output array, a running counter, and a cached length.

Per-iteration (`punycode.js:105-121`):

- Read one code unit and advance the counter (`punycode.js:106`).
- If the unit is a **high surrogate** (`0xD800..0xDBFF`) AND there is a next unit (`punycode.js:107`):
  - Read the next unit and advance the counter (`punycode.js:109`).
  - Check it is a **low surrogate** by masking with `0xFC00` and comparing to `0xDC00` (`punycode.js:110`). The mask isolates the top 6 bits; all low surrogates have the binary prefix `110111`.
  - If yes, combine into a scalar via `((high & 0x3FF) << 10) + (low & 0x3FF) + 0x10000` (`punycode.js:111`). This is the canonical UTF-16 decode formula; the `+ 0x10000` shifts into the supplementary-plane range.
  - If no, push the high surrogate alone (`punycode.js:115`) and rewind by one code unit (`punycode.js:116`) so the non-low-surrogate unit is re-examined on the next iteration — this permits lone high surrogates followed by lone low surrogates to round-trip.
- Otherwise, push the unit as-is (`punycode.js:119`).

Return the output array (`punycode.js:122`).

Exposed at `punycode.ucs2.decode` (`punycode.js:434`). Called internally by `encode` (`punycode.js:294`).

### 4.2 `ucs2encode(codePoints)` (`punycode.js:125-133`)

Single-line arrow: `codePoints => String.fromCodePoint(...codePoints)` (`punycode.js:133`). The host primitive `String.fromCodePoint` accepts integers in `0..0x10FFFF` and emits one code unit for scalars ≤ `0xFFFF` and a surrogate pair for larger scalars. Values in the surrogate range `0xD800..0xDFFF` are emitted as single code units, allowing lone-surrogate inputs produced by `ucs2decode` to round-trip byte-for-byte.

Exposed at `punycode.ucs2.encode` (`punycode.js:435`). Not called elsewhere in the module.

## 5. Bootstring primitives (`punycode.js:135-187`)

### 5.1 `basicToDigit(codePoint)` (`punycode.js:135-155`)

Maps a single ASCII code point to a base-36 digit value, or returns `base` (36) as a sentinel for "not a digit".

| Input range | Branch | Line | Digit value |
|---|---|---|---|
| `0x30..0x39` (`'0'..'9'`) | digits | `punycode.js:145-146` | `26..35` |
| `0x41..0x5A` (`'A'..'Z'`) | uppercase letters | `punycode.js:148-149` | `0..25` |
| `0x61..0x7A` (`'a'..'z'`) | lowercase letters | `punycode.js:151-152` | `0..25` |
| anything else | sentinel | `punycode.js:154` | `base` = 36 |

Case-insensitive: both `'A'` and `'a'` map to digit 0. Called at `punycode.js:238`.

### 5.2 `digitToBasic(digit, flag)` (`punycode.js:157-172`)

Inverse of `basicToDigit`. One-line branchless formula (`punycode.js:171`):

```
digit + 22 + 75 * (digit < 26) - ((flag != 0) << 5)
```

Derivation:

- For letters (`digit < 26`): `digit + 22 + 75 = digit + 97`. ASCII `'a'` is 97, so the formula yields `'a'..'z'`. When `flag` is non-zero, subtract `1 << 5 = 32` to reach `'A'..'Z'`.
- For digits (`digit >= 26`): `digit + 22`. Digit 26 → 48 = `'0'`; digit 35 → 57 = `'9'`. The `flag` subtraction still fires but has no case meaning for digits.

The `flag` is the mixed-case annotation of RFC 3492. This module **never sets it** — both call sites (`punycode.js:359, 364`) pass `0`. Consequence: `encode` always emits lowercase, and a port need not implement case flags. The Russian vector at `tests/tests.js:74-82` documents this with a lowercase expectation (`b1abfaaepdrnnbgefbadotcwatmq2g4l`).

### 5.3 `adapt(delta, numPoints, firstTime)` (`punycode.js:174-187`)

Bias adaptation per RFC 3492 §3.4.

1. Initialise `k = 0` (`punycode.js:180`).
2. `delta = firstTime ? floor(delta / damp) : delta >> 1` (`punycode.js:181`). `>> 1` is integer division by 2.
3. `delta += floor(delta / numPoints)` (`punycode.js:182`).
4. While `delta > (baseMinusTMin * tMax) >> 1` (i.e. > 455): `delta = floor(delta / baseMinusTMin); k += base` (`punycode.js:183-184`).
5. Return `floor(k + (baseMinusTMin + 1) * delta / (delta + skew))` (`punycode.js:186`), i.e. `floor(k + 36·delta / (delta + 38))`.

Called at `punycode.js:264` (decode) and `punycode.js:365` (encode).

## 6. `decode(input)` — label-level Punycode decoder (`punycode.js:189-281`)

Decodes a Punycode (ASCII-only) label body. Does NOT strip `xn--`; that is the caller's job (see `toUnicode`, §8.1).

### 6.1 State (`punycode.js:196-202`)

- `output` — array of decoded code points.
- `inputLength = input.length` — cached.
- `i = 0` — insertion cursor into `output`.
- `n = initialN` = 128 — current code point being emitted.
- `bias = initialBias` = 72 — current threshold bias.

### 6.2 Basic prefix copy (`punycode.js:208-219`)

Let `basic = input.lastIndexOf(delimiter)` (`punycode.js:208`). If `basic < 0`, set `basic = 0` (`punycode.js:209-211`). For each index `j` in `[0, basic)` (`punycode.js:213-219`):

- If `input.charCodeAt(j) >= 0x80`, throw `'not-basic'` (`punycode.js:215-216`).
- Otherwise, push the ASCII code unit onto `output` (`punycode.js:218`).

Note the use of `charCodeAt` (UCS-2 code units), not `codePointAt`: a basic-prefix character is by definition in `0x00..0x7F`, so the distinction is academic. The guard at `punycode.js:215` rejects any high-bit input as non-basic.

### 6.3 Main loop (`punycode.js:224-278`)

Start `index = basic > 0 ? basic + 1 : 0` (`punycode.js:224`). While `index < inputLength`:

**Read one generalised variable-length integer** (`punycode.js:231-261`). Save `oldi = i` (`punycode.js:231`). With `w = 1`, `k = base`, loop forever:

- If `index >= inputLength`, throw `'invalid-input'` (`punycode.js:234-236`).
- `digit = basicToDigit(input.charCodeAt(index++))` (`punycode.js:238`).
- If `digit >= base`, throw `'invalid-input'` (`punycode.js:240-242`).
- **Overflow guard A** (`punycode.js:243-245`): if `digit > floor((maxInt - i) / w)`, throw `'overflow'`. Protects the next line from overflow in `digit * w + i`.
- `i += digit * w` (`punycode.js:247`).
- `t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias)` (`punycode.js:248`).
- If `digit < t`, break (`punycode.js:250-251`).
- `baseMinusT = base - t` (`punycode.js:254`).
- **Overflow guard B** (`punycode.js:255-257`): if `w > floor(maxInt / baseMinusT)`, throw `'overflow'`. Protects `w *= baseMinusT`.
- `w *= baseMinusT` (`punycode.js:259`); `k += base` (loop update at `punycode.js:232`).

**Bias update and code-point advance** (`punycode.js:263-273`):

- `out = output.length + 1` (`punycode.js:263`).
- `bias = adapt(i - oldi, out, oldi == 0)` (`punycode.js:264`). Third argument distinguishes the first adaptation of a decode run.
- **Overflow guard C** (`punycode.js:268-270`): if `floor(i / out) > maxInt - n`, throw `'overflow'`. Protects `n += floor(i / out)`.
- `n += floor(i / out)` (`punycode.js:272`).
- `i %= out` (`punycode.js:273`).

**Insert** (`punycode.js:276`): `output.splice(i++, 0, n)` inserts `n` at position `i` and post-increments `i`.

### 6.4 Finalise (`punycode.js:280`)

Return `String.fromCodePoint(...output)`. Supplementary-plane scalars are automatically re-encoded as surrogate pairs by the host primitive.

### 6.5 Port notes

- All arithmetic is 32-bit-signed-friendly. A port MUST use `int32` (or widen + re-check) and MUST preserve all three overflow guards; the tests at `tests/tests.js:263-270` exercise guard paths.
- The three error strings (`'not-basic'`, `'invalid-input'`, `'overflow'`) should map to distinct error values in the target language so callers can distinguish them — the tests assert on the exact message via `RegExp` matching.

## 7. `encode(input)` — label-level Punycode encoder (`punycode.js:283-376`)

Encodes a Unicode label into its Punycode body. Does NOT prepend `xn--`; that is the caller's job (see `toASCII`, §8.2).

### 7.1 State (`punycode.js:290-302`)

- `output` — array of ASCII characters to be joined at the end.
- `input = ucs2decode(input)` (`punycode.js:294`) — surrogate-folded code-point array.
- `inputLength = input.length` (`punycode.js:297`).
- `n = initialN` = 128 (`punycode.js:300`).
- `delta = 0` (`punycode.js:301`).
- `bias = initialBias` = 72 (`punycode.js:302`).

### 7.2 Basic-code-point copy (`punycode.js:304-312`)

Iterate the code-point array; for each `currentValue < 0x80`, push `stringFromCharCode(currentValue)` onto `output` (`punycode.js:305-308`). Record `basicLength = output.length` (`punycode.js:311`) and start `handledCPCount = basicLength` (`punycode.js:312`).

### 7.3 Delimiter emission (`punycode.js:318-320`)

If `basicLength > 0`, append `delimiter` (`-`) to output (`punycode.js:319`). Labels with no basic code points therefore have no leading `-`.

### 7.4 Main loop (`punycode.js:323-374`)

While `handledCPCount < inputLength`:

**Find the next `m`** (`punycode.js:327-332`). Start `m = maxInt`; scan all code points; set `m` to the smallest value that is `>= n`.

**Advance `delta` and `n`** (`punycode.js:336-342`):

- `handledCPCountPlusOne = handledCPCount + 1` (`punycode.js:336`).
- **Overflow guard D** (`punycode.js:337-339`): if `m - n > floor((maxInt - delta) / handledCPCountPlusOne)`, throw `'overflow'`. Protects `delta += (m - n) * handledCPCountPlusOne`.
- `delta += (m - n) * handledCPCountPlusOne` (`punycode.js:341`).
- `n = m` (`punycode.js:342`).

**Emit characters equal to `n`** (`punycode.js:344-369`). For each code point `currentValue` in input order:

- If `currentValue < n`: `++delta`. **Overflow guard E** (`punycode.js:345-347`): if `delta > maxInt` after the increment, throw `'overflow'`.
- If `currentValue === n`: emit as a generalised variable-length integer.
  - `q = delta` (`punycode.js:350`).
  - With `k = base`, loop forever (`punycode.js:351`):
    - `t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias)` (`punycode.js:352`).
    - If `q < t`, break (`punycode.js:353-354`).
    - `qMinusT = q - t`; `baseMinusT = base - t` (`punycode.js:356-357`).
    - Push `stringFromCharCode(digitToBasic(t + qMinusT % baseMinusT, 0))` (`punycode.js:358-360`).
    - `q = floor(qMinusT / baseMinusT)` (`punycode.js:361`); `k += base` (`punycode.js:351`).
  - Push `stringFromCharCode(digitToBasic(q, 0))` (`punycode.js:364`).
  - `bias = adapt(delta, handledCPCountPlusOne, handledCPCount === basicLength)` (`punycode.js:365`). Third argument is `true` only for the first emission after the basic prefix.
  - `delta = 0` (`punycode.js:366`); `++handledCPCount` (`punycode.js:367`).

**Advance to next round** (`punycode.js:371-372`): `++delta; ++n`.

### 7.5 Finalise (`punycode.js:375`)

Return `output.join('')`.

### 7.6 Port notes

- All emitted digits are lowercase (flag = 0 at both `digitToBasic` call sites).
- Neither overflow guard D nor E is triggered by any fixture in the `testData.strings` bucket (see `specs/test-data-fixtures.md`); they are correctness insurance, not test-covered paths.

## 8. Domain/email wrappers (`punycode.js:378-414`)

Both wrappers delegate label handling to `mapDomain` (§3.3) and differ only in the callback.

### 8.1 `toUnicode(input)` (`punycode.js:378-395`)

Callback (`punycode.js:391-393`): if `regexPunycode.test(string)` then `decode(string.slice(4).toLowerCase())` else return the label unchanged.

- `string.slice(4)` strips the four-character `xn--` prefix.
- `.toLowerCase()` normalises any case variation in the encoded body before decoding (the fixture `xn--ZCKzah` at `tests/tests.js:140-149` exercises this).
- Non-`xn--` labels pass through verbatim — idempotent over already-Unicode input (`tests/tests.js:332-343`).

Exposed as `punycode.toUnicode` (`punycode.js:440`).

### 8.2 `toASCII(input)` (`punycode.js:397-414`)

Callback (`punycode.js:410-412`): if `regexNonASCII.test(string)` then `'xn--' + encode(string)` else return the label unchanged.

- Pure-ASCII labels (including an already-encoded `xn--…`) pass through unchanged, so `toASCII` is idempotent over *ASCII* input (`tests/tests.js:332-343`).
- Unlike `toUnicode`, there is no re-encoding guard: a label that is *both* `xn--`-prefixed *and* contains non-ASCII would be re-encoded. Not exercised by tests; see §11.
- IDNA2003 separator normalisation (U+3002 / U+FF0E / U+FF61 → U+002E) is provided by `mapDomain` before labels reach the callback (`tests/tests.js:221-242`).

Exposed as `punycode.toASCII` (`punycode.js:439`).

## 9. Public API object and module export (`punycode.js:418-443`)

The `punycode` object aggregates the exports:

| Key | Value | Line |
|---|---|---|
| `version` | `'2.3.1'` | `punycode.js:425` |
| `ucs2.decode` | `ucs2decode` | `punycode.js:434` |
| `ucs2.encode` | `ucs2encode` | `punycode.js:435` |
| `decode` | `decode` | `punycode.js:437` |
| `encode` | `encode` | `punycode.js:438` |
| `toASCII` | `toASCII` | `punycode.js:439` |
| `toUnicode` | `toUnicode` | `punycode.js:440` |

`module.exports = punycode;` at `punycode.js:443` is the single commonjs export statement. It is also the exact target of the ES6-build regex in `scripts/prepublish.js:6` — any rename or reflow of this line breaks the build (see `specs/src-scripts-prepublish.md`).

`version` is not under test and is explicitly out of port scope (`ralph/todo.md` §"Out of scope").

## 10. Overflow-guard inventory

Five distinct guards throw `RangeError('Overflow: input needs wider integers to process')`.

| # | Site | Condition | Operation protected | Test coverage |
|---|---|---|---|---|
| A | `punycode.js:243-245` | `digit > floor((maxInt - i) / w)` | `i += digit * w` (decode) | `tests/tests.js:263-270` via `punycode.decode('\x81')` — though that input triggers guard A through the `'\x81'` digit reading pathway (see §11 for exact mapping) |
| B | `punycode.js:255-257` | `w > floor(maxInt / baseMinusT)` | `w *= baseMinusT` (decode) | Not directly asserted; insurance only |
| C | `punycode.js:268-270` | `floor(i / out) > maxInt - n` | `n += floor(i / out)` (decode) | Not directly asserted; insurance only |
| D | `punycode.js:337-339` | `m - n > floor((maxInt - delta) / handledCPCountPlusOne)` | `delta += (m - n) * handledCPCountPlusOne` (encode) | Not exercised |
| E | `punycode.js:345-347` | `++delta > maxInt` | `delta += 1` (encode) | Not exercised |

A port MUST preserve all five guards verbatim even without direct test coverage. Widening to `int64` and removing the guards would change the observable error behaviour.

## 11. Uncovered and under-covered code paths

Flagged by the test-driven specs as intentional scope gaps.

1. **`regexPunycode` is case-sensitive** (`punycode.js:17, 391`). Input with uppercase `XN--` passes through `toUnicode` unchanged. No fixture covers this.
2. **`regexNonASCII` treats U+007F DEL as ASCII** (`punycode.js:18, 410`). The fixture `foo\x7F.example` at `tests/tests.js:216-219` covers the "DEL in an otherwise-ASCII label" case; DEL mixed with non-ASCII is untested.
3. **Multi-`@` emails** (`punycode.js:73, 78-79`). `split('@')` with input `a@b@c` yields `['a', 'b', 'c']`; only `parts[0]` and `parts[1]` are used, dropping `'c'`. Untested.
4. **`toASCII` on a label that is both `xn--`-prefixed and non-ASCII** (`punycode.js:410-411`). Would be re-encoded into nested `xn--xn--…`. Untested.
5. **Encode-side overflow guards** (`punycode.js:338, 346`). No fixture triggers them.

A port may match these behaviours as-is (the pragmatic default) or document deliberate divergence.

## 12. Within-file call graph

```
decode          → error, basicToDigit, adapt
encode          → ucs2decode, stringFromCharCode, digitToBasic, adapt, error
toUnicode       → mapDomain, regexPunycode, decode
toASCII         → mapDomain, regexNonASCII, encode
mapDomain       → regexSeparators, map
ucs2encode      → (host) String.fromCodePoint
ucs2decode      → (host) String#charCodeAt
basicToDigit    → (pure)
digitToBasic    → (pure)
adapt           → floor
error           → (host) RangeError, errors
map             → (pure)
```

Every public entry point is reachable; there are no dead helpers.

## 13. Port guidance (Go)

See `ralph/todo.md` for the commit-scoped ordering. For this file specifically, a porter should:

1. Land the Bootstring constants (`punycode.js:3-14`) verbatim as typed Go constants.
2. Translate `basicToDigit`, `digitToBasic`, `adapt`, and `error` first — they have no public-surface tests of their own but every other function depends on them.
3. Translate `ucs2decode`/`ucs2encode` against `[]rune` or a UTF-16 abstraction that preserves lone surrogates (`tests/tests.js:145-174`).
4. Translate `decode`, `encode`, then the wrappers `toUnicode`/`toASCII` and `mapDomain`.
5. Leave `version` out of the public surface unless consumers require it.
6. Leave `scripts/prepublish.js` entirely out of scope — it is a JS-only build step with no Go analogue.
