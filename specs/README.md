# Specs — Punycode behaviour, grounded in tests

This directory is the language-agnostic specification of `punycode.js`, derived from `tests/tests.js` (test-driven specs) and from the source files themselves (source-driven specs). Every claim is cited back to the test file or the reference implementation. It is the source-of-truth for the Go port at `port/` (see `TARGET.md`).

All files live at the top level here — the directory is intentionally flat. Test-driven specs map 1:1 to a Mocha `describe` block in `tests/tests.js`. Source-driven specs map 1:1 to a file under the project's source tree (there is no literal `src/` directory — the library file sits at project root).

## Test-driven index

One spec per Mocha `describe` block, plus a shared fixtures spec.

| Spec | Function under test | Test source (`tests/tests.js`) | Implementation (`punycode.js`) |
|---|---|---|---|
| [punycode-ucs2-decode.md](./punycode-ucs2-decode.md) | `punycode.ucs2.decode` | `245-271` | `101-123`, `433-436` |
| [punycode-ucs2-encode.md](./punycode-ucs2-encode.md) | `punycode.ucs2.encode` | `273-288` | `133`, `433-436` |
| [punycode-decode.md](./punycode-decode.md) | `punycode.decode` | `290-310` | `196-281`, `437` |
| [punycode-encode.md](./punycode-encode.md) | `punycode.encode` | `312-321` | `290-376`, `438` |
| [punycode-to-unicode.md](./punycode-to-unicode.md) | `punycode.toUnicode` | `323-344` | `389-395`, `440` |
| [punycode-to-ascii.md](./punycode-to-ascii.md) | `punycode.toASCII` | `346-371` | `408-414`, `439` |
| [test-data-fixtures.md](./test-data-fixtures.md) | Shared fixture tables (`strings`, `ucs2`, `domains`, `separators`) | `6-243` | — |

## Source-driven index

One spec per source file, studied line-by-line with citations back into the file.

| Spec | Source file | Lines | Role |
|---|---|---|---|
| [src-punycode.md](./src-punycode.md) | `punycode.js` | `1-443` | Library — constants, helpers, UCS-2 layer, Bootstring primitives, `decode`/`encode`, `toUnicode`/`toASCII`, public API object |
| [src-scripts-prepublish.md](./src-scripts-prepublish.md) | `scripts/prepublish.js` | `1-17` | Build step that regex-rewrites `punycode.js` into the generated ES-module variant `punycode.es6.js` |

## How the specs are written

- **Language-agnostic prose.** Specs describe behaviour, not JS syntax, so a Go (or any other) engineer can re-implement without reading the source.
- **Every fact is cited.** Claims carry a `path:line` or `path:line-line` reference into `tests/tests.js` or `punycode.js`. Readers can `Read` the cited location to verify.
- **Fixtures live in one place.** `test-data-fixtures.md` catalogues every `testData` record once; the function specs point back to it instead of re-listing vectors.
- **Error paths are documented where the tests live.** The two `RangeError` assertions at `tests/tests.js:255-270` appear inside the `punycode.ucs2.decode` describe block even though they target `punycode.decode`; that quirk is called out in both specs.
- **Source-driven specs complement test-driven ones.** The test-driven specs answer *what must this function do* from the fixtures' point of view. The source-driven specs answer *what is in this file, section by section, and how does it cooperate*. Both cite the same `punycode.js:line` locations.

## Reading order for the Go port

1. `test-data-fixtures.md` — learn the vector shape and the fidelity requirement.
2. `src-punycode.md` — get the full source-file map in one place (constants, primitives, call graph, overflow guards, scope gaps).
3. `punycode-ucs2-decode.md` + `punycode-ucs2-encode.md` — the UCS-2/UTF-16 primitives, most sensitive to language choice (Go strings are UTF-8).
4. `punycode-decode.md` + `punycode-encode.md` — the Bootstring core per RFC 3492.
5. `punycode-to-unicode.md` + `punycode-to-ascii.md` — the domain/email wrappers, including IDNA2003 separator normalisation.
6. `src-scripts-prepublish.md` — reference only; the build script is JS-only and out of port scope.

## Scope boundaries

- The specs cover only what `tests/tests.js` exercises. Uncovered code paths (e.g. `regexPunycode` case-sensitivity, multi-`@` emails) are noted as untested edge cases where relevant but not tested-against.
- The `version` string (`punycode.js:425`) is not under test and is not specified here.
