# punycode.toUnicode — Specification

## Purpose
`toUnicode` is a domain-level converter that transforms a Punycoded domain name or email address into its Unicode representation. It is idempotent: input that is already Unicode (or any string containing labels that do not start with the ACE prefix `xn--`) passes through unchanged. Only labels starting with `xn--` are decoded via the label-level Punycode `decode` primitive; all other labels — including plain ASCII, Unicode, or mixed labels — are copied verbatim to the output. See `punycode.js:378-395`.

## Contract
- Input: a domain name or email address, expressed as a string. The input may contain Unicode labels, ASCII labels, already-decoded labels, or Punycoded labels (prefix `xn--`), in any combination.
- Output: the same domain or email, with every label that starts with `xn--` replaced by its Unicode decoding. Labels that do not start with `xn--` are preserved byte-for-byte.
- Email handling: the input is split on `@` once; the portion before the first `@` (the local part) is preserved verbatim together with the `@` sign, and only the portion after the `@` is treated as the domain and processed label-by-label. See `punycode.js:72-80`.
- Label separators: the regex at `punycode.js:19` treats the four code points U+002E (`.`), U+3002 (IDEOGRAPHIC FULL STOP), U+FF0E (FULLWIDTH FULL STOP), and U+FF61 (HALFWIDTH IDEOGRAPHIC FULL STOP) as equivalent separators. `mapDomain` normalizes all of them to U+002E before splitting into labels. See `punycode.js:82-83`. Note: the `toUnicode` tests at `tests/tests.js:323-344` only exercise U+002E; the alternative separators are exercised by the `toASCII` tests.
- Errors: `toUnicode` itself raises no errors. A malformed `xn--` label may cause the underlying `decode` routine to throw a `RangeError` of type `invalid-input`, `not-basic`, or `overflow` (see `punycode.js:22-26` and `punycode.js:196-281`). No test in the `toUnicode` describe block exercises a malformed label.

## Algorithm (high level)
1. If the input contains an `@`, split it on `@` and treat the first piece plus a trailing `@` as a literal prefix, while the second piece becomes the domain to process. Otherwise, the whole input is the domain and the prefix is empty. See `punycode.js:72-80`.
2. In the domain, replace every separator character (U+002E, U+3002, U+FF0E, U+FF61) with U+002E. See `punycode.js:82`.
3. Split the domain on U+002E into an ordered list of labels. See `punycode.js:83`.
4. For every label, test whether it starts with the literal four-character ACE prefix `xn--` (case-sensitive — the regex at `punycode.js:17` is anchored and lowercase). If so, strip the four-character prefix, lowercase the remaining characters, and feed the result to the label-level `decode` routine. Otherwise, keep the label unchanged. See `punycode.js:389-395`.
5. Rejoin the decoded labels with U+002E and prepend the preserved email prefix (if any). See `punycode.js:84-85`.

## Behavior rules
- A simple Punycoded label is decoded and joined with its ASCII sibling: `xn--maana-pta.com` decodes to `mañana.com`. Fixture at `tests/tests.js:177-180`. Implementation at `punycode.js:389-395`.
- Trailing-dot domains round-trip unchanged: `example.com.` stays `example.com.`. Splitting on `.` yields the labels `example`, `com`, and an empty string; none starts with `xn--`, so each is copied verbatim, and the trailing empty label re-produces the trailing dot on rejoin. Fixture at `tests/tests.js:181-184`. Implementation at `punycode.js:83-84`.
- German umlaut: `xn--bcher-kva.com` decodes to `bücher.com`. Fixture at `tests/tests.js:185-188`.
- French accent: `xn--caf-dma.com` decodes to `café.com`. Fixture at `tests/tests.js:189-192`.
- Mixed ASCII and non-letter Unicode symbols within a single Punycoded label: `xn----dqo34k.com` decodes to `☃-⌘.com` (a snowman, ASCII hyphen, and command key). Fixture at `tests/tests.js:193-196`.
- Astral-plane (supplementary-plane) symbols are decoded correctly: `xn----dqo34kn65z.com` decodes to `U+D400☃-⌘.com`, where U+D400 is the starting code unit of a code point above U+FFFF. Fixture at `tests/tests.js:197-200`.
- Emoji labels work the same as any other non-ASCII label: `xn--ls8h.la` decodes to `U+1F4A9.la` (the pile-of-poo emoji followed by `.la`). Fixture at `tests/tests.js:201-205`.
- Non-printable ASCII control characters do not force any processing because the label does not start with `xn--`: `\0\x01\x02foo.bar` passes through byte-identical. Fixture at `tests/tests.js:206-210`. The regex gate at `punycode.js:391` fails and the identity branch at `punycode.js:392` returns the original label.
- Email addresses: the local part is preserved verbatim (including non-ASCII code points), and only the domain part after the `@` is decoded. Input `джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` becomes `джумла@джpумлатест.bрфa`. Fixture at `tests/tests.js:211-215`. Implementation at `punycode.js:72-80`.
- Labels containing DEL (U+007F) but not prefixed with `xn--` pass through unchanged: `foo\x7F.example` stays `foo\x7F.example`. This demonstrates that the `xn--` prefix check gates all other processing, and DEL is not treated as non-ASCII by the `regexPunycode` test. Fixture at `tests/tests.js:216-219`. Implementation at `punycode.js:391` (regex test against the full label).
- Idempotence over non-ACE strings: for every entry in `testData.strings` (`tests/tests.js:6-136`), the `toUnicode` test at `tests/tests.js:332-343` asserts that both `toUnicode(encoded)` returns `encoded` unchanged and `toUnicode(decoded)` returns `decoded` unchanged. None of those strings begins with `xn--`, so `mapDomain` splits them on `.` (often into a single label) and the identity branch at `punycode.js:392` returns each label unchanged. The hard-coded `it` description at `tests/tests.js:333` is reused for every such entry — even those without a `description` field in the fixture.

## Edge cases / invariants
- The `regexPunycode` test at `punycode.js:17` is case-sensitive. A label beginning with `XN--` (uppercase) would not be recognized as Punycode and would pass through unchanged. No fixture asserts this directly.
- The `xn--` detection runs against each individual label, not the full domain string, so a domain such as `foo.xn--bar.baz` would decode only the middle label while leaving `foo` and `baz` untouched. The label-level granularity is a direct consequence of splitting before matching at `punycode.js:83-84`.
- When the underlying `decode` is invoked, its input is first lowercased via `string.toLowerCase()` (`punycode.js:392`). This matches Punycode's case-insensitivity for the ACE payload.
- Email handling with multiple `@` characters is a subtle limitation. `split('@')` with more than one `@` produces more than two parts, but the implementation at `punycode.js:75-79` constructs the prefix as `parts[0] + '@'` and then assigns the domain as `parts[1]`, silently discarding any further `@`-separated pieces. A Go port should either document this behavior or make an explicit design choice; using a two-piece split (equivalent to a limit-2 split on `@`) is the faithful replication. No fixture exercises multi-`@` input.
- The domain splitting step always produces at least one label. An empty input string becomes a one-element list containing the empty string, which does not match `xn--` and is rejoined to yield the original empty string.

## Port notes (Go)
- Use `strings.SplitN(input, "@", 2)` to replicate the email split precisely — this yields either a one- or two-element slice and mirrors the JS `parts[0] + '@'` / `parts[1]` logic without surprise for multi-`@` strings.
- Normalize separators with a `strings.Replacer` or `strings.Map` that rewrites the four separator runes (U+002E, U+3002, U+FF0E, U+FF61) to U+002E before splitting on U+002E with `strings.Split`.
- Test for the ACE prefix with `strings.HasPrefix(label, "xn--")` (ASCII-only, case-sensitive) to mirror `regexPunycode`.
- For matching labels, strip the prefix with `label[4:]`, lowercase with `strings.ToLower`, and delegate to the Go port of the label-level `decode` primitive.
- Rejoin with `strings.Join(labels, ".")` and prepend the preserved email prefix.
- Preserve the exact test vectors — every fixture in the `domains` and `strings` tables must produce the documented output. See `TARGET.md`.

## Test-vector source
Domain vectors listed in `specs/test-data-fixtures.md` under the `domains` section; identity-pass-through vectors under `strings`; originals at `tests/tests.js:176-220` and `tests/tests.js:6-136`.
