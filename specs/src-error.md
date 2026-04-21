# Spec: `error(type)` private helper

Source function: `punycode.js:35-43`

---

## 1. Subject

**Function:** `error(type)` — a private error-throwing utility.

**Binding:** `error` is a module-scoped function defined at `punycode.js:35-43`, not exported, and used exclusively within the same module to signal error conditions.

**Signature:**

- Input: `type` — a string that must match a key in the `errors` object (`punycode.js:22-26`).
- Output: Never returns; always throws a `RangeError`.

---

## 2. Contract

Derived from the JSDoc at `punycode.js:35-40`.

The `error` function accepts a string error type key. It looks up the corresponding user-facing message in the `errors` object and throws a `RangeError` with that message. The function never returns normally; it is tail-called for all error conditions in the decoder and encoder.

The `errors` object (`punycode.js:22-26`) defines three error types:

| Error Key | Message |
|---|---|
| `'overflow'` | `'Overflow: input needs wider integers to process'` |
| `'not-basic'` | `'Illegal input >= 0x80 (not a basic code point)'` |
| `'invalid-input'` | `'Invalid input'` |

The function always uses `RangeError` as the error constructor (`punycode.js:42`), regardless of the error type.

---

## 3. Implementation

The implementation is a single-line throw statement at `punycode.js:42`:

```javascript
throw new RangeError(errors[type]);
```

The caller provides the error type string (e.g., `'overflow'`), which serves as a lookup key into the `errors` object. The corresponding message string is retrieved and passed to the `RangeError` constructor. If the provided type does not exist as a key in `errors`, the expression evaluates to `undefined`, yielding `RangeError: undefined`. However, this case does not occur in the codebase: all call sites pass valid keys.

---

## 4. Call sites

All invocations of `error(` within `punycode.js`:

#### `'not-basic'`

- `punycode.js:216` — inside the basic-code-point validation loop in `decode`. Thrown when any pre-delimiter input character has code point >= 0x80.

#### `'invalid-input'`

- `punycode.js:235` — inside the main decoding loop in `decode`. Thrown when the loop reaches the end of input while decoding a variable-length integer.
- `punycode.js:241` — inside the main decoding loop in `decode`. Thrown when a digit returned by `basicToDigit` equals or exceeds the `base` value (36), indicating invalid input.

#### `'overflow'`

- `punycode.js:244` — inside the main decoding loop in `decode`. Thrown when the increment `digit * w` would exceed `maxInt`, signaling integer overflow in the Punycode accumulator.
- `punycode.js:256` — inside the main decoding loop in `decode`. Thrown when the multiplier `w` would exceed `maxInt / baseMinusT` upon the next iteration, signaling overflow risk before the multiplication.
- `punycode.js:269` — inside the main decoding loop in `decode`. Thrown when the computed output length `floor(i / out)` exceeds `maxInt - n`, signaling overflow when updating the code point counter.
- `punycode.js:338` — inside the main encoding loop in `encode`. Thrown when the delta increment would exceed `maxInt` when handling the next code point range.
- `punycode.js:346` — inside the main encoding loop in `encode`. Thrown when incrementing the bias-adjusted position delta exceeds `maxInt`.

**Total: 8 call sites** (1 `'not-basic'`, 2 `'invalid-input'`, 5 `'overflow'`)

---

## 5. Observable behavior

When invoked, `error(type)` immediately throws a `RangeError` object. The error's `message` property contains the user-facing message from the `errors` table.

**Observable error messages:**

1. `'Overflow: input needs wider integers to process'` — thrown by overflow guards in `decode` and `encode`.
2. `'Illegal input >= 0x80 (not a basic code point)'` — thrown when a pre-delimiter character is out of the basic ASCII range.
3. `'Invalid input'` — thrown when a character cannot be decoded to a valid base-36 digit or when input ends prematurely.

All thrown errors are instances of the native `RangeError` class, making them catchable via `try/catch` blocks that filter on `RangeError`.

---

## 6. Edge cases

**Unknown error type:** If the caller were to invoke `error(unknownKey)`, the expression `errors[unknownKey]` would evaluate to `undefined`, and the thrown error would have message `'undefined'`. This case is not exercised in the source code, as all 8 call sites pass one of the three valid keys: `'overflow'`, `'not-basic'`, or `'invalid-input'`.

**No recovery path:** Because `error(type)` never returns, callers must assume the function is a terminator. There is no mechanism within the function to recover or suppress the error; all error conditions in the codebase are truly fatal.

---

## Implementation citations

| Citation | Location |
|---|---|
| `error` function definition | `punycode.js:35-43` |
| `errors` object | `punycode.js:22-26` |
| JSDoc for `error` | `punycode.js:35-40` |
| Error throw statement | `punycode.js:42` |
| `'overflow'` message string | `punycode.js:23` |
| `'not-basic'` message string | `punycode.js:24` |
| `'invalid-input'` message string | `punycode.js:25` |
| Call site: `'not-basic'` at decode line 216 | `punycode.js:216` |
| Call site: `'invalid-input'` at decode line 235 | `punycode.js:235` |
| Call site: `'invalid-input'` at decode line 241 | `punycode.js:241` |
| Call site: `'overflow'` at decode line 244 | `punycode.js:244` |
| Call site: `'overflow'` at decode line 256 | `punycode.js:256` |
| Call site: `'overflow'` at decode line 269 | `punycode.js:269` |
| Call site: `'overflow'` at encode line 338 | `punycode.js:338` |
| Call site: `'overflow'` at encode line 346 | `punycode.js:346` |

