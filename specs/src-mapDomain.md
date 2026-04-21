# Spec: `mapDomain(domain, callback)` private helper

Source block: `punycode.js:62-86`

---

## 1. Subject

**Function name:** `mapDomain`

**Scope:** Private helper function.

**Purpose:** Applies a per-label callback to every label of a domain name (or the domain part of an email address), joining labels back with `.` after transformation.

---

## 2. Contract

Derived from the JSDoc at `punycode.js:62-71` and the implementation at `punycode.js:72-86`.

**Input:** A string that is either a domain name or an email address of the form `local@domain`.

**Output:** A string with the callback applied to each label (dot-separated segment) of the domain. The local part of an email address (everything before the first `@`) is preserved unchanged.

**Behavior:**
- If the input contains an `@` character, split on `@` at `punycode.js:73`. If exactly two parts result (`punycode.js:75`), preserve `parts[0] + '@'` as a prefix to be prepended to the result (`punycode.js:78-79`), and use `parts[1]` as the domain to process.
- Normalize all RFC 3490 separator characters to the standard ASCII dot (`\x2E`) using `regexSeparators` replace at `punycode.js:82`. This handles the four separator characters (ideographic full stop `\u3002`, fullwidth stop `\uFF0E`, halfwidth stop `\uFF61`, and ASCII dot `\x2E`) defined at `punycode.js:19`, treating all as label boundaries.
- Split the normalized domain on `.` at `punycode.js:83`. The code avoids `split(regex)` and uses the replace approach for IE8 compatibility, as noted in the adjacent comment at `punycode.js:81`.
- Apply the `map` helper function (defined at `punycode.js:53-60`) to the labels array with the provided callback at `punycode.js:84`, then join the transformed labels back with `.`.
- Return the local part prefix (if present) concatenated with the encoded domain at `punycode.js:85`.

---

## 3. Implementation walkthrough

### 3.1 Email address handling — `punycode.js:73-79`

The function begins by splitting the input on the `@` character:

```javascript
const parts = domain.split('@');
let result = '';
if (parts.length > 1) {
  result = parts[0] + '@';
  domain = parts[1];
}
```

If the split yields more than one part, the first part (the local part of the email address) is preserved as `result = parts[0] + '@'`, and the domain component is reassigned to `parts[1]`. This keeps the local part intact while preparing the domain for label processing.

### 3.2 Separator normalization — `punycode.js:81-82`

The comment at `punycode.js:81` explains the IE8 compatibility constraint: `split(regex)` must be avoided. Instead, the code uses a replace operation:

```javascript
domain = domain.replace(regexSeparators, '\x2E');
```

This normalizes all RFC 3490 separator characters (defined at `punycode.js:19` as `/[\x2E\u3002\uFF0E\uFF61]/g`) to the ASCII dot character `\x2E`. The four separator characters are:
- `\x2E`: ASCII period (fullstop)
- `\u3002`: ideographic full stop (CJK)
- `\uFF0E`: fullwidth full stop (CJK variant)
- `\uFF61`: halfwidth ideographic full stop (CJK variant)

By normalizing all of these to `\x2E`, the domain is prepared for safe ASCII-only label splitting.

### 3.3 Label mapping — `punycode.js:83-84`

After normalization, the domain is split on the ASCII dot:

```javascript
const labels = domain.split('.');
const encoded = map(labels, callback).join('.');
```

The `map` helper (at `punycode.js:53-60`) invokes the callback on each label. The callback's return value becomes the new label. The transformed labels are joined back with `.`.

### 3.4 Result assembly — `punycode.js:85`

```javascript
return result + encoded;
```

The result concatenates the local-part prefix (either empty string or `'local@'`) with the encoded domain labels.

---

## 4. Edge cases

### 4.1 Multiple `@` characters

When the input contains more than two `@` characters, the `split('@')` at `punycode.js:73` returns an array with more than two elements. The condition at `punycode.js:75` checks `if (parts.length > 1)`, which is true. However, only `parts[0]` is used as the local part (assigned to `result` at `punycode.js:78`), and only `parts[1]` is used as the domain (assigned to `domain` at `punycode.js:79`). Any `@` characters beyond the first are effectively dropped: they become part of the discarded `parts[2]`, `parts[3]`, etc.

This is subtle behavior: an input like `'local@domain@extra'` will be transformed to `'local@<callback applied to domain>'`, losing the `@extra` suffix.

### 4.2 Empty input

When the input is an empty string, the label array will be `['']`, and `map([''], callback)` will return `[callback('')]`. The callback is invoked even on the empty label. Most callbacks short-circuit on empty labels, but this is not guaranteed.

### 4.3 Already-ASCII labels

The callback is invoked on every label, regardless of whether it is ASCII or non-ASCII. Callbacks such as the one in `toASCII` (at `punycode.js:409-412`) test each label with `regexNonASCII` and return the label unchanged if it is already ASCII. However, the callback is still called, and the overhead of testing and re-returning the same label is incurred for every ASCII label.

---

## 5. Callsites

### 5.1 `toUnicode(input)` — `punycode.js:389-395`

The `toUnicode` function calls `mapDomain` with a callback that decodes Punycode-encoded labels:

```javascript
const toUnicode = function(input) {
  return mapDomain(input, function(string) {
    return regexPunycode.test(string)
      ? decode(string.slice(4).toLowerCase())
      : string;
  });
};
```

The callback tests each label against `regexPunycode` (defined at `punycode.js:17` as `/^xn--/`) to detect Punycode-encoded labels. If a label is prefixed with `xn--`, the callback decodes it by slicing off the prefix and passing the remainder to the `decode` function. Otherwise, the label is returned unchanged.

### 5.2 `toASCII(input)` — `punycode.js:408-414`

The `toASCII` function calls `mapDomain` with a callback that encodes non-ASCII labels:

```javascript
const toASCII = function(input) {
  return mapDomain(input, function(string) {
    return regexNonASCII.test(string)
      ? 'xn--' + encode(string)
      : string;
  });
};
```

The callback tests each label against `regexNonASCII` (defined at `punycode.js:18` as `/[^\0-\x7F]/`) to detect non-ASCII characters. If a label contains non-ASCII, it is encoded via the `encode` function and prefixed with `xn--`. Otherwise, the label is returned unchanged.

---

## 6. References

- RFC 3490 defines the four separator characters for IDNA (Internationalized Domain Names in Applications).
- The comment at `punycode.js:81` notes the IE8 compatibility constraint that prevents use of regex split.
- The `regexSeparators` pattern is defined at `punycode.js:19`.
- The `map` utility is defined at `punycode.js:53-60`.

---

## 7. Implementation citations

| Symbol | Location |
|---|---|
| `mapDomain` function definition | `punycode.js:62-86` |
| JSDoc for `mapDomain` | `punycode.js:62-71` |
| `regexSeparators` pattern | `punycode.js:19` |
| `map` helper function | `punycode.js:53-60` |
| `regexPunycode` pattern | `punycode.js:17` |
| `regexNonASCII` pattern | `punycode.js:18` |
| `toUnicode` call site | `punycode.js:389-395` |
| `toASCII` call site | `punycode.js:408-414` |
| IE8 compatibility comment | `punycode.js:81` |

