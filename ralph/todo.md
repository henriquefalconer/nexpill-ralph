# Ralph TODO

The ULTIMATE GOAL (document every file in `tests/*` in flat `specs/*.md` with
implementation citations) is **complete**. `tests/` contains one file,
`tests/tests.js`, whose shared fixture and six `describe` suites are each
documented in `specs/`:

- `specs/overview.md` — suite map, runner, inverse-pair relationships, known discrepancy
- `specs/test-fixtures.md` — the shared `testData` object (strings/ucs2/domains/separators)
- `specs/ucs2-decode.md`, `specs/ucs2-encode.md`
- `specs/decode.md`, `specs/encode.md`
- `specs/to-unicode.md`, `specs/to-ascii.md`

Every claim cites `tests/tests.js:line` and/or `punycode.js:line`.

## Remaining / future work (only if scope expands)

- [ ] If new files are added under `tests/`, add a matching flat `specs/<unit>.md`
      and link it from `specs/overview.md`.
- [ ] Optional: spec the build/publish tooling (`scripts/prepublish.js`,
      `package.json:44`) and the ES6 variant (`punycode.es6.js`) — out of scope
      for the current goal, which is limited to `tests/*`.
- [ ] Optional: surface the documented test/code discrepancy (the
      `decode('\x81')` test labeled "Overflow" actually throws "Invalid input",
      `tests/tests.js:263-270` vs `punycode.js:240-242`) as a code/test fix if a
      build iteration is ever requested.
