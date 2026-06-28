# UCS-2 Conversion — `ucs2decode` / `ucs2encode` (`punycode.js:101-133`)

JavaScript strings are UCS-2 internally, exposing each surrogate half as a
separate "character." This pair converts between a JS string and an array of
true Unicode code points (UTF-16 semantics), per the documentation at
`punycode.js:88-100` and the reference at `punycode.js:95`
(<https://mathiasbynens.be/notes/javascript-encoding>). They are exported as
`punycode.ucs2.decode` / `punycode.ucs2.encode` (`punycode.js:434-435`).

## `ucs2decode(string)` (`punycode.js:101-123`)

- **Returns:** an array of numeric code points (`punycode.js:99`).
- **Loop:** walks the string by code unit with `counter`/`length`
  (`punycode.js:102-105`), reading `value = string.charCodeAt(counter++)`
  (`punycode.js:106`).
- **High surrogate detection** (`punycode.js:107`): `value >= 0xD800 && value <=
  0xDBFF && counter < length` — i.e. a high surrogate with at least one more code
  unit available.
- **Low surrogate test** (`punycode.js:109-110`): read
  `extra = string.charCodeAt(counter++)`, then `(extra & 0xFC00) == 0xDC00`
  checks the top six bits identify a low surrogate (0xDC00–0xDFFF).
- **Pair combination** (`punycode.js:111`):
  `((value & 0x3FF) << 10) + (extra & 0x3FF) + 0x10000` — take the 10 payload
  bits of each half, shift the high half left 10, add the low half, add the
  supplementary-plane offset 0x10000. The combined code point is pushed.
- **Unmatched high surrogate** (`punycode.js:113-116`): if `extra` is not a low
  surrogate, push the lone `value` and `counter--` (`punycode.js:116`) so the
  consumed code unit is reprocessed next iteration — it may itself begin a valid
  pair.
- **Non-surrogate** (`punycode.js:118-120`): push `value` directly.

**Internal use:** `encode` calls `ucs2decode(input)` (`punycode.js:294`) to turn
its argument into a code-point array before encoding.

## `ucs2encode(codePoints)` (`punycode.js:133`)

An arrow function: `codePoints => String.fromCodePoint(...codePoints)`. It
spreads the code-point array into `String.fromCodePoint`, reconstituting a JS
(UCS-2) string. The core `decode` function uses the identical idiom inline at
`punycode.js:280` (`String.fromCodePoint(...output)`) rather than calling
`ucs2encode`.
