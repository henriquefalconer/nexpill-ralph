# Spec: `punycode.decode` — test coverage

## 1. Subject

`punycode.decode(input: string): string`

Converts a Punycode (ASCII-only) string to a Unicode string. This is the raw RFC 3492 decoder — it does **not** strip any `xn--` prefix. Bound to the exported object at `punycode.js:437`.

---

## 2. Contract

Drawn from JSDoc at `punycode.js:189-195` and from observed test behaviour (`tests/tests.js:290-310`):

- The input must be a string of ASCII-only characters (all code points `< 0x80`).
- Characters that fall after the last `-` delimiter in the input are basic code points and are copied verbatim to the output.
- Characters before that delimiter are decoded as a generalized variable-length base-36 integer stream that encodes the positions and scalar values of the non-basic code points in the output.
- When the input contains no `-` delimiter, the entire string is treated as the encoded delta stream with an empty basic prefix (`basic = 0`).
- Digit characters are case-insensitive over the base-36 alphabet (`a-z`, `A-Z`, `0-9`).
- Any character outside the base-36 alphabet (`basicToDigit` returning `base`) causes an `invalid-input` `RangeError` (`punycode.js:240-242`).
- Any basic code point in the prefix that is `>= 0x80` causes a `not-basic` `RangeError` (`punycode.js:215-217`).
- Integer accumulation overflow causes an `overflow` `RangeError` (`punycode.js:243-245`, `punycode.js:255-257`, `punycode.js:268-270`).

---

## 3. Test cases

### 3.1 Parameterized loop — `testData.strings` vectors

The loop at `tests/tests.js:291-298` iterates every entry in `testData.strings` (`tests/tests.js:7-136`), calls `punycode.decode(object.encoded)`, and asserts deep equality with `object.decoded`.

There are 23 vectors, enumerated below with their source line ranges.

---

#### Vector 1 — `tests/tests.js:8-12`

| Field | Value |
|---|---|
| `description` | `'a single basic code point'` |
| `encoded` | `'Bach-'` |
| `decoded` | `'Bach'` |

The trailing `-` is the delimiter (`punycode.js:14`). The four characters before it are basic code points copied verbatim. The encoded delta stream is empty, so no non-basic code point is inserted.

---

#### Vector 2 — `tests/tests.js:13-17`

| Field | Value |
|---|---|
| `description` | `'a single non-ASCII character'` |
| `encoded` | `'tda'` |
| `decoded` | `'\xFC'` (U+00FC, LATIN SMALL LETTER U WITH DIAERESIS) |

No delimiter is present; `basic = 0` (`punycode.js:209-211`). The entire string `tda` is the delta stream, decoding to a single inserted code point U+00FC.

---

#### Vector 3 — `tests/tests.js:18-22`

| Field | Value |
|---|---|
| `description` | `'multiple non-ASCII characters'` |
| `encoded` | `'4can8av2009b'` |
| `decoded` | `'\xFC\xEB\xE4\xF6\u2665'` |

No delimiter present; `basic = 0`. Five non-ASCII code points are decoded from the delta stream.

---

#### Vector 4 — `tests/tests.js:23-27`

| Field | Value |
|---|---|
| `description` | `'mix of ASCII and non-ASCII characters'` |
| `encoded` | `'bcher-kva'` |
| `decoded` | `'b\xFCcher'` |

The last `-` in the input is the delimiter. The prefix `bcher` is copied as basic code points, then one non-ASCII code point (U+00FC) is inserted at position 1 by the delta stream `kva`.

---

#### Vector 5 — `tests/tests.js:28-32`

| Field | Value |
|---|---|
| `description` | `'long string with both ASCII and non-ASCII characters'` |
| `encoded` | `'Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal'` |
| `decoded` | `'Willst du die Bl\xFCthe des fr\xFChen, die Fr\xFCchte des sp\xE4teren Jahres'` |

Long German sentence. The delimiter separates the ASCII skeleton from the delta stream that inserts U+00FC and U+00E4 at their respective positions.

---

#### Vector 6 — `tests/tests.js:34-38` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Arabic (Egyptian)'` |
| `encoded` | `'egbpdaj6bu4bxfgehfvwxn'` |
| `decoded` | `'\u0644\u064A\u0647\u0645\u0627\u0628\u062A\u0643\u0644\u0645\u0648\u0634\u0639\u0631\u0628\u064A\u061F'` |

No delimiter present; all output code points are non-basic.

---

#### Vector 7 — `tests/tests.js:39-43` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Chinese (simplified)'` |
| `encoded` | `'ihqwcrb4cv8a8dqg056pqjye'` |
| `decoded` | `'\u4ED6\u4EEC\u4E3A\u4EC0\u4E48\u4E0D\u8BF4\u4E2D\u6587'` |

---

#### Vector 8 — `tests/tests.js:44-48` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Chinese (traditional)'` |
| `encoded` | `'ihqwctvzc91f659drss3x8bo0yb'` |
| `decoded` | `'\u4ED6\u5011\u7232\u4EC0\u9EBD\u4E0D\u8AAA\u4E2D\u6587'` |

---

#### Vector 9 — `tests/tests.js:49-53` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Czech'` |
| `encoded` | `'Proprostnemluvesky-uyb24dma41a'` |
| `decoded` | `'Pro\u010Dprost\u011Bnemluv\xED\u010Desky'` |

Mixed ASCII and non-ASCII. The delimiter is the `-` after `Proprostnemluvesky`.

---

#### Vector 10 — `tests/tests.js:54-58` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Hebrew'` |
| `encoded` | `'4dbcagdahymbxekheh6e0a7fei0b'` |
| `decoded` | `'\u05DC\u05DE\u05D4\u05D4\u05DD\u05E4\u05E9\u05D5\u05D8\u05DC\u05D0\u05DE\u05D3\u05D1\u05E8\u05D9\u05DD\u05E2\u05D1\u05E8\u05D9\u05EA'` |

---

#### Vector 11 — `tests/tests.js:59-63` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Hindi (Devanagari)'` |
| `encoded` | `'i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd'` |
| `decoded` | `'\u092F\u0939\u0932\u094B\u0917\u0939\u093F\u0928\u094D\u0926\u0940\u0915\u094D\u092F\u094B\u0902\u0928\u0939\u0940\u0902\u092C\u094B\u0932\u0938\u0915\u0924\u0947\u0939\u0948\u0902'` |

---

#### Vector 12 — `tests/tests.js:64-68` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Japanese (kanji and hiragana)'` |
| `encoded` | `'n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa'` |
| `decoded` | `'\u306A\u305C\u307F\u3093\u306A\u65E5\u672C\u8A9E\u3092\u8A71\u3057\u3066\u304F\u308C\u306A\u3044\u306E\u304B'` |

---

#### Vector 13 — `tests/tests.js:69-73` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Korean (Hangul syllables)'` |
| `encoded` | `'989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c'` |
| `decoded` | `'\uC138\uACC4\uC758\uBAA8\uB4E0\uC0AC\uB78C\uB4E4\uC774\uD55C\uAD6D\uC5B4\uB97C\uC774\uD574\uD55C\uB2E4\uBA74\uC5BC\uB9C8\uB098\uC88B\uC744\uAE4C'` |

---

#### Vector 14 — `tests/tests.js:83-87` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Russian (Cyrillic)'` |
| `encoded` | `'b1abfaaepdrnnbgefbadotcwatmq2g4l'` |
| `decoded` | `'\u043F\u043E\u0447\u0435\u043C\u0443\u0436\u0435\u043E\u043D\u0438\u043D\u0435\u0433\u043E\u0432\u043E\u0440\u044F\u0442\u043F\u043E\u0440\u0443\u0441\u0441\u043A\u0438'` |

**Note on mixed-case annotation** (`tests/tests.js:74-82`): The RFC 3492 §7.1 sample string uses mixed-case annotation (an optional RFC feature) and encodes to `b1abfaaepdrnnbgefbaDotcwatmq2g4l`. Because JavaScript provides no mechanism for mixed-case annotation, Punycode.js omits it; the test vector therefore uses the all-lowercase-encoded form `b1abfaaepdrnnbgefbadotcwatmq2g4l`. See [issue #3](https://github.com/mathiasbynens/punycode.js/issues/3).

---

#### Vector 15 — `tests/tests.js:88-92` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Spanish'` |
| `encoded` | `'PorqunopuedensimplementehablarenEspaol-fmd56a'` |
| `decoded` | `'Porqu\xE9nopuedensimplementehablarenEspa\xF1ol'` |

---

#### Vector 16 — `tests/tests.js:93-97` (RFC 3492 §7.1)

| Field | Value |
|---|---|
| `description` | `'Vietnamese'` |
| `encoded` | `'TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g'` |
| `decoded` | `'T\u1EA1isaoh\u1ECDkh\xF4ngth\u1EC3ch\u1EC9n\xF3iti\u1EBFngVi\u1EC7t'` |

---

#### Vector 17 — `tests/tests.js:98-101` (unnamed)

| Field | Value |
|---|---|
| `description` | *(none — test label falls back to `object.encoded`)* |
| `encoded` | `'3B-ww4c5e180e575a65lsy2b'` |
| `decoded` | `'3\u5E74B\u7D44\u91D1\u516B\u5148\u751F'` |

Mixed ASCII digits, ASCII letters, and CJK unified ideographs.

---

#### Vector 18 — `tests/tests.js:102-105` (unnamed)

| Field | Value |
|---|---|
| `description` | *(none)* |
| `encoded` | `'-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n'` |
| `decoded` | `'\u5B89\u5BA4\u5948\u7F8E\u6075-with-SUPER-MONKEYS'` |

The leading `-` in the encoded form is itself a delimiter character; `lastIndexOf('-')` finds the rightmost one, which separates the basic skeleton `-with-SUPER-MONKEYS` from the delta stream `pc58ag80a8qai00g7n9n`.

---

#### Vector 19 — `tests/tests.js:106-109` (unnamed)

| Field | Value |
|---|---|
| `description` | *(none)* |
| `encoded` | `'Hello-Another-Way--fc4qua05auwb3674vfr0b'` |
| `decoded` | `'Hello-Another-Way-\u305D\u308C\u305E\u308C\u306E\u5834\u6240'` |

The double `--` before the delta stream means the basic prefix ends with a literal `-`.

---

#### Vector 20 — `tests/tests.js:110-113` (unnamed)

| Field | Value |
|---|---|
| `description` | *(none)* |
| `encoded` | `'2-u9tlzr9756bt3uc0v'` |
| `decoded` | `'\u3072\u3068\u3064\u5C4B\u6839\u306E\u4E0B2'` |

---

#### Vector 21 — `tests/tests.js:114-117` (unnamed)

| Field | Value |
|---|---|
| `description` | *(none)* |
| `encoded` | `'MajiKoi5-783gue6qz075azm5e'` |
| `decoded` | `'Maji\u3067Koi\u3059\u308B5\u79D2\u524D'` |

---

#### Vector 22 — `tests/tests.js:118-121` (unnamed)

| Field | Value |
|---|---|
| `description` | *(none)* |
| `encoded` | `'de-jg4avhby1noc0d'` |
| `decoded` | `'\u30D1\u30D5\u30A3\u30FCde\u30EB\u30F3\u30D0'` |

---

#### Vector 23 — `tests/tests.js:122-125` (unnamed)

| Field | Value |
|---|---|
| `description` | *(none)* |
| `encoded` | `'d9juau41awczczp'` |
| `decoded` | `'\u305D\u306E\u30B9\u30D4\u30FC\u30C9\u3067'` |

---

#### Vector 24 — `tests/tests.js:131-135`

| Field | Value |
|---|---|
| `description` | `'ASCII string that breaks the existing rules for host-name labels'` |
| `encoded` | `'-> $1.00 <--'` |
| `decoded` | `'-> $1.00 <-'` |

The comment at `tests/tests.js:126-130` notes this is a pure-ASCII string that is not a realistic IDNA example. The trailing `-` in the encoded form is the Punycode delimiter (`punycode.js:14`). Everything before it — `-> $1.00 <-` — is the basic prefix and is copied verbatim. The delta stream is empty so no non-basic code point is inserted, producing a decoded output that is simply the basic prefix.

---

### 3.2 Static test — `tests/tests.js:299-301`

```
punycode.decode('ZZZ') === '\u7BA5'
```

The input `'ZZZ'` contains no `-` delimiter, so `lastIndexOf(delimiter)` returns `-1` and `basic` is set to `0` (`punycode.js:208-211`). The delta stream is the entire string. Each `Z` is processed by `basicToDigit` (`punycode.js:144-155`): code point `0x5A` satisfies the branch at `punycode.js:148-149` (`0x41 <= 0x5A < 0x5B`), returning `0x5A - 0x41 = 25`. This test demonstrates that the decoder is **case-insensitive** over the base-36 alphabet: uppercase letters produce the same digit value as their lowercase counterparts.

---

### 3.3 Error test — `tests/tests.js:302-309`

```
punycode.decode('ls8h=')  // throws RangeError
```

The input encodes the emoji domain label `ls8h` followed by `=` (U+003D, code point `0x3D`). When the main loop (`punycode.js:224`) reaches `=`, it calls `basicToDigit(0x3D)` (`punycode.js:238`). Code point `0x3D` satisfies none of the three range checks in `basicToDigit` (`punycode.js:145`, `punycode.js:148`, `punycode.js:151`), so the function returns `base` (36) (`punycode.js:154`). The check at `punycode.js:240-242` detects `digit >= base` and calls `error('invalid-input')` (`punycode.js:25`, `punycode.js:41-43`), which throws `RangeError: Invalid input`.

---

## 4. Implementation citations

| Construct | Location |
|---|---|
| `decode` function body | `punycode.js:196-281` |
| Bootstring constants (`base=36`, `tMin=1`, `tMax=26`, `skew=38`, `damp=700`, `initialBias=72`, `initialN=128`, `delimiter='-'`) | `punycode.js:7-14` |
| `basicToDigit` | `punycode.js:144-155` |
| `adapt` (bias adaptation, RFC 3492 §3.4) | `punycode.js:179-187` |
| `error` utility | `punycode.js:41-43` |
| `errors` map | `punycode.js:22-26` |
| Overflow guard — digit accumulation | `punycode.js:243-245` |
| Overflow guard — weight multiplication | `punycode.js:255-257` |
| Overflow guard — `n` increment | `punycode.js:268-270` |
| Output construction via `String.fromCodePoint(...output)` | `punycode.js:280` |
| Export binding | `punycode.js:437` |

---

## 5. Algorithm outline (for porters)

1. **Find delimiter** (`punycode.js:208-211`). Call `input.lastIndexOf('-')`. If the result is `< 0`, set `basic = 0` (no basic prefix); otherwise `basic` is the index of the rightmost `-`.

2. **Copy basic prefix** (`punycode.js:213-219`). For each character at index `j < basic`, assert `charCodeAt(j) < 0x80` (throw `not-basic` if violated), then append the code point to `output`.

3. **Main decoding loop** (`punycode.js:224-278`). Initialise `n = initialN` (128), `bias = initialBias` (72), `i = 0`. Start reading at `index = basic > 0 ? basic + 1 : 0`.

   a. **Decode variable-length integer** (`punycode.js:232-261`). For each base-36 position `k = base, 2*base, ...`:
      - Read one character; convert via `basicToDigit`; reject (`invalid-input`) if `>= base`.
      - Compute threshold `t`:
        - `t = tMin` if `k <= bias` (`punycode.js:248`)
        - `t = tMax` if `k >= bias + tMax` (`punycode.js:248`)
        - `t = k - bias` otherwise (`punycode.js:248`)
      - Accumulate: `i += digit * w`; apply overflow check (`punycode.js:243-245`).
      - If `digit < t`, break (this digit was the last in the sequence).
      - Multiply weight: `w *= base - t`; apply overflow check (`punycode.js:255-257`).

   b. **Adapt bias** (`punycode.js:264`). Call `adapt(i - oldi, output.length + 1, oldi == 0)`.

   c. **Derive code point** (`punycode.js:268-273`). Add `floor(i / out)` to `n`; apply overflow check (`punycode.js:268-270`). Reduce `i` modulo `out`.

   d. **Insert** (`punycode.js:276`). Splice code point `n` into `output` at position `i`, then increment `i`.

4. **Produce string** (`punycode.js:280`). Return `String.fromCodePoint(...output)`.
