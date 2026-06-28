# Configuration Block — Constants & Regexes (`punycode.js:1-31`)

The file opens with a module-level configuration block: a strict-mode directive,
the overflow sentinel, the RFC 3492 Bootstring parameters, three regexes, the
error-message table, and convenience shortcuts. Everything here is `const`.

## `'use strict'` (`punycode.js:1`)

The module runs in strict mode, which applies to every function defined in the
file (e.g. `decode`, `encode`, `adapt`). It forbids implicit globals, octal
literals, and duplicate parameter names, among other restrictions.

## Overflow sentinel `maxInt` (`punycode.js:4`)

`maxInt = 2147483647` (`0x7FFFFFFF`, i.e. 2³¹−1), the highest positive signed
32-bit value. It is the ceiling every overflow guard checks against:

- `decode`: `punycode.js:243`, `punycode.js:255`, `punycode.js:268`.
- `encode`: `punycode.js:327` (initializing `m`), `punycode.js:337`,
  `punycode.js:345`.

## Bootstring parameters (`punycode.js:6-14`)

| Constant | Value | Line | RFC 3492 role | Consumed at |
|---|---|---|---|---|
| `base` | 36 | `punycode.js:7` | radix (a–z + 0–9) | `154`, `183`, `232`, `240`, `254`, `351`, `357` |
| `tMin` | 1 | `punycode.js:8` | minimum threshold | via `baseMinusTMin` (`29`); `248`, `352` |
| `tMax` | 26 | `punycode.js:9` | maximum threshold | `183`, `248`, `352` |
| `skew` | 38 | `punycode.js:10` | bias skew | `186` |
| `damp` | 700 | `punycode.js:11` | first-pass bias damping | `181` |
| `initialBias` | 72 | `punycode.js:12` | starting bias | `202` (decode), `302` (encode) |
| `initialN` | 128 (`0x80`) | `punycode.js:13` | first non-basic code point | `201` (decode), `300` (encode) |
| `delimiter` | `'-'` (`\x2D`) | `punycode.js:14` | basic/extended separator | `208` (decode), `319` (encode) |

These are the parameter values mandated by RFC 3492 §5 for the Punycode profile
of Bootstring.

## Regular expressions (`punycode.js:16-19`)

- `regexPunycode = /^xn--/` (`punycode.js:17`) — matches the ACE prefix at the
  start of a label. Used by `toUnicode` (`punycode.js:391`) to decide whether a
  label is Punycoded.
- `regexNonASCII = /[^\0-\x7F]/` (`punycode.js:18`) — matches any character
  outside U+0000–U+007F. The inline comment notes U+007F DEL is *excluded* from
  the allowed set (i.e. it counts as ASCII/basic; only ≥ U+0080 matches). Used by
  `toASCII` (`punycode.js:410`) to decide whether a label needs encoding.
- `regexSeparators = /[\x2E。．｡]/g` (`punycode.js:19`) — the four
  RFC 3490 label separators: U+002E FULL STOP, U+3002 IDEOGRAPHIC FULL STOP,
  U+FF0E FULLWIDTH FULL STOP, U+FF61 HALFWIDTH IDEOGRAPHIC FULL STOP. The `g`
  flag makes the replacement global. Used by `mapDomain` (`punycode.js:82`) to
  normalize all separators to ASCII `.` before splitting.

## Error-message table `errors` (`punycode.js:21-26`)

Three keys, each mapping to the exact message string thrown by `error()`
(`punycode.js:42`):

| Key | Message | Thrown at |
|---|---|---|
| `overflow` | `'Overflow: input needs wider integers to process'` | `244`, `256`, `269`, `338`, `346` |
| `not-basic` | `'Illegal input >= 0x80 (not a basic code point)'` | `216` |
| `invalid-input` | `'Invalid input'` | `235`, `241` |

## Convenience shortcuts (`punycode.js:28-31`)

- `baseMinusTMin = base - tMin` = 35 (`punycode.js:29`) — precomputed for the
  bias loop and digit math; consumed at `183`, `184`, `186`.
- `floor = Math.floor` (`punycode.js:30`) — aliased for the many integer
  divisions in `adapt`, `decode`, and `encode`.
- `stringFromCharCode = String.fromCharCode` (`punycode.js:31`) — aliased; used
  by `encode` at `307`, `359`, `364`.
