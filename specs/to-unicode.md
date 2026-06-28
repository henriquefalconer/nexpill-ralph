# `punycode.toUnicode` — Behavior Specification

## Overview

`punycode.toUnicode(input)` converts a Punycode-encoded domain name or email address into its Unicode representation. Only labels that begin with the ACE prefix `xn--` are decoded; all other labels are returned unchanged. This means the function is safe to call on a string that has already been converted to Unicode — it is idempotent for non-Punycode input. (`punycode.js:389-395`)

---

## Mechanics

### Entry Point — `toUnicode` (`punycode.js:389-395`)

```js
const toUnicode = function(input) {
    return mapDomain(input, function(string) {
        return regexPunycode.test(string)
            ? decode(string.slice(4).toLowerCase())
            : string;
    });
};
```

The function delegates all structural parsing to `mapDomain`, passing a per-label callback. Inside the callback:

1. The label is tested against `regexPunycode` (`/^xn--/`, `punycode.js:17`). (`punycode.js:391`)
2. If the test passes, the 4-character prefix `xn--` is stripped via `.slice(4)`, the remainder is lowercased, and the result is passed to `decode`. (`punycode.js:392`)
3. If the test fails, the label is returned unchanged. (`punycode.js:393`)

### Domain Splitting — `mapDomain` (`punycode.js:72-86`)

`mapDomain(domain, callback)` handles both plain domain names and email addresses:

- It splits `domain` on `@` (`punycode.js:73`).
- If there is more than one part (i.e., an `@` is present), the local part (everything before `@`) is preserved as-is and the variable `domain` is reassigned to the portion after `@`. (`punycode.js:75-79`) This ensures only the domain portion of an email address is ever Punycode-decoded.
- All IDNA-recognised label separators — U+002E `.`, U+3002 `。`, U+FF0E `．`, U+FF61 `｡` — are normalized to ASCII `.` via a `replace` call against `regexSeparators` (`/[\x2E。．｡]/g`, `punycode.js:19`). (`punycode.js:82`)
- The normalized domain string is split on `.` to produce an array of labels. (`punycode.js:83`)
- Each label is passed to the callback via the private `map` utility (`punycode.js:53-60`), and the results are rejoined with `.`. (`punycode.js:84`)
- The local-part prefix (if any) is prepended and the final string is returned. (`punycode.js:85`)

### Private `map` utility (`punycode.js:53-60`)

A minimal `Array#map` replacement that iterates over an array in reverse order by index and collects callback return values into a new array of the same length. It is used internally so that `mapDomain` does not rely on `Array.prototype.map`.

### Core Decoder — `decode` (`punycode.js:196-281`)

`decode(input)` implements RFC 3492 Punycode decoding for a single label string (without the `xn--` prefix). It:

- Copies all basic (ASCII) code points appearing before the last delimiter `-` directly into the output array. (`punycode.js:208-219`)
- Runs the main generalized-variable-length-integer decoding loop to recover non-basic code points and insert them at computed positions. (`punycode.js:224-278`)
- Returns the final Unicode string via `String.fromCodePoint(...output)`. (`punycode.js:280`)

See [decode.md](decode.md) for the full specification of this function.

---

## Test Suite

### First Loop — Domain Fixtures (`tests/tests.js:323-331`)

```js
for (const object of testData.domains) {
    it(object.description || object.encoded, function() {
        assert.deepEqual(
            punycode.toUnicode(object.encoded),
            object.decoded
        );
    });
}
```

Each fixture in `testData.domains` (`tests/tests.js:176-220`) supplies an `encoded` value (the Punycode form) and a `decoded` value (the expected Unicode output). The assertion is `assert.deepEqual` (`tests/tests.js:326-329`). The fixtures and what each one exercises are:

| `encoded` (input) | `decoded` (expected) | What it exercises | Source lines |
|---|---|---|---|
| `xn--maana-pta.com` | `mañana.com` | Basic single-label IDNA decode; U+00F1 `ñ` recovery | `tests/tests.js:177-180` |
| `example.com.` | `example.com.` | Trailing-dot domain passthrough; label after trailing `.` is empty and has no `xn--` prefix, so it is returned unchanged | `tests/tests.js:181-184` |
| `xn--bcher-kva.com` | `bücher.com` | U+00FC `ü` in a mixed ASCII/non-ASCII label | `tests/tests.js:185-188` |
| `xn--caf-dma.com` | `café.com` | U+00E9 `é` at end of label | `tests/tests.js:189-192` |
| `xn----dqo34k.com` | `☃-⌘.com` (U+2603, U+2318) | Multi-symbol non-ASCII label with hyphen | `tests/tests.js:193-196` |
| `xn----dqo34kn65z.com` | `Ԁ☃-⌘.com` (U+D400, U+2603, U+2318) | Astral + BMP non-ASCII characters in a single label | `tests/tests.js:197-200` |
| `xn--ls8h.la` | `💩.la` (U+1F4A9, emoji) | Emoji domain; astral code point above U+FFFF decoded via surrogate pair U+D83D U+DCA9 | `tests/tests.js:201-205` |
| `\0\x01\x02foo.bar` | `\0\x01\x02foo.bar` | Non-printable ASCII characters in a label do not match `regexPunycode`; the label passes through unchanged | `tests/tests.js:206-210` |
| `джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` | `джумла@джумлатест.брфa` | Email address: the local part (`джумла`) is left intact; only the two domain labels (`xn--p-8sbkgc5ag7bhce` and `xn--ba-lmcq`) are decoded | `tests/tests.js:211-215` |
| `foo\x7F.example` | `foo\x7F.example` | U+007F DEL character in a label; no `xn--` prefix, so the label passes through unchanged | `tests/tests.js:216-219` |

### Second Loop — Strings Fixture Idempotency (`tests/tests.js:332-343`)

```js
for (const object of testData.strings) {
    it('does not convert names (or other strings) that don\'t start with `xn--`', function() {
        assert.deepEqual(
            punycode.toUnicode(object.encoded),
            object.encoded
        );
        assert.deepEqual(
            punycode.toUnicode(object.decoded),
            object.decoded
        );
    });
}
```

Every entry in `testData.strings` (`tests/tests.js:7-136`) is a raw Punycode label string (not a domain), covering scripts including Arabic, Chinese, Czech, Hebrew, Hindi, Japanese, Korean, Russian, Spanish, Vietnamese, and various mixed-ASCII strings. None of these strings begins with `xn--`. Consequently:

- Calling `toUnicode(object.encoded)` returns `object.encoded` unchanged, because no label matches `regexPunycode`. (`tests/tests.js:334-337`)
- Calling `toUnicode(object.decoded)` returns `object.decoded` unchanged for the same reason. (`tests/tests.js:338-341`)

Both assertions use `assert.deepEqual`. (`tests/tests.js:334`, `tests/tests.js:338`) This loop proves that `toUnicode` is a no-op on any input that contains no `xn--`-prefixed labels, establishing idempotency for already-decoded or non-IDNA strings.

---

## Assertion Mechanics

All assertions in this test suite use `assert.deepEqual` from Node's built-in `assert` module (imported at `tests/tests.js:3`). Because all values compared are primitive strings, `deepEqual` is equivalent to strict equality here. It is used uniformly throughout the `toUnicode`, `toASCII`, `decode`, and `encode` suites for consistency. Relevant call sites: `tests/tests.js:327` (domain loop) and `tests/tests.js:334`, `tests/tests.js:338` (strings loop).

---

## Cross-References

- **`punycode.decode`** ([decode.md](decode.md), `punycode.js:196-281`, tested at `tests/tests.js:290-310`): the core RFC 3492 decoder that `toUnicode` invokes after stripping the `xn--` prefix. Refer to its spec for the full decoding algorithm, basic code point handling, bias adaptation, and error conditions (`RangeError: Illegal input >= 0x80`, `RangeError: Overflow`, `RangeError: Invalid input`).
- **`punycode.toASCII`** ([to-ascii.md](to-ascii.md), `punycode.js:408-414`, tested at `tests/tests.js:346-371`): the inverse operation. It uses the same `mapDomain` infrastructure but applies `encode` to labels containing non-ASCII characters (detected via `regexNonASCII`, `/[^\0-\x7F]/`, `punycode.js:18`) and prepends `xn--`. It additionally supports IDNA2003 separator normalisation as verified by the `testData.separators` fixture (`tests/tests.js:363-370`).
- **`punycode.encode`** ([encode.md](encode.md), `punycode.js:290-376`, tested at `tests/tests.js:312-321`): the single-label encoder, symmetric counterpart to `decode`.
