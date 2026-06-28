# `punycode.toASCII` — Behavior Specification

## Overview

`punycode.toASCII(input)` converts a Unicode domain name or email address to its ACE (ASCII-Compatible Encoding) form as specified by IDNA (Internationalized Domain Names in Applications). Only labels that contain at least one non-ASCII character are transformed; each such label is Punycode-encoded and prefixed with `xn--`. Labels that are already pure ASCII are returned unchanged. The function is idempotent on already-ASCII input. (`punycode.js:408-414`)

```js
// punycode.js:408-414
const toASCII = function(input) {
    return mapDomain(input, function(string) {
        return regexNonASCII.test(string)
            ? 'xn--' + encode(string)
            : string;
    });
};
```

---

## Mechanics

### Step 1 — Email address splitting (`mapDomain`, `punycode.js:73-80`)

`toASCII` delegates all label-splitting and reassembly to the private helper `mapDomain(domain, callback)` (`punycode.js:72-86`). `mapDomain` first calls `domain.split('@')` (`punycode.js:73`). If the result has more than one part the local-part (everything before `@`) is preserved verbatim and prepended to the final result, while only the domain portion (everything after `@`) continues through the pipeline (`punycode.js:75-79`). This ensures that a Unicode username such as `джумла` in `джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` is never Punycode-encoded.

### Step 2 — Separator normalization (`regexSeparators`, `punycode.js:19`, `82`)

The constant `regexSeparators` is defined as:

```js
// punycode.js:19
const regexSeparators = /[\x2E。．｡]/g; // RFC 3490 separators
```

It matches all four label-separator code points recognized by RFC 3490:

| Code point | Character | Name |
|---|---|---|
| U+002E | `.` | FULL STOP |
| U+3002 | `。` | IDEOGRAPHIC FULL STOP |
| U+FF0E | `．` | FULLWIDTH FULL STOP |
| U+FF61 | `｡` | HALFWIDTH IDEOGRAPHIC FULL STOP |

After the email split, `mapDomain` replaces every occurrence of any of these four code points with a plain ASCII full stop `\x2E` (`punycode.js:82`). This normalization step is what makes IDNA2003 separator backwards compatibility possible (see the third test loop below).

### Step 3 — Label splitting and mapping (`punycode.js:83-84`)

The normalized domain string is split on `.` into an array of labels (`punycode.js:83`). The private `map` utility applies the `toASCII` callback to every label and the results are rejoined with `.` (`punycode.js:84`).

### Step 4 — Per-label ASCII test (`regexNonASCII`, `punycode.js:18`, `410-412`)

The constant `regexNonASCII` is defined as:

```js
// punycode.js:18
const regexNonASCII = /[^\0-\x7F]/; // Note: U+007F DEL is excluded too.
```

The character class `[^\0-\x7F]` matches any code unit whose value is greater than U+007F. Critically, U+007F (DEL) is **not** matched because the range `\0-\x7F` includes it. This is the reason a label such as `foo\x7F` (containing the DEL character) passes through unchanged rather than being Punycode-encoded (`punycode.js:18`).

For each label the callback tests `regexNonASCII.test(string)` (`punycode.js:410`):

- **Match (non-ASCII present):** The label is passed to `encode(string)` and the result is prefixed with `'xn--'` (`punycode.js:411`). See [encode.md](encode.md) for the full Bootstring algorithm.
- **No match (pure ASCII):** The label is returned unchanged (`punycode.js:412`).

---

## Test Suite: `describe('punycode.toASCII', ...)` (`tests/tests.js:346-371`)

The suite contains three independent `for` loops, each targeting a different fixture array.

---

### First loop — Unicode-to-ACE domain conversion (`tests/tests.js:347-354`)

```js
// tests/tests.js:347-354
for (const object of testData.domains) {
    it(object.description || object.decoded, function() {
        assert.deepEqual(
            punycode.toASCII(object.decoded),
            object.encoded
        );
    });
}
```

Each fixture in `testData.domains` (`tests/tests.js:176-220`) provides a `decoded` Unicode domain and an `encoded` ACE domain. The assertion (`assert.deepEqual`, `tests/tests.js:349`) verifies that `toASCII(decoded) === encoded`.

**Key fixtures and their behavior:**

| Fixture location | `decoded` | `encoded` | Notes |
|---|---|---|---|
| `tests/tests.js:178-179` | `mañana.com` | `xn--maana-pta.com` | Basic non-ASCII label encoding; `com` is ASCII so it passes through |
| `tests/tests.js:182-183` | `example.com.` | `example.com.` | Trailing dot (empty label) is preserved intact |
| `tests/tests.js:186-187` | `bücher.com` | `xn--bcher-kva.com` | German umlaut |
| `tests/tests.js:189-190` | `café.com` | `xn--caf-dma.com` | Accented ASCII letter |
| `tests/tests.js:193-194` | `☃-⌘.com` | `xn----dqo34k.com` | Non-BMP/BMP symbols |
| `tests/tests.js:202-204` | `💩.la` (Emoji, U+1F4A9) | `xn--ls8h.la` | Astral-plane emoji encoded via surrogate pair `💩`; `la` is ASCII |
| `tests/tests.js:207-209` | `\0\x01\x02foo.bar` | `\0\x01\x02foo.bar` | Non-printable control characters U+0000–U+0002 are within `\0-\x7F`; `regexNonASCII` does not match; entire domain passes through unchanged |
| `tests/tests.js:212-214` | `джумла@джрумлатест.bрфa` | `джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` | Email address: local part `джумла` (before `@`) is preserved verbatim; only domain labels are encoded |
| `tests/tests.js:217-218` | `foo\x7F.example` | `foo\x7F.example` | U+007F DEL falls within the `\0-\x7F` range in `regexNonASCII`; the label is not matched and passes through as-is |

---

### Second loop — ASCII passthrough / idempotence (`tests/tests.js:355-362`)

```js
// tests/tests.js:355-362
for (const object of testData.strings) {
    it('does not convert domain names (or other strings) that are already in ASCII', function() {
        assert.deepEqual(
            punycode.toASCII(object.encoded),
            object.encoded
        );
    });
}
```

This loop iterates over `testData.strings` (`tests/tests.js:7-136`), which are fixtures designed primarily for the low-level `encode`/`decode` API. Each fixture's `encoded` field is an already-ASCII Punycode string (e.g., `'Bach-'`, `'tda'`, `'egbpdaj6bu4bxfgehfvwxn'`). When fed to `toASCII`, none of these strings contain characters that match `regexNonASCII`; therefore `toASCII` returns every label unchanged. The assertion (`assert.deepEqual`, `tests/tests.js:357`) verifies that `toASCII(encoded) === encoded`, confirming the function is a no-op on already-ASCII input.

---

### Third loop — IDNA2003 separator backwards compatibility (`tests/tests.js:363-370`)

```js
// tests/tests.js:363-370
for (const object of testData.separators) {
    it('supports IDNA2003 separators for backwards compatibility', function() {
        assert.deepEqual(
            punycode.toASCII(object.decoded),
            object.encoded
        );
    });
}
```

Each fixture in `testData.separators` (`tests/tests.js:221-242`) encodes the same logical domain `mañana<sep>com` using a different one of the four RFC 3490 separator code points. All four must produce the canonical output `xn--maana-pta.com`.

| Fixture location | `decoded` | Separator used | `encoded` |
|---|---|---|---|
| `tests/tests.js:223-225` | `mañana\x2Ecom` | U+002E FULL STOP | `xn--maana-pta.com` |
| `tests/tests.js:228-230` | `mañana。com` | U+3002 IDEOGRAPHIC FULL STOP | `xn--maana-pta.com` |
| `tests/tests.js:233-235` | `mañana．com` | U+FF0E FULLWIDTH FULL STOP | `xn--maana-pta.com` |
| `tests/tests.js:238-240` | `mañana｡com` | U+FF61 HALFWIDTH IDEOGRAPHIC FULL STOP | `xn--maana-pta.com` |

All four cases succeed because `regexSeparators` (`punycode.js:19`) matches all four code points and `mapDomain` replaces each with `\x2E` (`punycode.js:82`) before splitting into labels. After normalization the pipeline is identical for all four inputs. The assertion (`assert.deepEqual`, `tests/tests.js:365`) confirms the result in each case.

---

## Assertion Mechanics

All three loops use `assert.deepEqual` (Node.js built-in `assert` module, imported at `tests/tests.js:3`). For string values `assert.deepEqual` performs strict value equality, making it equivalent to `assert.strictEqual` for these tests. The relevant call sites are:

- First loop: `tests/tests.js:349`
- Second loop: `tests/tests.js:357`
- Third loop: `tests/tests.js:365`

---

## Cross-References

- **`encode` spec** ([encode.md](encode.md)) — `toASCII` calls `encode(string)` (`punycode.js:411`) for every label that contains non-ASCII characters. The `encode` function (`punycode.js:290-376`) implements the RFC 3492 Bootstring algorithm: it emits all basic (ASCII) code points first, appends a delimiter `-` if any basic code points were present, then encodes each non-basic code point using a generalized variable-length integer scheme driven by the `adapt` bias function. The `encode` spec should be consulted for the full details of that transformation.
- **`toUnicode` spec** ([to-unicode.md](to-unicode.md)) — `toUnicode` (`punycode.js:389-395`) is the inverse direction: it calls `mapDomain` with a callback that strips the `xn--` prefix from any label matching `regexPunycode` and passes the remainder to `decode`. A round-trip `toUnicode(toASCII(input))` should yield the original Unicode domain for any well-formed input.
