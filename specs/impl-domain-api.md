# Domain-Level Functions & Public API (`punycode.js:389-443`)

The two public domain/email converters and the exported `punycode` object. Both
converters delegate label handling to `mapDomain`
([impl-helpers.md](impl-helpers.md)) and are idempotent.

## `toUnicode(input)` (`punycode.js:389-395`)

Converts a Punycode domain name or email address to Unicode; non-Punycoded parts
pass through, so re-running on already-Unicode input is a no-op
(`punycode.js:378-388`).

- `mapDomain(input, callback)` (`punycode.js:390`).
- Per label (`punycode.js:391-393`): if `regexPunycode.test(string)` (matches
  `^xn--`, `punycode.js:17`), return `decode(string.slice(4).toLowerCase())` —
  strip the 4-char `xn--` prefix, lowercase, and [`decode`](impl-decode.md).
  Otherwise return the label unchanged.

## `toASCII(input)` (`punycode.js:408-414`)

Converts a Unicode domain name or email address to Punycode; ASCII-only parts
pass through, so it is idempotent on already-ASCII input (`punycode.js:397-407`).

- `mapDomain(input, callback)` (`punycode.js:409`).
- Per label (`punycode.js:410-412`): if `regexNonASCII.test(string)` (matches any
  char ≥ U+0080, `punycode.js:18`), return `'xn--' + encode(string)` — the only
  place the `xn--` prefix is added ([`encode`](impl-encode.md) returns the bare
  label). Otherwise return the label unchanged.

## Public API object `punycode` (`punycode.js:419-441`)

| Member | Value / target | Line |
|---|---|---|
| `version` | `'2.3.1'` | `punycode.js:425` |
| `ucs2.decode` | `ucs2decode` | `punycode.js:434` |
| `ucs2.encode` | `ucs2encode` | `punycode.js:435` |
| `decode` | `decode` | `punycode.js:437` |
| `encode` | `encode` | `punycode.js:438` |
| `toASCII` | `toASCII` | `punycode.js:439` |
| `toUnicode` | `toUnicode` | `punycode.js:440` |

`version` mirrors `package.json:3` (`"2.3.1"`). The `ucs2` members are documented
at `punycode.js:426-432`.

## Export (`punycode.js:443`)

`module.exports = punycode` — the object above is the sole CommonJS export. The
build script [`scripts/prepublish.js`](impl-prepublish.md) rewrites exactly this
line to produce the ES6 variant.
