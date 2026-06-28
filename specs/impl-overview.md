# Implementation Overview — punycode.js

This is the index for the **implementation** specs of the punycode.js library —
the source code itself, as opposed to the test-suite specs (`overview.md`,
`encode.md`, `decode.md`, etc.) which document `tests/tests.js`. Every claim in
these specs cites the implementation as `punycode.js:<line>` (or
`scripts/prepublish.js:<line>` / `package.json:<line>`).

## Scope: what "the source" is

The project has **no `src/` directory**. The entire codec is a single CommonJS
file, `punycode.js` (`punycode.js:1-444`, 443 lines of code), declared in strict
mode (`punycode.js:1`). The only other source file is the build script
`scripts/prepublish.js` (`scripts/prepublish.js:1-17`), which derives the ES6
variant. These two files constitute the project source and are each documented
below.

- Package metadata: name `"punycode"` (`package.json:2`), version `"2.3.1"`
  (`package.json:3`, mirrored at `punycode.js:425`), description "A robust
  Punycode converter that fully complies to RFC 3492 and RFC 5891…"
  (`package.json:4`), CommonJS entry `"main": "punycode.js"` (`package.json:6`).
- Build script wiring: `"build": "node scripts/prepublish.js"` (`package.json:44`).

## Implementation spec map

| Spec | Covers | Implementation |
|---|---|---|
| [impl-constants.md](impl-constants.md) | `'use strict'`, `maxInt`, Bootstring params, regexes, `errors`, shortcuts | `punycode.js:1-31` |
| [impl-helpers.md](impl-helpers.md) | Generic private helpers `error`, `map`, `mapDomain` | `punycode.js:41-86` |
| [impl-ucs2.md](impl-ucs2.md) | `ucs2decode` / `ucs2encode` (UCS-2 ↔ code-point arrays) | `punycode.js:101-123`, `punycode.js:133` |
| [impl-bootstring-helpers.md](impl-bootstring-helpers.md) | `basicToDigit`, `digitToBasic`, `adapt` | `punycode.js:144-187` |
| [impl-decode.md](impl-decode.md) | `decode` (Punycode → Unicode) | `punycode.js:196-281` |
| [impl-encode.md](impl-encode.md) | `encode` (Unicode → Punycode) | `punycode.js:290-376` |
| [impl-domain-api.md](impl-domain-api.md) | `toUnicode`, `toASCII`, public `punycode` object, `module.exports` | `punycode.js:389-443` |
| [impl-prepublish.md](impl-prepublish.md) | ES6-variant build tooling | `scripts/prepublish.js:1-17` |

## Module layout (top to bottom)

1. **Configuration block** (`punycode.js:1-31`) — strict-mode directive, the
   overflow sentinel `maxInt`, the eight RFC 3492 Bootstring parameters, three
   regexes, the `errors` message table, and convenience shortcuts. See
   [impl-constants.md](impl-constants.md).
2. **Generic helpers** (`punycode.js:41-86`) — `error`, `map`, `mapDomain`. See
   [impl-helpers.md](impl-helpers.md).
3. **UCS-2 conversion** (`punycode.js:101-133`) — `ucs2decode`, `ucs2encode`.
   See [impl-ucs2.md](impl-ucs2.md).
4. **Bootstring digit/bias helpers** (`punycode.js:144-187`) — `basicToDigit`,
   `digitToBasic`, `adapt`. See [impl-bootstring-helpers.md](impl-bootstring-helpers.md).
5. **Core codec** (`punycode.js:196-376`) — `decode`
   ([impl-decode.md](impl-decode.md)) and `encode`
   ([impl-encode.md](impl-encode.md)).
6. **Domain-level API** (`punycode.js:389-414`) — `toUnicode`, `toASCII`. See
   [impl-domain-api.md](impl-domain-api.md).
7. **Public API + export** (`punycode.js:419-443`) — the `punycode` object and
   `module.exports`. See [impl-domain-api.md](impl-domain-api.md).

## Public API surface (`punycode.js:419-441`)

The sole export (`module.exports = punycode`, `punycode.js:443`) is an object with:

- `version` — `'2.3.1'` (`punycode.js:425`).
- `ucs2.decode` → `ucs2decode` (`punycode.js:434`), `ucs2.encode` → `ucs2encode`
  (`punycode.js:435`).
- `decode` (`punycode.js:437`), `encode` (`punycode.js:438`), `toASCII`
  (`punycode.js:439`), `toUnicode` (`punycode.js:440`).

## Internal call graph

- `encode` calls `ucs2decode` (`punycode.js:294`), `digitToBasic`
  (`punycode.js:359`, `punycode.js:364`), `adapt` (`punycode.js:365`),
  `stringFromCharCode` (`punycode.js:307`, `359`, `364`), and `error`
  (`punycode.js:338`, `346`).
- `decode` calls `basicToDigit` (`punycode.js:238`), `adapt` (`punycode.js:264`),
  and `error` (`punycode.js:216`, `235`, `241`, `244`, `256`, `269`); it builds
  output with `String.fromCodePoint` (`punycode.js:280`).
- `toUnicode` calls `mapDomain` + `decode` (`punycode.js:390-393`); `toASCII`
  calls `mapDomain` + `encode` (`punycode.js:409-412`). `mapDomain` calls `map`
  (`punycode.js:84`).

## Relationship to the test specs and to the ES6 build

- The test specs in this directory document `tests/tests.js` against this same
  implementation; cross-reference the `Implementation` column of
  [overview.md](overview.md).
- The ES6 module `punycode.es6.js` is **generated** from this file by
  `scripts/prepublish.js` and is gitignored (`.gitignore:2`); it is not present
  in the working tree. See [impl-prepublish.md](impl-prepublish.md).
