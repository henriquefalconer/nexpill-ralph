# Test Vectors: `testData` Fixture

## Overview

The `testData` object defined at `tests/tests.js:6-243` is the shared fixture that drives all parameterized tests in the file. It has four categories: `strings`, `ucs2`, `domains`, and `separators`. The `strings` category is consumed by `describe('punycode.decode')` (lines 290-310) and `describe('punycode.encode')` (lines 312-321), and also supplies the passthrough/no-op assertions inside `describe('punycode.toUnicode')` (lines 332-343) and `describe('punycode.toASCII')` (lines 355-362). The `ucs2` category is consumed by `describe('punycode.ucs2.decode')` (lines 245-271) and `describe('punycode.ucs2.encode')` (lines 273-288). The `domains` category is consumed by `describe('punycode.toUnicode')` (lines 323-344) and `describe('punycode.toASCII')` (lines 346-371). The `separators` category is consumed only by the final loop in `describe('punycode.toASCII')` (lines 363-370), which verifies that alternative Unicode dot separators are normalised to the standard ASCII full stop before encoding.

---

## Category: `strings`

**Source:** `tests/tests.js:7-136`

**Purpose:** Tests round-trip encoding and decoding of Punycode labels (single DNS label, not a full domain). Each vector supplies a Unicode string as `decoded` and its Bootstring/Punycode label as `encoded`. The `encode` function (`punycode.js`) must produce `encoded` from `decoded`; the `decode` function must reproduce `decoded` from `encoded`.

**Shape:** `decoded` is a JavaScript string (Unicode). `encoded` is a JavaScript string (ASCII Punycode label, no `xn--` prefix). Most vectors include a `description` field; vectors at lines 99-125 omit it.

**Note on mixed-case annotation** (`tests/tests.js:74-82`): Punycode.js does not implement the optional mixed-case annotation described in RFC 3492. The RFC sample for Russian (Cyrillic) would encode to `b1abfaaepdrnnbgefbaDotcwatmq2g4l` with annotation, but the correct expected value here is the all-lowercase `b1abfaaepdrnnbgefbadotcwatmq2g4l`. See also the linked issue at `tests/tests.js:81` (https://github.com/mathiasbynens/punycode.js/issues/3).

**RFC 3492 reference:** Vectors at `tests/tests.js:33-125` correspond to the sample strings in RFC 3492 section 7.1 (cited inline at `tests/tests.js:33`).

### Vectors

| # | Description | `decoded` | `encoded` | Source |
|---|-------------|-----------|-----------|--------|
| 1 | a single basic code point | `Bach` | `Bach-` | `tests/tests.js:8-12` |
| 2 | a single non-ASCII character | `\xFC` (ü) | `tda` | `tests/tests.js:13-17` |
| 3 | multiple non-ASCII characters | `\xFC\xEB\xE4\xF6\u2665` | `4can8av2009b` | `tests/tests.js:18-22` |
| 4 | mix of ASCII and non-ASCII characters | `b\xFCcher` | `bcher-kva` | `tests/tests.js:23-27` |
| 5 | long string with both ASCII and non-ASCII characters | `Willst du die Bl\xFCthe des fr\xFChen, die Fr\xFCchte des sp\xE4teren Jahres` | `Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal` | `tests/tests.js:28-32` |
| 6 | Arabic (Egyptian) | `\u0644\u064A\u0647\u0645\u0627\u0628\u062A\u0643\u0644\u0645\u0648\u0634\u0639\u0631\u0628\u064A\u061F` | `egbpdaj6bu4bxfgehfvwxn` | `tests/tests.js:33-38` |
| 7 | Chinese (simplified) | `\u4ED6\u4EEC\u4E3A\u4EC0\u4E48\u4E0D\u8BF4\u4E2d\u6587` | `ihqwcrb4cv8a8dqg056pqjye` | `tests/tests.js:39-43` |
| 8 | Chinese (traditional) | `\u4ED6\u5011\u7232\u4EC0\u9EBD\u4E0D\u8AAA\u4E2D\u6587` | `ihqwctvzc91f659drss3x8bo0yb` | `tests/tests.js:44-48` |
| 9 | Czech | `Pro\u010Dprost\u011Bnemluv\xED\u010Desky` | `Proprostnemluvesky-uyb24dma41a` | `tests/tests.js:49-53` |
| 10 | Hebrew | `\u05DC\u05DE\u05D4\u05D4\u05DD\u05E4\u05E9\u05D5\u05D8\u05DC\u05D0\u05DE\u05D3\u05D1\u05E8\u05D9\u05DD\u05E2\u05D1\u05E8\u05D9\u05EA` | `4dbcagdahymbxekheh6e0a7fei0b` | `tests/tests.js:54-58` |
| 11 | Hindi (Devanagari) | `\u092F\u0939\u0932\u094B\u0917\u0939\u093F\u0928\u094D\u0926\u0940\u0915\u094D\u092F\u094B\u0902\u0928\u0939\u0940\u0902\u092C\u094B\u0932\u0938\u0915\u0924\u0947\u0939\u0948\u0902` | `i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd` | `tests/tests.js:59-63` |
| 12 | Japanese (kanji and hiragana) | `\u306A\u305C\u307F\u3093\u306A\u65E5\u672C\u8A9E\u3092\u8A71\u3057\u3066\u304F\u308C\u306A\u3044\u306E\u304B` | `n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa` | `tests/tests.js:64-68` |
| 13 | Korean (Hangul syllables) | `\uC138\uACC4\uC758\uBAA8\uB4E0\uC0AC\uB78C\uB4E4\uC774\uD55C\uAD6D\uC5B4\uB97C\uC774\uD574\uD55C\uB2E4\uBA74\uC5BC\uB9C8\uB098\uC88B\uC744\uAE4C` | `989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c` | `tests/tests.js:69-73` |
| 14 | Russian (Cyrillic) — no mixed-case annotation (see note above) | `\u043F\u043E\u0447\u0435\u043C\u0443\u0436\u0435\u043E\u043D\u0438\u043D\u0435\u0433\u043E\u0432\u043E\u0440\u044F\u0442\u043F\u043E\u0440\u0443\u0441\u0441\u043A\u0438` | `b1abfaaepdrnnbgefbadotcwatmq2g4l` | `tests/tests.js:83-87` |
| 15 | Spanish | `Porqu\xE9nopuedensimplementehablarenEspa\xF1ol` | `PorqunopuedensimplementehablarenEspaol-fmd56a` | `tests/tests.js:88-92` |
| 16 | Vietnamese | `T\u1EA1isaoh\u1ECDkh\xF4ngth\u1EC3ch\u1EC9n\xF3iti\u1EBFngVi\u1EC7t` | `TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g` | `tests/tests.js:93-97` |
| 17 | _(none)_ | `3\u5E74B\u7D44\u91D1\u516B\u5148\u751F` | `3B-ww4c5e180e575a65lsy2b` | `tests/tests.js:98-101` |
| 18 | _(none)_ | `\u5B89\u5BA4\u5948\u7F8E\u6075-with-SUPER-MONKEYS` | `-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n` | `tests/tests.js:102-105` |
| 19 | _(none)_ | `Hello-Another-Way-\u305D\u308C\u305E\u308C\u306E\u5834\u6240` | `Hello-Another-Way--fc4qua05auwb3674vfr0b` | `tests/tests.js:106-109` |
| 20 | _(none)_ | `\u3072\u3068\u3064\u5C4B\u6839\u306E\u4E0B2` | `2-u9tlzr9756bt3uc0v` | `tests/tests.js:110-113` |
| 21 | _(none)_ | `Maji\u3067Koi\u3059\u308B5\u79D2\u524D` | `MajiKoi5-783gue6qz075azm5e` | `tests/tests.js:114-117` |
| 22 | _(none)_ | `\u30D1\u30D5\u30A3\u30FCde\u30EB\u30F3\u30D0` | `de-jg4avhby1noc0d` | `tests/tests.js:118-121` |
| 23 | _(none)_ | `\u305D\u306E\u30B9\u30D4\u30FC\u30C9\u3067` | `d9juau41awczczp` | `tests/tests.js:122-125` |
| 24 | ASCII string that breaks the existing rules for host-name labels (see inline comment at `tests/tests.js:126-130`) | `-> $1.00 <-` | `-> $1.00 <--` | `tests/tests.js:131-135` |

---

## Category: `ucs2`

**Source:** `tests/tests.js:137-175`

**Purpose:** Tests the low-level UCS-2/UTF-16 codec (`punycode.ucs2.decode` and `punycode.ucs2.encode`). These vectors focus on edge cases involving surrogate pairs and lone surrogates — the comment at `tests/tests.js:137-139` explains that every individual Unicode symbol is assumed tested elsewhere; this category covers only the tricky multi-unit combinations.

**Shape:** `decoded` is an array of Unicode code-point integers (decimal or `0x`-prefixed hex). `encoded` is a JavaScript string that contains the corresponding UTF-16 code units (including raw surrogates where applicable). The relationship is: `ucs2.decode(encoded)` → `decoded`; `ucs2.encode(decoded)` → `encoded`.

### Vectors

| # | Description | `decoded` (code points) | `encoded` (UTF-16 string) | Source |
|---|-------------|-------------------------|---------------------------|--------|
| 1 | Consecutive astral symbols | `[127829, 119808, 119558, 119638]` | `\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56` | `tests/tests.js:140-144` |
| 2 | U+D800 (high surrogate) followed by non-surrogates | `[55296, 97, 98]` | `\uD800ab` | `tests/tests.js:145-149` |
| 3 | U+DC00 (low surrogate) followed by non-surrogates | `[56320, 97, 98]` | `\uDC00ab` | `tests/tests.js:150-154` |
| 4 | High surrogate followed by another high surrogate | `[0xD800, 0xD800]` | `\uD800\uD800` | `tests/tests.js:155-159` |
| 5 | Unmatched high surrogate, followed by a surrogate pair, followed by an unmatched high surrogate | `[0xD800, 0x1D306, 0xD800]` | `\uD800\uD834\uDF06\uD800` | `tests/tests.js:160-164` |
| 6 | Low surrogate followed by another low surrogate | `[0xDC00, 0xDC00]` | `\uDC00\uDC00` | `tests/tests.js:165-169` |
| 7 | Unmatched low surrogate, followed by a surrogate pair, followed by an unmatched low surrogate | `[0xDC00, 0x1D306, 0xDC00]` | `\uDC00\uD834\uDF06\uDC00` | `tests/tests.js:170-174` |

---

## Category: `domains`

**Source:** `tests/tests.js:176-220`

**Purpose:** Tests full domain name conversion between Unicode (`decoded`) and ACE/IDNA (`encoded`). Each label that contains non-ASCII characters is rendered with an `xn--` prefix in the encoded form; labels that are already ASCII (or contain characters that cannot be Punycode-encoded, such as non-printable bytes) are passed through unchanged.

**Shape:** Both `decoded` and `encoded` are JavaScript strings representing complete domain names (dot-separated labels, possibly with a trailing dot). Most vectors omit `description`.

**Notable annotations:**

- `tests/tests.js:181`: inline comment links to https://github.com/mathiasbynens/punycode.js/issues/17, covering the trailing-dot passthrough case (`example.com.` → `example.com.`).
- `tests/tests.js:216`: inline comment links to https://github.com/mathiasbynens/punycode.js/pull/115, covering the DEL character (`\x7F`) passthrough case.

### Vectors

| # | Description | `decoded` | `encoded` | Source |
|---|-------------|-----------|-----------|--------|
| 1 | _(none)_ | `ma\xF1ana.com` | `xn--maana-pta.com` | `tests/tests.js:177-180` |
| 2 | _(none)_ — trailing dot passthrough (issue #17) | `example.com.` | `example.com.` | `tests/tests.js:181-184` |
| 3 | _(none)_ | `b\xFCcher.com` | `xn--bcher-kva.com` | `tests/tests.js:185-188` |
| 4 | _(none)_ | `caf\xE9.com` | `xn--caf-dma.com` | `tests/tests.js:189-192` |
| 5 | _(none)_ | `\u2603-\u2318.com` | `xn----dqo34k.com` | `tests/tests.js:193-196` |
| 6 | _(none)_ | `\uD400\u2603-\u2318.com` | `xn----dqo34kn65z.com` | `tests/tests.js:197-200` |
| 7 | Emoji | `\uD83D\uDCA9.la` | `xn--ls8h.la` | `tests/tests.js:201-205` |
| 8 | Non-printable ASCII | `\0\x01\x02foo.bar` | `\0\x01\x02foo.bar` (passthrough) | `tests/tests.js:206-210` |
| 9 | Email address | `\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a` | `\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` | `tests/tests.js:211-215` |
| 10 | _(none)_ — DEL character passthrough (pull #115) | `foo\x7F.example` | `foo\x7F.example` (passthrough) | `tests/tests.js:216-219` |

---

## Category: `separators`

**Source:** `tests/tests.js:221-242`

**Purpose:** Tests that `punycode.toASCII` recognises all four IDNA2003-compatible label separators and normalises them to a standard U+002E FULL STOP before encoding. All four vectors encode the same base domain (`ma\xF1ana` + separator + `com`) and expect the same `encoded` output (`xn--maana-pta.com`), differing only in which separator character appears in `decoded`.

**Shape:** `decoded` is a JavaScript string containing one of the four Unicode separator code points; `encoded` is always a standard dot-separated ASCII domain. Every vector has a `description`.

**The four recognised separators:**

- `tests/tests.js:223`: U+002E — FULL STOP (standard ASCII dot, `\x2E`)
- `tests/tests.js:228`: U+3002 — IDEOGRAPHIC FULL STOP
- `tests/tests.js:233`: U+FF0E — FULLWIDTH FULL STOP
- `tests/tests.js:238`: U+FF61 — HALFWIDTH IDEOGRAPHIC FULL STOP

### Vectors

| # | Description | `decoded` (separator code point) | `encoded` | Source |
|---|-------------|----------------------------------|-----------|--------|
| 1 | Using U+002E as separator | `ma\xF1ana\x2Ecom` | `xn--maana-pta.com` | `tests/tests.js:222-226` |
| 2 | Using U+3002 as separator | `ma\xF1ana\u3002com` | `xn--maana-pta.com` | `tests/tests.js:227-231` |
| 3 | Using U+FF0E as separator | `ma\xF1ana\uFF0Ecom` | `xn--maana-pta.com` | `tests/tests.js:232-236` |
| 4 | Using U+FF61 as separator | `ma\xF1ana\uFF61com` | `xn--maana-pta.com` | `tests/tests.js:237-241` |

---

## Cross-references

The table below maps each `testData` category to the `describe` blocks that iterate over it.

| Category | `describe` block | Lines | How it is used |
|----------|-----------------|-------|----------------|
| `strings` | `punycode.decode` | `tests/tests.js:290-310` | Iterates `testData.strings`; asserts `decode(encoded) === decoded` for each vector (`tests/tests.js:291-298`). |
| `strings` | `punycode.encode` | `tests/tests.js:312-321` | Iterates `testData.strings`; asserts `encode(decoded) === encoded` for each vector (`tests/tests.js:313-320`). |
| `strings` | `punycode.toUnicode` (passthrough loop) | `tests/tests.js:332-343` | Iterates `testData.strings`; asserts that neither `toUnicode(encoded)` nor `toUnicode(decoded)` transforms values that do not start with `xn--`. |
| `strings` | `punycode.toASCII` (passthrough loop) | `tests/tests.js:355-362` | Iterates `testData.strings`; asserts that `toASCII(encoded)` returns `encoded` unchanged (strings that are already ASCII). |
| `ucs2` | `punycode.ucs2.decode` | `tests/tests.js:245-271` | Iterates `testData.ucs2`; asserts `ucs2.decode(encoded)` equals `decoded` code-point array (`tests/tests.js:246-254`). |
| `ucs2` | `punycode.ucs2.encode` | `tests/tests.js:273-288` | Iterates `testData.ucs2`; asserts `ucs2.encode(decoded)` equals `encoded` string (`tests/tests.js:274-281`). |
| `domains` | `punycode.toUnicode` | `tests/tests.js:323-344` | Iterates `testData.domains`; asserts `toUnicode(encoded) === decoded` (`tests/tests.js:324-331`). |
| `domains` | `punycode.toASCII` | `tests/tests.js:346-371` | Iterates `testData.domains`; asserts `toASCII(decoded) === encoded` (`tests/tests.js:347-354`). |
| `separators` | `punycode.toASCII` (separator loop) | `tests/tests.js:363-370` | Iterates `testData.separators`; asserts `toASCII(decoded) === encoded`, verifying that all IDNA2003 separator variants are normalised before encoding. |
