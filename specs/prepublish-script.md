# prepublish script

## Purpose

Generate `punycode.es6.js` from `punycode.js` so that the package exposes an ES-module entry point. The `"module"` field in `package.json` (package.json:8) points to `punycode.es6.js`; bundlers and ESM-aware runtimes use this file in preference to the CommonJS `"main"` entry. The script is invoked once at build time — it is not part of the runtime library.

## Inputs

- **Source file:** `../punycode.js` resolved relative to `scripts/prepublish.js` via `path.resolve(__dirname, '../punycode.js')` (scripts/prepublish.js:9). Read as UTF-8 text.
- **Guard regex:** `/module\.exports = punycode;/` (scripts/prepublish.js:6). Used to locate the CommonJS export line that must be replaced.

## Output

- **Output file:** `../punycode.es6.js` resolved relative to `scripts/prepublish.js` via `path.resolve(__dirname, '../punycode.es6.js')` (scripts/prepublish.js:17). Written as UTF-8 text, overwriting any prior contents.

## Behavior

Steps execute top-to-bottom with no branching beyond the guard check:

1. **Read source** — Read `punycode.js` synchronously as UTF-8 text and assign the result to `sourceContents` (scripts/prepublish.js:9).

2. **Guard check** — Test `sourceContents` against the regex `/module\.exports = punycode;/` (scripts/prepublish.js:11). The regex is non-global; `RegExp.prototype.test` returns `true` if the pattern appears anywhere in the string.

3. **Throw on mismatch** — If the regex does not match, throw `Error('The underlying library has changed. Please update the prepublish script.')` and halt (scripts/prepublish.js:12). This protects against silently producing a malformed output file if `punycode.js` is ever restructured.

4. **Rewrite the export line** — Call `String.prototype.replace` with the (non-global) regex and the replacement string (scripts/prepublish.js:15):

   ```
   export { ucs2decode, ucs2encode, decode, encode, toASCII, toUnicode };
   export default punycode;
   ```

   The replacement string is defined as a single string literal containing a literal newline (`\n`) between the two export statements (scripts/prepublish.js:7).

   The matched substring `module.exports = punycode;` (punycode.js:443) is substituted in-place; all other characters of `sourceContents` are preserved verbatim.

5. **Write output** — Write the resulting text to `punycode.es6.js` synchronously via `fs.writeFileSync` (scripts/prepublish.js:17).

## Invariants

- **Single replacement.** `String.prototype.replace` with a non-global regex replaces only the first occurrence of the pattern (scripts/prepublish.js:15). If the pattern `module.exports = punycode;` appeared more than once in `punycode.js`, only the first match would be rewritten. In the current source there is exactly one occurrence (punycode.js:443).
- **Always overwrite.** `fs.writeFileSync` truncates and rewrites the output file unconditionally (scripts/prepublish.js:17). There is no incremental or diff-based update.
- **No module bundling.** The script performs plain text substitution; it does not parse or transpile JavaScript. The rest of `punycode.js` is emitted unchanged, meaning `punycode.es6.js` still uses `var`, `const`, and CommonJS-style internal helpers. Only the final export declaration is converted to ESM syntax.
- **Named exports are fixed.** The six names in the named-export statement — `ucs2decode`, `ucs2encode`, `decode`, `encode`, `toASCII`, `toUnicode` — are hard-coded in the replacement string (scripts/prepublish.js:7). They correspond exactly to the properties assigned to the `punycode` object in `punycode.js` (punycode.js:430-441).

## NPM wiring

```
npm run build  →  node scripts/prepublish.js
```

Defined in `package.json` (package.json:44):

```json
"build": "node scripts/prepublish.js"
```

`punycode.es6.js` is listed in the `"files"` array (package.json:40) alongside `punycode.js`, so it is included in the published npm package.

## Relation to Go port

Go uses a single-file module model and has no concept of separate CommonJS and ESM entry points. This build step has no direct analogue in a Go port. However, the set of public identifiers it hard-codes is authoritative: any Go port must expose exactly the same six names as exported symbols:

| JS export | Go export (suggested) |
|---|---|
| `ucs2decode` | `UCS2Decode` |
| `ucs2encode` | `UCS2Encode` |
| `decode` | `Decode` |
| `encode` | `Encode` |
| `toASCII` | `ToASCII` |
| `toUnicode` | `ToUnicode` |

For the full behavioural contract of each symbol, see the API specs listed under Cross-References below. The module-level surface (default export and named exports) is also documented in `punycode-module.md` (if present).

## Cross-References

- `punycode-module.md` — top-level module surface; documents the `punycode` object and default export that this script wraps in ESM syntax.
- [punycode-ucs2-decode.md](./punycode-ucs2-decode.md) — spec for `ucs2decode` (one of the six named exports).
- [punycode-ucs2-encode.md](./punycode-ucs2-encode.md) — spec for `ucs2encode`.
- [punycode-decode.md](./punycode-decode.md) — spec for `decode`.
- [punycode-encode.md](./punycode-encode.md) — spec for `encode`.
- [punycode-to-ascii.md](./punycode-to-ascii.md) — spec for `toASCII`.
- [punycode-to-unicode.md](./punycode-to-unicode.md) — spec for `toUnicode`.
- [test-fixtures.md](./test-fixtures.md) — test data shared across all API specs.
