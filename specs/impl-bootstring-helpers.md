# Bootstring Digit/Bias Helpers — `basicToDigit`, `digitToBasic`, `adapt` (`punycode.js:144-187`)

Three private helpers implementing the digit codec and bias adaptation of
RFC 3492. They are the shared arithmetic core called by both `decode` and
`encode`.

## `basicToDigit(codePoint)` (`punycode.js:144-155`)

Maps an ASCII code point to a digit value in `[0, 35]`, or returns `base` (36) to
signal "not a digit" (the inverse of the RFC 3492 digit set).

| Input range | Condition | Result | Effect |
|---|---|---|---|
| `'0'`–`'9'` (0x30–0x39) | `punycode.js:145` | `26 + (codePoint - 0x30)` (`punycode.js:146`) | 26..35 |
| `'A'`–`'Z'` (0x41–0x5A) | `punycode.js:148` | `codePoint - 0x41` (`punycode.js:149`) | 0..25 |
| `'a'`–`'z'` (0x61–0x7A) | `punycode.js:151` | `codePoint - 0x61` (`punycode.js:152`) | 0..25 |
| anything else | — | `base` (36) (`punycode.js:154`) | invalid sentinel |

Note the decode/encode digit set is case-insensitive (upper and lower letters map
identically). **Caller:** `decode` (`punycode.js:238`); the `digit >= base` check
at `punycode.js:240` turns the sentinel into an `'invalid-input'` error.

## `digitToBasic(digit, flag)` (`punycode.js:168-172`)

The inverse of `basicToDigit`, computed branchlessly (`punycode.js:171`):

```
digit + 22 + 75 * (digit < 26) - ((flag != 0) << 5)
```

- `digit + 22` base offset.
- `75 * (digit < 26)`: the boolean `digit < 26` coerces to 1/0. Digits 0–25
  (letters) get +75 → land on `'a'`..`'z'` (97–122); digits 26–35 get +0 → land
  on `'0'`..`'9'` (48–57).
- `((flag != 0) << 5)` subtracts 0x20 when `flag` is non-zero, converting the
  lowercase letter to uppercase. Behavior is undefined for digits 26–35 with a
  set flag (`punycode.js:164-166`).

**Callers:** `encode` only, always with `flag = 0` (lowercase output):
`punycode.js:359` and `punycode.js:364`.

## `adapt(delta, numPoints, firstTime)` (`punycode.js:179-187`)

Bias adaptation per RFC 3492 §3.4 (`punycode.js:175-176`).

1. **Initial scale** (`punycode.js:181`):
   `delta = firstTime ? floor(delta / damp) : delta >> 1` — heavy damping by
   `damp` (700) the first time, otherwise halve.
2. **Per-point add** (`punycode.js:182`): `delta += floor(delta / numPoints)`.
3. **Scaling loop** (`punycode.js:183-185`): while
   `delta > baseMinusTMin * tMax >> 1` (i.e. `> 455`), do
   `delta = floor(delta / baseMinusTMin)` and `k += base`.
4. **Final bias** (`punycode.js:186`):
   `return floor(k + (baseMinusTMin + 1) * delta / (delta + skew))`
   = `floor(k + 36 * delta / (delta + 38))`.

**Callers and `firstTime` argument:**

- `decode` (`punycode.js:264`): `adapt(i - oldi, out, oldi == 0)` — `firstTime`
  is true on the first inserted code point.
- `encode` (`punycode.js:365`): `adapt(delta, handledCPCountPlusOne,
  handledCPCount === basicLength)` — `firstTime` is true for the first non-basic
  code point.
