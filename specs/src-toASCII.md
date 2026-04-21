# Spec: `punycode.toASCII()` — High-Level Unicode Domain/Email to Punycode Converter

**Source:** `punycode.js:397-414`  
**Public binding:** `punycode.toASCII` at `punycode.js:439`

---

## 1. Subject

`toASCII(input)` is a high-level public function that converts a Unicode domain name or email address to its ASCII-Compatible Encoding (ACE) Punycode form. Each label (dot-separated segment) containing at least one non-ASCII character is passed through the `encode` function and prefixed with `xn--`. Labels that are already purely ASCII pass through unchanged, making the function idempotent against ASCII-only input.

---

## 2. Contract

Derived from the JSDoc at `punycode.js:397-407`:

- **Input:** A Unicode string representing a domain name or email address (`punycode.js:403-404`).
- **Output:** The Punycode (ACE) representation of the given domain name or email address (`punycode.js:405-406`).
- **Idempotent on ASCII input:** Because the function only converts non-ASCII parts, calling it on a string that is already ASCII produces the same string unchanged (`punycode.js:399-401`).

---

## 3. Implementation Walkthrough

### 3.1 Top-Level Flow

The function `toASCII` at `punycode.js:408-414` delegates entirely to `mapDomain(input, callback)` at `punycode.js:409`. The callback is an inline function defined at `punycode.js:409-413`.

### 3.2 Callback Logic

For each domain label processed by `mapDomain`, the callback at `punycode.js:409-413` executes:

1. **Test for non-ASCII characters** (`punycode.js:410`): Invoke `regexNonASCII.test(string)`, where `regexNonASCII = /[^\0-\x7F]/` (`punycode.js:18`). This regex matches any character outside the range U+0000 through U+007F.

2. **If non-ASCII is found** (`punycode.js:411`): Return `'xn--' + encode(string)`. The string is Punycode-encoded by the `encode` function (`punycode.js:290-376`), and the result is prefixed with the ACE prefix `xn--`.

3. **If no non-ASCII is found** (`punycode.js:412`): Return the label unchanged.

### 3.3 mapDomain's Role

The `mapDomain` function at `punycode.js:72-86` handles:

- **Email address splitting** (`punycode.js:73-80`): Splits on `@` and preserves the local part (left of `@`) verbatim, processing only the domain part (right of `@`).
- **RFC 3490 separator normalization** (`punycode.js:82`): Replaces all four separator characters — U+002E (`.`), U+3002 (Ideographic full stop), U+FF0E (Fullwidth full stop), U+FF61 (Halfwidth ideographic full stop) — with U+002E using `regexSeparators = /[\x2E\u3002\uFF0E\uFF61]/g` (`punycode.js:19`).
- **Label splitting and mapping** (`punycode.js:83-84`): Splits the normalized domain on `.` and applies the callback to each label.
- **Rejoining** (`punycode.js:84-85`): Joins the transformed labels with `.` and prepends any preserved local part.

### 3.4 encode Function

The `encode` function at `punycode.js:290-376` implements RFC 3492 Bootstring encoding. It converts a Unicode label to a Punycode ASCII string. If the label contains characters with code points >= 0x80, they are converted to a base-36 delta-encoded suffix. The function may throw a `RangeError` with message `'Overflow: input needs wider integers to process'` (`punycode.js:338, 346`).

---

## 4. Edge Cases

### 4.1 Pure-ASCII Labels, Including ACE Prefixes

A label composed entirely of ASCII characters (U+0000-U+007F) passes through unchanged, even if it begins with `xn--` (`punycode.js:410-412`). The implementation does not strip or re-apply ACE prefixes in the `toASCII` direction; pre-existing `xn--` prefixes in ASCII-only labels are never altered.

**Example:** `xn--maana-pta` (already-encoded ASCII) → `xn--maana-pta` (unchanged).

### 4.2 The DEL Character (U+007F)

The regex `regexNonASCII = /[^\0-\x7F]/` (`punycode.js:18`) matches characters strictly outside the range U+0000-U+007F. U+007F (DEL) is the upper boundary and falls inside the range, so it does NOT match. The inline comment confirms: "U+007F DEL is excluded too" (`punycode.js:18`).

As a result, a label like `foo\x7F` (containing only DEL and ASCII) is returned unchanged because `regexNonASCII.test('foo\x7F')` is false.

**Example:** `foo\x7F.example` → `foo\x7F.example` (unchanged).

**Porting note:** A regex such as `/[^\x00-\x7E]/` or a code-point threshold `> 127` would incorrectly classify DEL as non-ASCII and produce wrong output.

### 4.3 Mixed-Script Labels

If a label contains any character with a code point >= 0x80, `regexNonASCII` matches and the entire label is encoded. The encoding preserves basic ASCII characters within the label but appends Punycode-encoded data after a `-` delimiter.

**Example:** `café` contains `é` (U+00E9), so the entire label is encoded as `caf-dma`, then prefixed with `xn--` to yield `xn--caf-dma`.

### 4.4 Email Address Handling

When the input contains `@`, `mapDomain` splits it at the first `@` and preserves the local part verbatim. Only the domain part (to the right of `@`) is subject to label processing.

**Example:** `user@café.com` → `user@xn--caf-dma.com`. The local part `user` is unchanged.

### 4.5 RFC 3490 Separator Normalization

All four separator characters are normalized to U+002E before label splitting. This ensures that domains using non-standard separators (e.g., U+3002) are correctly split into labels.

**Example:** `mañana\u3002com` (with Ideographic full stop) is normalized to `mañana.com`, then processed as two labels: `mañana` → `xn--maana-pta` and `com` → `com`, yielding `xn--maana-pta.com`.

---

## 5. Error Handling

Any error thrown by the `encode` function propagates to the caller. The primary error is `'overflow'` (thrown at `punycode.js:338, 346`), which raises a `RangeError` with message `'Overflow: input needs wider integers to process'`.

**Example:** Extremely long or special Unicode strings that cause delta to exceed `maxInt` (2147483647) will trigger an overflow error.

---

## 6. Callers

- **Public API binding** (`punycode.js:439`): The function is exported as `punycode.toASCII` in the public `punycode` object.
- **No internal callers:** Within `punycode.js`, `toASCII` is not called by any other function.

---

## 7. References

- **RFC 3490 (Internationalizing Domain Names in Applications):** Defines the ToASCII operation and the ACE prefix `xn--`.
- **RFC 3492 (Punycode: A Bootstring algorithm for representing Unicode with ASCII):** Specifies the Bootstring encoding algorithm used by the `encode` function.
- **RFC 5891 (Internationalized Domain Names for Applications):** Updates and clarifies IDNA processing.

---

## 8. Cross-References

| Component | Source | Purpose |
|-----------|--------|---------|
| `mapDomain(domain, callback)` | `punycode.js:72-86` | Splits and processes domain labels, handling email addresses and RFC 3490 separators |
| `regexNonASCII` | `punycode.js:18` | `/[^\0-\x7F]/` — detects non-ASCII characters |
| `regexSeparators` | `punycode.js:19` | `/[\x2E\u3002\uFF0E\uFF61]/g` — RFC 3490 separator characters |
| `encode(input)` | `punycode.js:290-376` | Converts a Unicode label to Punycode ASCII |
| `map(array, callback)` | `punycode.js:53-60` | Generic utility function used by `mapDomain` to iterate labels |

