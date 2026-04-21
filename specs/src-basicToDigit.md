# Spec: `basicToDigit(codePoint)` — Punycode character-to-digit converter

Source function: `punycode.js:135-155`

---

## 1. Subject

**Function:** `basicToDigit`

**Purpose:** Inverse of `digitToBasic` (`punycode.js:157-175`). Maps an ASCII code point to its Punycode digit value in the range `[0, 36)`, or returns `base` (36) as a sentinel for "not a valid Punycode digit."

**Context:** Used during Punycode decoding to convert ASCII characters in the variable-length-integer portion of an encoded label back into their numeric Punycode digit values. See `decode` call site at `punycode.js:238`.

---

## 2. Contract

**JSDoc (lines 136–142):**

- **Input:** numeric code point (non-negative integer, typically in the ASCII range 0–127).
- **Output:** 
  - Integer in `[0, base - 1]` (i.e., 0–35) if the code point is a valid Punycode basic digit (a letter A–Z, a–z, or digit 0–9).
  - `base` (36) if the code point is not a valid Punycode digit.

**Interpretation:** The function never throws or raises an error; invalid code points are handled by the caller (see **Callsites and error handling**, below).

---

## 3. Character mapping

The function maps three ranges of ASCII code points to Punycode digits:

### 3.1 Digits 0–9 map to digit values 26–35

**Condition (line 145):** `codePoint >= 0x30 && codePoint < 0x3A`

**Mapping (line 146):** `return 26 + (codePoint - 0x30)`

- `0x30`–`0x39` are the ASCII code points for the characters `'0'`–`'9'`.
- Subtracting `0x30` yields 0–9.
- Adding 26 yields 26–35.
- Thus `'0'` → 26, `'1'` → 27, …, `'9'` → 35.

### 3.2 Uppercase letters A–Z map to digit values 0–25

**Condition (line 148):** `codePoint >= 0x41 && codePoint < 0x5B`

**Mapping (line 149):** `return codePoint - 0x41`

- `0x41`–`0x5A` are the ASCII code points for the characters `'A'`–`'Z'`.
- Subtracting `0x41` yields 0–25.
- Thus `'A'` → 0, `'B'` → 1, …, `'Z'` → 25.

### 3.3 Lowercase letters a–z map to digit values 0–25

**Condition (line 151):** `codePoint >= 0x61 && codePoint < 0x7B`

**Mapping (line 152):** `return codePoint - 0x61`

- `0x61`–`0x7A` are the ASCII code points for the characters `'a'`–`'z'`.
- Subtracting `0x61` yields 0–25.
- Thus `'a'` → 0, `'b'` → 1, …, `'z'` → 25.

### 3.4 Everything else returns base (36)

**Default (line 154):** `return base` (which is 36, defined at `punycode.js:7`)

Any code point that does not fall into the three ranges above — including non-ASCII characters (>= 128), control characters, punctuation (`!`, `@`, `-`, `.`, etc.), and whitespace — is rejected and returns `base`.

---

## 4. Case-insensitivity

Both uppercase and lowercase letters map to the **same digit range 0–25**. This means:
- `'A'` and `'a'` both map to digit 0.
- `'B'` and `'b'` both map to digit 1.
- And so on.

**Decoder normalization:** The `decode` function (line 392) calls `toLowerCase()` before passing the input to the main decode loop, so the input is already lowercase by the time `basicToDigit` is called. However, `basicToDigit` itself tolerates both cases, providing a defensive layer: if a caller were to pass an uppercase letter, it would still be correctly decoded.

---

## 5. Callsites and error handling

**Call site:** `punycode.js:238` — within the main variable-length-integer decoding loop in the `decode` function.

```javascript
const digit = basicToDigit(input.charCodeAt(index++));

if (digit >= base) {
  error('invalid-input');
}
```

(Lines 238–242.)

**Error path:** If `digit >= base` (i.e., the character is not a valid Punycode digit), the decoder throws a `RangeError` with message `'Invalid input'` (from `errors['invalid-input']` at `punycode.js:25`).

---

## 6. Edge cases

### 6.1 Non-Punycode-digit basic code points are rejected

Characters such as:
- Hyphen/minus (`-`, 0x2D)
- Period (`.`, 0x2E)
- Underscore (`_`, 0x5F)
- Exclamation mark (`!`, 0x21)
- At sign (`@`, 0x40)
- Other punctuation or symbols

All fall outside the three mapped ranges and return `base`, triggering an error when encountered in the variable-length-integer portion of a Punycode label. This ensures that malformed or unexpected input is rejected early.

### 6.2 Non-ASCII characters (>= 0x80) are rejected

Characters with code points 0x80 and above are not in any of the three mapped ranges and return `base`. However, Punycode decoders typically also have an earlier check (at `punycode.js:215–216`) that rejects any non-ASCII character in the basic part of the encoded label before it ever reaches `basicToDigit`.

### 6.3 Control characters and whitespace are rejected

Codes 0x00–0x1F (control characters) and 0x7F (DEL), as well as space (0x20) and other whitespace, are outside the three ranges and return `base`.

---

## 7. References

**RFC 3492 Section 5** — "Parameter values for Punycode":

> The basic code point set is the ASCII letters and digits: A–Z (0x41–0x5A), a–z (0x61–0x7A), and 0–9 (0x30–0x39). In a Punycode string, these characters represent digit values 0–25 (letters) and 26–35 (digits) respectively.

**Implementation constants:**
- `base` constant: `punycode.js:7` (value: 36)

---

## 8. Implementation notes

The function is declared as a const arrow function at `punycode.js:144`:

```javascript
const basicToDigit = function(codePoint) { … };
```

**Strictness:**

- No input validation or error handling; relies on the caller (`decode`) to check the returned value against `base`.
- No side effects; the function is pure and deterministic.
- No overflow or underflow risks; the return value is always in `[0, 36]`.

**Performance:**

The function uses a series of range checks and arithmetic operations. In JavaScript engines with branch prediction, the most common cases (digits and lowercase letters) should inline efficiently.

