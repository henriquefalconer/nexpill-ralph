# Spec: `digitToBasic(digit, flag)` — Punycode digit to ASCII code point

Source location: `punycode.js:157-172`

---

## 1. Subject

**Function:** `digitToBasic(digit, flag)` — the inverse of `basicToDigit()` (`punycode.js:144-155`).

**Purpose:** Maps a Punycode digit in the range `[0, 35]` back to an ASCII code point. A bit-flag parameter controls whether the output character should be in uppercase (for alphabetic digits only) or lowercase.

**Callsites:** This function is called exactly twice, both within the `encode` function:
- `punycode.js:359` — `digitToBasic(t + qMinusT % baseMinusT, 0)` to emit intermediate digits of a generalized variable-length integer.
- `punycode.js:364` — `digitToBasic(q, 0)` to emit the final terminator digit.

Both callsites pass `flag = 0`, ensuring the output is always lowercase ASCII.

---

## 2. Contract

**JSDoc:** `punycode.js:157-167`

**Signature:**
```javascript
digitToBasic(digit, flag)
```

**Inputs:**
- `digit` (Number): A numeric value in the range `[0, 35]`, representing a digit in the Punycode base-36 alphabet.
- `flag` (Number): A control flag. Non-zero requests uppercase output; zero (or falsy) requests lowercase. Only meaningful for `digit` in `[0, 25]` (alphabetic).

**Output:**
- Returns a Number (ASCII code point).
- For `digit` in `[0, 25]`: returns ASCII code for `a..z` (code points 0x61–0x7A = 97–122) or `A..Z` (code points 0x41–0x5A = 65–90), depending on `flag`.
- For `digit` in `[26, 35]`: returns ASCII code for `0..9` (code points 0x30–0x39 = 48–57); `flag` is ignored.

**Undefined behavior:**
- If `flag != 0` and `digit >= 26`, the behavior is undefined per the JSDoc. Numeric digits (`0..9`) have no uppercase form; passing a non-zero `flag` with such a digit yields control characters (see §3).

---

## 3. Implementation arithmetic

**Function code:** `punycode.js:168-172`

```javascript
const digitToBasic = function(digit, flag) {
	//  0..25 map to ASCII a..z or A..Z
	// 26..35 map to ASCII 0..9
	return digit + 22 + 75 * (digit < 26) - ((flag != 0) << 5);
};
```

**The one-line formula at line 171:**
```
digit + 22 + 75 * (digit < 26) - ((flag != 0) << 5)
```

**Explanation:**

The formula collapses two conditional branches into one expression, exploiting ASCII digit and letter offsets.

**Case 1: Alphabetic digits (digit < 26)**

When `digit < 26` (true, coerced to `1`):
- The term `75 * (digit < 26)` evaluates to `75 * 1 = 75`.
- Result before flag adjustment: `digit + 22 + 75 = digit + 97`.
- Since ASCII `'a'` = 0x61 = 97, this becomes `digit + 'a'` code, mapping `[0, 25]` to ASCII `[97, 122]` = `'a'..'z'`.
- If `flag != 0` (true, coerced to `1`): subtract `((flag != 0) << 5) = 1 << 5 = 32` (the ASCII offset from lowercase to uppercase).
  - Result: `digit + 97 - 32 = digit + 65`, which is `digit + 'A'` code, mapping `[0, 25]` to ASCII `[65, 90]` = `'A'..'Z'`.
- If `flag == 0`: the subtracted term is `0`, so output is `digit + 97` = `'a'..'z'`.

**Case 2: Numeric digits (digit >= 26)**

When `digit >= 26` (false, coerced to `0`):
- The term `75 * (digit < 26)` evaluates to `75 * 0 = 0`.
- Result before flag adjustment: `digit + 22 + 0 = digit + 22`.
- Since ASCII `'0'` = 0x30 = 48, and `digit - 26` ranges over `[0, 9]`, we have `digit + 22 = (digit - 26) + 48 = (digit - 26) + '0'` code.
  - This maps `digit ∈ [26, 35]` to ASCII `[48, 57]` = `'0'..'9'`.
- The `flag` term still subtracts `((flag != 0) << 5)` if non-zero, but doing so on numeric digits yields control characters (code points in the range `[16, 25]` = 0x10–0x19 = DLE–EM), which is meaningless and forbidden by the contract's "undefined behavior" clause.

**Why this works:**
- ASCII lowercase letters (`a–z`) occupy 0x61–0x7A (97–122).
- ASCII uppercase letters (`A–Z`) occupy 0x41–0x5A (65–90).
- ASCII digits (`0–9`) occupy 0x30–0x39 (48–57).
- The offsets 22, 75, 32, and the boolean multiplier exploit these gaps to produce the correct codes in one arithmetic expression.

---

## 4. Relationship to `basicToDigit()`

**Inverse function:** `basicToDigit()` (`punycode.js:144-155`)

`basicToDigit()` maps an ASCII code point back to a digit:
- 0x30–0x39 (`'0'..'9'`) → 26–35
- 0x41–0x5A (`'A'..'Z'`) → 0–25
- 0x61–0x7B (`'a'..'z'`) → 0–25
- Anything else → `base` (36, an invalid sentinel)

`digitToBasic()` is the inverse: given a digit and a flag, it produces the corresponding ASCII code point, ignoring case-folding on digits.

---

## 5. Call sites and usage

### Call site 1: `punycode.js:359`

```javascript
stringFromCharCode(digitToBasic(t + qMinusT % baseMinusT, 0))
```

**Context:** This line is inside the variable-length integer encoding loop of the `encode` function (`punycode.js:302–369`). It emits intermediate digits of the encoded run of deltas.

**Arguments:**
- `digit = t + qMinusT % baseMinusT`: a computed digit in `[0, 35]`.
- `flag = 0`: always requests lowercase output.

### Call site 2: `punycode.js:364`

```javascript
output.push(stringFromCharCode(digitToBasic(q, 0)))
```

**Context:** This line immediately follows the loop at line 359, emitting the final terminator digit of the run.

**Arguments:**
- `digit = q`: the remaining quotient, in `[0, 35]`.
- `flag = 0`: always requests lowercase output.

---

## 6. Edge cases

**All callsites use `flag = 0`:** Both `punycode.js:359` and `punycode.js:364` pass `flag = 0`, so uppercase conversion never occurs in practice during encoding. The `flag` parameter exists for API completeness and potential future use.

**Undefined behavior with `flag != 0` and `digit >= 26`:**
If called with `flag != 0` and `digit ∈ [26, 35]`, the arithmetic produces:
```
digit + 22 - 32 = digit - 10 ∈ [16, 25] = 0x10–0x19 = DLE–EM (control characters)
```
These control characters are not valid Punycode output and should never be produced. Implementations should avoid this condition, or document that doing so is undefined.

---

## 7. RFC 3492 reference

This function implements the inverse of the digit-to-code-point mapping defined in RFC 3492 §5, specifically section 5.4 ("Decoding Punycode"):

> Increment `n` by the product of `delta` and (`base` minus `t`), where `t` is the weight of the most recently inserted code point; note that the Punycode digits for the default initial values in section 3.4 are:
>
> ```
> 0..25 map to ASCII a–z
> 26..35 map to ASCII 0–9
> ```

The function encodes these mappings in both directions, with `digitToBasic()` serving as the forward direction (digit → code point) during encoding.

---

## 8. Implementation citations

| Item | Location |
|---|---|
| Function definition | `punycode.js:168-172` |
| JSDoc / contract | `punycode.js:157-167` |
| Inverse function `basicToDigit()` | `punycode.js:144-155` |
| Call site (intermediate digit) | `punycode.js:359` |
| Call site (terminator digit) | `punycode.js:364` |
| Encoding function context | `punycode.js:302-369` |
