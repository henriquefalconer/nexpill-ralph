# punycode.toASCII

## Purpose

Convert a Unicode domain name or email address into its ASCII Compatible Encoding (ACE) form suitable for DNS queries and storage. Labels containing any non-ASCII code point are Punycode-encoded and prefixed with `xn--`; ASCII-only labels pass through unchanged. This function implements the ToASCII operation of IDNA2003 (Internationalized Domain Names in Applications).

## Signature

**Input:** A string representing a domain name or email address, which may contain Unicode characters, multiple labels separated by dots, and optional email syntax (local-part@domain).

**Output:** An ASCII-safe string where non-ASCII labels have been encoded and prefixed with `xn--`, and ASCII labels are unchanged. Label separators are normalized to U+002E (full stop).

## Behavior

### Input parsing and separator normalization

If the input contains an `@` character, split on the **first** `@` only: the local part (everything before the `@`) is preserved verbatim and not processed further; only the domain part (after the `@`) undergoes label processing (punycode.js:72–80, tests.js:211–215).

All label separators in the domain are normalized to U+002E (full stop). The function recognizes four IDNA2003 separators:
- U+002E (full stop / period)
- U+3002 (ideographic full stop)
- U+FF0E (fullwidth full stop)
- U+FF61 (halfwidth ideographic full stop)

All occurrences are replaced with U+002E before further processing (punycode.js:19, 82).

### Label-by-label encoding

After normalization, the domain is split into labels on U+002E. Each label is independently processed:

**If the label contains any code unit outside the range U+0000–U+007F:** Apply the `punycode.encode` function (punycode.js:290–376) to the label and prepend the ASCII prefix `xn--` to the result (punycode.js:409–411). The detection regex at punycode.js:18 is `/[^\0-\x7F]/`, so every byte in 0x00–0x7F — including all C0 control codes and U+007F DEL — is treated as ASCII and does NOT trigger encoding.

**Otherwise (label is pure ASCII):** Leave the label unchanged (punycode.js:411–412).

### Reconstruction

Labels are rejoined with U+002E (punycode.js:84). If an email address was detected, prepend the original local part followed by `@` (punycode.js:75–80, 85).

## Quirks and Edge Cases

### Non-printable ASCII characters
Non-printable ASCII code points (U+0000–U+001F and U+007F DEL) are valid ASCII and are **not** encoded. A label like `\0\x01\x02foo` or `foo\x7F` remains unchanged because all bytes fall within the ASCII range (punycode.js:18; tests.js:206–210, 216–219).

### Trailing dots
An empty label (produced by a trailing dot in the input) is preserved in the output, maintaining FQDN syntax. For example, `example.com.` encodes to `example.com.` (tests.js:181–184).

### Emoji and surrogate pairs
Surrogate pairs and astral plane characters are handled by the underlying `punycode.encode` function. The domain `\uD83D\uDCA9.la` (pile of poo emoji) encodes to `xn--ls8h.la` (tests.js:201–205).

### Email addresses
The local part of an email address is never encoded, even if it contains non-ASCII characters. For example, `\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a` yields `\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq`, where the Cyrillic local part is preserved verbatim and the domain labels are encoded (tests.js:211–215; punycode.js:75–80).

### Idempotence
Punycode strings (those containing only ASCII) pass through unchanged because the regex check for non-ASCII characters returns false, so labels are not re-encoded (tests.js:355–362). A domain already in full ACE form (e.g., `xn--maana-pta.com`) remains unchanged because each label is pure ASCII.

### Separator normalization cascade
All four IDNA2003 separators normalize to `.` **before** label splitting, ensuring consistent encoding. The four test fixtures in `testData.separators` — using U+002E, U+3002, U+FF0E, and U+FF61 respectively — all produce the same encoded output `xn--maana-pta.com` (tests.js:221–241; punycode.js:19, 82).

## Examples

All examples are drawn from `testData.domains` and `testData.separators` in tests.js:

| Input (decoded) | Output (encoded) | Citation |
|---|---|---|
| `ma\xF1ana.com` | `xn--maana-pta.com` | tests.js:177–180 |
| `example.com.` | `example.com.` | tests.js:181–184 |
| `b\xFCcher.com` | `xn--bcher-kva.com` | tests.js:185–188 |
| `caf\xE9.com` | `xn--caf-dma.com` | tests.js:189–192 |
| `\u2603-\u2318.com` | `xn----dqo34k.com` | tests.js:193–196 |
| `\uD400\u2603-\u2318.com` | `xn----dqo34kn65z.com` | tests.js:197–200 |
| `\uD83D\uDCA9.la` (emoji) | `xn--ls8h.la` | tests.js:201–205 |
| `\0\x01\x02foo.bar` (non-printable ASCII) | `\0\x01\x02foo.bar` | tests.js:206–210 |
| `\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a` (email) | `\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` | tests.js:211–215 |
| `foo\x7F.example` (DEL as ASCII) | `foo\x7F.example` | tests.js:216–219 |
| `ma\xF1ana\x2Ecom` (U+002E separator) | `xn--maana-pta.com` | tests.js:222–225 |
| `ma\xF1ana\u3002com` (U+3002 separator) | `xn--maana-pta.com` | tests.js:227–230 |
| `ma\xF1ana\uFF0Ecom` (U+FF0E separator) | `xn--maana-pta.com` | tests.js:232–235 |
| `ma\xF1ana\uFF61com` (U+FF61 separator) | `xn--maana-pta.com` | tests.js:237–240 |

## Cross-References

- [test-fixtures.md](test-fixtures.md) — Complete test data definitions
- [punycode-encode.md](punycode-encode.md) — Bootstring encoding algorithm applied to individual labels
- [punycode-to-unicode.md](punycode-to-unicode.md) — Inverse operation (ACE to Unicode)

## Implementation Notes

The function is exported at punycode.js:439 and defined at punycode.js:408–414. It delegates label processing to the `mapDomain` helper (punycode.js:72–86), which handles email addresses and separator normalization. The non-ASCII detection regex at punycode.js:18 matches any code unit outside U+0000–U+007F, correctly including all control characters and DEL (U+007F) as ASCII-safe.
