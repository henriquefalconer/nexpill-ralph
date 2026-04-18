# punycode-internal-helpers

## Purpose

This spec documents three private helper functions that are used internally throughout `punycode.js` but are not part of the public API. None of these functions are exported. A Go port must implement equivalent logic, though the exact function signatures may differ.

## `error(type)` — Throw a named RangeError

**Declaration:** `punycode.js:41–43`

### Signature

- **Input:** `type` — a string key; must be one of `'overflow'`, `'not-basic'`, or `'invalid-input'`.
- **Output:** Never returns. Always throws.

### Behavior

Looks up the error key `type` in the module-level `errors` table (punycode.js:22–26) and immediately throws a `RangeError` whose message is the looked-up string. The three possible messages are:

| `type` value | Thrown message |
|---|---|
| `'overflow'` | `'Overflow: input needs wider integers to process'` |
| `'not-basic'` | `'Illegal input >= 0x80 (not a basic code point)'` |
| `'invalid-input'` | `'Invalid input'` |

In a Go port, the equivalent is to return an error value (e.g., `errors.New(...)` or a sentinel `var`) using the exact message strings above. Callers of this helper — `decode` and `encode` — must propagate the error to their own callers.

### Call Sites

- `decode` calls `error('not-basic')` when a pre-delimiter character has code point >= 0x80 (punycode.js:216–217).
- `decode` calls `error('invalid-input')` when input ends mid-integer or when a character is not a valid Bootstring digit (punycode.js:235–236, 241–242).
- `decode` calls `error('overflow')` when any arithmetic would exceed `maxInt` (punycode.js:244–245, 256–257, 269–270).
- `encode` calls `error('overflow')` in two places during delta computation (punycode.js:337–338, 345–346).

See `punycode-decode.md` and `punycode-encode.md` for full error-condition details.

---

## `map(array, callback)` — Reverse-iterating array map

**Declaration:** `punycode.js:53–60`

### Signature

- **Input:** `array` — any array-like object with a `length` property and integer-keyed elements.
- **Input:** `callback` — a function that accepts one element and returns a transformed value.
- **Output:** A new array of the same length, where each element is the result of applying `callback` to the corresponding element of `array`.

### Behavior

The function creates a new result array and fills it by iterating the input array from the last element down to the first (punycode.js:56–58):

1. Set `length` to `array.length`.
2. While `length` is greater than zero:
   a. Decrement `length` by 1.
   b. Set `result[length]` to `callback(array[length])`.
3. Return `result`.

The iteration order is from index `length - 1` down to index `0`. Because the output is written at the same index as the input is read, the result array ends up with elements in the same positional order as the input array — the reverse iteration does not reverse the output. However, the callback is invoked in reverse order: the callback for the last element is called first, and the callback for the first element is called last.

**Consequence for side-effecting callbacks:** If the callback function has side effects that depend on call order, those effects will occur in reverse order relative to the input array. In practice, `mapDomain` passes `map` a callback that encodes each domain label independently, so this ordering difference has no observable effect on the final result.

**Implementation note:** This function is a private reimplementation of `Array.prototype.map`. It was written to avoid compatibility issues with older environments. In a Go port, use a standard forward-iterating loop; the callback in every call site is pure (no cross-element side effects), so the iteration order does not matter for correctness.

### Call Sites

- `mapDomain` (punycode.js:84) passes an array of domain labels and an encoding/decoding callback. The result is an array of transformed labels that are then joined with `.`.

---

## `mapDomain(domain, callback)` — Domain-label iterator with email and separator support

**Declaration:** `punycode.js:72–86`

### Signature

- **Input:** `domain` — a string containing a domain name or an email address.
- **Input:** `callback` — a function that accepts a single domain label (string) and returns a transformed label (string).
- **Output:** A string with each domain label replaced by the result of calling `callback` on it, with label separators normalized to U+002E.

### Behavior

The function performs the following steps in order:

#### Step 1: Extract email local part (punycode.js:73–80)

Split the input string on the `@` character (using the first occurrence only, since `.split('@')` in the underlying implementation splits on all occurrences and the result is then handled by checking `parts.length > 1`).

- If the split produces more than one part (i.e., the input contains at least one `@`):
  - Prepend `parts[0] + '@'` to the result and discard everything before and including the first `@`.
  - Set the domain to process to `parts[1]` (i.e., everything after the first `@`).
  - The local part is preserved verbatim; it is never passed through `callback`.
- If there is no `@`, process the entire input as a domain.

**Rationale (punycode.js:76–78):** In email addresses, only the domain name portion (after `@`) should be Punycode-processed. The local part (everything before `@`) is left intact. This is consistent with IDNA2003 and RFC 5321.

#### Step 2: Normalize separators (punycode.js:82)

Replace every occurrence of any IDNA2003 separator character in the domain with U+002E (full stop). The four recognized separators are defined by `regexSeparators` (punycode.js:19): U+002E, U+3002 (ideographic full stop), U+FF0E (fullwidth full stop), and U+FF61 (halfwidth ideographic full stop).

**Implementation note (punycode.js:81):** The source uses `.replace(regexSeparators, '\x2E')` rather than `.split(regex)` to avoid a known bug in Internet Explorer 8 that drops empty strings when splitting on a regex. In a Go port, a simple `strings.NewReplacer` or regex replacement achieves the same result without this concern.

#### Step 3: Split into labels (punycode.js:83)

Split the normalized domain on the literal `.` character (U+002E). This produces an array of label strings. An empty label will be produced for a trailing dot (e.g., `"example.com."` splits into `["example", "com", ""]`); this empty label is preserved and passed to `callback`, which must handle it.

#### Step 4: Map each label through callback (punycode.js:84)

Apply `callback` to each label using the `map` helper. Each label is processed independently. The result is an array of transformed labels.

#### Step 5: Rejoin and prepend local part (punycode.js:84–85)

Join the transformed labels with `'.'` (U+002E). Prepend the email local part (including `@`) if one was extracted in Step 1. Return the combined string.

### Call Sites

- `toUnicode` (punycode.js:390–394): passes a callback that tests each label against `regexPunycode` and, if it matches, strips the `xn--` prefix, lowercases the remainder, and calls `decode`. See `punycode-to-unicode.md`.
- `toASCII` (punycode.js:409–413): passes a callback that tests each label against `regexNonASCII` and, if it matches, prepends `xn--` and calls `encode`. See `punycode-to-ascii.md`.

## Cross-References

- [punycode-module.md](./punycode-module.md) — `errors` table, `regexSeparators`, and `regexNonASCII` patterns referenced here
- [punycode-decode.md](./punycode-decode.md) — full list of error conditions thrown via `error()`
- [punycode-encode.md](./punycode-encode.md) — overflow conditions thrown via `error()`
- [punycode-to-ascii.md](./punycode-to-ascii.md) — uses `mapDomain` for label-by-label encoding
- [punycode-to-unicode.md](./punycode-to-unicode.md) — uses `mapDomain` for label-by-label decoding
