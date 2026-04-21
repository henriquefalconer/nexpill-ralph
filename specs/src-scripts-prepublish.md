# Spec: `scripts/prepublish.js` — ES6 module synthesizer

Source file: `scripts/prepublish.js:1-17` (whole file)

---

## 1. Subject

**Script purpose:** A small build-time Node.js script that transforms the CommonJS `punycode.js` into an ES6 module `punycode.es6.js` by replacing the `module.exports` line with named and default ES exports.

**Invocation:** `npm run build` maps to `node scripts/prepublish.js` (`package.json:44`).

---

## 2. Implementation walkthrough

### 2.1 Module setup and imports — `scripts/prepublish.js:1-4`

The script begins with `'use strict';` (`scripts/prepublish.js:1`) and imports Node built-ins:
- `const fs = require('fs');` (`scripts/prepublish.js:3`)
- `const path = require('path');` (`scripts/prepublish.js:4`)

These are used later to read the source file and write the output.

### 2.2 Regex and output template — `scripts/prepublish.js:6-7`

Defines a regex to match the single export line:
```javascript
const regex = /module\.exports = punycode;/;
```
(`scripts/prepublish.js:6`)

This matches the exact pattern in `punycode.js:443`:
```javascript
module.exports = punycode;
```

Defines the ES6 replacement string:
```javascript
const output = 'export { ucs2decode, ucs2encode, decode, encode, toASCII, toUnicode };\nexport default punycode;';
```
(`scripts/prepublish.js:7`)

This provides both named exports (the public API methods) and a default export of the entire `punycode` object.

### 2.3 Read source file — `scripts/prepublish.js:9`

```javascript
const sourceContents = fs.readFileSync(path.resolve(__dirname, '../punycode.js'), 'utf-8');
```

Reads the entire CommonJS source relative to the script's directory, stored as a UTF-8 string.

### 2.4 Safety check — `scripts/prepublish.js:11-13`

```javascript
if (!regex.test(sourceContents)) {
	throw new Error('The underlying library has changed. Please update the prepublish script.');
}
```

If the regex does NOT match, throws an error. This is a critical guard: if `punycode.js` has been reformatted (spaces added, identifier names changed, etc.) or if the export line has been removed entirely, the build will fail rather than silently produce a broken ES6 module.

### 2.5 Transform and write — `scripts/prepublish.js:15, 17`

```javascript
const outputContents = sourceContents.replace(regex, output);

fs.writeFileSync(path.resolve(__dirname, '../punycode.es6.js'), outputContents);
```

Replaces the single matched line with the ES6 export statement and writes the result to `punycode.es6.js` in the repo root.

---

## 3. Why this exists

The library maintains a **single source of truth** in `punycode.js` (CommonJS format) and synthesizes the ES6 variant during the build rather than maintaining two parallel files. This avoids duplication and the risk of divergence.

`package.json` declares both:
- `"jsnext:main": "punycode.es6.js"` (`package.json:7`)
- `"module": "punycode.es6.js"` (`package.json:8`)

This signals to ESM-aware bundlers (webpack, rollup, etc.) that `punycode.es6.js` is the correct entry point for tree-shaking and module graph optimization.

The generated file `punycode.es6.js` is listed in `package.json:40` and included in the published package, but it is **not checked into git** — `.gitignore:2` explicitly ignores it as a generated artifact.

---

## 4. Invariants and contracts

### 4.1 Source file must be unchanged

The regex at `scripts/prepublish.js:6` is tightly coupled to the exact format of `punycode.js:443`. Any reformatting breaks the build:
- Adding spaces: `module.exports  =  punycode;` would not match.
- Changing the identifier: `module.exports = punycodeModule;` would not match.
- Removing the line entirely (e.g., using `export` directly in `punycode.js`) would cause the safety check to throw.

### 4.2 Named export list must stay in sync with the `punycode` object

The hard-coded export list at `scripts/prepublish.js:7`:
```javascript
export { ucs2decode, ucs2encode, decode, encode, toASCII, toUnicode }
```

must match the keys of the `punycode` object literal defined at `punycode.js:419-441`:
- `'version'` — version string (not exported as named export, only via default)
- `'ucs2'` — object with `decode` and `encode` methods
- `'decode'` — Punycode decoder
- `'encode'` — Punycode encoder
- `'toASCII'` — internationalized domain name to ASCII
- `'toUnicode'` — ASCII to internationalized domain name

The selected exports cover the main API methods but intentionally omit `version` and the `ucs2` container object (consumers import `ucs2decode` / `ucs2encode` directly or use the default export's `punycode.ucs2` property).

**Risk:** If a new method is added to the `punycode` object without updating the named export list in the script, ESM consumers who rely on named imports will silently fail to access the new method.

### 4.3 Silent success on match

The script produces no stdout or logging on successful execution (`scripts/prepublish.js:15-17`). It is silent if the regex matches and the file is written. This is appropriate for a build step where silence indicates success, but means errors must be caught via exit code.

---

## 5. Edge cases and failure modes

### 5.1 Caller-visible renaming of `punycode`

If `punycode.js` renames the exported identifier from `punycode` to something else (e.g., `punycodeLib`), the regex test at `scripts/prepublish.js:11-13` will catch it and throw. However, if someone were to rename `punycode` only in the `export default` statement at `scripts/prepublish.js:7` (e.g., `export default punycodeLib`), the mismatch would silently produce an ES6 module that still exports `punycode`, leading to runtime errors in consumers.

### 5.2 No validation of export list correctness

The script does not introspect or validate that the named export list actually exists as keys in the `punycode` object. If someone manually edits the list (e.g., adding a typo like `toUNICODE` instead of `toUnicode`), the ES6 build will silently reference a non-existent property, and consumers will get `undefined`.

### 5.3 File I/O errors

If `punycode.js` cannot be read (permissions, missing file) or if `punycode.es6.js` cannot be written (permissions, read-only filesystem), `fs.readFileSync` or `fs.writeFileSync` will throw and crash the build. These errors propagate to the caller (`npm run build`).

---

## 6. Implementation citations

| Item | Location |
|---|---|
| Script entry point and strict mode | `scripts/prepublish.js:1` |
| `fs` module import | `scripts/prepublish.js:3` |
| `path` module import | `scripts/prepublish.js:4` |
| Regex pattern for `module.exports = punycode;` | `scripts/prepublish.js:6` |
| ES6 export template | `scripts/prepublish.js:7` |
| Read source file | `scripts/prepublish.js:9` |
| Safety check on regex match | `scripts/prepublish.js:11-13` |
| Replace matched line | `scripts/prepublish.js:15` |
| Write ES6 module | `scripts/prepublish.js:17` |
| Build script entry in package.json | `package.json:44` |
| Module entry points (ESM) | `package.json:7-8` |
| Published files list (includes punycode.es6.js) | `package.json:37-40` |
| Exported `module.exports = punycode;` in source | `punycode.js:443` |
| `punycode` object definition | `punycode.js:419-441` |
| gitignore of generated ES6 build | `.gitignore:2` |

---

## 7. Related artifacts

- **Source file being transformed:** `punycode.js` (CommonJS, the single source of truth)
- **Generated output:** `punycode.es6.js` (ES6 module, generated only at build time)
- **Build system hook:** `npm run build` in `package.json:44`
- **Package metadata:** `package.json` declares `module` and `jsnext:main` fields pointing to the generated file

