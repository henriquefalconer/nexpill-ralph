# Spec: `decode(input)` — Punycode decoder (RFC 3492 §6.2)

Source function: `punycode.js:189-281`

---

## 1. Subject

**Function:** `decode(input)` converts a Punycode ASCII-only string (the label body without the `xn--` prefix) into a Unicode string, implementing the RFC 3492 §6.2 decoding algorithm.

**Binding:** `punycode.decode = decode` at `punycode.js:437`, in the public object returned by the module.

**Input:** A Punycode-encoded ASCII string. No `xn--` prefix should be present; it is the caller's responsibility (e.g., `toUnicode` at `punycode.js:392`) to strip the prefix before passing to `decode`.

**Output:** A Unicode string.

---

## 2. Contract

Drawn from JSDoc at `punycode.js:189-195`:

- Input: a Punycode string of ASCII-only symbols (code points `< 0x80`).
- Output: the resulting string of Unicode symbols.
- Throws `RangeError` with one of three error messages (from `punycode.js:22-26`):
  - `'overflow'` — "Overflow: input needs wider integers to process"
  - `'not-basic'` — "Illegal input >= 0x80 (not a basic code point)"
  - `'invalid-input'` — "Invalid input"

---

## 3. Implementation walkthrough

### 3.1 State initialization (`punycode.js:196-202`)

```javascript
const output = [];
const inputLength = input.length;
let i = 0;
let n = initialN;
let bias = initialBias;
```

- `output` — accumulator array for code points (initially empty).
- `inputLength` — cached length of the input.
- `i` — accumulator for the decoded position; used to track where to insert the next code point.
- `n` — the current code point being inserted (initialized to `initialN = 128` at `punycode.js:13`).
- `bias` — the adaptive bias (initialized to `initialBias = 72` at `punycode.js:12`); adjusted after each iteration via the `adapt()` function to favor either small or large deltas depending on the distribution of insertions.

### 3.2 Find the basic-prefix delimiter (`punycode.js:208-211`)

```javascript
let basic = input.lastIndexOf(delimiter);
if (basic < 0) {
    basic = 0;
}
```

- Find the index of the rightmost `-` (the `delimiter` at `punycode.js:14`) in the input.
- If no `-` is found, `lastIndexOf` returns `-1`, so `basic` is set to `0` (empty basic prefix).
- Otherwise, `basic` is the index of the rightmost `-`.

**Effect:** Characters at indices `[0, basic)` form the basic code-point prefix and are copied verbatim to the output. Characters at indices `[basic+1, inputLength)` form the encoded delta stream.

### 3.3 Copy basic code points (`punycode.js:213-219`)

```javascript
for (let j = 0; j < basic; ++j) {
    if (input.charCodeAt(j) >= 0x80) {
        error('not-basic');
    }
    output.push(input.charCodeAt(j));
}
```

- Iterate from index `0` to `basic - 1`.
- For each character, verify it is ASCII (code point `< 0x80`). If not, throw `RangeError('not-basic')` via `error('not-basic')` at `punycode.js:216`.
- Push the code point to `output`.

**Purpose:** Pre-populate `output` with the basic ASCII characters that appear before the delimiter. These are non-encoded code points and require no further processing.

### 3.4 Main decoding loop (`punycode.js:224-278`)

The outer loop runs from `index = basic > 0 ? basic + 1 : 0` (just after the delimiter, or 0 if no delimiter) to `index < inputLength`. Each iteration decodes one Punycode-encoded insertion point and code point.

#### 3.4a Variable-length integer decoder (`punycode.js:231-261`)

This inner loop decodes a generalized variable-length base-36 integer, accumulating it into `i`:

```javascript
const oldi = i;
for (let w = 1, k = base; /* no condition */; k += base) {
    if (index >= inputLength) {
        error('invalid-input');
    }
    const digit = basicToDigit(input.charCodeAt(index++));
    if (digit >= base) {
        error('invalid-input');
    }
    if (digit > floor((maxInt - i) / w)) {
        error('overflow');
    }
    i += digit * w;
    const t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias);
    if (digit < t) {
        break;
    }
    const baseMinusT = base - t;
    if (w > floor(maxInt / baseMinusT)) {
        error('overflow');
    }
    w *= baseMinusT;
}
```

**Execution steps:**

1. **Save starting position** (`punycode.js:231`): `oldi = i`. This records the value of `i` before accumulating the delta, so we can later compute `delta = i - oldi`.

2. **Loop over base-36 positions** (`punycode.js:232`): Initialize `w = 1` and `k = base`. The variable `w` scales each decoded digit by its positional weight in the base-36 encoding. The variable `k` tracks the "position" in the variable-length encoding: `k = base, 2*base, 3*base, ...` Bias is applied to `k` to compute the threshold `t` for whether this digit is the final digit in the sequence.

3. **Bounds check** (`punycode.js:234-236`): If `index >= inputLength`, we have run past the end of the input without finding a terminator. Throw `RangeError('invalid-input')`.

4. **Decode digit** (`punycode.js:238`): Call `basicToDigit(input.charCodeAt(index++))` to convert the next character code point to a base-36 digit (0–25 for a–z/A–Z, 26–35 for 0–9). Increment `index` to advance past this character. The function `basicToDigit` (at `punycode.js:144-155`) returns `base` (36) if the character is not a valid base-36 digit.

5. **Validate digit** (`punycode.js:240-242`): If `digit >= base`, the character was not a valid base-36 symbol. Throw `RangeError('invalid-input')`.

6. **Overflow check (digit accumulation)** (`punycode.js:243-245`): Before adding `digit * w` to `i`, check that the result will not exceed `maxInt` (0x7FFFFFFF, the max signed 32-bit int). Specifically, if `digit > floor((maxInt - i) / w)`, then `i + digit * w > maxInt`. Throw `RangeError('overflow')`.

7. **Accumulate delta** (`punycode.js:247`): Add `digit * w` to `i`.

8. **Compute threshold** (`punycode.js:248`): The threshold `t` determines whether this digit is the last. It is computed as:
   - `t = tMin` (1) if `k <= bias`
   - `t = tMax` (26) if `k >= bias + tMax`
   - `t = k - bias` otherwise
   
   These bounds ensure `tMin <= t <= tMax` (i.e., `1 <= t <= 26`).

9. **Terminator test** (`punycode.js:250-252`): If `digit < t`, this was the final digit in the sequence. Break out of the inner loop.

10. **Overflow check (weight multiplication)** (`punycode.js:255-257`): Before multiplying `w` by `base - t`, check that the result will not exceed `maxInt`. Specifically, if `w > floor(maxInt / baseMinusT)`, then `w * baseMinusT > maxInt`. Throw `RangeError('overflow')`.

11. **Update weight** (`punycode.js:259`): Multiply `w` by `(base - t)` to scale the next digit's contribution.

#### 3.4b Bias adaptation (`punycode.js:263-264`)

```javascript
const out = output.length + 1;
bias = adapt(i - oldi, out, oldi == 0);
```

- `out = output.length + 1` is the total number of code points in the output so far (including the one about to be inserted).
- Call `adapt(delta, numPoints, firstTime)` (defined at `punycode.js:179-187`) to compute the updated bias. The function takes:
  - `delta = i - oldi` (the difference between the current and starting value of `i`, representing the distance the code point moved through the output array).
  - `numPoints = out` (the total count of code points).
  - `firstTime = (oldi == 0)` (whether this is the first iteration; used to apply a damping factor).
  - Returns the new bias value.

**Purpose:** The bias adapts after each insertion to reflect whether insertions are clustered close together (favor small delta values; increase bias) or spread apart (favor large delta values; decrease bias).

#### 3.4c Derive the code point (`punycode.js:268-273`)

```javascript
if (floor(i / out) > maxInt - n) {
    error('overflow');
}
n += floor(i / out);
i %= out;
```

- **Overflow guard** (`punycode.js:268-270`): Check that adding `floor(i / out)` to `n` will not exceed `maxInt`. If `floor(i / out) > maxInt - n`, throw `RangeError('overflow')`.
- **Update code point** (`punycode.js:272`): Increment `n` by `floor(i / out)`. This "unwraps" the code point that was encoded in the delta stream.
- **Reduce position** (`punycode.js:273`): Set `i = i % out`. This reduces `i` modulo `out`, the current count of code points, preparing it for the insertion step.

**Rationale:** The Punycode algorithm encodes both the code point value and its insertion position in a single delta stream. The coded value is recovered by dividing by the position count; the position is recovered by taking the remainder.

#### 3.4d Insert code point (`punycode.js:276`)

```javascript
output.splice(i++, 0, n);
```

- Call `Array.splice(i, 0, n)` to insert code point `n` at position `i` in `output` (without removing any elements).
- Increment `i` for the next iteration (though the loop will reset `i` to the new delta at the top of the next outer iteration).

**Effect:** Each code point is inserted at the position determined by the decoded delta stream.

### 3.5 Assemble and return the string (`punycode.js:280`)

```javascript
return String.fromCodePoint(...output);
```

- Call `String.fromCodePoint()` with the spread `output` array to convert the array of code points into a single Unicode string.
- Return the result.

**Note:** `String.fromCodePoint` properly handles astral code points (> 0xFFFF) by emitting the correct UTF-16 surrogate pairs.

---

## 4. Error conditions and overflow guards

### 4.1 `'not-basic'` error

**Location:** `punycode.js:215-217`

**Condition:** A character in the basic-code-point prefix (before the last `-`) has a code point >= 0x80 (non-ASCII).

**Handler:** Throw `RangeError('not-basic')` via `error('not-basic')` at `punycode.js:216`.

**Rationale:** Punycode encoding assumes basic code points are ASCII. Any non-ASCII character in the prefix is malformed.

### 4.2 `'invalid-input'` errors

**Locations:** `punycode.js:234-236`, `punycode.js:240-242`

**Conditions:**
1. The variable-length integer stream is truncated: the main loop reaches `index >= inputLength` without finding a digit with value `< threshold` to terminate the integer.
2. A character in the delta stream is not a valid base-36 digit (i.e., `basicToDigit` returns `base = 36`).

**Handler:** Throw `RangeError('invalid-input')` via `error('invalid-input')`.

**Rationale:** Both conditions indicate the input is not a valid Punycode string.

### 4.3 `'overflow'` errors

**Locations:** `punycode.js:243-245`, `punycode.js:255-257`, `punycode.js:268-270`

**Conditions:**
1. (**Digit accumulation overflow**, `punycode.js:243-245`): `digit * w` would cause `i` to exceed `maxInt = 0x7FFFFFFF` when added to the current value of `i`.
2. (**Weight multiplication overflow**, `punycode.js:255-257`): `w * (base - t)` would exceed `maxInt`.
3. (**Code point increment overflow**, `punycode.js:268-270`): Adding `floor(i / out)` to `n` would exceed `maxInt`.

**Handler:** Throw `RangeError('overflow')` via `error('overflow')`.

**Rationale:** These guards prevent integer overflow when processing pathologically large or malformed inputs, per RFC 3492.

---

## 5. Helper functions and constants

### 5.1 `basicToDigit(codePoint)` — `punycode.js:144-155`

Converts a single character code point to its base-36 digit value:

- `0x30–0x39` (digits `0–9`) → `26–35`
- `0x41–0x5A` (uppercase `A–Z`) → `0–25`
- `0x61–0x7A` (lowercase `a–z`) → `0–25`
- Any other code point → `base = 36` (invalid)

This function is called at `punycode.js:238` to decode each character in the delta stream.

### 5.2 `adapt(delta, numPoints, firstTime)` — `punycode.js:179-187`

Adjusts the bias after each insertion, returning a new bias value. Called at `punycode.js:264`.

- **Parameters:**
  - `delta`: the distance the last code point moved (`i - oldi`).
  - `numPoints`: the total count of code points in the output so far.
  - `firstTime`: boolean; `true` if this is the first iteration (applies a damping factor).
- **Returns:** the updated bias value.
- **Purpose:** Heuristically tune bias to reflect the distribution of insertions in the output, improving compression ratios.

### 5.3 `error(type)` — `punycode.js:41-43`

Throws a `RangeError` with the message corresponding to `type`:

```javascript
function error(type) {
    throw new RangeError(errors[type]);
}
```

Called at `punycode.js:216`, `punycode.js:235`, `punycode.js:241`, `punycode.js:244`, `punycode.js:256`, and `punycode.js:269` with `type` values `'not-basic'`, `'invalid-input'`, or `'overflow'`.

### 5.4 Bootstring constants — `punycode.js:7-14`

```javascript
const base = 36;        // base for variable-length integers
const tMin = 1;         // minimum threshold
const tMax = 26;        // maximum threshold
const skew = 38;        // skew parameter for bias computation
const damp = 700;       // damping factor for bias on first iteration
const initialBias = 72; // starting bias
const initialN = 128;   // starting code point (first non-basic)
const delimiter = '-';  // basic/delta separator
```

Also `maxInt = 0x7FFFFFFF` at `punycode.js:4` (max signed 32-bit int).

---

## 6. Callers

### 6.1 `toUnicode()` callback — `punycode.js:389-395`

The `toUnicode` function wraps `mapDomain` with a callback that decodes Punycode labels:

```javascript
const toUnicode = function(input) {
    return mapDomain(input, function(string) {
        return regexPunycode.test(string)
            ? decode(string.slice(4).toLowerCase())
            : string;
    });
};
```

At `punycode.js:392`, if a label matches `regexPunycode` (i.e., starts with `xn--` at `punycode.js:17`), the four-character prefix is stripped via `.slice(4)`, the result is lowercased, and `decode` is called to produce the Unicode form.

**Note:** `toUnicode` is the public API for decoding domain names; `decode` is the raw RFC 3492 decoder.

### 6.2 Public binding — `punycode.js:437`

```javascript
'decode': decode,
```

Bound in the exported object, allowing external callers to use `punycode.decode(input)`.

---

## 7. References

- **RFC 3492:** "Punycode: A Bootstring algorithm for representing Unicode with ASCII" (March 2002). Sections 3–6 define the algorithm; §6.2 specifically covers decoding.
- **Source:** https://tools.ietf.org/html/rfc3492
- **Implementation notes:** The algorithm processes a variable-length, bias-adaptive base-36 integer stream to derive both code point values and insertion positions in a single delta stream, inserting code points into an initially ASCII-only output array.
