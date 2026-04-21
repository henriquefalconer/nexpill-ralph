# Spec: `encode(input)` — Punycode encoder

**Source:** `punycode.js:283-376`  
**Public binding:** `punycode.js:438`  
**Cross-references:** `ucs2decode` (`punycode.js:101-123`), `digitToBasic` (`punycode.js:168-172`), `adapt` (`punycode.js:179-187`), `error` (`punycode.js:41-43`), constants (`punycode.js:1-31`)  
**Reference:** RFC 3492 §6.3

---

## 1. Subject

Converts a Unicode string (a single domain name label) into a Punycode ASCII string. The output is a plain ASCII string suitable for use in domain names; the `xn--` ACE (ASCII Compatible Encoding) prefix is **not** prepended by this function. That responsibility belongs to `toASCII` at `punycode.js:408-414`, which calls `encode` and prepends the prefix (`punycode.js:411`).

The function is exposed to callers as `punycode.encode` at `punycode.js:438`.

---

## 2. Contract

Derived from JSDoc at `punycode.js:283-289`:

- **Input** — a string of Unicode symbols (e.g. a domain name label).
- **Output** — the resulting Punycode string of ASCII-only symbols.
- **Errors** — throws `RangeError('overflow')` when an intermediate delta value would exceed `maxInt = 2147483647` (`punycode.js:4`).
- **Member** — `@memberOf punycode`.

---

## 3. Implementation Walkthrough

### 3.1 Input Normalization

**Lines:** `punycode.js:291-294`

```
const output = [];
input = ucs2decode(input);
const inputLength = input.length;
```

The input string is converted to an array of Unicode code points via `ucs2decode` (`punycode.js:101-123`). This transformation is essential: it decodes UTF-16 surrogate pairs (e.g. `\uD83C\uDF55`, a single emoji) into their corresponding astral code points (e.g. `127829`). Subsequent code-point comparisons and counts operate on this logical array, not the raw UTF-16 code units. Omitting this step causes astral characters (those > U+FFFF) to be miscounted and miscoded.

### 3.2 State Initialization

**Lines:** `punycode.js:300-302`

```
let n = initialN;        // 128 (0x80)
let delta = 0;
let bias = initialBias;  // 72
```

The encoder maintains three state variables:
- `n`: the code point threshold. Initially 128 (the boundary between basic and non-basic code points). Advanced through the main loop.
- `delta`: the accumulated "distance" to encode for non-basic code points. Reset to 0 after each non-basic code point is emitted.
- `bias`: the encoding bias for generalized variable-length integer (VLQ) compression. Computed by the `adapt` function (`punycode.js:179-187`) and applied during digit emission.

### 3.3 Basic Code Point Collection

**Lines:** `punycode.js:305-309`

```
for (const currentValue of input) {
	if (currentValue < 0x80) {
		output.push(stringFromCharCode(currentValue));
	}
}
```

The first pass over the input collects all basic code points (code points `< 0x80`) and appends them literally as ASCII characters. Non-basic code points are skipped; they are handled in the main encoding loop.

### 3.4 Basic-Section Bookkeeping

**Lines:** `punycode.js:311-312`

```
const basicLength = output.length;
let handledCPCount = basicLength;
```

Record the number of basic code points in the output. `handledCPCount` tracks the total number of code points (basic + non-basic) processed so far. Both quantities are used in overflow guards and for the delimiter and main-loop logic.

### 3.5 Delimiter Emission

**Lines:** `punycode.js:318-320`

```
if (basicLength) {
	output.push(delimiter);  // '-'
}
```

Emit the Punycode delimiter `-` (U+002D) if any basic code points were output. When the input is all non-basic (e.g. `'\xFC'`), `basicLength === 0` and the delimiter is omitted entirely. This rule produces:
- `'Bach-'` for ASCII input (delimiter appended),
- `'tda'` for all-non-ASCII input (no delimiter),
- `'-> $1.00 <--'` for the ASCII edge case (double hyphen: content + delimiter at the end).

### 3.6 Main Encoding Loop

**Lines:** `punycode.js:323-374`

The loop runs until all input code points have been handled.

#### 3.6.1 Find Next-Larger Code Point

**Lines:** `punycode.js:327-332`

```
let m = maxInt;
for (const currentValue of input) {
	if (currentValue >= n && currentValue < m) {
		m = currentValue;
	}
}
```

Scan the entire input array to find the smallest code point `m` such that `m >= n`. This code point has not been processed yet and will be the target of the next encoding step. The search is O(N) and runs on every iteration of the outer loop, making the overall algorithm **O(N²)** in the number of code points.

#### 3.6.2 Delta Increment with Overflow Guard

**Lines:** `punycode.js:336-342`

```
const handledCPCountPlusOne = handledCPCount + 1;
if (m - n > floor((maxInt - delta) / handledCPCountPlusOne)) {
	error('overflow');
}
delta += (m - n) * handledCPCountPlusOne;
n = m;
```

Compute the distance `m - n` and check whether advancing delta by `(m - n) * handledCPCountPlusOne` would exceed `maxInt`. The check at `punycode.js:337-339` is equivalent to:

```
(m - n) * handledCPCountPlusOne + delta > maxInt
```

If overflow would occur, throw `RangeError('overflow')` via `error('overflow')` at `punycode.js:338`. Otherwise, update `delta` and `n`.

#### 3.6.3 Per-Code-Point Inner Loop

**Lines:** `punycode.js:344-369`

For each code point in the input:

**Increment for code points < n:**

```javascript
if (currentValue < n && ++delta > maxInt) {
	error('overflow');
}
```

Every code point smaller than `n` causes `delta` to increment by 1. If this increment would exceed `maxInt`, throw `RangeError('overflow')` at `punycode.js:346-347`.

**Encode code point === n:**

```javascript
if (currentValue === n) {
	let q = delta;
	for (let k = base; ; k += base) {
		const t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias);
		if (q < t) {
			break;
		}
		const qMinusT = q - t;
		const baseMinusT = base - t;
		output.push(
			stringFromCharCode(digitToBasic(t + qMinusT % baseMinusT, 0))
		);
		q = floor(qMinusT / baseMinusT);
	}
	output.push(stringFromCharCode(digitToBasic(q, 0)));
	bias = adapt(delta, handledCPCountPlusOne, handledCPCount === basicLength);
	delta = 0;
	++handledCPCount;
}
```

When a code point equals the current threshold `n`, the accumulated `delta` is encoded as a generalized VLQ (variable-length quantity) and emitted digit by digit:

1. Initialize `q = delta` (the value to encode).
2. Inner loop: for each successive position `k` (stepping by `base` each iteration), compute the threshold `t`:
   - If `k <= bias`: `t = tMin` (1).
   - If `k >= bias + tMax`: `t = tMax` (26).
   - Otherwise: `t = k - bias`.
3. When `q < t`, the loop terminates (the final digit is `q`).
4. Otherwise, emit `t + (q - t) % (base - t)` as an ASCII character via `digitToBasic(..., 0)` (flag = 0 means lowercase), then divide `q` by `base - t` for the next iteration.
5. After the loop, emit the final digit `q` via `digitToBasic(q, 0)`.
6. Update `bias` using the `adapt` function (`punycode.js:179-187`), reset `delta = 0`, and increment `handledCPCount`.

The `digitToBasic` function at `punycode.js:168-172` always maps digits 0–35 to lowercase letters a–z and digits 0–9. The `flag` parameter controls case; it is always `0` in `encode`, ensuring lowercase-only output.

#### 3.6.4 Advance Threshold

**Lines:** `punycode.js:371-372`

```
++delta;
++n;
```

At the end of each iteration, increment both `delta` (to account for the code point just processed) and `n` (to advance the threshold for the next iteration).

### 3.7 Assembly

**Line:** `punycode.js:375`

```
return output.join('');
```

Concatenate the output array into a single string and return.

---

## 4. Overflow Conditions

The function checks for overflow in two places:

1. **Delta-advance overflow guard** (`punycode.js:337-339`): when computing `delta += (m - n) * handledCPCountPlusOne`, verify that the result does not exceed `maxInt`. Overflow here is rare for realistic domain labels but possible for pathological or malicious inputs.

2. **Per-code-point delta increment overflow** (`punycode.js:345-347`): when incrementing `delta` for each code point `< n`, check that the result does not exceed `maxInt`. This guard is checked for every iteration of the inner loop and is the more likely source of an overflow error on adversarial input.

Both paths throw `RangeError` via `error('overflow')` with the message `'Overflow: input needs wider integers to process'` (from `punycode.js:23`).

---

## 5. Non-Obvious Details

### 5.1 No `xn--` Prefix

The `encode` function outputs only the Punycode payload, never the `xn--` ACE prefix. This responsibility is reserved for the higher-level `toASCII` function (`punycode.js:408-414`), which calls `encode` and prepends the prefix:

```javascript
// punycode.js:411
return 'xn--' + encode(string);
```

### 5.2 O(N²) Complexity

The main encoding loop (line `punycode.js:323`) runs once per distinct non-basic code point. Inside it, a full linear scan of the input array searches for the next-larger code point (lines `punycode.js:327-332`). For an input with N distinct code points, this yields O(N²) comparisons. For typical domain labels (short and ASCII-heavy), this is negligible. For synthetic worst-case inputs (all distinct non-basic code points), the cost rises quadratically.

### 5.3 Empty Basic Prefix

When all input code points are non-basic (e.g. `'\xFC'`), the basic code point loop outputs nothing (lines `punycode.js:305-309`), `basicLength === 0` (line `punycode.js:311`), and the delimiter is omitted (line `punycode.js:318`). The main loop then encodes the entire input and outputs only the encoded payload.

### 5.4 Lowercase-Only Output

The `digitToBasic` function is always called with `flag = 0` (lines `punycode.js:359`, `punycode.js:364`). This means the output is always lowercase. RFC 3492 §6.3 specifies that implementations *may* use mixed case for checksum purposes, but Punycode.js does not. JavaScript provides no standard way to encode or parse case-sensitive checksums in domain names, so the library follows the all-lowercase convention.

---

## 6. Callers

1. **`toASCII`** (`punycode.js:408-414`): when encoding a domain name component, `toASCII` tests whether the label contains non-ASCII code points. If so, it calls `encode(string)` and prepends `'xn--'` to produce the full ACE label.

2. **Public API** (`punycode.js:438`): the `encode` function is bound to the public object as `punycode.encode`, allowing external callers to encode individual labels.

---

## 7. Key Dependencies

| Symbol | Location | Purpose |
|--------|----------|---------|
| `ucs2decode` | `punycode.js:101-123` | Convert UTF-16 input to code-point array |
| `digitToBasic` | `punycode.js:168-172` | Map digit integers to ASCII characters |
| `adapt` | `punycode.js:179-187` | Compute bias for VLQ encoding |
| `error` | `punycode.js:41-43` | Throw RangeError with message lookup |
| `maxInt` | `punycode.js:4` | Overflow threshold (2147483647) |
| `initialN` | `punycode.js:13` | Initial code-point threshold (128) |
| `initialBias` | `punycode.js:12` | Initial bias (72) |
| `base` | `punycode.js:7` | VLQ base (36) |
| `tMin` | `punycode.js:8` | Minimum threshold increment (1) |
| `tMax` | `punycode.js:9` | Maximum threshold increment (26) |
| `damp` | `punycode.js:11` | Bias-adaptation damping factor (700) |
| `skew` | `punycode.js:10` | Bias-adaptation skew (38) |
| `delimiter` | `punycode.js:14` | Punycode label separator (`'-'`) |

---

## 8. Test Binding

The `encode` function is tested via the `describe('punycode.encode')` block at `tests/tests.js:312-321`. See `specs/test-encode.md` for the full test specification, including 23 parameterized test vectors covering ASCII, non-ASCII, mixed-script, RFC 3492 reference samples, and edge cases.

