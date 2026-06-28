# `decode` — Punycode → Unicode (`punycode.js:196-281`)

`decode(input)` converts a **bare** Punycode ASCII label (no `xn--` prefix) to a
Unicode string (`punycode.js:189-196`). It works on code units, not UCS-2
("Don't use UCS-2." — `punycode.js:197`), and returns
`String.fromCodePoint(...output)` (`punycode.js:280`). It is the inverse of
[`encode`](impl-encode.md).

## State (`punycode.js:198-202`)

- `output = []` — accumulator of decoded code points (`punycode.js:198`).
- `inputLength = input.length` (`punycode.js:199`).
- `i = 0` — the running `<n,i>` index (`punycode.js:200`).
- `n = initialN` (128) (`punycode.js:201`); `bias = initialBias` (72)
  (`punycode.js:202`).

## Step 1 — Basic code points (`punycode.js:208-219`)

- `basic = input.lastIndexOf(delimiter)`; if `< 0`, clamp to `0`
  (`punycode.js:208-211`). This is the count of literal basic characters before
  the last `'-'`.
- Copy each of the first `basic` characters to `output`
  (`punycode.js:213-219`). Any character `>= 0x80` in this region throws
  `error('not-basic')` (`punycode.js:215-216`).

## Step 2 — Main loop (`punycode.js:224-278`)

Start `index` at `basic > 0 ? basic + 1 : 0` (skip the delimiter if basics were
copied) (`punycode.js:224`). Each iteration decodes one generalized
variable-length integer into `i`, then derives the next code point.

### Variable-length integer (`punycode.js:231-261`)

- Save `oldi = i` (`punycode.js:231`); loop with weight `w = 1`, `k = base`,
  `k += base` each round (`punycode.js:232`).
- `index >= inputLength` → `error('invalid-input')` (`punycode.js:234-236`).
- `digit = basicToDigit(input.charCodeAt(index++))` (`punycode.js:238`); if
  `digit >= base` → `error('invalid-input')` (`punycode.js:240-242`).
- Overflow guard `digit > floor((maxInt - i) / w)` → `error('overflow')`
  (`punycode.js:243-245`); then `i += digit * w` (`punycode.js:247`).
- Threshold `t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias)`
  (`punycode.js:248`). If `digit < t`, **break** — integer complete
  (`punycode.js:250-252`).
- Else `baseMinusT = base - t`, guard `w > floor(maxInt / baseMinusT)` →
  `error('overflow')` (`punycode.js:254-257`), then `w *= baseMinusT`
  (`punycode.js:259`).

### Derive the code point (`punycode.js:263-276`)

- `out = output.length + 1` (`punycode.js:263`);
  `bias = adapt(i - oldi, out, oldi == 0)` (`punycode.js:264`).
- Overflow guard `floor(i / out) > maxInt - n` → `error('overflow')`
  (`punycode.js:268-270`).
- `n += floor(i / out)` (`punycode.js:272`); `i %= out` (`punycode.js:273`).
- `output.splice(i++, 0, n)` — insert `n` at position `i`, then advance `i`
  (`punycode.js:276`).

## Error paths (exact messages)

| Condition | Line | Message |
|---|---|---|
| non-basic char in basic region | `216` | `not-basic` |
| input exhausted mid-integer | `235` | `invalid-input` |
| digit ≥ base | `241` | `invalid-input` |
| `i += digit*w` would overflow | `244` | `overflow` |
| `w *= baseMinusT` would overflow | `256` | `overflow` |
| `n += floor(i/out)` would overflow | `269` | `overflow` |

### Note — `decode('\x81')` throws `invalid-input`, not overflow

`'\x81'` (U+0081) has no delimiter, so `basic = 0` (`punycode.js:209-210`) and
the main loop starts at `index = 0` (`punycode.js:224`).
`basicToDigit(0x81)` returns `base` (`punycode.js:154`), so the `digit >= base`
check at `punycode.js:240-242` throws **`invalid-input`**. The test that labels
this case "Overflow" (`tests/tests.js:263-270`) is therefore mislabeled relative
to the actual message — see [overview.md](overview.md) and [decode.md](decode.md).
