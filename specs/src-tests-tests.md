# `tests/tests.js` — source-file specification

This spec is the source-driven view of the Mocha test suite. Every section cites the exact line range in `tests/tests.js` (372 lines) so a Go porter can verify each claim against the JS reference suite. The test-driven specs in this directory (one per Mocha `describe` block) describe *what each test asserts about the library*; this spec describes *what the file contains and how it is wired*.

Read this alongside:

- `specs/src-punycode.md` — source-file specification of the library under test.
- `specs/test-data-fixtures.md` — table of every fixture in the four `testData` buckets.
- `specs/punycode-decode.md`, `specs/punycode-encode.md`, `specs/punycode-to-unicode.md`, `specs/punycode-to-ascii.md`, `specs/punycode-ucs2-decode.md`, `specs/punycode-ucs2-encode.md` — per-`describe` behaviour.

## File overview

- **Path:** `tests/tests.js`.
- **Length:** 372 lines (`tests/tests.js:1-372`).
- **Language level:** Strict-mode CommonJS (`tests/tests.js:1`). No `module.exports` — the file is loaded as a Mocha test script, not as a module with exports.
- **Runtime dependencies:** Node's built-in `assert` module (`tests/tests.js:3`) and the library under test imported as `../punycode.js` (`tests/tests.js:4`).
- **Test framework:** Mocha (provides the global `describe` and `it` functions). The framework is not `require`d; it is supplied by the runner. `package.json` wires `mocha tests/tests.js` to the `test` script.
- **Shape:** one large fixture table (`tests/tests.js:6-243`) followed by six top-level `describe` blocks (`tests/tests.js:245-371`). No nested `describe`s; no `before`/`after` hooks.

## 1. Module preamble (`tests/tests.js:1-4`)

`'use strict';` at `tests/tests.js:1` (same semantics as `punycode.js:1`; see `specs/src-punycode.md` §1).

`tests/tests.js:3` requires Node's built-in `assert`. The suite uses two members: `assert.deepEqual` (value/structural equality, used for almost every assertion) and `assert.throws` (used only by the two RangeError tests at `tests/tests.js:255-270`). Because `assert.deepEqual` performs a coercive (`==`) comparison on leaves, all comparisons in this suite happen between values of the same type (string-vs-string or number-array-vs-number-array), so the coercion is not load-bearing and a port may use strict equality.

`tests/tests.js:4` loads `../punycode.js` — the library at the project root. The relative path is the only place this file binds to the library's filesystem location.

## 2. The `testData` fixture object (`tests/tests.js:6-243`)

A single top-level `const` named `testData` holds four fixture buckets. Each bucket is an array of objects; each object has a `decoded` value, an `encoded` value, and (usually) a `description`. The buckets are consumed by the `describe` blocks in §3 via `for...of` loops.

The full table (every fixture, with code-point breakdown) is documented in `specs/test-data-fixtures.md`. This section describes only the *structure* of each bucket as it appears in the file.

### 2.1 `testData.strings` (`tests/tests.js:7-136`)

Twenty-three label-level fixtures. Each entry has shape `{ description?, decoded, encoded }` where:

- `decoded` is a JavaScript string containing the Unicode label (may include non-BMP characters via `\uXXXX` surrogate pairs).
- `encoded` is the Punycode body — the result of `punycode.encode(decoded)` — without the `xn--` prefix.
- `description` is a human-readable label used as the Mocha test name.

The first five entries (`tests/tests.js:8-32`) are short synthetic vectors. Lines `tests/tests.js:33` and `tests/tests.js:74-82` carry inline comments that link to RFC 3492 §7.1 and to `mathiasbynens/punycode.js#3` respectively; the comment block at `tests/tests.js:74-82` documents that mixed-case annotation is intentionally NOT implemented (see `specs/src-punycode.md` §5.2 for the corresponding source-side note). Entries `tests/tests.js:34-97` are the eleven RFC 3492 §7.1 sample strings (Arabic, both Chinese variants, Czech, Hebrew, Hindi, Japanese, Korean, Russian, Spanish, Vietnamese). Entries `tests/tests.js:98-125` are six Japanese mix-in vectors without descriptions. The final entry at `tests/tests.js:131-135` is a deliberately-malformed-host-name fixture (`'-> $1.00 <-'`) labelled "ASCII string that breaks the existing rules for host-name labels".

Consumed by `describe('punycode.decode')` (`tests/tests.js:290-310`), `describe('punycode.encode')` (`tests/tests.js:312-321`), and the idempotence loops inside `describe('punycode.toUnicode')` (`tests/tests.js:332-343`) and `describe('punycode.toASCII')` (`tests/tests.js:355-362`).

### 2.2 `testData.ucs2` (`tests/tests.js:137-175`)

Seven fixtures targeting the UCS-2 / surrogate layer. Shape: `{ description, decoded, encoded }` where:

- `decoded` is an **array of integer code points** (NOT a string).
- `encoded` is a JavaScript string of UTF-16 code units.

The fixtures cover one consecutive-astral-symbols case (`tests/tests.js:140-144`) and six lone- or paired-surrogate edge cases (`tests/tests.js:145-174`). Hex literals (`0xD800`, `0x1D306`, etc.) and decimal literals (`55296`, `127829`) are mixed; both forms are pure integer notation and the values match exactly when read.

Consumed by `describe('punycode.ucs2.decode')` (`tests/tests.js:245-271`) and `describe('punycode.ucs2.encode')` (`tests/tests.js:273-288`).

### 2.3 `testData.domains` (`tests/tests.js:176-220`)

Ten domain-level fixtures. Shape: `{ description?, decoded, encoded }` where:

- `decoded` is the unicode-domain form (e.g. `'mañana.com'`).
- `encoded` is the ACE form with `xn--` labels where required (e.g. `'xn--maana-pta.com'`).

Three entries carry inline comments linking to upstream issues: `tests/tests.js:181` (issue #17, `example.com.`), `tests/tests.js:216` (PR #115, `foo\x7F.example`). Notable shapes covered:

- A trailing-dot domain that must round-trip unchanged (`tests/tests.js:181-184`).
- A multi-label all-Unicode-prefix domain (`tests/tests.js:197-200`).
- An emoji TLD-style fixture (`tests/tests.js:201-205`).
- An email address with a non-ASCII local part and non-ASCII domain (`tests/tests.js:211-215`).
- Two pure-ASCII fixtures with control characters that must pass through unchanged (`tests/tests.js:206-210`, `tests/tests.js:216-219`).

Consumed by `describe('punycode.toUnicode')` (`tests/tests.js:323-331`) and `describe('punycode.toASCII')` (`tests/tests.js:346-354`).

### 2.4 `testData.separators` (`tests/tests.js:221-242`)

Four fixtures, one per IDNA2003 label separator (U+002E, U+3002, U+FF0E, U+FF61). All four `decoded` values produce the **same** `encoded` value `xn--maana-pta.com`, exercising the normalisation performed by `mapDomain` via `regexSeparators` at `punycode.js:19, 82` (see `specs/src-punycode.md` §3.3).

Consumed only by the third loop inside `describe('punycode.toASCII')` (`tests/tests.js:363-369`).

## 3. The six `describe` blocks (`tests/tests.js:245-371`)

The file has exactly six top-level Mocha `describe` blocks, one per public API entry point. They appear in this order:

| # | Block | Lines | Bucket(s) consumed | Per-block spec |
|---|---|---|---|---|
| 1 | `punycode.ucs2.decode` | `tests/tests.js:245-271` | `ucs2` | `specs/punycode-ucs2-decode.md` |
| 2 | `punycode.ucs2.encode` | `tests/tests.js:273-288` | `ucs2` | `specs/punycode-ucs2-encode.md` |
| 3 | `punycode.decode` | `tests/tests.js:290-310` | `strings` | `specs/punycode-decode.md` |
| 4 | `punycode.encode` | `tests/tests.js:312-321` | `strings` | `specs/punycode-encode.md` |
| 5 | `punycode.toUnicode` | `tests/tests.js:323-344` | `domains`, `strings` | `specs/punycode-to-unicode.md` |
| 6 | `punycode.toASCII` | `tests/tests.js:346-371` | `domains`, `strings`, `separators` | `specs/punycode-to-ascii.md` |

Every block follows one of two patterns; both are described in §4. Per-fixture behavioural claims live in the per-block specs above — this section only inventories the file's structure.

### 3.1 `describe('punycode.ucs2.decode')` (`tests/tests.js:245-271`)

Loops `testData.ucs2` (`tests/tests.js:246-254`) using each fixture's `description` as the test name and asserting `punycode.ucs2.decode(encoded) deepEqual decoded`.

After the loop, two extra `it` cases assert `RangeError` is thrown:

- `tests/tests.js:255-262`: `assert.throws(() => punycode.decode('\x81-'), RangeError)`. **Quirk:** the test name says "Illegal input >= 0x80 (not a basic code point)" and the block is named `punycode.ucs2.decode`, but the function actually invoked is `punycode.decode` — this exercises the basic-prefix guard at `punycode.js:215-216`, NOT the UCS-2 decoder. See `specs/punycode-ucs2-decode.md` "Cross-cutting RangeError tests" for the full explanation.
- `tests/tests.js:263-270`: `assert.throws(() => punycode.decode('\x81'), RangeError)`. Same misplacement quirk; this one exercises overflow guard A at `punycode.js:243-245` (see `specs/src-punycode.md` §10).

A port should treat both assertions as tests of `decode`, not of `ucs2.decode`.

### 3.2 `describe('punycode.ucs2.encode')` (`tests/tests.js:273-288`)

Loops `testData.ucs2` (`tests/tests.js:274-281`) using each fixture's `description` as the test name and asserting `punycode.ucs2.encode(decoded) deepEqual encoded`.

After the loop, an inline non-loop block (`tests/tests.js:282-287`) defines `const codePoints = [0x61, 0x62, 0x63]` at module-evaluation time, calls `punycode.ucs2.encode(codePoints)`, then inside the `it` body asserts that (a) the result is `'abc'` and (b) `codePoints` is still `[0x61, 0x62, 0x63]` — i.e. that `ucs2.encode` does not mutate its input array. See `specs/punycode-ucs2-encode.md`. Notable: the call to `ucs2.encode` happens at module evaluation, not inside the `it` callback, so the assertion only covers post-call state; a port may verify either by re-reading the array in the test body.

### 3.3 `describe('punycode.decode')` (`tests/tests.js:290-310`)

Loops `testData.strings` (`tests/tests.js:291-298`). Test name is `description || encoded` (some entries lack a `description` and fall back to the `encoded` body).

Two extra `it` cases follow:

- `tests/tests.js:299-301`: `assert.deepEqual(punycode.decode('ZZZ'), '\u7BA5')` — exercises the case-insensitive uppercase branch of `basicToDigit` at `punycode.js:148-149` (see `specs/src-punycode.md` §5.1).
- `tests/tests.js:302-309`: `assert.throws(() => punycode.decode('ls8h='), RangeError)` — exercises the `'invalid-input'` branch at `punycode.js:240-242`.

### 3.4 `describe('punycode.encode')` (`tests/tests.js:312-321`)

Loops `testData.strings` (`tests/tests.js:313-320`). Test name is `description || decoded`. No extra cases. The simplest of the six blocks.

### 3.5 `describe('punycode.toUnicode')` (`tests/tests.js:323-344`)

Two consecutive loops:

- `tests/tests.js:324-331` iterates `testData.domains` and asserts `punycode.toUnicode(encoded) deepEqual decoded`.
- `tests/tests.js:332-343` iterates `testData.strings` and, for each, asserts both `punycode.toUnicode(encoded) deepEqual encoded` and `punycode.toUnicode(decoded) deepEqual decoded`. This is the "does not convert names that don't start with `xn--`" idempotence check.

**Quirk:** in the second loop, every iteration registers an `it` with the **same** name (`'does not convert names (or other strings) that don\'t start with \`xn--\`'`, `tests/tests.js:333`). Mocha tolerates duplicate test names within a `describe` and runs each registration as a separate test; the reporter output therefore shows the same line many times. A port may either generate distinct names (preferred) or replicate the duplicate-name behaviour.

### 3.6 `describe('punycode.toASCII')` (`tests/tests.js:346-371`)

Three consecutive loops:

- `tests/tests.js:347-354` iterates `testData.domains` and asserts `punycode.toASCII(decoded) deepEqual encoded`. Test name is `description || decoded`.
- `tests/tests.js:355-362` iterates `testData.strings` and asserts `punycode.toASCII(encoded) deepEqual encoded` (idempotence over already-ASCII input). All iterations share the same name `'does not convert domain names (or other strings) that are already in ASCII'` — same duplicate-name quirk as §3.5.
- `tests/tests.js:363-370` iterates `testData.separators` and asserts `punycode.toASCII(decoded) deepEqual encoded`. All iterations share the name `'supports IDNA2003 separators for backwards compatibility'`. This is the only consumer of `testData.separators`.

## 4. Helper patterns

Two patterns recur across the file; both are pure Mocha + `assert` idioms.

### 4.1 The `describe`/`it` loop pattern

Used in every fixture-driven block. Skeleton:

```
describe('<api>', function() {
    for (const object of testData.<bucket>) {
        it(object.description || object.<fallback>, function() {
            assert.deepEqual(<api>(object.<input>), object.<expected>);
        });
    }
    // optional extra it() cases follow
});
```

The loop runs at module evaluation (when Mocha loads the file) and registers one `it` per fixture. The `it` callbacks are invoked later, when Mocha runs the suite. Closure capture of `object` is correct because `for...of` with `const` (the form used at `tests/tests.js:246, 274, 291, 313, 324, 332, 347, 355, 363`) creates a fresh binding per iteration.

The fallback for the `it` name varies by block: §3.3 uses `encoded`; §3.4 and §3.6 first loop use `decoded`. See the per-block specs for the exact conventions.

### 4.2 The `assert.throws(fn, ErrorClass)` pattern

Used only at `tests/tests.js:255-270` and `tests/tests.js:302-309`. Each call passes (a) a thunk that invokes the library and (b) the `RangeError` constructor; `assert.throws` succeeds if the thunk throws an instance of that class (the message text is not asserted, only the type). The library always throws via `error()` at `punycode.js:35-43`, which constructs `RangeError(errors[type])` — so the error type is invariant across the three message kinds (`'overflow'`, `'not-basic'`, `'invalid-input'`).

A port should expose at least the type/class distinction these tests rely on; if it surfaces distinct error variants per message kind (recommended; see `specs/src-punycode.md` §6.5), the test translation should assert on the variant.

## 5. Cross-references to the implementation

| `tests/tests.js` block | Library entry point | `punycode.js` lines | Source-spec section |
|---|---|---|---|
| `punycode.ucs2.decode` (§3.1) | `ucs2decode` | `punycode.js:88-123` | `specs/src-punycode.md` §4.1 |
| `punycode.ucs2.encode` (§3.2) | `ucs2encode` | `punycode.js:125-133` | `specs/src-punycode.md` §4.2 |
| `punycode.decode` (§3.3) | `decode` | `punycode.js:189-281` | `specs/src-punycode.md` §6 |
| `punycode.encode` (§3.4) | `encode` | `punycode.js:283-376` | `specs/src-punycode.md` §7 |
| `punycode.toUnicode` (§3.5) | `toUnicode` | `punycode.js:378-395` | `specs/src-punycode.md` §8.1 |
| `punycode.toASCII` (§3.6) | `toASCII` | `punycode.js:397-414` | `specs/src-punycode.md` §8.2 |

The two `RangeError` tests in §3.1 reach `decode`, not `ucs2decode` (the misplacement quirk). The case-insensitive `'ZZZ'` test in §3.3 reaches `basicToDigit` at `punycode.js:148-149`. The IDNA2003-separator loop in §3.6 reaches `mapDomain` and `regexSeparators` at `punycode.js:19, 82`.

## 6. Test-coverage gaps

The suite covers every public entry point but leaves several library paths unexercised. The canonical list lives in `specs/src-punycode.md` §11; in summary:

1. **`regexPunycode` is case-sensitive** — no fixture covers an `XN--` (uppercase) prefix at `tests/tests.js:176-220`.
2. **U+007F DEL mixed with non-ASCII** — covered as DEL-only (`tests/tests.js:216-219`); the mixed case is absent.
3. **Multi-`@` email addresses** — `mapDomain` silently drops everything after the second `@`; no fixture in `testData.domains` (`tests/tests.js:176-220`) exercises this.
4. **`toASCII` of an `xn--` label that itself contains non-ASCII** — would re-encode into `xn--xn--…`; no fixture exercises this.
5. **Encode-side overflow guards D and E** at `punycode.js:337-339, 345-347` — no fixture in `testData.strings` (`tests/tests.js:7-136`) is large enough to trigger them.
6. **Decode-side overflow guards B and C** at `punycode.js:255-257, 268-270` — only guard A is reached by `tests/tests.js:263-270`.

A port may match these as-is (the pragmatic default) or add coverage. See `specs/src-punycode.md` §10-§11 and `ralph/todo.md` for the canonical scope decisions.

## 7. Within-file structural summary

```
preamble                     tests/tests.js:1-4
testData.strings             tests/tests.js:7-136     (consumed by §3.3, §3.4, §3.5, §3.6)
testData.ucs2                tests/tests.js:137-175   (consumed by §3.1, §3.2)
testData.domains             tests/tests.js:176-220   (consumed by §3.5, §3.6)
testData.separators          tests/tests.js:221-242   (consumed by §3.6)
describe punycode.ucs2.decode  tests/tests.js:245-271
describe punycode.ucs2.encode  tests/tests.js:273-288
describe punycode.decode       tests/tests.js:290-310
describe punycode.encode       tests/tests.js:312-321
describe punycode.toUnicode    tests/tests.js:323-344
describe punycode.toASCII      tests/tests.js:346-371
EOF                            tests/tests.js:372
```

Every `describe` is reachable via `mocha tests/tests.js`; every fixture in `testData` is reached by at least one `describe` (cross-checked via the consumed-by annotations above).

## 8. Port guidance (Go)

For the suite specifically, a Go porter should:

1. Translate `testData` into a single Go fixtures file (one struct per fixture, four package-level slices). Keep the literal `decoded`/`encoded` values byte-identical to `tests/tests.js:6-243`; in particular preserve the integer arrays in `testData.ucs2` as `[]rune` or `[]uint32`, not as Go strings.
2. Translate each `describe` block into a Go test function; translate each `it` into a `t.Run(name, ...)` subtest so the per-fixture granularity survives.
3. For the two misplaced `RangeError` tests at `tests/tests.js:255-270`, place the translated assertions under the `decode` test function (not `ucs2.decode`), and document the JS-side misplacement in a comment.
4. For the duplicate-name idempotence loops at `tests/tests.js:332-343` and `tests/tests.js:355-362`, generate distinct `t.Run` names (e.g. include the fixture index or the `encoded` value) so Go's `-run` filter remains usable.
5. Skip translating coverage-gap notes — a port test suite need not exercise paths that the JS suite leaves unexercised, unless the port chooses to harden them. Either choice is acceptable; document the choice.
