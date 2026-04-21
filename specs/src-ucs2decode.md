# Spec: `ucs2decode(string)` — Surrogate-pair-aware UCS-2 to code-points decoder

Source function: `punycode.js:101-123`

---

## 1. Subject

**Function:** `ucs2decode(string)` converts a UCS-2 string (JavaScript's native UTF-16 / UCS-2 representation) into an array of Unicode code points, combining surrogate pairs into astral code points (≥ 0x10000) while preserving unmatched surrogates as raw code-unit values.

**Binding:** `punycode.ucs2.decode = ucs2decode` at `punycode.js:434`, inside the public `punycode.ucs2` object declared at `punycode.js:433-436`.

---

## 2. Contract

Derived from the JSDoc at `punycode.js:88-100` and the implementation at `punycode.js:101-123`.

### Input

- **Type:** JavaScript string (UCS-2 / UTF-16 code units)
- **Semantics:** a sequence of 16-bit unsigned integers (`charCodeAt` values), including high surrogates (0xD800–0xDBFF), low surrogates (0xDC00–0xDFFF), and non-surrogate code units (BMP characters).

### Output

- **Type:** Array of non-negative integers
- **Semantics:** An array where each entry is either:
  - A Unicode scalar value (≥ 0x10000) formed from a high-low surrogate pair; or
  - A raw code-unit value (including unmatched surrogates or BMP characters).

### Behavior summary

The function does not throw on malformed input. It iterates over code units once with a mutable cursor (`counter`), examining each code unit in order:

1. If the current code unit is a **high surrogate (0xD800–0xDBFF)** and is immediately followed by a **low surrogate (0xDC00–0xDFFF)**, the pair is combined into a single astral code point via the formula at `punycode.js:111`. Both code units are consumed.

2. If the current code unit is a **high surrogate** but is **not** followed by a low surrogate, the high-surrogate code unit is emitted as-is, the cursor is decremented so the following code unit is re-examined in the next iteration, and iteration continues.

3. In all other cases (non-surrogates, lone low surrogates), the code unit is emitted directly without special handling.

---

## 3. Implementation walkthrough

### 3.1 Initialization (`punycode.js:101-104`)

```javascript
function ucs2decode(string) {
	const output = [];
	let counter = 0;
	const length = string.length;
```

- `output` accumulates the result code points.
- `counter` tracks the current position in the input string.
- `length` caches `string.length` to avoid repeated property access.

### 3.2 Main loop (`punycode.js:105-121`)

```javascript
	while (counter < length) {
		const value = string.charCodeAt(counter++);
```

For each iteration, the code unit at `counter` is read via `charCodeAt` at `punycode.js:106`, and `counter` is immediately post-incremented.

### 3.3 High-surrogate branch (`punycode.js:107-117`)

```javascript
		if (value >= 0xD800 && value <= 0xDBFF && counter < length) {
			// It's a high surrogate, and there is a next character.
			const extra = string.charCodeAt(counter++);
			if ((extra & 0xFC00) == 0xDC00) { // Low surrogate.
				output.push(((value & 0x3FF) << 10) + (extra & 0x3FF) + 0x10000);
			} else {
				// It's an unmatched surrogate; only append this code unit, in case the
				// next code unit is the high surrogate of a surrogate pair.
				output.push(value);
				counter--;
			}
```

**Condition at `:107`:** The code unit is a high surrogate (range 0xD800–0xDBFF) and a next code unit is available (`counter < length`).

**Valid pair (`:108-111`):** The following code unit is read at `:109` and tested against the low-surrogate mask `(extra & 0xFC00) == 0xDC00` at `:110`. If it is a low surrogate:

- Extract the 10-bit payload from the high surrogate: `value & 0x3FF`
- Extract the 10-bit payload from the low surrogate: `extra & 0x3FF`
- Combine via `((value & 0x3FF) << 10) + (extra & 0x3FF) + 0x10000` at `:111`
- Push the result; both code units have been consumed.

**Unmatched high surrogate (`:112-117`):** If the following code unit is not a low surrogate:

- Emit the high-surrogate code unit as-is at `:115`
- Decrement `counter` at `:116` so the next iteration re-examines the code unit that was not a low surrogate.

### 3.4 Non-surrogate / lone low surrogate branch (`punycode.js:118-120`)

```javascript
		} else {
			output.push(value);
		}
```

Any code unit outside the high-surrogate range (including non-surrogates and lone low surrogates) is pushed directly.

### 3.5 Return (`punycode.js:122`)

```javascript
	return output;
```

---

## 4. Edge cases and invariants

### 4.1 Lone high surrogate at end of string

- **Input:** `'\uD800'` (high surrogate, no following code unit)
- **Condition:** `value >= 0xD800 && value <= 0xDBFF` is true, but `counter < length` is false (no next character).
- **Behavior:** The condition at `:107` fails, control falls through to `:118-120`, and the high-surrogate code unit (0xD800) is emitted as-is.
- **Output:** `[0xD800]` (55296)

### 4.2 Lone low surrogate

- **Input:** `'\uDC00'` (low surrogate)
- **Condition:** `value >= 0xD800 && value <= 0xDBFF` is false (not in high-surrogate range).
- **Behavior:** Control falls through to `:118-120` and the low-surrogate code unit (0xDC00) is emitted directly.
- **Output:** `[0xDC00]` (56320)

### 4.3 High surrogate followed by non-low-surrogate

- **Input:** `'\uD800\uD800'` (two high surrogates back-to-back)
- **Iteration 1:** `value = 0xD800`, next character is `0xD800`. Condition at `:107` is true (high surrogate exists and next exists). Read `extra = 0xD800` at `:109`. Test `(0xD800 & 0xFC00) == 0xDC00` at `:110` yields false (not a low surrogate). Emit `0xD800` at `:115`, decrement counter at `:116`.
- **Iteration 2:** `value = 0xD800` (re-examined). Condition at `:107` is true. Read `extra = ?` at `:109` (depends on what follows). If nothing follows or the next is not a low surrogate, emit `0xD800` again.
- **Output:** `[0xD800, 0xD800]`

### 4.4 Unmatched high surrogate, surrogate pair, unmatched high surrogate

- **Input:** `'\uD800\uD834\uDF06\uD800'` (as per `specs/test-ucs2-decode.md`, Vector 5)
- **Iteration 1:** `value = 0xD800`, next is `0xD834`. Read `0xD834` at `:109`. Test `(0xD834 & 0xFC00) == 0xDC00` is false. Emit `0xD800` at `:115`, decrement at `:116`.
- **Iteration 2:** `value = 0xD834`, next is `0xDF06`. Test succeeds; combine to astral code point 0x1D306 at `:111`.
- **Iteration 3:** `value = 0xD800`, no next character. Condition `:107` fails; emit `0xD800` via `:118-120`.
- **Output:** `[0xD800, 0x1D306, 0xD800]`

### 4.5 Empty string

- **Input:** `''` (empty)
- **Behavior:** `length = 0`, loop condition `counter < length` is never true, output is never populated.
- **Output:** `[]`

---

## 5. Callers

### 5.1 `encode(input)` at `punycode.js:290-294`

The `encode` function calls `ucs2decode(input)` at `:294` to normalize the input string into an array of Unicode code points before applying the Punycode encoding algorithm.

---

## 6. References

- JSDoc citation: `punycode.js:88-100` (function definition with JSDoc)
- **Encoding reference:** <https://mathiasbynens.be/notes/javascript-encoding> (cited in JSDoc at `:95`)
- **Public binding:** `punycode.js:434`
- **Test specification:** `specs/test-ucs2-decode.md`

---

## 7. Implementation citations

| Concept | Location |
|---|---|
| Function definition | `punycode.js:101-123` |
| JSDoc | `punycode.js:88-100` |
| Public binding `punycode.ucs2.decode` | `punycode.js:434` |
| Initialization | `punycode.js:102-104` |
| Main loop start | `punycode.js:105` |
| Code unit read | `punycode.js:106` |
| High-surrogate condition | `punycode.js:107` |
| Low-surrogate read | `punycode.js:109` |
| Low-surrogate test mask | `punycode.js:110` |
| Surrogate pair combination formula | `punycode.js:111` |
| Unmatched surrogate emission | `punycode.js:115` |
| Cursor rollback | `punycode.js:116` |
| Non-surrogate emission | `punycode.js:119` |
| Return statement | `punycode.js:122` |
| Caller: `encode` function | `punycode.js:294` |
