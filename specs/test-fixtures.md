# `testData` Fixture Specification — punycode.js Test Suite

## Overview

The `testData` object (`tests/tests.js:6-243`) is the single shared fixture that backs all five `describe` blocks in the test file. It is declared as a `const` at module scope and contains four named arrays: `strings`, `ucs2`, `domains`, and `separators`. Each array holds plain objects with at minimum a `decoded` field and an `encoded` field; most also carry a `description` string used as the `it()` label (`tests/tests.js:292`, `313`, `325`, `349`).

---

## Array 1: `strings` (`tests/tests.js:7-136`)

### Purpose

`strings` tests the low-level Punycode codec (`punycode.decode` / `punycode.encode`) in isolation — that is, without the `xn--` prefix and without domain-label splitting. It covers a representative sweep of Unicode scripts, pure-ASCII edge cases, and every RFC 3492 §7.1 sample string.

### Field semantics

| Field | Meaning |
|---|---|
| `decoded` | The raw Unicode string (the human-readable form). |
| `encoded` | The bare Punycode output — no `xn--` prefix. Basic-only strings get a trailing hyphen-delimiter appended by the encoder (e.g. `'Bach'` → `'Bach-'`, `tests/tests.js:11`). |

The trailing hyphen is significant: `punycode.encode` appends the `delimiter` constant (`'-'`, `punycode.js:14`) after the run of basic code points whenever that run is non-empty (`punycode.js:319`).

### Test suite consumers

- `describe('punycode.decode', …)` (`tests/tests.js:290-310`): passes `object.encoded` to `punycode.decode` and asserts equality to `object.decoded`.
- `describe('punycode.encode', …)` (`tests/tests.js:312-321`): passes `object.decoded` to `punycode.encode` and asserts equality to `object.encoded`.
- `describe('punycode.toUnicode', …)` (`tests/tests.js:323-344`): uses `strings` in a secondary loop (`tests/tests.js:332-343`) to assert that strings not starting with `xn--` are returned unchanged by `toUnicode` — i.e. it verifies the `regexPunycode` guard at `punycode.js:17` and `punycode.js:391-393`.
- `describe('punycode.toASCII', …)` (`tests/tests.js:346-371`): uses `strings` in a secondary loop (`tests/tests.js:355-360`) to assert that already-ASCII encoded strings are returned unchanged by `toASCII` — verifying the `regexNonASCII` guard at `punycode.js:18` and `punycode.js:410-412`.

### Notable fixtures

**Single basic code point** (`tests/tests.js:8-12`): `'Bach'` → `'Bach-'`. Demonstrates that a string composed entirely of basic (ASCII) code points still gets the delimiter appended. The delimiter is stripped by `decode` via `input.lastIndexOf(delimiter)` at `punycode.js:208`.

**Single non-ASCII character** (`tests/tests.js:13-17`): `'\xFC'` (U+00FC, ü) → `'tda'`. No delimiter because there are no basic code points at all.

**RFC 3492 §7.1 sample strings** (`tests/tests.js:33-125`): The block comment at `tests/tests.js:33` references `https://tools.ietf.org/html/rfc3492#section-7.1`. The fourteen multilingual samples that follow (Arabic, Chinese Simplified, Chinese Traditional, Czech, Hebrew, Hindi/Devanagari, Japanese, Korean, Russian, Spanish, Vietnamese, and four Japanese/mixed strings without `description` fields) are the canonical interoperability test vectors from the RFC.

**Mixed-case annotation caveat — Russian (Cyrillic)** (`tests/tests.js:74-87`): The block comment at `tests/tests.js:74-82` explains a deliberate divergence from the RFC sample. RFC 3492 encodes the Russian string with mixed-case annotation (`b1abfaaepdrnnbgefbaDotcwatmq2g4l`), which encodes the case of the original ASCII characters in the delta stream. JavaScript has no mechanism to detect this at the language level, so the library always emits lowercase: `'b1abfaaepdrnnbgefbadotcwatmq2g4l'` (`tests/tests.js:86`). The `digitToBasic` function at `punycode.js:168-172` controls case: it subtracts `0x20` only when `flag != 0`, and `encode` always passes `flag = 0` at `punycode.js:359` and `punycode.js:364`.

**ASCII string that breaks host-name rules** (`tests/tests.js:131-135`): `'-> $1.00 <-'` → `'-> $1.00 <--'`. The comment at `tests/tests.js:126-130` notes this is not a realistic IDNA label; it exercises the encoder on punctuation and spaces. The double trailing hyphen is the basic-code-point delimiter (the input ends in `'-'`, a basic character, so the encoder appends another `'-'`).

---

## Array 2: `ucs2` (`tests/tests.js:137-175`)

### Purpose

`ucs2` tests the `punycode.ucs2.decode` and `punycode.ucs2.encode` utility methods, which bridge JavaScript's internal UCS-2 string representation and an array of true Unicode code points. The comment at `tests/tests.js:138-139` states that individual Unicode symbols are tested elsewhere; this array focuses exclusively on multi-symbol combinations, particularly surrogate edge cases.

### Field semantics

| Field | Meaning |
|---|---|
| `decoded` | An **array of numeric Unicode code points** (integers). |
| `encoded` | A JavaScript **UCS-2 string** — what JavaScript's `String` type actually stores, including raw surrogate halves where applicable. |

The directionality is intentionally reversed from intuition: `ucs2.decode` takes the UCS-2 `encoded` string as **input** and returns the code-point `decoded` array; `ucs2.encode` takes the `decoded` array and returns the `encoded` string. The test harness reflects this: `tests/tests.js:249` calls `punycode.ucs2.decode(object.encoded)` expecting `object.decoded`, and `tests/tests.js:277` calls `punycode.ucs2.encode(object.decoded)` expecting `object.encoded`.

### Test suite consumers

- `describe('punycode.ucs2.decode', …)` (`tests/tests.js:245-271`).
- `describe('punycode.ucs2.encode', …)` (`tests/tests.js:273-288`).

### Implementation references

`ucs2decode` (`punycode.js:101-123`) inspects each `charCodeAt` value. When a value falls in the high-surrogate range `0xD800–0xDBFF` and is followed by a low surrogate (`(extra & 0xFC00) == 0xDC00`, `punycode.js:110`), the pair is combined into a single astral code point via `((value & 0x3FF) << 10) + (extra & 0x3FF) + 0x10000` (`punycode.js:111`). Unmatched surrogates fall through to `output.push(value)` at `punycode.js:115` or `punycode.js:119`.

`ucs2encode` is simply `codePoints => String.fromCodePoint(...codePoints)` (`punycode.js:133`), which handles astral code points natively but will throw on invalid code points.

### Notable fixtures

**Consecutive astral symbols** (`tests/tests.js:141-144`): Code points `[127829, 119808, 119558, 119638]` map to four surrogate pairs: `'🍕𝐀𝌆𝍖'`. U+1F355 (SLICE OF PIZZA), U+1D400, U+1D306, U+1D356 — all above U+FFFF, requiring surrogate pairs in UCS-2.

**U+D800 (high surrogate) followed by non-surrogates** (`tests/tests.js:146-149`): `decoded: [55296, 97, 98]` / `encoded: '\uD800ab'`. The lone high surrogate 0xD800 is followed by ordinary ASCII characters `a` and `b`. `ucs2decode` will read 0xD800, find that `0x61` (`a`) is not a low surrogate (since `(0x61 & 0xFC00) != 0xDC00`), decrement `counter` (`punycode.js:116`), and push 0xD800 as a standalone code point.

**U+DC00 (low surrogate) followed by non-surrogates** (`tests/tests.js:151-154`): `decoded: [56320, 97, 98]` / `encoded: '\uDC00ab'`. A lone low surrogate is never the start of a pair-detection branch (the `if` at `punycode.js:107` only triggers for high surrogates), so 0xDC00 is pushed directly via `punycode.js:119`.

**High surrogate followed by another high surrogate** (`tests/tests.js:156-159`): `decoded: [0xD800, 0xD800]`. The first 0xD800 is high, the second is also high (not a low surrogate), so both are emitted individually as unmatched surrogates.

**Unmatched high surrogate, surrogate pair, unmatched high surrogate** (`tests/tests.js:161-164`): `decoded: [0xD800, 0x1D306, 0xD800]` / `encoded: '\uD800𝌆\uD800'`. Interleaves an unmatched high surrogate before and after a valid pair (U+1D306, encoded as `𝌆`). The valid pair is recombined to 0x1D306 by `punycode.js:111`.

**Low surrogate followed by another low surrogate** (`tests/tests.js:166-169`): `decoded: [0xDC00, 0xDC00]`. Both pass through unchanged.

**Unmatched low surrogate, surrogate pair, unmatched low surrogate** (`tests/tests.js:171-174`): `decoded: [0xDC00, 0x1D306, 0xDC00]`. Mirrors the high-surrogate case above.

---

## Array 3: `domains` (`tests/tests.js:176-220`)

### Purpose

`domains` tests `punycode.toUnicode` and `punycode.toASCII` at the full domain-name level, including the `xn--` ACE prefix, label splitting at dots, and the email-address special case. Each fixture is a complete domain name (or email address) rather than a bare label.

### Field semantics

| Field | Meaning |
|---|---|
| `decoded` | The Unicode (human-readable) form of the domain, e.g. `'mañana.com'` (`tests/tests.js:178`). |
| `encoded` | The full ACE (ASCII Compatible Encoding) form with `xn--` prefixes on internationalised labels, e.g. `'xn--maana-pta.com'` (`tests/tests.js:179`). Labels that are already ASCII pass through unchanged. |

`toASCII` prepends `'xn--'` only when `regexNonASCII` matches (`punycode.js:410-412`). `toUnicode` strips the prefix and lowercases before decoding when `regexPunycode` matches (`punycode.js:391-394`).

### Test suite consumers

- `describe('punycode.toUnicode', …)` (`tests/tests.js:323-344`): primary loop (`tests/tests.js:324-331`) passes `object.encoded` and expects `object.decoded`.
- `describe('punycode.toASCII', …)` (`tests/tests.js:346-371`): primary loop (`tests/tests.js:347-354`) passes `object.decoded` and expects `object.encoded`.

### Notable fixtures

**Trailing-dot domain** (`tests/tests.js:181-184`): `decoded: 'example.com.'` / `encoded: 'example.com.'`. A GitHub issue reference (`https://github.com/mathiasbynens/punycode.js/issues/17`) at `tests/tests.js:181` explains the motivation. `mapDomain` splits on `'.'` (`punycode.js:83`), so a trailing dot produces an empty final label; `regexNonASCII` does not match the empty string, so it is returned unchanged. The round-trip output therefore retains the trailing dot.

**Emoji domain** (`tests/tests.js:202-205`): `'💩.la'` → `'xn--ls8h.la'`. The label is the PILE OF POO emoji (U+1F4A9), an astral code point above U+FFFF. The encoder processes it via the surrogate-pair path in `ucs2decode` (`punycode.js:107-116`).

**Non-printable ASCII** (`tests/tests.js:207-210`): `decoded: '\0\x01\x02foo.bar'` / `encoded: '\0\x01\x02foo.bar'`. The label `'\0\x01\x02foo'` consists entirely of characters below 0x80 so `regexNonASCII` (`punycode.js:18`) does not match, and the string passes through `toASCII` unmodified. This fixture documents that the library does not validate label content beyond the ASCII/non-ASCII boundary.

**U+007F DEL** (`tests/tests.js:216-219`): `decoded: 'foo\x7F.example'` / `encoded: 'foo\x7F.example'`. The inline comment references `https://github.com/mathiasbynens/punycode.js/pull/115`. U+007F is below 0x80 and thus passes the `regexNonASCII` test unchanged. Notably, `regexNonASCII` is defined as `/[^\0-\x7F]/` (`punycode.js:18`), which explicitly excludes U+007F (i.e. DEL is in the ASCII range `\0-\x7F`), so the label is treated as already-ASCII.

**Email address** (`tests/tests.js:212-215`): The local part `'джумла'` (Cyrillic) is preserved verbatim before the `@`; only the domain portion is Punycoded: `'xn--p-8sbkgc5ag7bhce.xn--ba-lmcq'`. The `mapDomain` function at `punycode.js:72-86` splits on `'@'` first (`punycode.js:73-79`) and only processes `parts[1]` through the encoding callback, leaving `parts[0] + '@'` intact in `result`.

---

## Array 4: `separators` (`tests/tests.js:221-242`)

### Purpose

`separators` verifies that `toASCII` normalises all four IDNA2003 full-stop separator code points to a plain ASCII dot (`U+002E`) before encoding. This tests backwards compatibility with label boundaries defined in RFC 3490.

### Field semantics

| Field | Meaning |
|---|---|
| `decoded` | A Unicode domain string where the dot separator is one of the four recognised separator characters. |
| `encoded` | The canonical ACE form using only ASCII `'.'` (U+002E) as the separator. |

All four fixtures share the same base domain (`mañana` + separator + `com`) and the same expected output (`'xn--maana-pta.com'`), varying only the separator character used in `decoded`.

### Test suite consumers

- `describe('punycode.toASCII', …)` (`tests/tests.js:346-371`): the third loop (`tests/tests.js:363-370`) iterates over `testData.separators` and asserts `punycode.toASCII(object.decoded) === object.encoded`.

### Separator code points

The four fixtures correspond exactly to the four characters in `regexSeparators` at `punycode.js:19`:

```
/[\x2E。．｡]/g  // RFC 3490 separators
```

| Fixture | Separator used in `decoded` | Code point | Name |
|---|---|---|---|
| `tests/tests.js:222-226` | `\x2E` | U+002E | FULL STOP (standard ASCII dot) |
| `tests/tests.js:227-231` | `。` | U+3002 | IDEOGRAPHIC FULL STOP |
| `tests/tests.js:232-236` | `．` | U+FF0E | FULLWIDTH FULL STOP |
| `tests/tests.js:237-241` | `｡` | U+FF61 | HALFWIDTH IDEOGRAPHIC FULL STOP |

The normalisation is performed inside `mapDomain` at `punycode.js:82`:

```js
domain = domain.replace(regexSeparators, '\x2E');
```

This replaces every occurrence of any of the four separator characters with a plain `'.'` before splitting into labels, so all four inputs produce identical label arrays and therefore identical ACE output.
