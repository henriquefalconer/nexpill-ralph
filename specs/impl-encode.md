# `encode` — Unicode → Punycode (`punycode.js:290-376`)

`encode(input)` converts a Unicode label to a **bare** Punycode ASCII string —
**without** the `xn--` prefix, which is added only by `toASCII`
(`punycode.js:411`). Documentation: `punycode.js:283-290`; return
`output.join('')` (`punycode.js:375`). It is the inverse of
[`decode`](impl-decode.md).

## Input prep & state (`punycode.js:294-302`)

- `input = ucs2decode(input)` — convert to a code-point array
  ([impl-ucs2.md](impl-ucs2.md)) (`punycode.js:294`); cache `inputLength`
  (`punycode.js:297`).
- `n = initialN` (128) (`punycode.js:300`); `delta = 0` (`punycode.js:301`);
  `bias = initialBias` (72) (`punycode.js:302`).

## Step 1 — Basic code points (`punycode.js:305-320`)

- For each `currentValue < 0x80`, push `stringFromCharCode(currentValue)`
  (`punycode.js:305-309`).
- `basicLength = output.length`; `handledCPCount = basicLength`
  (`punycode.js:311-312`).
- If `basicLength` is non-zero, push the `delimiter` `'-'`
  (`punycode.js:318-320`).

## Step 2 — Main loop (`punycode.js:323-374`)

Runs while `handledCPCount < inputLength` (`punycode.js:323`).

### Find next minimum code point ≥ n (`punycode.js:327-332`)

`m = maxInt` (`punycode.js:327`); scan all input, setting `m` to the smallest
`currentValue` with `currentValue >= n && currentValue < m`
(`punycode.js:328-332`).

### Advance delta to `<m,0>` (`punycode.js:336-342`)

- `handledCPCountPlusOne = handledCPCount + 1` (`punycode.js:336`).
- Overflow guard `m - n > floor((maxInt - delta) / handledCPCountPlusOne)` →
  `error('overflow')` (`punycode.js:337-339`).
- `delta += (m - n) * handledCPCountPlusOne` (`punycode.js:341`); `n = m`
  (`punycode.js:342`).

### Per-code-point pass (`punycode.js:344-369`)

For each `currentValue`:

- If `currentValue < n`, `++delta`; guard `++delta > maxInt` →
  `error('overflow')` (`punycode.js:345-347`).
- If `currentValue === n` (`punycode.js:348`), emit a generalized
  variable-length integer for `q = delta` (`punycode.js:350`):
  - Loop `k = base`, `k += base` (`punycode.js:351`).
  - `t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias)`
    (`punycode.js:352`). If `q < t`, **break** (`punycode.js:353-355`).
  - `qMinusT = q - t`, `baseMinusT = base - t`; push
    `stringFromCharCode(digitToBasic(t + qMinusT % baseMinusT, 0))`
    (`punycode.js:356-360`); `q = floor(qMinusT / baseMinusT)`
    (`punycode.js:361`).
  - After the loop, push the final digit
    `stringFromCharCode(digitToBasic(q, 0))` (`punycode.js:364`).
  - `bias = adapt(delta, handledCPCountPlusOne, handledCPCount === basicLength)`
    (`punycode.js:365`); `delta = 0` (`punycode.js:366`); `++handledCPCount`
    (`punycode.js:367`).

### Loop tail (`punycode.js:371-372`)

`++delta`; `++n`.

## Output (`punycode.js:375`)

`return output.join('')` — the bare Punycode string. `toASCII` prepends `'xn--'`
to this result (`punycode.js:411`). See [impl-domain-api.md](impl-domain-api.md).
