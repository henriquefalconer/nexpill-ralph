# Spec: `punycode.toUnicode` — `toUnicode(input): string`

**Source:** `punycode.js:389-395`

**Public binding:** `punycode.js:440`

---

## 1. Subject

User-facing function that converts a Punycoded domain name or email address to its Unicode form. Only labels prefixed with `xn--` are decoded; other labels pass through unchanged. The function is idempotent: calling it on input that is already Unicode returns that input unmodified, because no label will match the `xn--` prefix test.

---

## 2. Contract

From the JSDoc at `punycode.js:378-388`:

- Converts a Punycode string representing a domain name or an email address to Unicode.
- Only the Punycoded parts of the input will be converted.
- It does not matter if the function is called on a string that has already been converted to Unicode.
- Parameter `input` (`String`): the Punycoded domain name or email address to convert to Unicode.
- Returns (`String`): the Unicode representation of the given Punycode string.

---

## 3. Implementation Walkthrough

**Function location:** `punycode.js:389-395`

```javascript
const toUnicode = function(input) {
  return mapDomain(input, function(string) {
    return regexPunycode.test(string)
      ? decode(string.slice(4).toLowerCase())
      : string;
  });
};
```

### 3.1 Delegation to `mapDomain`

At `punycode.js:390`, `toUnicode` delegates entirely to `mapDomain(input, callback)` at `punycode.js:72-86`. The `mapDomain` function splits the input on label boundaries (`.` and three additional Unicode dot variants) and applies the provided callback to each label.

### 3.2 Per-Label Callback

The callback passed to `mapDomain` at `punycode.js:390-394` operates as follows:

1. **Prefix test** (`punycode.js:391`): Test each label against `regexPunycode` at `punycode.js:17`, which is `/^xn--/`. This regex is lowercase and case-sensitive.
   
2. **If match** (`punycode.js:392`): Strip the 4-character `xn--` prefix via `.slice(4)`, lowercase the remainder via `.toLowerCase()` (per RFC 3492 case-insensitivity), then decode the result by calling `decode(string.slice(4).toLowerCase())` at `punycode.js:392`. The `decode` function is defined at `punycode.js:196-281`.

3. **If no match** (`punycode.js:393`): Return the label unchanged.

### 3.3 Email Handling

When `mapDomain` receives input containing `@`, it splits on `@` at `punycode.js:73`, preserves the local part (everything before the first `@`) in `result` at `punycode.js:78`, and applies the callback only to the domain portion at `punycode.js:84`. The local part is concatenated back untouched at `punycode.js:85`.

---

## 4. Edge Cases

### 4.1 Already-Unicode Labels

Labels that do not start with `xn--` pass through unchanged (`punycode.js:393`). A call to `toUnicode('café.com')` returns `'café.com'` unmodified because neither label matches the prefix test.

### 4.2 Mixed-Case Prefix

The `regexPunycode` test at `punycode.js:391` is case-sensitive and anchored to lowercase. A label prefixed with `XN--` (uppercase) or `Xn--` (mixed case) will NOT match. The label is therefore returned unchanged. This is an implementation quirk: the RFC 3490 ToUnicode operation applies case-folding before testing the prefix, but this implementation does not. No test vectors in `specs/test-toUnicode.md` exercise uppercase or mixed-case `xn--` prefixes.

### 4.3 Trailing Dot

An input like `'example.com.'` splits into labels `['example', 'com', '']`. The empty final label does not match `regexPunycode`, so it is returned as-is and the trailing dot is preserved in the output. Implementations must not strip or special-case trailing dots.

### 4.4 Separator Normalization

Before splitting into labels, `mapDomain` normalizes all four Unicode dot variants (U+002E, U+3002, U+FF0E, U+FF61) to U+002E using `regexSeparators` at `punycode.js:19` and `punycode.js:82`. This normalization is transparent to the caller.

---

## 5. Error Handling

Any errors thrown by `decode` propagate to the caller. The `decode` function at `punycode.js:196-281` may throw the following errors:

- `'not-basic'` (thrown at `punycode.js:216`): A non-basic code point was encountered in the basic portion of the encoded string.
- `'invalid-input'` (thrown at `punycode.js:235`, `punycode.js:241`): The input is malformed or incomplete.
- `'overflow'` (thrown at `punycode.js:244`, `punycode.js:256`, `punycode.js:269`): An integer overflow occurred during decoding.

---

## 6. Call Graph

- **Internal callers:** None. `toUnicode` is not called from elsewhere within `punycode.js`.
- **Public binding:** `punycode.toUnicode` at `punycode.js:440`.
- **Dependencies:**
  - `mapDomain(domain, callback)` at `punycode.js:72-86`
  - `regexPunycode` at `punycode.js:17`
  - `decode(input)` at `punycode.js:196-281`

---

## 7. Standards References

- **RFC 3490** (Internationalizing Domain Names in Applications): Defines the ToUnicode operation, which maps a domain to a Unicode string or reports error.
- **RFC 5891** (Internationalized Domain Names for Applications): Supersedes RFC 3490 and refines the IDNA algorithm.
- **RFC 3492** (Punycode): Defines the codec used by `decode`. Case-insensitivity is referenced in the base-36 alphabet treatment.

---

## 8. Implementation Notes

The `.toLowerCase()` call at `punycode.js:392` applies only to the post-prefix remainder, not to the prefix itself. The prefix matching at `punycode.js:391` is case-sensitive and requires exact lowercase `xn--`. This asymmetry means that an input like `'XN--abc123'` will not be decoded, because the uppercase prefix will not match `regexPunycode`.
