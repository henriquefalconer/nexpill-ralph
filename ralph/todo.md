# Ralph Todo

Specs for this repository live under `specs/*.md` (flat layout). Two stages have been completed:

- **Stage 1 — test-derived specs** (from `tests/tests.js`, one per describe block).
- **Stage 2 — source-derived specs** (from `punycode.js` and `scripts/prepublish.js`).

Every behavioral or structural claim cites its source with `path:line`. The two stages are complementary: Stage 1 describes observed behavior; Stage 2 describes implementation structure (constants, helpers, build pipeline).

## Priority queue

_No outstanding work._

The ultimate goal of the current plan iteration ("study every file in src/* using separate subagents per file and link the implementation as citations in the specification (flat structure)") is satisfied. There is no literal `src/` directory; the effective source files (per `package.json` `"files"` list and `"main"` entry) are:

- `punycode.js` (the library implementation, 443 lines)
- `scripts/prepublish.js` (the ESM build step, 18 lines)

Each source file was studied by its own Sonnet subagent and produced flat-layout specs under `specs/`.

## Specs inventory (flat)

### Stage 1 — test-derived (from `tests/tests.js`)

- `specs/test-fixtures.md` — the four shared `testData` arrays (strings, ucs2, domains, separators).
- `specs/punycode-ucs2-decode.md` — `punycode.ucs2.decode` behavior + the two colocated `punycode.decode` error tests.
- `specs/punycode-ucs2-encode.md` — `punycode.ucs2.encode` behavior + non-mutation guarantee.
- `specs/punycode-decode.md` — full Bootstring decode algorithm and error conditions.
- `specs/punycode-encode.md` — full Bootstring encode algorithm and round-trip guarantee.
- `specs/punycode-to-unicode.md` — domain/email ACE-to-Unicode wrapper.
- `specs/punycode-to-ascii.md` — domain/email Unicode-to-ACE wrapper with IDNA2003 separator normalization.

### Stage 2 — source-derived (from `punycode.js` and `scripts/prepublish.js`)

- `specs/punycode-module.md` — module-level constants, Bootstring parameters, regex patterns, error-message table, convenience shortcuts, public-API object shape, and `module.exports` wiring.
- `specs/punycode-internal-helpers.md` — private helpers: `error()` (throws `RangeError`), `map()` (reverse-iterating array map), `mapDomain()` (email-aware label splitter with separator normalization).
- `specs/punycode-bootstring-digit.md` — `basicToDigit()` and `digitToBasic()` (the Bootstring digit ↔ ASCII code-point conversions shared by `decode` and `encode`).
- `specs/punycode-bias-adapt.md` — `adapt()` (RFC 3492 §3.4 bias adaptation used by both `decode` and `encode`).
- `specs/prepublish-script.md` — the `npm run build` step that rewrites `module.exports = punycode;` into ESM named+default exports to generate `punycode.es6.js`.

## If future iterations revisit this

- **When `tests/tests.js` changes:** regenerate the affected `specs/punycode-*.md` test-derived specs (keyed by describe block).
- **When `punycode.js` changes:** regenerate the relevant Stage-2 spec(s):
  - Constant/regex/export changes → `punycode-module.md`.
  - Helper changes → `punycode-internal-helpers.md`, `punycode-bootstring-digit.md`, or `punycode-bias-adapt.md`.
  - Behavioral changes also touch the Stage-1 specs that cite the affected line ranges.
- **When `scripts/prepublish.js` changes:** regenerate `prepublish-script.md`.
- **When a new source or test file appears:** add a new spec file under `specs/` (flat layout, named after the logical unit).
- **Next stage (per README.md:151):** a separate plan iteration will distill these specs into a prioritized Go-port plan. No source/test/config edits are authorized under plan mode — it writes markdown only.
