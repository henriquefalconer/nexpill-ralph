# Build Tooling — `scripts/prepublish.js` (`scripts/prepublish.js:1-17`)

A 17-line Node script that derives the ES6 module variant `punycode.es6.js` from
the CommonJS source `punycode.js`. Strict mode (`scripts/prepublish.js:1`). Wired
as the `build` npm script: `"build": "node scripts/prepublish.js"`
(`package.json:44`).

## Behavior

1. **Read source** (`scripts/prepublish.js:9`):
   `fs.readFileSync(path.resolve(__dirname, '../punycode.js'), 'utf-8')`.
2. **Integrity check** (`scripts/prepublish.js:6`, `11-13`): the regex
   `/module\.exports = punycode;/` (`scripts/prepublish.js:6`) is tested against
   the source; if absent, throw
   `'The underlying library has changed. Please update the prepublish script.'`
   (`scripts/prepublish.js:12`). This guards against the source export line
   (`punycode.js:443`) drifting.
3. **Rewrite the export** (`scripts/prepublish.js:7`, `15`): replace the matched
   `module.exports = punycode;` with the `output` string (`scripts/prepublish.js:7`):
   ```
   export { ucs2decode, ucs2encode, decode, encode, toASCII, toUnicode };
   export default punycode;
   ```
   Named exports: `ucs2decode`, `ucs2encode`, `decode`, `encode`, `toASCII`,
   `toUnicode`; plus a default export of the `punycode` object.
4. **Write output** (`scripts/prepublish.js:17`):
   `fs.writeFileSync(path.resolve(__dirname, '../punycode.es6.js'), outputContents)`.

## The ES6 variant artifact

- `punycode.es6.js` is **generated**, not committed: it is gitignored
  (`.gitignore:2`) and absent from the working tree until `npm run build` runs.
- `package.json` points ES6-aware consumers at it: `"jsnext:main"`
  (`package.json:7`) and `"module"` (`package.json:8`) both reference
  `punycode.es6.js`, and it is listed under `"files"` for npm packaging
  (`package.json:37-41`). The CommonJS `"main"` remains `punycode.js`
  (`package.json:6`).

The named exports the script injects correspond to the same functions the
CommonJS build exposes via the `punycode` object — see
[impl-domain-api.md](impl-domain-api.md).
