# Ralph TODO

Two documentation goals are now **complete** in flat `specs/*.md`, every claim
citing `path:line`:

## 1. Implementation specs (`src/*` / the source) — COMPLETE

There is no literal `src/` directory; the project source is `punycode.js` (the
whole codec) plus `scripts/prepublish.js` (ES6 build tooling). One subagent
studied each logical source unit; results captured as flat `impl-*.md`:

- `specs/impl-overview.md` — index, module layout, public API, call graph, build
- `specs/impl-constants.md` — `'use strict'`, `maxInt`, Bootstring params, regexes, `errors`, shortcuts (`punycode.js:1-31`)
- `specs/impl-helpers.md` — `error`, `map`, `mapDomain` (`punycode.js:41-86`)
- `specs/impl-ucs2.md` — `ucs2decode` / `ucs2encode` (`punycode.js:101-133`)
- `specs/impl-bootstring-helpers.md` — `basicToDigit`, `digitToBasic`, `adapt` (`punycode.js:144-187`)
- `specs/impl-decode.md` — `decode` (`punycode.js:196-281`)
- `specs/impl-encode.md` — `encode` (`punycode.js:290-376`)
- `specs/impl-domain-api.md` — `toUnicode`, `toASCII`, public API, export (`punycode.js:389-443`)
- `specs/impl-prepublish.md` — `scripts/prepublish.js:1-17`

## 2. Test specs (`tests/*`) — COMPLETE (prior iterations)

`specs/overview.md`, `test-fixtures.md`, `ucs2-decode.md`, `ucs2-encode.md`,
`decode.md`, `encode.md`, `to-unicode.md`, `to-ascii.md` document
`tests/tests.js` with `tests/tests.js:line` and `punycode.js:line` citations.

## Remaining / future work (only if scope expands)

- [ ] If new source files are added (e.g. a real `src/` dir), add a matching
      flat `specs/impl-<unit>.md` and link it from `specs/impl-overview.md`.
- [ ] If new files are added under `tests/`, add a matching `specs/<unit>.md`
      and link it from `specs/overview.md`.
- [ ] Optional code/test fix: the `decode('\x81')` test labeled "Overflow"
      actually throws "Invalid input" (`tests/tests.js:263-270` vs
      `punycode.js:240-242`; documented in `specs/impl-decode.md` and
      `specs/decode.md`). Surface as a build iteration only if requested.
