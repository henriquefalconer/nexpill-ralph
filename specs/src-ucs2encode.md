# Spec: `punycode.ucs2.encode` — array of code points → UCS-2 string

Source: `punycode.js:125-133`

---

## 1. Subject

**Function:** Arrow function `ucs2encode` at `punycode.js:133`

**Signature:**
```javascript
const ucs2encode = codePoints => String.fromCodePoint(...codePoints);
```

**Purpose:** Transform an array of numeric Unicode code points into a UCS-2 string (JavaScript's native string format, which uses UTF-16 encoding with surrogate pairs for astral code points).

---

## 2. Contract

Derived from JSDoc at `punycode.js:125-132` and implementation at `punycode.js:133`.

**Input:** An array of numeric Unicode code points (`punycode.js:130`).

**Output:** A Unicode string in UCS-2 format (`punycode.js:131`). `String.fromCodePoint` automatically emits surrogate pairs for code points above 0xFFFF (astral code points).

**Semantics:**
- Accepts one argument: `codePoints`, an array of integers.
- Spreads the array into variadic arguments via the spread operator `...codePoints` (`punycode.js:133`).
- Delegates encoding entirely to the native `String.fromCodePoint` function.
- Does not validate the input array; invalid or out-of-range code points are handled (or rejected) by `String.fromCodePoint`.
- Does not mutate the input array.

**Public binding:** Exposed as `punycode.ucs2.encode` at `punycode.js:435`.

---

## 3. Implementation

The function at `punycode.js:133` is a single-line arrow function:

```javascript
const ucs2encode = codePoints => String.fromCodePoint(...codePoints);
```

**Mechanism:**
1. The spread operator `...codePoints` unpacks the array into a comma-separated list of arguments.
2. Each argument is passed to `String.fromCodePoint`.
3. `String.fromCodePoint` converts each code point to its corresponding character(s):
   - Code points 0x0–0xFFFF (BMP) → single code unit.
   - Code points 0x10000–0x10FFFF (astral) → surrogate pair (two code units).
4. All characters are concatenated into a single string and returned.

**No validation:** The function performs no checks on the input array or its elements. `String.fromCodePoint` enforces validity rules (see edge cases below).

---

## 4. Edge cases

### 4.1 Empty array
**Input:** `[]`
**Output:** Empty string `''`
**Reason:** `String.fromCodePoint()` with no arguments returns an empty string.

### 4.2 Code points in surrogate range (0xD800–0xDFFF)
**Input:** `[0xD800]` or any code point in 0xD800–0xDFFF
**Output:** A string containing that surrogate code unit (e.g., `'\uD800'`)
**Behavior:** `String.fromCodePoint` emits a lone surrogate without validation or rejection. This is a deliberate deviation from strict UTF-16 validity; the function treats its input as raw UCS-2 code units and makes no attempt to validate unpaired surrogates.

### 4.3 Code points exceeding maximum Unicode value (> 0x10FFFF)
**Input:** `[0x110000]` or any value > 0x10FFFF
**Output:** Throws `RangeError`
**Reason:** `String.fromCodePoint` validates all arguments and throws `RangeError` for values outside the range [0x0, 0x10FFFF].

### 4.4 Very large arrays
**Input:** Arrays with millions of elements (exact threshold depends on the JavaScript engine)
**Output:** May throw `RangeError` due to exceeding maximum function argument count
**Reason:** The spread operator `...codePoints` converts the array into a variadic argument list. Most JavaScript engines have a practical limit on the number of arguments a function can accept (commonly 65,536 or similar). Exceeding this limit results in a `RangeError`.

---

## 5. Callers

**Direct internal calls:** None. `ucs2encode` is not called by any other internal function in `punycode.js`.

**Public API:** `punycode.ucs2.encode` at `punycode.js:435` is the sole binding and entry point.

**Redundancy note:** The `decode` function at `punycode.js:280` also uses `String.fromCodePoint` directly (`return String.fromCodePoint(...output);`) rather than calling `ucs2encode`. This represents a subtle redundancy: both functions delegate to the same built-in, and `decode` could theoretically call `ucs2encode` instead. However, the direct call is retained, likely for simplicity and to avoid an extra function call overhead.

---

## 6. Implementation citations

| Item | Location |
|---|---|
| Function definition | `punycode.js:133` |
| JSDoc contract | `punycode.js:125-132` |
| Parameter description (`codePoints`) | `punycode.js:130` |
| Return type description (UCS-2 string) | `punycode.js:131` |
| Public binding (`punycode.ucs2.encode`) | `punycode.js:435` |
| Redundant use in `decode` function | `punycode.js:280` |

