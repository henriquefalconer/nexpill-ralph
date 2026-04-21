# Spec: `punycode.toASCII` — `describe('punycode.toASCII')` block

Source: `tests/tests.js:346-371`

---

## 1. Subject

**Signature:** `punycode.toASCII(input: string): string`

`toASCII` is a high-level domain/email converter. It accepts a Unicode string representing a domain name or email address and returns the equivalent ASCII-Compatible Encoding (ACE) form. Each label (dot-separated segment) that contains at least one non-ASCII character is Punycode-encoded and prefixed with `xn--`. Labels that are already pure ASCII are left unchanged. The function is defined at `punycode.js:408-414` and exported at `punycode.js:439`.

---

## 2. Contract

Derived from the JSDoc at `punycode.js:397-407`:

- **Input:** A Unicode string representing a domain name or email address (`punycode.js:403-404`).
- **Output:** The Punycode representation of the given domain name or email address (`punycode.js:405-406`).
- **Idempotent on ASCII input:** Because the function only converts non-ASCII parts, calling it on a string that is already in ASCII produces the same string unchanged (`punycode.js:399-401`).
- **Email-aware:** When the input contains `@`, only the domain part (to the right of `@`) is subject to conversion; the local part is preserved verbatim (`punycode.js:75-79` via `mapDomain`).
- **IDNA2003 separator normalization:** Before splitting into labels, all four RFC 3490 separator characters — U+002E, U+3002, U+FF0E, U+FF61 — are normalized to U+002E (`punycode.js:82`), so that any of those characters acts as a label boundary.

---

## 3. Test Cases

### 3.1 Domain Conversion Loop (`tests/tests.js:347-354`)

Iterates `testData.domains` (`tests/tests.js:176-220`). For each entry, calls `punycode.toASCII(object.decoded)` and asserts strict equality with `object.encoded`. There are 10 entries.

| # | `decoded` | `encoded` | Notes | Source |
|---|-----------|-----------|-------|--------|
| 1 | `mañana.com` | `xn--maana-pta.com` | `ñ` (U+00F1) triggers encoding of the first label. | `tests/tests.js:177-180` |
| 2 | `example.com.` | `example.com.` | Trailing dot preserved; all labels are ASCII, so no conversion occurs. | `tests/tests.js:181-184` |
| 3 | `bücher.com` | `xn--bcher-kva.com` | `ü` (U+00FC) in the first label. | `tests/tests.js:185-188` |
| 4 | `café.com` | `xn--caf-dma.com` | `é` (U+00E9) in the first label. | `tests/tests.js:189-192` |
| 5 | `☃-⌘.com` | `xn----dqo34k.com` | Two non-ASCII code points (U+2603, U+2318) in the first label. | `tests/tests.js:193-196` |
| 6 | `\uD400☃-⌘.com` | `xn----dqo34kn65z.com` | Lone high surrogate U+D400 combined with U+2603 and U+2318. | `tests/tests.js:197-200` |
| 7 | `\uD83D\uDCA9.la` (pile of poo emoji, U+1F4A9) | `xn--ls8h.la` | Astral-plane emoji encoded as a surrogate pair in UCS-2; `ucs2decode` reassembles it before encoding. | `tests/tests.js:201-205` |
| 8 | `\0\x01\x02foo.bar` | `\0\x01\x02foo.bar` | All characters are in the range U+0000-U+007F; `regexNonASCII` does not match; the label is returned unchanged. | `tests/tests.js:206-210` |
| 9 | `\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a` | `\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` | Email address: the local part (Cyrillic characters before `@`) is preserved verbatim; only the two domain labels to the right of `@` are Punycode-encoded. | `tests/tests.js:211-215` |
| 10 | `foo\x7F.example` | `foo\x7F.example` | DEL (U+007F) is within the range `\0-\x7F` matched by `regexNonASCII = /[^\0-\x7F]/` (`punycode.js:18`), so it does NOT trigger encoding; the label passes through unchanged. See porting note below. | `tests/tests.js:216-219` |

### 3.2 ASCII Passthrough Loop (`tests/tests.js:355-362`)

Iterates `testData.strings` (`tests/tests.js:7-136`), which contains 23 entries. For each, calls `punycode.toASCII(object.encoded)` and asserts that the result equals `object.encoded` — that is, the already-encoded (ASCII-only) string is returned unchanged.

Because every `encoded` value in `testData.strings` is composed exclusively of characters in the range U+0000-U+007F, `regexNonASCII` (`punycode.js:18`) never matches any label, and the inner callback at `punycode.js:409-413` always returns the label as-is via the `else` branch (`punycode.js:412`). This loop covers 23 iterations across the RFC 3492 sample strings and other ASCII fixtures.

### 3.3 IDNA2003 Separator Loop (`tests/tests.js:363-370`)

Iterates `testData.separators` (`tests/tests.js:221-242`). For each, calls `punycode.toASCII(object.decoded)` and asserts equality with `object.encoded`. All four entries decode the domain `mañana<separator>com` and expect the encoded form `xn--maana-pta.com`:

| # | Separator | Code point | `decoded` | `encoded` | Source |
|---|-----------|------------|-----------|-----------|--------|
| 1 | Full stop | U+002E | `mañana.com` | `xn--maana-pta.com` | `tests/tests.js:222-226` |
| 2 | Ideographic full stop | U+3002 | `mañana\u3002com` | `xn--maana-pta.com` | `tests/tests.js:227-231` |
| 3 | Fullwidth full stop | U+FF0E | `mañana\uFF0Ecom` | `xn--maana-pta.com` | `tests/tests.js:232-236` |
| 4 | Halfwidth ideographic full stop | U+FF61 | `mañana\uFF61com` | `xn--maana-pta.com` | `tests/tests.js:237-241` |

All four separators are recognized by `regexSeparators = /[\x2E\u3002\uFF0E\uFF61]/g` at `punycode.js:19`. Before label splitting, `mapDomain` replaces every match with U+002E (`punycode.js:82`), so the subsequent `domain.split('.')` at `punycode.js:83` correctly yields two labels in every case.

---

## 4. Implementation Citations

- **`toASCII` body:** `punycode.js:408-414`. The function delegates entirely to `mapDomain`, passing a callback that tests each label with `regexNonASCII` and, if the test passes, returns `'xn--' + encode(label)`.
- **`mapDomain`:** `punycode.js:72-86`.
  - Splits the input on `@` to isolate the local part of an email address (`punycode.js:73-80`).
  - Replaces all RFC 3490 separators with U+002E via `regexSeparators` (`punycode.js:82`).
  - Splits the (now normalized) domain on `.` and maps each label through the callback (`punycode.js:83-84`).
  - Rejoins with `.` and prepends any preserved local part (`punycode.js:84-85`).
- **`regexNonASCII`:** `punycode.js:18` — `/[^\0-\x7F]/`. Matches any character NOT in the range U+0000 through U+007F. This is the sole gating condition that determines whether a label is encoded.
- **`regexSeparators`:** `punycode.js:19` — `/[\x2E\u3002\uFF0E\uFF61]/g`. Covers exactly the four separator characters tested in the IDNA2003 separator loop.
- **`encode`:** `punycode.js:290-376`. Called by the `toASCII` callback when a label fails the ASCII test; produces the raw Punycode token that is then prefixed with `xn--`.
- **Binding in public API:** `punycode.js:439`.

---

## 5. Porting Notes

- **DEL character (U+007F) is treated as ASCII.** `regexNonASCII = /[^\0-\x7F]/` at `punycode.js:18` matches characters strictly outside the range U+0000-U+007F. U+007F (DEL) is the upper boundary of that range and therefore falls inside it, so it does NOT match. The inline comment at `punycode.js:18` — "U+007F DEL is excluded too" — confirms this is intentional. As a result, the vector `foo\x7F.example` → `foo\x7F.example` round-trips unchanged (`tests/tests.js:216-219`). A port that uses a regex such as `/[^\x00-\x7E]/` or a code-point threshold of `> 127` would misclassify DEL and produce incorrect output for this case.

- **Pure-ASCII labels are never re-encoded.** The callback at `punycode.js:410-412` tests each label independently. If a label is already ASCII — even if it happens to begin with `xn--` — it is returned unchanged. The implementation does not strip or re-process existing ACE prefixes in the `toASCII` direction. The ASCII passthrough loop at `tests/tests.js:355-362` exercises this property across 23 inputs.

- **IDNA2003 separator normalization is required.** A naive implementation that splits only on U+002E (`.`) would fail the separator loop at `tests/tests.js:363-370`. The `regexSeparators` replacement at `punycode.js:82` must precede the `split('.')` call at `punycode.js:83`. The four separators covered are exactly those defined in RFC 3490 and encoded in `regexSeparators` at `punycode.js:19`.

- **Email handling requires splitting on `@` before label processing.** Without the `@`-split logic in `mapDomain` (`punycode.js:73-80`), the local part of an email address would be treated as domain labels and erroneously Punycode-encoded. The email test case at `tests/tests.js:211-215` verifies that the Cyrillic local part is preserved verbatim and only the domain labels are transformed.
