# Ralph Todo

Specs for the test suite are now in place under `specs/*.md` (flat layout). They cover every describe block in `tests/tests.js` plus the shared fixtures, with citations back to `tests/tests.js:<line>` and `punycode.js:<line>`.

## Priority queue

_No outstanding work._

The ultimate goal ("study every file in tests/* using separate subagents and document in /specs/*.md and link the implementation as citations in the specification (flat structure)") is satisfied:

- `tests/` contains only `tests/tests.js` (371 lines).
- Specs produced (all under `specs/`):
  - `test-fixtures.md` — the four shared `testData` arrays (strings, ucs2, domains, separators).
  - `punycode-ucs2-decode.md` — `punycode.ucs2.decode` behavior + the two colocated `punycode.decode` error tests.
  - `punycode-ucs2-encode.md` — `punycode.ucs2.encode` behavior + non-mutation guarantee.
  - `punycode-decode.md` — full Bootstring decode algorithm and error conditions.
  - `punycode-encode.md` — full Bootstring encode algorithm and round-trip guarantee.
  - `punycode-to-unicode.md` — domain/email ACE-to-Unicode wrapper.
  - `punycode-to-ascii.md` — domain/email Unicode-to-ACE wrapper with IDNA2003 separator normalization.
- Every behavioral claim cites `punycode.js:<line>`; every test-derived claim cites `tests/tests.js:<line>`.

## If future iterations revisit this:

- Re-run when `tests/tests.js` changes: regenerate the affected `specs/punycode-*.md` only (specs are keyed by describe block, so changes localize cleanly).
- Re-run when a new test file appears under `tests/`: add a new spec file named after the logical unit under `specs/` (flat layout).
- Keep `test-fixtures.md` in lockstep with the `testData` object; sibling specs reference its sections.
