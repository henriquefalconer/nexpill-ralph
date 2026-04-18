# punycode-bias-adapt

## Purpose

This spec documents the `adapt` function, which implements the bias adaptation algorithm defined in RFC 3492 §3.4. The function adjusts the current bias value after each variable-length integer is decoded or encoded, so that the threshold values used in subsequent iterations stay proportional to the actual distribution of code points in the input. It is a shared subroutine called by both `decode` (see `punycode-decode.md`) and `encode` (see `punycode-encode.md`).

## `adapt(delta, numPoints, firstTime)` — RFC 3492 §3.4 bias adaptation

**Declaration:** `punycode.js:179–187`

**RFC reference:** https://tools.ietf.org/html/rfc3492#section-3.4

### Signature

- **Input:** `delta` — a non-negative integer; the generalized delta value from the most recently decoded or encoded variable-length integer. In the decoder, this is `i - oldi` (the change in the accumulator); in the encoder, this is the value of `delta` at the point `adapt` is called.
- **Input:** `numPoints` — a positive integer; the number of code points processed so far (output length + 1 at the moment of the call).
- **Input:** `firstTime` — a boolean; true when this is the very first call to `adapt` for a given string (i.e., the first non-basic code point has just been handled), false on all subsequent calls.
- **Output:** A non-negative integer representing the new bias value.

### Algorithm

The function modifies `delta` in several stages, then uses the result to compute the new bias. All divisions are integer divisions (floor), using the `floor` shortcut alias for `Math.floor` (punycode.js:30).

#### Step 1: Scale delta for the first iteration (punycode.js:181)

If `firstTime` is true:

```
delta = floor(delta / damp)
```

where `damp = 700` (punycode.js:11). This dampens the initial delta because the first delta is typically much larger than subsequent ones (it represents the cost of reaching the first non-basic code point from `initialN = 128`).

If `firstTime` is false:

```
delta = delta >> 1
```

which is equivalent to `floor(delta / 2)`. This is an integer right-shift by 1 bit, halving `delta`. The source uses the bitwise shift as a performance optimization; the semantics are identical to floored division by 2 for non-negative integers.

#### Step 2: Adjust delta by the number of code points (punycode.js:182)

```
delta = delta + floor(delta / numPoints)
```

This accounts for the fact that `delta` has been increased by `numPoints` on each of the previous `numPoints` iterations. Dividing by `numPoints` converts the cumulative delta to a per-symbol delta.

#### Step 3: Iterative reduction loop (punycode.js:183–185)

Initialize an accumulator `k = 0`.

While the following condition holds:

```
delta > floor(baseMinusTMin * tMax / 2)
```

where `baseMinusTMin = base - tMin = 35` (punycode.js:29), `tMax = 26` (punycode.js:9):

- The threshold is `floor(35 * 26 / 2) = floor(455) = 455`.
- Perform: `delta = floor(delta / baseMinusTMin)` (i.e., `floor(delta / 35)`).
- Add `base` (36) to `k`: `k = k + 36`.

This loop counts how many complete "rounds" of `base` units fit into `delta`, reducing `delta` toward the range `[0, 455]`. Each iteration increments `k` by `base`, preparing the final formula.

**Note:** The source code expresses the loop condition as `delta > baseMinusTMin * tMax >> 1` (punycode.js:183). The `>> 1` is a bitwise right shift by 1 applied to the product `baseMinusTMin * tMax = 35 * 26 = 910`; the result is `910 >> 1 = 455`. This is equivalent to `floor(910 / 2) = 455`. The condition is therefore `delta > 455`.

#### Step 4: Compute and return the new bias (punycode.js:186)

```
return floor(k + (baseMinusTMin + 1) * delta / (delta + skew))
```

where `skew = 38` (punycode.js:10) and `baseMinusTMin + 1 = 36`.

The expression `(baseMinusTMin + 1) * delta / (delta + skew)` = `36 * delta / (delta + 38)` is a rational interpolation that maps the reduced `delta` into an offset in the range `[0, 35)`. Adding `k` gives the final bias.

**Division order:** Because all values are integers and the source uses `Math.floor`, the multiplication `(baseMinusTMin + 1) * delta` is performed before the division by `(delta + skew)`. In a Go port, use integer arithmetic: `(baseMinusTMin+1)*delta / (delta+skew)` with integer division (truncation toward zero, which equals floor for non-negative values).

### Full Algorithm Summary (language-agnostic)

```
function adapt(delta, numPoints, firstTime):
    if firstTime:
        delta = floor(delta / 700)       // damp = 700
    else:
        delta = floor(delta / 2)
    delta = delta + floor(delta / numPoints)
    k = 0
    while delta > 455:                   // floor((base-tMin)*tMax/2) = floor(35*26/2) = 455
        delta = floor(delta / 35)        // floor(delta / (base - tMin))
        k = k + 36                       // k += base
    return floor(k + 36 * delta / (delta + 38))  // k + (base-tMin+1)*delta/(delta+skew)
```

The magic numbers inline above are derived from the constants as shown. A faithful port must use the constant expressions (`damp`, `base`, `tMin`, `tMax`, `skew`) rather than the numeric literals so that the algorithm's relationship to RFC 3492 remains clear.

### Call Sites

#### In `decode` (punycode.js:264)

```
bias = adapt(i - oldi, out, oldi == 0);
```

Where:
- `i - oldi` is the delta value decoded in the current iteration (the change to the accumulator).
- `out` is `output.length + 1` at the point of the call.
- `oldi == 0` is true only on the first iteration (before the first non-basic code point is processed).

#### In `encode` (punycode.js:365)

```
bias = adapt(delta, handledCPCount + 1, handledCPCount === basicLength);
```

Where:
- `delta` is the current delta value at the moment the code point `n` was found and emitted.
- `handledCPCount + 1` is the number of code points handled after this emission.
- `handledCPCount === basicLength` is true only on the first call (when the first non-basic code point has just been encoded).

After this call, `delta` is reset to 0 in the encoder.

## Constants Used

All constants are defined at module scope (see `punycode-module.md`):

| Constant | Value | Line |
|---|---|---|
| `base` | 36 | punycode.js:7 |
| `tMin` | 1 | punycode.js:8 |
| `tMax` | 26 | punycode.js:9 |
| `skew` | 38 | punycode.js:10 |
| `damp` | 700 | punycode.js:11 |
| `baseMinusTMin` | 35 (`base - tMin`) | punycode.js:29 |
| `floor` | integer floor function | punycode.js:30 |

## Cross-References

- [punycode-module.md](./punycode-module.md) — all Bootstring constants and the `floor`/`baseMinusTMin` shortcuts used here
- [punycode-decode.md](./punycode-decode.md) — calls `adapt` after each variable-length integer is decoded to update the bias for subsequent iterations
- [punycode-encode.md](./punycode-encode.md) — calls `adapt` after each code point is encoded to update the bias for subsequent iterations
- [punycode-bootstring-digit.md](./punycode-bootstring-digit.md) — digit conversion helpers that work alongside `adapt` in the encode/decode loops
