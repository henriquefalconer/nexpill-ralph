# punycode — Go port

This is a Go port of [punycode.js](https://github.com/mathiasbynens/punycode.js). For full behaviour specification see [`specs/README.md`](../specs/README.md).

## Symbol mapping

| Go symbol | JS symbol | `punycode.js` line |
|---|---|---|
| `Decode` | `punycode.decode` | 196–281 |
| `Encode` | `punycode.encode` | 283–376 |
| `ToUnicode` | `punycode.toUnicode` | 389–395 |
| `ToASCII` | `punycode.toASCII` | 408–414 |
| `UCS2Decode` | `punycode.ucs2.decode` | 101–123 |
| `UCS2Encode` | `punycode.ucs2.encode` | 125–133 |
| `ErrOverflow` | `errors['overflow']` | 23 |
| `ErrNotBasic` | `errors['not-basic']` | 24 |
| `ErrInvalidInput` | `errors['invalid-input']` | 25 |
