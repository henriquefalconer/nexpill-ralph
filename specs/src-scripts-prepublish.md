# `scripts/prepublish.js` — source-file specification

This spec is the source-driven view of the build script that produces the ES-module variant shipped to npm. Every claim cites the exact line in `scripts/prepublish.js` (17 lines total). For the library itself, see `specs/src-punycode.md`.

## File overview

- **Path:** `scripts/prepublish.js`.
- **Length:** 17 lines (`scripts/prepublish.js:1-17`).
- **Role:** Generate `punycode.es6.js` from `punycode.js` by rewriting the CommonJS export line into ES-module `export` syntax.
- **Invoked by:** `npm run build` — `package.json:44` (`"build": "node scripts/prepublish.js"`).
- **Inputs:** reads `../punycode.js` (i.e. `punycode.js` at project root) at `scripts/prepublish.js:9`.
- **Outputs:** writes `../punycode.es6.js` at `scripts/prepublish.js:17`.
- **Source control:** `punycode.es6.js` is NOT checked in — it is a generated artefact produced at publish time. It is listed in `package.json:37-41` under `files`, so it ships in the npm tarball; the `package.json:6-8` entries `"jsnext:main"` and `"module"` point ES-module-aware consumers at it.
- **Port scope:** out of scope for the Go port (see `ralph/todo.md` §"Out of scope" and `TARGET.md`). This file is a JS-only build step with no analogue in a Go module.

## 1. Line-by-line behaviour

### 1.1 Preamble (`scripts/prepublish.js:1`)

`'use strict';` — strict mode for the whole script. Same semantics as `punycode.js:1`.

### 1.2 Imports (`scripts/prepublish.js:3-4`)

- `fs` (`scripts/prepublish.js:3`) — Node's synchronous file-system module. Used at `scripts/prepublish.js:9` (read) and `scripts/prepublish.js:17` (write).
- `path` (`scripts/prepublish.js:4`) — Node's path utilities. Used at `scripts/prepublish.js:9` and `scripts/prepublish.js:17` to resolve `../punycode.js` and `../punycode.es6.js` relative to this script's directory.

### 1.3 Transformation constants (`scripts/prepublish.js:6-7`)

- **Regex** (`scripts/prepublish.js:6`): `/module\.exports = punycode;/`. The backslash escapes the literal `.` and the regex carries no flags — it is used as a single-occurrence match. It targets the exact line `module.exports = punycode;` in `punycode.js:443`.
- **Replacement string** (`scripts/prepublish.js:7`): `'export { ucs2decode, ucs2encode, decode, encode, toASCII, toUnicode };\nexport default punycode;'`. Two statements joined by a literal `\n` newline.

The replacement MUST keep the six named exports in sync with `punycode.js:433-440`. If the library adds, removes, or renames a public export, this line must be updated by hand.

### 1.4 Source read (`scripts/prepublish.js:9`)

`fs.readFileSync(path.resolve(__dirname, '../punycode.js'), 'utf-8')` — synchronous, UTF-8-decoded read of the library file. `__dirname` is the directory of `scripts/prepublish.js` (i.e. `<project>/scripts`), so `../punycode.js` resolves to `<project>/punycode.js`. Failure modes: I/O error or missing file throws and aborts the script.

### 1.5 Sanity guard (`scripts/prepublish.js:11-13`)

```
if (!regex.test(sourceContents)) {
    throw new Error('The underlying library has changed. Please update the prepublish script.');
}
```

If the expected `module.exports = punycode;` line has been renamed, reflowed, or removed, the script throws before writing anything. This is a fail-fast invariant: without this guard the replace at `scripts/prepublish.js:15` would silently return the source unchanged and the ES-module build would be broken (still shaped like CommonJS) — consumers of `punycode.es6.js` would fail at import time. The guard converts that latent failure into an immediate build-time error with a human-actionable message.

### 1.6 Replacement (`scripts/prepublish.js:15`)

`sourceContents.replace(regex, output)` performs the single-occurrence substitution. The regex has no `g` flag; `String#replace` stops at the first match, which is exactly what is wanted because `punycode.js:443` is the sole `module.exports` line.

Before (from `punycode.js:443`):
```
module.exports = punycode;
```

After (the tail of `punycode.es6.js`):
```
export { ucs2decode, ucs2encode, decode, encode, toASCII, toUnicode };
export default punycode;
```

All other bytes of `punycode.js` pass through unchanged.

### 1.7 Output write (`scripts/prepublish.js:17`)

`fs.writeFileSync(path.resolve(__dirname, '../punycode.es6.js'), outputContents)` — synchronous write, creating or overwriting `<project>/punycode.es6.js`. No explicit encoding argument is given; `writeFileSync` defaults to UTF-8 for string inputs.

## 2. Invariants enforced by the script

1. **The library exposes exactly one `module.exports = punycode;` statement** — enforced by `scripts/prepublish.js:11-13`. Violated if `punycode.js:443` is reformatted.
2. **The six named exports in the replacement match the keys of the `punycode` object** at `punycode.js:433-440`. Not enforced by code — maintained by hand. Any drift produces a `punycode.es6.js` whose named-export list disagrees with the actual library surface.
3. **The generated file is idempotent** — running `npm run build` twice yields identical output, because the script reads, transforms, and writes without depending on prior state.

## 3. Project-workflow context

| Concern | Source | Value |
|---|---|---|
| Build command | `package.json:44` | `node scripts/prepublish.js` |
| CommonJS entry | `package.json:6` | `punycode.js` |
| ES-module entry | `package.json:7-8` | `punycode.es6.js` (generated) |
| Published tarball | `package.json:37-41` | `LICENSE-MIT.txt`, `punycode.js`, `punycode.es6.js` |

The generated `punycode.es6.js` must exist on disk at publish time, or it will be missing from the npm tarball. It is not tracked in git.

## 4. Out-of-scope for the Go port

Per `TARGET.md` and `ralph/todo.md`, the Go port targets only the public Punycode API exposed by `punycode.js`. This build script:

- Has no behavioural contract tested by `tests/tests.js`.
- Produces a JS-only artefact (`punycode.es6.js`) that a Go module cannot consume.
- Contains no algorithmic logic that would be re-used by a port.

A port therefore does not need to replicate `scripts/prepublish.js` — it is documented here for completeness only.
