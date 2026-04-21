# Spec: `punycode.toUnicode` — `describe('punycode.toUnicode')` block

**Source:** `tests/tests.js:323-344`

---

## 1. Subject

`punycode.toUnicode(input: string): string`

High-level domain/email converter. Splits the input on label boundaries (`.` and the three additional Unicode dot variants), inspects each label, decodes any label that carries the `xn--` ACE prefix, and leaves all other labels unchanged. The function is idempotent: calling it on a string that is already Unicode returns that string unmodified, because none of its labels will match `xn--`.

---

## 2. Contract

From the JSDoc at `punycode.js:378-388`:

- Converts a Punycode string representing a domain name or an email address to Unicode.
- Only the Punycoded parts of the input are converted.
- It does not matter if the function is called on a string that has already been converted to Unicode.
- Parameter `input` (`String`): the Punycoded domain name or email address to convert.
- Returns (`String`): the Unicode representation of the given Punycode string.

---

## 3. Test Cases

### 3.1 Domain vector loop (`tests/tests.js:324-331`)

Iterates `testData.domains` (`tests/tests.js:176-220`). For each entry, the test calls `punycode.toUnicode(object.encoded)` and asserts strict deep equality against `object.decoded` (`tests/tests.js:326-329`).

The ten domain vectors are:

| # | `encoded` (input) | `decoded` (expected output) | Notes |
|---|-------------------|-----------------------------|-------|
| 1 | `xn--maana-pta.com` | `mañana.com` | `tests/tests.js:177-180` |
| 2 | `example.com.` | `example.com.` | Trailing dot preserved; no label starts with `xn--`, so nothing is decoded. References GitHub issue #17 via comment (`tests/tests.js:181-184`). |
| 3 | `xn--bcher-kva.com` | `bücher.com` | `tests/tests.js:185-188` |
| 4 | `xn--caf-dma.com` | `café.com` | `tests/tests.js:189-192` |
| 5 | `xn----dqo34k.com` | `\u2603-\u2318.com` | `tests/tests.js:193-196` |
| 6 | `xn----dqo34kn65z.com` | `\uD400\u2603-\u2318.com` | `tests/tests.js:197-200` |
| 7 | `xn--ls8h.la` | Emoji: U+1F4A9 followed by `.la` | `tests/tests.js:201-205`; description field: `'Emoji'`. |
| 8 | `\0\x01\x02foo.bar` | `\0\x01\x02foo.bar` | Non-printable ASCII passthrough. No label begins with `xn--`, so the string is returned unchanged. `tests/tests.js:206-210`; description field: `'Non-printable ASCII'`. |
| 9 | `\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` | `\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a` | Email address: the local part before `@` is left untouched; each domain label carrying `xn--` is decoded independently. `tests/tests.js:211-215`; description field: `'Email address'`. |
| 10 | `foo\x7F.example` | `foo\x7F.example` | DEL character passthrough. No label matches `xn--`. References PR #115 via comment (`tests/tests.js:216-219`). |

### 3.2 Non-conversion loop (`tests/tests.js:332-343`)

Iterates `testData.strings` (`tests/tests.js:7-136`), which contains 23 entries covering individual Punycode-encoded strings (not full domain names). For each entry, the test makes two assertions (`tests/tests.js:334-341`):

1. `punycode.toUnicode(object.encoded) === object.encoded` — the raw encoded string (e.g. `'Bach-'`, `'tda'`, `'egbpdaj6bu4bxfgehfvwxn'`) is returned as-is, because none of these values begin with the `xn--` label prefix.
2. `punycode.toUnicode(object.decoded) === object.decoded` — the Unicode form (e.g. `'Bach'`, `'\xFC'`) is likewise returned unchanged, since it also carries no `xn--` prefix.

This contrasts with `punycode.decode`, which would transform the `encoded` values into their corresponding `decoded` forms. `toUnicode` is label-aware: it splits input on `.` and only acts on labels that match `regexPunycode` (`punycode.js:17`). Bare Punycode strings without the `xn--` prefix are therefore treated as opaque pass-through text.

---

## 4. Implementation Citations

- **`toUnicode`** — `punycode.js:389-395`. Delegates entirely to `mapDomain` with an inline callback. The callback tests each label against `regexPunycode` (`punycode.js:17`); matching labels have their four-character `xn--` prefix removed via `.slice(4)` and are decoded with `decode(string.slice(4).toLowerCase())` (`punycode.js:392`); non-matching labels are returned unchanged.

- **`mapDomain`** — `punycode.js:72-86`. Splits off any `@` prefix first (`punycode.js:73-80`), preserving the local part in `result`. Then replaces all Unicode separator variants with U+002E via `regexSeparators` (`punycode.js:82`), splits on `.` to produce individual labels (`punycode.js:83`), maps the callback over all labels (`punycode.js:84`), and rejoins with `.` before prepending the preserved local-part prefix.

- **`regexPunycode`** — `/^xn--/` at `punycode.js:17`. Controls which labels are decoded. Only labels whose first four characters are exactly `xn--` (lowercase) match.

- **`regexSeparators`** — `/[\x2E\u3002\uFF0E\uFF61]/g` at `punycode.js:19`. Used in `mapDomain` at `punycode.js:82` to normalize all four RFC 3490 dot variants to U+002E before splitting.

- **`decode`** — `punycode.js:196-281`. Called on the post-prefix remainder (`string.slice(4).toLowerCase()`) for each matching label.

- **Public binding** — `punycode.js:440`. `toUnicode` is exposed as the `'toUnicode'` property on the exported object.

---

## 5. Porting Notes

- **`toLowerCase()` before decode** (`punycode.js:392`). The base-36 alphabet used by `decode` is already case-insensitive, but the `xn--` prefix detection via `regexPunycode` (`punycode.js:17`) is case-sensitive (lowercase only). A caller passing `XN--foo` would not have the prefix detected, and the label would be returned unchanged. The `.toLowerCase()` call applies to the post-prefix text only, not to the prefix itself. The test suite does not exercise uppercase `XN--` input.

- **Separator normalization** (`punycode.js:19`, `punycode.js:82`). All four Unicode dot variants (U+002E, U+3002, U+FF0E, U+FF61) are mapped to U+002E before splitting into labels. This mechanism is shared with `toASCII`. The `toUnicode` test suite does not exercise alternative separators directly (those vectors reside in `testData.separators` at `tests/tests.js:221-242`, which is used only by `toASCII` tests). Ports must apply the same normalization step.

- **Trailing dot / empty final label** (`tests/tests.js:181-184`). The input `'example.com.'` splits into `['example', 'com', '']`. The empty string does not match `regexPunycode`, so it is returned as-is and the trailing dot is preserved in the rejoined output. Implementations must not special-case or strip trailing dots before label splitting.

- **Email handling** (`tests/tests.js:211-215`). `mapDomain` detects `@` and routes only the domain portion through label processing (`punycode.js:73-80`). The local part (everything before `@`) is concatenated back verbatim. Ports must preserve this separation.
