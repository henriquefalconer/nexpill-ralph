# Spec: `describe('punycode.encode')` — `tests/tests.js:312-321`

## 1. Subject

`punycode.encode(input: string): string`

Converts a Unicode string into its Punycode representation. The output is a plain ASCII string; the `xn--` ACE prefix is **not** prepended (that is handled by `toASCII` at `punycode.js:408-414`). The function is bound to the public API at `punycode.js:438`.

## 2. Contract

Derived from JSDoc at `punycode.js:283-289`:

- **Input** — a string of Unicode symbols (e.g. a domain name label).
- **Output** — the resulting Punycode string of ASCII-only symbols.
- The function is a member of `punycode` (`@memberOf punycode`).

## 3. Test Cases

The entire `describe` block is a single parameterized loop (`tests/tests.js:313-320`). There are no separate static assertions and no error-path tests in this block. For each element `object` of `testData.strings` (`tests/tests.js:7-136`), the suite creates one `it` case whose title is `object.description` when present, or `object.decoded` otherwise (`tests/tests.js:314`). The assertion is:

```
assert.deepEqual(punycode.encode(object.decoded), object.encoded)
```

(`tests/tests.js:315-317`)

This is the exact inverse of the `punycode.decode` suite, running against the same 23 vectors.

### 3.1 Basic ASCII

| # | `decoded` | `encoded` | Source |
|---|-----------|-----------|--------|
| 1 | `'Bach'` | `'Bach-'` | `tests/tests.js:8-12` |

Description: "a single basic code point" (`tests/tests.js:9`). The trailing `-` is the Punycode delimiter emitted because the basic section is non-empty (`punycode.js:318-320`).

### 3.2 Single Non-ASCII Character

| # | `decoded` | `encoded` | Source |
|---|-----------|-----------|--------|
| 2 | `'\xFC'` (U+00FC) | `'tda'` | `tests/tests.js:13-17` |

Description: "a single non-ASCII character" (`tests/tests.js:14`). No basic code points, so no delimiter is written (`punycode.js:318`).

### 3.3 Multiple Non-ASCII Characters

| # | `decoded` | `encoded` | Source |
|---|-----------|-----------|--------|
| 3 | `'\xFC\xEB\xE4\xF6\u2665'` | `'4can8av2009b'` | `tests/tests.js:18-22` |

Description: "multiple non-ASCII characters" (`tests/tests.js:19`).

### 3.4 Mixed ASCII and Non-ASCII

| # | `decoded` | `encoded` | Source |
|---|-----------|-----------|--------|
| 4 | `'b\xFCcher'` | `'bcher-kva'` | `tests/tests.js:23-27` |
| 5 | `'Willst du die Bl\xFCthe des fr\xFChen, die Fr\xFCchte des sp\xE4teren Jahres'` | `'Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal'` | `tests/tests.js:28-32` |
| 14 | `'Porqu\xE9nopuedensimplementehablarenEspa\xF1ol'` | `'PorqunopuedensimplementehablarenEspaol-fmd56a'` | `tests/tests.js:88-92` |
| 15 | `'T\u1EA1isaoh\u1ECDkh\xF4ngth\u1EC3ch\u1EC9n\xF3iti\u1EBFngVi\u1EC7t'` | `'TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g'` | `tests/tests.js:93-97` |

Descriptions: "mix of ASCII and non-ASCII characters" (`tests/tests.js:24`), "long string with both ASCII and non-ASCII characters" (`tests/tests.js:29`), "Spanish" (`tests/tests.js:89`), "Vietnamese" (`tests/tests.js:94`).

### 3.5 RFC 3492 Section 7.1 Samples

The vectors at `tests/tests.js:33-87` are drawn from the reference examples in RFC 3492 §7.1.

| # | Description | `decoded` | `encoded` | Source |
|---|-------------|-----------|-----------|--------|
| 6 | Arabic (Egyptian) | `'\u0644\u064A\u0647\u0645\u0627\u0628\u062A\u0643\u0644\u0645\u0648\u0634\u0639\u0631\u0628\u064A\u061F'` | `'egbpdaj6bu4bxfgehfvwxn'` | `tests/tests.js:34-38` |
| 7 | Chinese (simplified) | `'\u4ED6\u4EEC\u4E3A\u4EC0\u4E48\u4E0D\u8BF4\u4E2d\u6587'` | `'ihqwcrb4cv8a8dqg056pqjye'` | `tests/tests.js:39-43` |
| 8 | Chinese (traditional) | `'\u4ED6\u5011\u7232\u4EC0\u9EBD\u4E0D\u8AAA\u4E2D\u6587'` | `'ihqwctvzc91f659drss3x8bo0yb'` | `tests/tests.js:44-48` |
| 9 | Czech | `'Pro\u010Dprost\u011Bnemluv\xED\u010Desky'` | `'Proprostnemluvesky-uyb24dma41a'` | `tests/tests.js:49-53` |
| 10 | Hebrew | `'\u05DC\u05DE\u05D4\u05D4\u05DD\u05E4\u05E9\u05D5\u05D8\u05DC\u05D0\u05DE\u05D3\u05D1\u05E8\u05D9\u05DD\u05E2\u05D1\u05E8\u05D9\u05EA'` | `'4dbcagdahymbxekheh6e0a7fei0b'` | `tests/tests.js:54-58` |
| 11 | Hindi (Devanagari) | `'\u092F\u0939\u0932\u094B\u0917\u0939\u093F\u0928\u094D\u0926\u0940\u0915\u094D\u092F\u094B\u0902\u0928\u0939\u0940\u0902\u092C\u094B\u0932\u0938\u0915\u0924\u0947\u0939\u0948\u0902'` | `'i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd'` | `tests/tests.js:59-63` |
| 12 | Japanese (kanji and hiragana) | `'\u306A\u305C\u307F\u3093\u306A\u65E5\u672C\u8A9E\u3092\u8A71\u3057\u3066\u304F\u308C\u306A\u3044\u306E\u304B'` | `'n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa'` | `tests/tests.js:64-68` |
| 13 | Korean (Hangul syllables) | `'\uC138\uACC4\uC758\uBAA8\uB4E0\uC0AC\uB78C\uB4E4\uC774\uD55C\uAD6D\uC5B4\uB97C\uC774\uD574\uD55C\uB2E4\uBA74\uC5BC\uB9C8\uB098\uC88B\uC744\uAE4C'` | `'989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c'` | `tests/tests.js:69-73` |
| 14 | Russian (Cyrillic) | `'\u043F\u043E\u0447\u0435\u043C\u0443\u0436\u0435\u043E\u043D\u0438\u043D\u0435\u0433\u043E\u0432\u043E\u0440\u044F\u0442\u043F\u043E\u0440\u0443\u0441\u0441\u043A\u0438'` | `'b1abfaaepdrnnbgefbadotcwatmq2g4l'` | `tests/tests.js:83-87` |

The Russian entry carries a note (`tests/tests.js:74-82`) explaining that the RFC §7.1 reference output uses mixed-case annotation (`b1abfaaepdrnnbgefbaDotcwatmq2g4l`) but Punycode.js cannot replicate this because JavaScript provides no way to do so. The library always produces lowercase, so the expected value is `b1abfaaepdrnnbgefbadotcwatmq2g4l` (`tests/tests.js:86`). See Section 4 below for the implementation reason.

### 3.6 Unnamed Mixed-Script Vectors (RFC 3492 §7.1, continued)

These eight vectors have no `description` field; the test title falls back to `object.decoded` (`tests/tests.js:314`).

| # | `decoded` | `encoded` | Source |
|---|-----------|-----------|--------|
| 16 | `'3\u5E74B\u7D44\u91D1\u516B\u5148\u751F'` | `'3B-ww4c5e180e575a65lsy2b'` | `tests/tests.js:98-101` |
| 17 | `'\u5B89\u5BA4\u5948\u7F8E\u6075-with-SUPER-MONKEYS'` | `'-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n'` | `tests/tests.js:102-105` |
| 18 | `'Hello-Another-Way-\u305D\u308C\u305E\u308C\u306E\u5834\u6240'` | `'Hello-Another-Way--fc4qua05auwb3674vfr0b'` | `tests/tests.js:106-109` |
| 19 | `'\u3072\u3068\u3064\u5C4B\u6839\u306E\u4E0B2'` | `'2-u9tlzr9756bt3uc0v'` | `tests/tests.js:110-113` |
| 20 | `'Maji\u3067Koi\u3059\u308B5\u79D2\u524D'` | `'MajiKoi5-783gue6qz075azm5e'` | `tests/tests.js:114-117` |
| 21 | `'\u30D1\u30D5\u30A3\u30FCde\u30EB\u30F3\u30D0'` | `'de-jg4avhby1noc0d'` | `tests/tests.js:118-121` |
| 22 | `'\u305D\u306E\u30B9\u30D4\u30FC\u30C9\u3067'` | `'d9juau41awczczp'` | `tests/tests.js:122-125` |

### 3.7 ASCII Edge Case

| # | Description | `decoded` | `encoded` | Source |
|---|-------------|-----------|-----------|--------|
| 23 | ASCII string that breaks the existing rules for host-name labels | `'-> $1.00 <-'` | `'-> $1.00 <--'` | `tests/tests.js:131-135` |

The comment at `tests/tests.js:126-130` notes that this is not a realistic IDNA example because IDNA never encodes pure-ASCII labels. The input contains only basic code points, so the basic section covers the entire string. Because `basicLength` is non-zero, the delimiter `-` is appended (`punycode.js:318-319`), producing the double-hyphen suffix visible in the expected output. The main loop is never entered because `handledCPCount === inputLength` from the start (`punycode.js:323`).

## 4. Implementation Citations

**Function definition:** `punycode.js:290-376`. Bound to the public API at `punycode.js:438`.

**Step 1 — UCS-2 to code-point array** (`punycode.js:294`): the input string is passed through `ucs2decode` (defined at `punycode.js:101-123`). This converts surrogate pairs into a single code point each, so the encoder always operates on Unicode scalar values rather than raw UTF-16 code units.

**Step 2 — Basic code-point collection** (`punycode.js:305-309`): a loop over the decoded array pushes every value `< 0x80` as a character literal using `stringFromCharCode` (aliased at `punycode.js:31`, called at `punycode.js:307`).

**Step 3 — Delimiter** (`punycode.js:318-320`): the delimiter `-` is written to the output only when `basicLength > 0`. When all input code points are non-basic the delimiter is suppressed entirely. This rule produces the `'-> $1.00 <--'` result for the ASCII edge case (`tests/tests.js:131-135`) and explains the absence of a trailing `-` in all-non-ASCII outputs such as `'tda'` and `'4can8av2009b'`.

**Step 4 — Main encoding loop** (`punycode.js:323-374`):

- The smallest code point `m` that is `>= n` and not yet handled is found by a linear scan (`punycode.js:327-332`).
- `delta` is advanced by `(m - n) * handledCPCountPlusOne` to move the decoder's state forward, with an overflow guard against `maxInt = 2147483647` (`punycode.js:4`, `punycode.js:336-341`).
- For each code point in the input that equals `n`, the current `delta` is encoded as a generalized variable-length integer and emitted digit-by-digit (`punycode.js:344-369`). After each emission the bias is updated via `adapt` and `delta` is reset to zero (`punycode.js:365-366`).

**`digitToBasic`** (`punycode.js:168-172`): converts an integer digit in `[0, 35]` to its ASCII character. The `flag` parameter controls uppercase vs. lowercase. In `encode`, every call to `digitToBasic` passes `flag = 0` (literally `0`, at `punycode.js:359` and `punycode.js:364`), so the output is always lowercase. This is the direct reason the Russian RFC §7.1 sample encodes to all-lowercase `b1abfaaepdrnnbgefbadotcwatmq2g4l` rather than the mixed-case RFC reference `b1abfaaepdrnnbgefbaDotcwatmq2g4l` (`tests/tests.js:74-86`).

**`adapt`** (`punycode.js:179-187`): bias-adaptation function per RFC 3492 §3.4, called after each non-basic code point is emitted (`punycode.js:365`).

## 5. Algorithm Outline (for Porters)

1. Convert the input string to an array of Unicode code points via `ucs2decode` (`punycode.js:101-123`, called at `punycode.js:294`).
2. Emit all code points `< 0x80` literally as ASCII characters (`punycode.js:305-309`).
3. If any basic code points were emitted, append the delimiter `-` (`punycode.js:318-320`).
4. Initialize `n = 128`, `delta = 0`, `bias = 72` (`punycode.js:300-302`).
5. Repeat until all code points have been handled (`punycode.js:323`):
   a. Find the smallest non-basic code point `m >= n` still present in the input (`punycode.js:327-332`).
   b. Increase `delta` by `(m - n) * (handledCPCount + 1)`, checking for overflow (`punycode.js:336-341`).
   c. Set `n = m`.
   d. For each position in the input where the code point equals `n`, encode `delta` as a generalized VLQ and emit the digits; adapt the bias; reset `delta` to zero; increment `handledCPCount` (`punycode.js:344-369`).
   e. Increment `delta` and `n` (`punycode.js:371-372`).
6. Return the joined output array (`punycode.js:375`).

## 6. Porting Notes

- **Zero-basic-length path**: when `basicLength === 0` the delimiter must be omitted. Emitting it unconditionally would prepend a spurious `-` to the encoded string. See `punycode.js:318` and the all-non-ASCII test vectors (e.g. `tests/tests.js:13-17`).
- **Lowercase-only output**: the `flag` argument to `digitToBasic` is always `0` (`punycode.js:359`, `punycode.js:364`). Do not replicate the RFC mixed-case flag. The Russian vector at `tests/tests.js:83-87` serves as the canonical regression test for this contract.
- **Overflow constant**: the 32-bit signed maximum `maxInt = 2147483647` (`punycode.js:4`) is used in both the delta-advance overflow guard (`punycode.js:337`) and the per-code-point delta increment guard (`punycode.js:345`). Ports to languages with wider native integers must clamp to this value explicitly.
- **Surrogate pairs**: `ucs2decode` (`punycode.js:101-123`) must be applied before any code-point comparison. Omitting this step causes surrogate halves (two `< 0x80`? No — they are in `[0xD800, 0xDFFF]`) to be treated as two separate non-basic code points rather than one combined scalar value, producing incorrect output for strings containing characters above U+FFFF.
