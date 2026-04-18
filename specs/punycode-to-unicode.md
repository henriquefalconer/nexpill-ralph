# punycode.toUnicode

## Purpose

Convert a Punycode-encoded domain name or email address to its Unicode representation, decoding only the labels prefixed with `xn--` (ACE prefix) and leaving other labels untouched.

## Signature

**Input:** a domain name or email address as a string.

**Output:** the same string with Punycode (IDNA) labels decoded to Unicode.

## Behavior

### Email Address Handling

The function accepts both plain domain names and email addresses. If the input contains `@`, the function splits on the first `@` character: the local part (everything before `@`) is preserved verbatim, and only the domain part (everything after `@`) undergoes Punycode decoding. The result is rejoined as `local@processedDomain` (punycode.js:72-80).

### Label Separator Normalization

Before processing, the function normalizes label separators in the domain part. Any occurrence of U+002E (full stop `.`), U+3002 (ideographic full stop), U+FF0E (fullwidth full stop), or U+FF61 (halfwidth ideographic full stop) is replaced with U+002E, the standard label separator (punycode.js:19, 82). This normalization follows IDNA2003 (tests.js:221-242 defines the separator fixtures).

### Label Splitting

The normalized domain is split on U+002E (`.`) into individual labels (punycode.js:83).

### Per-Label Decoding

For each label, the function applies the following logic (punycode.js:391-393):

- If the label matches the regular expression `^xn--` (case-sensitive), the `xn--` prefix is stripped, the remainder is converted to lowercase, and the result is passed to `punycode.decode` to recover the original Unicode string.
- If the label does not match `^xn--`, it is returned unchanged.

### Label Rejoin

The processed labels are rejoined with U+002E (`.`) to reconstruct the domain (punycode.js:84).

## Quirks and Edge Cases

### Trailing Dot

A domain with a trailing dot (e.g., `example.com.`) is handled correctly: the empty label at the end is preserved, as the trailing dot is converted to a U+002E separator and splits into an empty label (tests.js:181-184).

### Already-Unicode Input

When a label does not contain `xn--`, it passes through unchanged. This means the function is idempotent on Unicode input: applying it twice produces the same result as applying it once (tests.js:332-342).

### Already-ASCII Non-IDNA Input

ASCII labels that do not begin with `xn--` are passed through unchanged. This allows the function to safely process mixed-case ASCII domains without data loss (tests.js:332-342).

### Non-Printable ASCII Labels

Labels containing non-printable ASCII characters (U+0000 through U+001F, and other characters outside the printable range) are passed through as-is, because such sequences do not match the `^xn--` pattern (tests.js:206-210; example: `\0\x01\x02foo.bar`).

### Email Local Part Not Decoded

The local part of an email address (the part before `@`) is never Punycode-decoded, even if it contains sequences that resemble Punycode payloads. Only the domain part is processed. This is enforced by the email splitting logic in `mapDomain` (punycode.js:72-80; example: tests.js:211-215).

### Lowercase Conversion

After stripping the `xn--` prefix, the remainder of the label is converted to lowercase before decoding. This ensures that uppercase ASCII characters within a Punycode payload are handled correctly (punycode.js:392; related to uppercase payload test in `punycode-decode.md`).

## Examples

All examples are from `testData.domains` (tests.js:175-219):

### ma\xF1ana.com

- **Encoded:** `xn--maana-pta.com`
- **Decoded:** `ma\xF1ana.com`
- **Notes:** Single Punycode label containing the Spanish ñ character (tests.js:177-180).

### b\xFCcher.com

- **Encoded:** `xn--bcher-kva.com`
- **Decoded:** `b\xFCcher.com`
- **Notes:** Punycode label with German umlaut (tests.js:185-187).

### caf\xE9.com

- **Encoded:** `xn--caf-dma.com`
- **Decoded:** `caf\xE9.com`
- **Notes:** Single Punycode label with accented é character (tests.js:188-191).

### \u2603-\u2318.com (Snowman and Place of Interest Sign)

- **Encoded:** `xn----dqo34k.com`
- **Decoded:** `\u2603-\u2318.com`
- **Notes:** Punycode label containing only symbols and a literal hyphen; mixed case-sensitivity across labels (tests.js:192-195).

### \uD400\u2603-\u2318.com

- **Encoded:** `xn----dqo34kn65z.com`
- **Decoded:** `\uD400\u2603-\u2318.com`
- **Notes:** Multi-character Punycode label including surrogate pairs and symbols (tests.js:197-200).

### Emoji domain (\uD83D\uDCA9.la)

- **Encoded:** `xn--ls8h.la`
- **Decoded:** `\uD83D\uDCA9.la`
- **Notes:** Punycode label containing an emoji character (pile of poo); second label `la` is passed through unchanged (tests.js:201-205).

### Non-Printable ASCII (\0\x01\x02foo.bar)

- **Encoded:** `\0\x01\x02foo.bar`
- **Decoded:** `\0\x01\x02foo.bar`
- **Notes:** Labels without `xn--` prefix are returned unchanged; non-printable characters are preserved (tests.js:206-210).

### Email Address

- **Encoded:** `\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq`
- **Decoded:** `\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a`
- **Notes:** Local part (Cyrillic characters before `@`) is preserved unchanged; only the domain part is decoded label-by-label (tests.js:211-215).

### foo\x7F.example

- **Encoded:** `foo\x7F.example`
- **Decoded:** `foo\x7F.example`
- **Notes:** Labels containing DEL (U+007F) without `xn--` are passed through unchanged (tests.js:215-218).

### example.com. (Trailing Dot)

- **Encoded:** `example.com.`
- **Decoded:** `example.com.`
- **Notes:** The trailing dot creates an empty label; both empty and non-empty labels are processed correctly (tests.js:181-184).

## Cross-References

- [test-fixtures.md](./test-fixtures.md) — complete test data for `testData.domains` and `testData.strings`
- [punycode-decode.md](./punycode-decode.md) — specification for the underlying `punycode.decode` function
- [punycode-to-ascii.md](./punycode-to-ascii.md) — specification for the inverse function `punycode.toASCII`
