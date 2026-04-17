# Punycode Library — Test Specification

This document captures every behavior asserted by `tests/tests.js`. Every claim cites `tests/tests.js:<line>`. The spec is language-agnostic: it describes what is observed of the public API, not any particular implementation.

The public API under test exposes five operations:

- `ucs2.decode(string) -> array of integer code points`
- `ucs2.encode(array of integer code points) -> string`
- `decode(string) -> string` (raw Punycode decoder; no `xn--` prefix, no label splitting)
- `encode(string) -> string` (raw Punycode encoder)
- `toUnicode(string) -> string` (domain-level decoder; splits on label separators, processes `xn--` labels)
- `toASCII(string) -> string` (domain-level encoder; splits on label separators, prefixes `xn--` on non-ASCII labels)

All string comparisons are exact (structural equality) (tests/tests.js:3, tests/tests.js:248, tests/tests.js:276, tests/tests.js:293, tests/tests.js:315, tests/tests.js:326, tests/tests.js:349).

---

## Fixture Data

### `testData.strings` — raw Punycode pairs (tests/tests.js:7–136)

Used by `decode` (decoded ↔ encoded), `encode` (decoded → encoded), and negatively by `toUnicode`/`toASCII` (pass-through since none of these encoded forms is prefixed `xn--`).

| # | description | decoded | encoded | line |
|---|---|---|---|---|
| 1 | a single basic code point | `Bach` | `Bach-` | tests/tests.js:8 |
| 2 | a single non-ASCII character | `\xFC` | `tda` | tests/tests.js:13 |
| 3 | multiple non-ASCII characters | `\xFC\xEB\xE4\xF6\u2665` | `4can8av2009b` | tests/tests.js:18 |
| 4 | mix of ASCII and non-ASCII characters | `b\xFCcher` | `bcher-kva` | tests/tests.js:23 |
| 5 | long string with both ASCII and non-ASCII characters | `Willst du die Bl\xFCthe des fr\xFChen, die Fr\xFCchte des sp\xE4teren Jahres` | `Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal` | tests/tests.js:28 |
| 6 | Arabic (Egyptian) | `\u0644\u064A\u0647\u0645\u0627\u0628\u062A\u0643\u0644\u0645\u0648\u0634\u0639\u0631\u0628\u064A\u061F` | `egbpdaj6bu4bxfgehfvwxn` | tests/tests.js:34 |
| 7 | Chinese (simplified) | `\u4ED6\u4EEC\u4E3A\u4EC0\u4E48\u4E0D\u8BF4\u4E2d\u6587` | `ihqwcrb4cv8a8dqg056pqjye` | tests/tests.js:39 |
| 8 | Chinese (traditional) | `\u4ED6\u5011\u7232\u4EC0\u9EBD\u4E0D\u8AAA\u4E2D\u6587` | `ihqwctvzc91f659drss3x8bo0yb` | tests/tests.js:44 |
| 9 | Czech | `Pro\u010Dprost\u011Bnemluv\xED\u010Desky` | `Proprostnemluvesky-uyb24dma41a` | tests/tests.js:49 |
| 10 | Hebrew | `\u05DC\u05DE\u05D4\u05D4\u05DD\u05E4\u05E9\u05D5\u05D8\u05DC\u05D0\u05DE\u05D3\u05D1\u05E8\u05D9\u05DD\u05E2\u05D1\u05E8\u05D9\u05EA` | `4dbcagdahymbxekheh6e0a7fei0b` | tests/tests.js:54 |
| 11 | Hindi (Devanagari) | `\u092F\u0939\u0932\u094B\u0917\u0939\u093F\u0928\u094D\u0926\u0940\u0915\u094D\u092F\u094B\u0902\u0928\u0939\u0940\u0902\u092C\u094B\u0932\u0938\u0915\u0924\u0947\u0939\u0948\u0902` | `i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd` | tests/tests.js:59 |
| 12 | Japanese (kanji and hiragana) | `\u306A\u305C\u307F\u3093\u306A\u65E5\u672C\u8A9E\u3092\u8A71\u3057\u3066\u304F\u308C\u306A\u3044\u306E\u304B` | `n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa` | tests/tests.js:64 |
| 13 | Korean (Hangul syllables) | `\uC138\uACC4\uC758\uBAA8\uB4E0\uC0AC\uB78C\uB4E4\uC774\uD55C\uAD6D\uC5B4\uB97C\uC774\uD574\uD55C\uB2E4\uBA74\uC5BC\uB9C8\uB098\uC88B\uC744\uAE4C` | `989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c` | tests/tests.js:69 |
| 14 | Russian (Cyrillic) | `\u043F\u043E\u0447\u0435\u043C\u0443\u0436\u0435\u043E\u043D\u0438\u043D\u0435\u0433\u043E\u0432\u043E\u0440\u044F\u0442\u043F\u043E\u0440\u0443\u0441\u0441\u043A\u0438` | `b1abfaaepdrnnbgefbadotcwatmq2g4l` | tests/tests.js:83 |
| 15 | Spanish | `Porqu\xE9nopuedensimplementehablarenEspa\xF1ol` | `PorqunopuedensimplementehablarenEspaol-fmd56a` | tests/tests.js:88 |
| 16 | Vietnamese | `T\u1EA1isaoh\u1ECDkh\xF4ngth\u1EC3ch\u1EC9n\xF3iti\u1EBFngVi\u1EC7t` | `TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g` | tests/tests.js:93 |
| 17 | (no description) | `3\u5E74B\u7D44\u91D1\u516B\u5148\u751F` | `3B-ww4c5e180e575a65lsy2b` | tests/tests.js:98 |
| 18 | (no description) | `\u5B89\u5BA4\u5948\u7F8E\u6075-with-SUPER-MONKEYS` | `-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n` | tests/tests.js:102 |
| 19 | (no description) | `Hello-Another-Way-\u305D\u308C\u305E\u308C\u306E\u5834\u6240` | `Hello-Another-Way--fc4qua05auwb3674vfr0b` | tests/tests.js:106 |
| 20 | (no description) | `\u3072\u3068\u3064\u5C4B\u6839\u306E\u4E0B2` | `2-u9tlzr9756bt3uc0v` | tests/tests.js:110 |
| 21 | (no description) | `Maji\u3067Koi\u3059\u308B5\u79D2\u524D` | `MajiKoi5-783gue6qz075azm5e` | tests/tests.js:114 |
| 22 | (no description) | `\u30D1\u30D5\u30A3\u30FCde\u30EB\u30F3\u30D0` | `de-jg4avhby1noc0d` | tests/tests.js:118 |
| 23 | (no description) | `\u305D\u306E\u30B9\u30D4\u30FC\u30C9\u3067` | `d9juau41awczczp` | tests/tests.js:122 |
| 24 | ASCII string that breaks the existing rules for host-name labels | `-> $1.00 <-` | `-> $1.00 <--` | tests/tests.js:131 |

Notes:

- Vector 14 (Russian) is documented as intentionally omitting RFC 3492 mixed-case annotation. The RFC's example encodes to `b1abfaaepdrnnbgefbaDotcwatmq2g4l`; this library emits `b1abfaaepdrnnbgefbadotcwatmq2g4l` (all lowercase) and the test asserts the lowercase form (tests/tests.js:74–87).

### `testData.ucs2` — UCS-2 round-trip vectors (tests/tests.js:137–175)

Every Unicode symbol is exercised individually elsewhere; this table holds the multi-symbol / surrogate edge cases. The `decoded` side is an array of integer code points; the `encoded` side is a string of UTF-16 code units.

| # | description | decoded (code points) | encoded (UTF-16 code units) | line |
|---|---|---|---|---|
| 1 | Consecutive astral symbols | `[127829, 119808, 119558, 119638]` | `\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56` | tests/tests.js:140 |
| 2 | U+D800 (high surrogate) followed by non-surrogates | `[55296, 97, 98]` | `\uD800ab` | tests/tests.js:145 |
| 3 | U+DC00 (low surrogate) followed by non-surrogates | `[56320, 97, 98]` | `\uDC00ab` | tests/tests.js:150 |
| 4 | High surrogate followed by another high surrogate | `[0xD800, 0xD800]` | `\uD800\uD800` | tests/tests.js:155 |
| 5 | Unmatched high surrogate, followed by a surrogate pair, followed by an unmatched high surrogate | `[0xD800, 0x1D306, 0xD800]` | `\uD800\uD834\uDF06\uD800` | tests/tests.js:160 |
| 6 | Low surrogate followed by another low surrogate | `[0xDC00, 0xDC00]` | `\uDC00\uDC00` | tests/tests.js:165 |
| 7 | Unmatched low surrogate, followed by a surrogate pair, followed by an unmatched low surrogate | `[0xDC00, 0x1D306, 0xDC00]` | `\uDC00\uD834\uDF06\uDC00` | tests/tests.js:170 |

### `testData.domains` — domain-level round-trip vectors (tests/tests.js:176–220)

| # | description | decoded | encoded | line |
|---|---|---|---|---|
| 1 | (no description) | `ma\xF1ana.com` | `xn--maana-pta.com` | tests/tests.js:177 |
| 2 | (no description) | `example.com.` | `example.com.` | tests/tests.js:181 |
| 3 | (no description) | `b\xFCcher.com` | `xn--bcher-kva.com` | tests/tests.js:185 |
| 4 | (no description) | `caf\xE9.com` | `xn--caf-dma.com` | tests/tests.js:189 |
| 5 | (no description) | `\u2603-\u2318.com` | `xn----dqo34k.com` | tests/tests.js:193 |
| 6 | (no description) | `\uD400\u2603-\u2318.com` | `xn----dqo34kn65z.com` | tests/tests.js:197 |
| 7 | Emoji | `\uD83D\uDCA9.la` | `xn--ls8h.la` | tests/tests.js:201 |
| 8 | Non-printable ASCII | `\0\x01\x02foo.bar` | `\0\x01\x02foo.bar` | tests/tests.js:206 |
| 9 | Email address | `\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a` | `\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` | tests/tests.js:211 |
| 10 | (no description) | `foo\x7F.example` | `foo\x7F.example` | tests/tests.js:216 |

### `testData.separators` — IDNA2003 label separators (tests/tests.js:221–242)

All four vectors use the same `encoded` form `xn--maana-pta.com`; only the separator character in the `decoded` input differs.

| # | description | decoded | encoded | line |
|---|---|---|---|---|
| 1 | Using U+002E as separator | `ma\xF1ana\x2Ecom` | `xn--maana-pta.com` | tests/tests.js:222 |
| 2 | Using U+3002 as separator | `ma\xF1ana\u3002com` | `xn--maana-pta.com` | tests/tests.js:227 |
| 3 | Using U+FF0E as separator | `ma\xF1ana\uFF0Ecom` | `xn--maana-pta.com` | tests/tests.js:232 |
| 4 | Using U+FF61 as separator | `ma\xF1ana\uFF61com` | `xn--maana-pta.com` | tests/tests.js:237 |

---

## `punycode.ucs2.decode`

Group defined at tests/tests.js:245.

### Vector behaviors

For each vector in `testData.ucs2`, the function `ucs2.decode(encoded)` returns an array structurally equal to `decoded` (tests/tests.js:246–253). Per-vector citations:

- Consecutive astral symbols — tests/tests.js:140
- U+D800 followed by non-surrogates — tests/tests.js:145
- U+DC00 followed by non-surrogates — tests/tests.js:150
- High surrogate followed by another high surrogate — tests/tests.js:155
- Unmatched high surrogate bracketing a surrogate pair — tests/tests.js:160
- Low surrogate followed by another low surrogate — tests/tests.js:165
- Unmatched low surrogate bracketing a surrogate pair — tests/tests.js:170

### Error behaviors

Two additional behaviors are declared inside this group but they invoke `punycode.decode` (not `punycode.ucs2.decode`). Preserve this as observed; do not relocate.

- Invoking `decode('\x81-')` raises a range error. The `it(...)` label text is `"throws RangeError: Illegal input >= 0x80 (not a basic code point)"`, but only the error class is asserted — no error message is checked (tests/tests.js:255–262).
- Invoking `decode('\x81')` raises a range error. The `it(...)` label text is `"throws RangeError: Overflow: input needs wider integers to process"`, but only the error class is asserted — no error message is checked (tests/tests.js:263–270).

---

## `punycode.ucs2.encode`

Group defined at tests/tests.js:273.

### Vector behaviors

For each vector in `testData.ucs2`, the function `ucs2.encode(decoded)` returns a string structurally equal to `encoded` (tests/tests.js:274–281). Per-vector citations are the same as for `ucs2.decode` above (tests/tests.js:140, 145, 150, 155, 160, 165, 170).

### Invariant: does not mutate its input

Given an input array of integer code points `[0x61, 0x62, 0x63]`:

- The call returns the string `"abc"` (tests/tests.js:283, tests/tests.js:285).
- After the call, the input array remains structurally equal to `[0x61, 0x62, 0x63]` — the function must not mutate its argument (tests/tests.js:282–287).

---

## `punycode.decode`

Group defined at tests/tests.js:290.

### Vector behaviors

For each vector in `testData.strings`, `decode(encoded)` returns a string equal to `decoded` (tests/tests.js:291–298). All 24 string vectors from the table above apply.

### Case-insensitive input

`decode('ZZZ')` returns `\u7BA5` — uppercase ASCII letters are accepted identically to lowercase (tests/tests.js:299–301).

### Malformed input

`decode('ls8h=')` raises a range error. The `it(...)` label is `"throws RangeError: Invalid input"`, but only the error class is asserted — no error message is checked (tests/tests.js:302–309).

---

## `punycode.encode`

Group defined at tests/tests.js:312.

### Vector behaviors

For each vector in `testData.strings`, `encode(decoded)` returns a string equal to `encoded` (tests/tests.js:313–320). All 24 string vectors from the table above apply; this is the inverse direction of `decode`.

---

## `punycode.toUnicode`

Group defined at tests/tests.js:323.

### Vector behaviors (happy path)

For each vector in `testData.domains`, `toUnicode(encoded)` returns a string equal to `decoded` (tests/tests.js:324–330). Per-vector citations: tests/tests.js:177, 181, 185, 189, 193, 197, 201, 206, 211, 216.

Observations captured by these vectors:

- Domains with no `xn--` labels are returned exactly as input (tests/tests.js:181–184, 206–210, 216–219).
- A trailing empty label (trailing dot) is preserved (tests/tests.js:181–184).
- Labels prefixed `xn--` are decoded; labels without the prefix are passed through unchanged within the same domain (tests/tests.js:211–215).

### Invariant: pass-through for non-`xn--` inputs

For every vector in `testData.strings`, the function is called twice — once on the `encoded` field and once on the `decoded` field — and in both cases returns its argument unchanged (tests/tests.js:332–343). Since none of the `testData.strings` vectors begins with the literal ASCII sequence `xn--`, the observed rule is:

- If (and apparently only if) an input label begins with `xn--`, that label is Punycode-decoded; otherwise the label is returned unchanged.
- Pass-through applies to both already-ASCII strings and strings containing non-ASCII characters, as long as no label begins with `xn--`.

---

## `punycode.toASCII`

Group defined at tests/tests.js:346.

### Vector behaviors (happy path)

For each vector in `testData.domains`, `toASCII(decoded)` returns a string equal to `encoded` (tests/tests.js:347–353). Per-vector citations: tests/tests.js:177, 181, 185, 189, 193, 197, 201, 206, 211, 216.

Observations captured by these vectors:

- A label containing only ASCII characters is emitted verbatim with no `xn--` prefix (tests/tests.js:181–184, 206–210, 216–219).
- A label containing any non-ASCII code point is emitted as `xn--` followed by its Punycode encoding (tests/tests.js:177–180, 185–188, 189–192, 193–196, 197–200, 201–205).
- In an email-style input, the portion before `@` is a user part and is emitted unchanged; only the domain portion after `@` is processed label-by-label (tests/tests.js:211–215).
- A non-printable ASCII code point (including `\x7F` and the C0 controls `\0`, `\x01`, `\x02`) is not treated as non-ASCII and does not trigger `xn--` encoding (tests/tests.js:206–210, 216–219).
- A trailing empty label (trailing dot) is preserved (tests/tests.js:181–184).

### Invariant: pass-through for already-ASCII input

For every vector in `testData.strings`, `toASCII(encoded)` returns the input unchanged (tests/tests.js:355–362). Only the `encoded` field — which is pure ASCII in every `testData.strings` vector — is tested here; the `decoded` side is NOT asserted in this group.

### Invariant: IDNA2003 label separators

For every vector in `testData.separators`, `toASCII(decoded)` returns `encoded` (tests/tests.js:363–370). This asserts that when splitting a domain into labels, any of the following four code points is recognized as a separator:

- U+002E (full stop)
- U+3002 (ideographic full stop)
- U+FF0E (fullwidth full stop)
- U+FF61 (halfwidth ideographic full stop)

All four vectors produce the same encoded output `xn--maana-pta.com`, so the emitted separator in the encoded form is always U+002E regardless of which of the four separators appeared in the input (tests/tests.js:221–242).

---

## Summary of explicit invariants

- `ucs2.encode` does not mutate its input array (tests/tests.js:282–287).
- `ucs2.decode` and `ucs2.encode` round-trip surrogate edge cases: lone high surrogate, lone low surrogate, consecutive-high, consecutive-low, lone surrogate bracketing an astral code point on either side, and sequences of astral code points (tests/tests.js:140, 145, 150, 155, 160, 165, 170).
- `decode` is case-insensitive over ASCII letters (tests/tests.js:299–301).
- `decode` raises a range error on `'\x81-'`, `'\x81'`, and `'ls8h='`; only the error class is asserted, not any error message (tests/tests.js:255–270, 302–309).
- `toUnicode` passes through any input whose labels do not start with `xn--`, whether ASCII-only or containing non-ASCII code points (tests/tests.js:332–343).
- `toASCII` passes through inputs that are already pure ASCII (per the `encoded` side of `testData.strings`) (tests/tests.js:355–362).
- `toASCII` and `toUnicode` split domains on any of U+002E, U+3002, U+FF0E, U+FF61 on input, and emit U+002E on output (tests/tests.js:221–242).
- Domain-level encoding only prefixes `xn--` on labels that contain non-ASCII code points; pure-ASCII labels (including labels containing C0 controls and U+007F) are passed through unchanged (tests/tests.js:181–184, 206–210, 216–219).
- The Russian (Cyrillic) vector intentionally does not use RFC 3492 mixed-case annotation; the expected encoding is all-lowercase `b1abfaaepdrnnbgefbadotcwatmq2g4l`, not the RFC sample `b1abfaaepdrnnbgefbaDotcwatmq2g4l` (tests/tests.js:74–87).
