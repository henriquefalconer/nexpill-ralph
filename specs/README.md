# Specs — Punycode behaviour, grounded in tests

This directory is the language-agnostic specification of `punycode.js`, derived from `tests/tests.js` with every claim cited back to either the test file or the reference implementation at `punycode.js`. It is the source-of-truth for the Go port at `port/` (see `TARGET.md`).

All files live at the top level here — the directory is intentionally flat. Each spec maps 1:1 to a Mocha `describe` block in `tests/tests.js`.

## Index

| Spec | Function under test | Test source (`tests/tests.js`) | Implementation (`punycode.js`) |
|---|---|---|---|
| [punycode-ucs2-decode.md](./punycode-ucs2-decode.md) | `punycode.ucs2.decode` | `245-271` | `101-123`, `433-436` |
| [punycode-ucs2-encode.md](./punycode-ucs2-encode.md) | `punycode.ucs2.encode` | `273-288` | `133`, `433-436` |
| [punycode-decode.md](./punycode-decode.md) | `punycode.decode` | `290-310` | `196-281`, `437` |
| [punycode-encode.md](./punycode-encode.md) | `punycode.encode` | `312-321` | `290-376`, `438` |
| [punycode-to-unicode.md](./punycode-to-unicode.md) | `punycode.toUnicode` | `323-344` | `389-395`, `440` |
| [punycode-to-ascii.md](./punycode-to-ascii.md) | `punycode.toASCII` | `346-371` | `408-414`, `439` |
| [test-data-fixtures.md](./test-data-fixtures.md) | Shared fixture tables (`strings`, `ucs2`, `domains`, `separators`) | `6-243` | — |

## How the specs are written

- **Language-agnostic prose.** Specs describe behaviour, not JS syntax, so a Go (or any other) engineer can re-implement without reading the source.
- **Every fact is cited.** Claims carry a `path:line` or `path:line-line` reference into `tests/tests.js` or `punycode.js`. Readers can `Read` the cited location to verify.
- **Fixtures live in one place.** `test-data-fixtures.md` catalogues every `testData` record once; the function specs point back to it instead of re-listing vectors.
- **Error paths are documented where the tests live.** The two `RangeError` assertions at `tests/tests.js:255-270` appear inside the `punycode.ucs2.decode` describe block even though they target `punycode.decode`; that quirk is called out in both specs.

## Reading order for the Go port

1. `test-data-fixtures.md` — learn the vector shape and the fidelity requirement.
2. `punycode-ucs2-decode.md` + `punycode-ucs2-encode.md` — the UCS-2/UTF-16 primitives, most sensitive to language choice (Go strings are UTF-8).
3. `punycode-decode.md` + `punycode-encode.md` — the Bootstring core per RFC 3492.
4. `punycode-to-unicode.md` + `punycode-to-ascii.md` — the domain/email wrappers, including IDNA2003 separator normalisation.

## Scope boundaries

- The specs cover only what `tests/tests.js` exercises. Uncovered code paths (e.g. `regexPunycode` case-sensitivity, multi-`@` emails) are noted as untested edge cases where relevant but not tested-against.
- The `version` string (`punycode.js:425`) is not under test and is not specified here.
