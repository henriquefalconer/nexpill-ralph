# punycode.toASCII — Specification

## Purpose

`toASCII` is the domain-level converter that transforms a Unicode domain name or email address to its ASCII-Compatible Encoding (ACE) form. It operates label-by-label: every label that contains at least one non-ASCII code point is replaced with `xn--` concatenated with the Punycode-encoded form of that label, while pure-ASCII labels — including already-`xn--`-prefixed ACE labels — are passed through unchanged. The function is idempotent over pure-ASCII input: feeding already-encoded output back into it produces the same string. Email addresses are handled specially: only the domain half is processed, and the local part (before `@`) is kept verbatim regardless of its contents. See `punycode.js:397-414` for the function header and body, and `punycode.js:439` for the public API wiring.

## Contract

- Input: a single Unicode string representing a domain name (for example `bücher.com`) or an email address (for example `user@bücher.com`).
- Output: an ASCII string. Each non-ASCII label is replaced with `xn--<encode(label)>`; ASCII labels are returned unchanged; an optional `local@` prefix is prepended back to the result.
- "Non-ASCII" for the purpose of label selection is defined by the regex `/[^\0-\x7F]/` at `punycode.js:18`. The inline comment on that line explicitly notes "U+007F DEL is excluded too", i.e. DEL (0x7F) is treated as ASCII and does not trigger ACE encoding. The fixture `foo\x7F.example` at `tests/tests.js:216-219` confirms this: it passes through unchanged.
- Email handling: the input is split once on `@` (`punycode.js:73`); if there are two or more parts, the first part plus `@` is kept as a verbatim prefix and only the second part is treated as the domain to encode (`punycode.js:75-80`). When the local part itself contains non-ASCII characters — as with the Cyrillic `джумла` at `tests/tests.js:211-215` — those characters are preserved untouched; only the two labels of the domain half are encoded.
- Label separators: the four characters U+002E FULL STOP, U+3002 IDEOGRAPHIC FULL STOP, U+FF0E FULLWIDTH FULL STOP, and U+FF61 HALFWIDTH IDEOGRAPHIC FULL STOP are all recognized as label separators. They are normalized to U+002E before splitting, via `regexSeparators` at `punycode.js:19` applied at `punycode.js:82`.
- No error is raised for empty labels or for inputs without `@`; the function always returns a string. Errors originating from the underlying `encode` routine (overflow) propagate out unchanged.

## Algorithm (high level)

1. If the input contains `@`, split it into `local` and `domain`; remember `local@` as a verbatim prefix and continue with only the `domain` portion (`punycode.js:72-80`).
2. Replace every occurrence of any of the four recognized separator characters with U+002E (`punycode.js:82`).
3. Split the normalized domain string on U+002E to get an ordered list of labels (`punycode.js:83`).
4. For each label, test it against the non-ASCII regex at `punycode.js:18`. If the label contains any code point outside the range U+0000..U+007F, output `"xn--"` concatenated with the Punycode encoding of the label; otherwise output the label unchanged (`punycode.js:408-413`).
5. Re-join the resulting labels with U+002E (`punycode.js:84`) and prepend the remembered `local@` prefix if one was captured in step 1 (`punycode.js:85`).

The non-ASCII test at step 4 operates on the JavaScript UCS-2 string; the call into `encode` at `punycode.js:411` internally performs surrogate-pair decoding via `ucs2decode` at `punycode.js:294`, so astral-plane characters (including emoji) are handled correctly.

## Behavior rules

### A. Domain fixtures (tests.js:347-354)

The describe block iterates `testData.domains` and asserts `toASCII(decoded) === encoded` for each entry.

- `mañana.com` encodes to `xn--maana-pta.com`: the first label contains `ñ` (U+00F1), fails the ASCII regex, and is ACE-encoded; the second label `com` is pure ASCII and passes through (`tests/tests.js:177-180`, `punycode.js:410-411`).
- `example.com.` (trailing dot, pure ASCII throughout) round-trips unchanged. After separator normalization, splitting on `.` yields the labels `example`, `com`, and an empty string; all three are pure ASCII so all three pass through, and rejoining with `.` restores the trailing dot (`tests/tests.js:181-184`, `punycode.js:82-84`).
- `bücher.com` encodes to `xn--bcher-kva.com` (`tests/tests.js:185-188`, `punycode.js:410-411`).
- `café.com` encodes to `xn--caf-dma.com` (`tests/tests.js:189-192`, `punycode.js:410-411`).
- `☃-⌘.com` (U+2603 SNOWMAN, U+2318 PLACE OF INTEREST SIGN) encodes to `xn----dqo34k.com`. The double hyphen after `xn--` is the Punycode delimiter separating the (empty) basic-code-point prefix from the extended data (`tests/tests.js:193-196`).
- `\uD400☃-⌘.com` encodes to `xn----dqo34kn65z.com`. The leading U+D400 together with subsequent code points forms a valid astral scalar value only via surrogate-pair decoding inside `encode`, exercising the `ucs2decode` path at `punycode.js:294` (`tests/tests.js:197-200`).
- `💩.la` (U+1F4A9 PILE OF POO, represented in JavaScript source as the surrogate pair `\uD83D\uDCA9`) encodes to `xn--ls8h.la`. The surrogate pair is recombined into a single code point by `ucs2decode` before Punycode encoding (`tests/tests.js:201-205`, `punycode.js:294`).
- `\0\x01\x02foo.bar` round-trips unchanged. Although the leading bytes are non-printable, they are all below 0x80, so the first label fails the non-ASCII regex and passes through the identity branch (`tests/tests.js:206-210`, `punycode.js:410` regex does not match).
- `джумла@джpумлатест.bрфa` encodes to `джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq`. The split on `@` at `punycode.js:73` separates the Cyrillic local part from the domain; the local part is preserved as literal Unicode (the local part is never punycoded, even though it contains non-ASCII characters); only the two labels of the domain are each ACE-encoded (`tests/tests.js:211-215`, `punycode.js:72-80`).
- `foo\x7F.example` round-trips unchanged. U+007F DEL is below 0x80 and therefore matches the range `\0-\x7F` in the regex at `punycode.js:18`; the inline comment on that line confirms the intent. The first label is classified as ASCII and passes through (`tests/tests.js:216-219`, `punycode.js:18`, `punycode.js:410-412`).

### B. ASCII-passthrough fixtures (tests.js:355-362)

The same `describe` block then iterates `testData.strings` and asserts `toASCII(encoded) === encoded` — that is, each entry's `encoded` field survives a round trip through `toASCII` unchanged. This is the idempotence guarantee: already-ASCII output is not re-encoded. The invariant holds because every `encoded` value in `testData.strings` at `tests/tests.js:6-136` is pure ASCII by construction (each is the output of the Punycode basic-string emission loop at `punycode.js:304-320` plus base-36 digits), so the non-ASCII regex at `punycode.js:18` does not match any label and the identity branch at `punycode.js:412` fires for every label. Note that `testData.strings` entries are raw Punycode strings, not domain names; they contain no `.` separators, so most of them encode as a single label during the loop.

### C. Separator-equivalence fixtures (tests.js:363-370)

Finally the block iterates `testData.separators` and asserts that `mañana.com` written with each of the four alternative separators produces the same ACE output `xn--maana-pta.com`. The four separators exercised are U+002E (`tests/tests.js:222-226`), U+3002 IDEOGRAPHIC FULL STOP (`tests/tests.js:227-231`), U+FF0E FULLWIDTH FULL STOP (`tests/tests.js:232-236`), and U+FF61 HALFWIDTH IDEOGRAPHIC FULL STOP (`tests/tests.js:237-241`). All four map to the same ACE form because `regexSeparators` at `punycode.js:19` normalizes any of them to U+002E before label splitting (applied at `punycode.js:82`). This is the RFC 3490 / IDNA2003 "backward-compatibility separators" rule, as named by the test's own `it` description at `tests/tests.js:364`.

## Edge cases and invariants

- Empty labels (for example the final empty string produced by a trailing `.`) are pure-ASCII by definition — an empty string contains no non-ASCII characters — so they survive the identity branch at `punycode.js:412` and are rejoined unchanged. `example.com.` at `tests/tests.js:181-184` is the only fixture that exercises this path.
- Already-ACE input (a label that begins with `xn--`) is only safe from double-encoding when the label contains no non-ASCII characters. All `testData.strings` fixtures satisfy that condition, so the identity branch fires. A hypothetical label that is both `xn--`-prefixed and contains non-ASCII characters would still be encoded again — this is unsymmetric with `toUnicode`, which specifically detects the `xn--` prefix at `punycode.js:391`. No fixture exercises this asymmetry.
- Multi-`@` input: `split('@')` at `punycode.js:73` produces more than two parts in that case. The implementation uses only `parts[0]` as the local prefix and `parts[1]` as the domain (`punycode.js:78-79`); any additional `@`-delimited segments are silently discarded. Not exercised by any fixture.
- ASCII control characters in the 0x00..0x1F range pass through unchanged when not mixed with any byte at or above 0x80 within the same label; see `\0\x01\x02foo.bar` at `tests/tests.js:206-210`.
- U+007F DEL is explicitly ASCII-class for this regex and never triggers encoding; see the comment at `punycode.js:18` and the fixture at `tests/tests.js:216-219`.

## Port notes (Go)

- Iterate the input string once as runes. If the input contains `@`, split on the first `@` only and retain `local@` as a verbatim prefix; work on the remainder. See `TARGET.md` for the module layout under `port/`.
- Build a rune-level transformation that maps U+002E, U+3002, U+FF0E, and U+FF61 all to U+002E. A `strings.Map` over the four separator runes is sufficient; a regex is not needed.
- Split the normalized domain on U+002E into a slice of labels. Empty labels must be preserved so that round-tripping trailing-dot input works.
- For each label, test "any rune greater than 0x7F" by scanning runes (decoded from UTF-8). Remember that the JavaScript regex treats surrogate halves U+D800..U+DFFF as non-ASCII because their numeric values exceed 0x7F; Go's rune iteration decodes surrogate pairs into a single scalar value, and lone surrogates in UTF-8 input are invalid — align with the Go `encode` port's decision on lone surrogates. If any rune in the label is greater than 0x7F, prepend the four-byte ASCII literal `xn--` and delegate to the Go port of `encode`; otherwise emit the label verbatim.
- Join the resulting labels with `.` and prepend the captured `local@` prefix (if any) to produce the final output string.
- No error is raised by `toASCII` itself; only `encode` can return an overflow error, and it must be propagated.

## Test-vector source

Domain vectors listed in `specs/test-data-fixtures.md` under the `domains` section; ACE-passthrough vectors under `strings`; separator-equivalence vectors under `separators`. Originals at `tests/tests.js:176-242` and `tests/tests.js:6-136`.
