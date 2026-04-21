# `adapt(delta, numPoints, firstTime)` — Bootstring Bias Adaptation

## Subject

The `adapt()` function at `punycode.js:178` implements the bias adaptation mechanism from RFC 3492 §3.4. This function is called once per code-point insertion (in `encode`) or extraction (in `decode`) to update the Bootstring `bias` used by the generalized variable-length integer coder. The bias affects how delta values are encoded/decoded in the Punycode algorithm.

Reference: [RFC 3492 §3.4](https://tools.ietf.org/html/rfc3492#section-3.4) (`punycode.js:175`)

## Contract

The function signature (`punycode.js:178`):

```javascript
const adapt = function(delta, numPoints, firstTime) { ... }
```

### Parameters

- **`delta`** — The delta value being adapted (an integer representing the distance in code point values between successive characters).
- **`numPoints`** — The total number of code points handled so far, including the current one. Used as a divisor to normalize delta across all processed code points.
- **`firstTime`** — A boolean flag. When `true`, delta is divided by `damp` (700); when `false`, delta is shifted right by 1 bit (equivalent to dividing by 2).

### Return Value

A new bias value (integer) to be used for the next variable-length integer encoding/decoding operation.

## Implementation Walkthrough

### Step 1: Initialize `k`

At `punycode.js:179`, `k` is initialized to 0. This will accumulate the bias offset across the loop iterations.

```javascript
let k = 0;
```

### Step 2: Scale Delta

At `punycode.js:180`, delta is scaled differently depending on whether this is the first adaptation:

```javascript
delta = firstTime ? floor(delta / damp) : delta >> 1;
```

- If `firstTime` is true, divide delta by `damp` (`700` at `punycode.js:10`), using `floor()` to truncate.
- If `firstTime` is false, right-shift delta by 1 bit (`delta >> 1`), which is equivalent to `floor(delta / 2)`.

This scaling reduces the delta's magnitude for the next step.

### Step 3: Normalize by Code Point Count

At `punycode.js:181`, delta is further adjusted by the number of code points:

```javascript
delta += floor(delta / numPoints);
```

This is the RFC 3492 operation: `delta := delta + delta div numPoints`. It distributes the delta proportionally across all code points processed so far.

### Step 4: Convergence Loop

At `punycode.js:182`, a loop runs while `delta > baseMinusTMin * tMax >> 1`:

```javascript
for (/* no initialization */; delta > baseMinusTMin * tMax >> 1; k += base) {
    delta = floor(delta / baseMinusTMin);
}
```

With constants from `punycode.js:7-10,28`:
- `baseMinusTMin = 35` (at `punycode.js:28`)
- `tMax = 26` (at `punycode.js:9`)
- `base = 36` (at `punycode.js:7`)

The loop condition evaluates to `delta > ((36 - 1) * 26) >> 1 = 455`.

Each iteration:
- Divides delta by `baseMinusTMin` (35), using `floor()`.
- Increments k by `base` (36).

This loop ensures delta converges to a small enough value to be representable with the available code range.

### Step 5: Compute Final Bias

At `punycode.js:185`, the final bias is computed:

```javascript
return floor(k + (baseMinusTMin + 1) * delta / (delta + skew));
```

With constants:
- `baseMinusTMin + 1 = 36` (at `punycode.js:28`; `baseMinusTMin = 35`)
- `skew = 38` (at `punycode.js:10`)

This formula applies a weighted correction to the accumulated `k` offset based on the remaining `delta` and the skew constant, producing the final bias for use in the next encoding/decoding iteration.

## Constants Referenced

| Constant | Value | Location | Purpose |
|----------|-------|----------|---------|
| `base` | 36 | `punycode.js:7` | Base for the digit set (0-9, a-z) |
| `tMin` | 1 | `punycode.js:8` | Minimum threshold value |
| `tMax` | 26 | `punycode.js:9` | Maximum threshold value |
| `skew` | 38 | `punycode.js:10` | Skew factor for bias adjustment |
| `damp` | 700 | `punycode.js:11` | Damping factor for first-time scaling |
| `baseMinusTMin` | 35 | `punycode.js:28` | Precomputed value: `base - tMin` |

## Callsites

### In `decode()` at `punycode.js:264`

```javascript
bias = adapt(i - oldi, out, oldi == 0);
```

- **`delta`**: `i - oldi` — the delta accumulated since the last bias adaptation.
- **`numPoints`**: `out` — the output length at this point (total code points decoded so far, plus one).
- **`firstTime`**: `oldi == 0` — true if and only if `oldi` is 0, meaning no non-basic character has yet been consumed.

### In `encode()` at `punycode.js:365`

```javascript
bias = adapt(delta, handledCPCountPlusOne, handledCPCount === basicLength);
```

- **`delta`**: `delta` — the accumulated delta for the current code point being encoded.
- **`numPoints`**: `handledCPCountPlusOne` — the total count of code points handled so far, plus one.
- **`firstTime`**: `handledCPCount === basicLength` — true if and only if no non-basic code point has yet been emitted (i.e., we are still encoding the first non-basic character).

## References

- RFC 3492: Punycode: A Bootstring algorithm for representing Unicode with ASCII: [Section 3.4 — Bias adaptation](https://tools.ietf.org/html/rfc3492#section-3.4)
