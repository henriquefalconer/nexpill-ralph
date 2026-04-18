# Test-Data Fixtures — Specification

This file documents the shared test vectors defined in `tests/tests.js` as the top-level `testData` object (`tests/tests.js:6-243`). Every function-level spec in this directory references these vectors. The Go port at `port/` must preserve every vector verbatim — see `TARGET.md`.

## Purpose

`testData` is a single object literal with four named buckets (`tests/tests.js:6-243`), each holding an array of `{description?, decoded, encoded}` records. The Mocha `describe` blocks iterate these arrays to assert input/output pairs for every public function. Citations below point at the line where each vector is defined.

## Bucket: `strings` (tests/tests.js:7-136)

Label-level Punycode vectors. Each record has a `decoded` Unicode string and an `encoded` Punycode string (no `xn--` prefix). Used by `punycode.decode` (`tests/tests.js:290-310`) and `punycode.encode` (`tests/tests.js:312-321`), and by the identity-passthrough branches of `punycode.toUnicode` (`tests/tests.js:332-343`) and `punycode.toASCII` (`tests/tests.js:355-362`). 25 entries.

- `Bach` ↔ `Bach-` — a single basic code point (`tests/tests.js:7-12`). Exercises the trailing-delimiter rule in `encode` (`punycode.js:318-320`).
- `\u00FC` ↔ `tda` — a single non-ASCII character (`tests/tests.js:13-17`). No basic prefix → no delimiter.
- `\u00FC\u00EB\u00E4\u00F6\u2665` ↔ `4can8av2009b` — multiple non-ASCII characters including a BMP symbol (`tests/tests.js:18-22`).
- `b\u00FCcher` ↔ `bcher-kva` — canonical ASCII+non-ASCII mix (`tests/tests.js:23-27`).
- Long German sentence with multiple umlauts ↔ `Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal` (`tests/tests.js:28-32`).
- RFC 3492 §7.1 canonical vectors, all cited in the test file:
  - Arabic (Egyptian) (`tests/tests.js:33-38`)
  - Chinese simplified (`tests/tests.js:39-43`)
  - Chinese traditional (`tests/tests.js:44-48`)
  - Czech (`tests/tests.js:49-53`)
  - Hebrew (`tests/tests.js:54-58`)
  - Hindi Devanagari (`tests/tests.js:59-63`)
  - Japanese kanji+hiragana (`tests/tests.js:64-68`)
  - Korean Hangul syllables (`tests/tests.js:69-73`)
  - Russian Cyrillic (`tests/tests.js:83-87`) — the explanatory comment at `tests/tests.js:74-82` documents that Punycode.js intentionally does NOT emit the mixed-case annotation from the RFC sample; the expected output is all-lowercase `b1abfaaepdrnnbgefbadotcwatmq2g4l`. Implementation detail: `digitToBasic` is always called with `flag=0` at `punycode.js:358-364`.
  - Spanish (`tests/tests.js:88-92`)
  - Vietnamese (`tests/tests.js:93-97`)
- Six anonymous entries (no `description`) with Japanese / emoji-adjacent content (`tests/tests.js:98-125`). Because `description` is absent, the `it` title falls back to `object.encoded` (`tests/tests.js:292`, `:313`) or a hard-coded string for the identity loops.
- `-> $1.00 <-` ↔ `-> $1.00 <--` — an ASCII-only string that deliberately breaks hostname-label rules (`tests/tests.js:126-135`). Round-trips exactly. The trailing `--` is: one literal `-` plus the Bootstring basic/extended delimiter.

## Bucket: `ucs2` (tests/tests.js:137-175)

UCS-2 / UTF-16 surrogate handling vectors. Each record holds a `decoded` array of integer code points and an `encoded` string (the UTF-16 text representation). Used by `punycode.ucs2.decode` (`tests/tests.js:245-271`) and `punycode.ucs2.encode` (`tests/tests.js:273-288`). 7 entries.

- Consecutive astral (supplementary-plane) symbols: `[127829, 119808, 119558, 119638]` ↔ the four-emoji-equivalent UTF-16 string (`tests/tests.js:140-144`). Exercises surrogate-pair combination in `ucs2decode` (`punycode.js:110-111`) and auto-surrogate splitting in `ucs2encode` (`punycode.js:133`).
- U+D800 (unmatched high surrogate) followed by ASCII `ab`: `[55296, 97, 98]` ↔ `\uD800ab` (`tests/tests.js:145-149`). Exercises the unmatched-surrogate fallback at `punycode.js:112-117`.
- U+DC00 (unmatched low surrogate) followed by ASCII `ab`: `[56320, 97, 98]` ↔ `\uDC00ab` (`tests/tests.js:150-154`).
- Two consecutive high surrogates: `[0xD800, 0xD800]` ↔ `\uD800\uD800` (`tests/tests.js:155-159`).
- Unmatched high + valid surrogate pair + unmatched high: `[0xD800, 0x1D306, 0xD800]` ↔ `\uD800\uD834\uDF06\uD800` (`tests/tests.js:160-164`).
- Two consecutive low surrogates: `[0xDC00, 0xDC00]` ↔ `\uDC00\uDC00` (`tests/tests.js:165-169`).
- Unmatched low + valid surrogate pair + unmatched low: `[0xDC00, 0x1D306, 0xDC00]` ↔ `\uDC00\uD834\uDF06\uDC00` (`tests/tests.js:170-174`).

In addition, the `ucs2.encode` describe block asserts input-array immutability: encoding `[0x61, 0x62, 0x63]` yields `"abc"` and leaves the source array unchanged (`tests/tests.js:282-287`). The `ucs2.decode` describe block also houses two error-path tests that target `punycode.decode` (not `ucs2.decode`): the `not-basic` RangeError on `'\x81-'` (`tests/tests.js:255-262`, impl `punycode.js:215-216`) and the `overflow` RangeError on `'\x81'` (`tests/tests.js:263-270`, impl `punycode.js:243-244`).

## Bucket: `domains` (tests/tests.js:176-220)

Domain- and email-level vectors. Each record has a `decoded` full domain/email and an `encoded` ACE form. Used by `punycode.toUnicode` (`tests/tests.js:323-344`) and `punycode.toASCII` (`tests/tests.js:346-371`). 10 entries.

- `ma\u00F1ana.com` ↔ `xn--maana-pta.com` (`tests/tests.js:177-180`).
- `example.com.` ↔ `example.com.` — trailing dot preserved as an empty last label (`tests/tests.js:181-184`). Hat-tips GitHub issue #17 (`tests/tests.js:181`).
- `b\u00FCcher.com` ↔ `xn--bcher-kva.com` (`tests/tests.js:185-188`).
- `caf\u00E9.com` ↔ `xn--caf-dma.com` (`tests/tests.js:189-192`).
- `\u2603-\u2318.com` ↔ `xn----dqo34k.com` — snowman and Place-of-Interest sign (`tests/tests.js:193-196`).
- `\uD400\u2603-\u2318.com` ↔ `xn----dqo34kn65z.com` — adds a supplementary-plane-adjacent Hangul syllable (`tests/tests.js:197-200`).
- `\uD83D\uDCA9.la` ↔ `xn--ls8h.la` — emoji (`tests/tests.js:201-205`).
- `\0\x01\x02foo.bar` ↔ same — non-printable ASCII passes through unchanged because `regexNonASCII` excludes 0x00-0x7F (`tests/tests.js:206-210`, regex at `punycode.js:18`).
- `\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a` ↔ `\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` — email; only the domain part is ACE-encoded, local part stays Unicode (`tests/tests.js:211-215`). Email handling at `punycode.js:72-80`.
- `foo\x7F.example` ↔ same — DEL (0x7F) is treated as ASCII by `regexNonASCII` per the explicit comment at `punycode.js:18`; fixture cites PR #115 (`tests/tests.js:216-219`).

## Bucket: `separators` (tests/tests.js:221-242)

IDNA2003 backward-compatibility separator equivalence. Each record encodes `ma\u00F1ana<SEP>com` with a different label separator character; all must produce the same ACE output `xn--maana-pta.com`. Used only by `punycode.toASCII` (`tests/tests.js:363-370`). 4 entries.

- U+002E `.` (FULL STOP) (`tests/tests.js:222-226`).
- U+3002 `\u3002` (IDEOGRAPHIC FULL STOP) (`tests/tests.js:227-231`).
- U+FF0E `\uFF0E` (FULLWIDTH FULL STOP) (`tests/tests.js:232-236`).
- U+FF61 `\uFF61` (HALFWIDTH IDEOGRAPHIC FULL STOP) (`tests/tests.js:237-241`).

All four are recognised by `regexSeparators` at `punycode.js:19` and normalised to U+002E by `mapDomain` at `punycode.js:82` before the domain is split into labels.

## Cross-reference: which bucket is consumed by which spec

| Bucket | Consumer describe block(s) | Spec file |
|---|---|---|
| `strings` | `punycode.decode` (`tests/tests.js:290-310`), `punycode.encode` (`tests/tests.js:312-321`), `punycode.toUnicode` identity loop (`tests/tests.js:332-343`), `punycode.toASCII` identity loop (`tests/tests.js:355-362`) | `specs/punycode-decode.md`, `specs/punycode-encode.md`, `specs/punycode-to-unicode.md`, `specs/punycode-to-ascii.md` |
| `ucs2` | `punycode.ucs2.decode` (`tests/tests.js:245-271`), `punycode.ucs2.encode` (`tests/tests.js:273-288`) | `specs/punycode-ucs2-decode.md`, `specs/punycode-ucs2-encode.md` |
| `domains` | `punycode.toUnicode` (`tests/tests.js:323-331`), `punycode.toASCII` (`tests/tests.js:347-354`) | `specs/punycode-to-unicode.md`, `specs/punycode-to-ascii.md` |
| `separators` | `punycode.toASCII` only (`tests/tests.js:363-370`) | `specs/punycode-to-ascii.md` |

## Fidelity requirement for the Go port

Every record in every bucket must be reproduced in the Go test suite with identical `decoded` and `encoded` values. `TARGET.md` states: *"every RFC test vector from `tests/tests.js` must pass in Go."*

Special care is required for:
- **Lone-surrogate fixtures in `ucs2`** (`tests/tests.js:145-174`): these values are not valid Unicode scalars and cannot be represented in a Go `string` (which is UTF-8). A faithful port must use `[]uint16` or an explicit UTF-16 code-unit representation to reproduce them.
- **Trailing-dot domain** (`tests/tests.js:181-184`): preserve empty labels when splitting on `.`.
- **Email local part** (`tests/tests.js:211-215`): never punycode the local part, even if it is Unicode.
- **Uppercase-Z digit** (`tests/tests.js:299-301`): case-insensitive digit decoding when decoding (`punycode.js:144-155`), but always-lowercase digit emission when encoding (`punycode.js:358-364`).
