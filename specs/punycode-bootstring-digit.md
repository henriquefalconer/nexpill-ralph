# punycode-bootstring-digit

## Purpose

This spec documents the two private digit-conversion helpers that translate between Bootstring digit values (integers 0–35) and their ASCII code-point representations. These helpers are the bridge between the raw integer arithmetic of the Bootstring algorithm and the ASCII characters stored in an encoded Punycode label. Both functions are private and not exported.

## `basicToDigit(codePoint)` — ASCII code point to Bootstring digit value

**Declaration:** `punycode.js:144–155`

### Signature

- **Input:** `codePoint` — a non-negative integer representing an ASCII code point.
- **Output:** An integer in the range 0–35 if the input is a valid Bootstring digit character, or 36 (the value of `base`) if it is not.

### Behavior

The function maps ASCII code points to their corresponding Bootstring digit values according to three ranges. The checks are performed in order; only the first matching range applies.

#### Range 1: ASCII digits '0'–'9' (punycode.js:145–147)

- Condition: `codePoint >= 0x30` and `codePoint < 0x3A`
- Output: `26 + (codePoint - 0x30)`
- Maps '0' (0x30) → 26, '1' (0x31) → 27, ..., '9' (0x39) → 35.

#### Range 2: Uppercase ASCII letters 'A'–'Z' (punycode.js:148–150)

- Condition: `codePoint >= 0x41` and `codePoint < 0x5B`
- Output: `codePoint - 0x41`
- Maps 'A' (0x41) → 0, 'B' (0x42) → 1, ..., 'Z' (0x5A) → 25.

#### Range 3: Lowercase ASCII letters 'a'–'z' (punycode.js:151–153)

- Condition: `codePoint >= 0x61` and `codePoint < 0x7B`
- Output: `codePoint - 0x61`
- Maps 'a' (0x61) → 0, 'b' (0x62) → 1, ..., 'z' (0x7A) → 25.

#### Fallback (punycode.js:154)

- Any code point not matching any of the three ranges returns `base` (36).
- A return value of 36 signals "not a valid Bootstring digit" to the caller.

### Case Insensitivity

Uppercase letters 'A'–'Z' and lowercase letters 'a'–'z' both map to the digit range 0–25, producing identical digit values for corresponding letters. For example, 'A' and 'a' both yield 0; 'Z' and 'z' both yield 25. This makes `basicToDigit` case-insensitive for letter inputs.

The Bootstring decoder exploits this: an encoded Punycode label with uppercase letters decodes to the same Unicode string as the same label with lowercase letters (confirmed by the test in `punycode-decode.md` for `'ZZZ'` vs `'zzz'`).

### Summary Table

| Code point range | Characters | Digit values |
|---|---|---|
| 0x30–0x39 | '0'–'9' | 26–35 |
| 0x41–0x5A | 'A'–'Z' | 0–25 |
| 0x61–0x7A | 'a'–'z' | 0–25 |
| all others | — | 36 (invalid sentinel) |

### Call Sites

- `decode` (punycode.js:238): converts each character of the variable-length-integer suffix into a digit. If the result is >= `base` (36), `decode` calls `error('invalid-input')`.

---

## `digitToBasic(digit, flag)` — Bootstring digit value to ASCII code point

**Declaration:** `punycode.js:168–172`

### Signature

- **Input:** `digit` — an integer in the range 0–35 (a valid Bootstring digit value).
- **Input:** `flag` — an integer; zero for lowercase output, non-zero for uppercase output.
- **Output:** An integer representing the ASCII code point for the given digit.

### Behavior

The entire function is implemented as a single expression (punycode.js:171):

```
digit + 22 + 75 * (digit < 26) - ((flag != 0) << 5)
```

To understand this expression, consider the three contributing terms:

#### Term 1: Base value `digit + 22`

When `digit < 26` (letters): `digit + 22 + 75 = digit + 97`. This is the ASCII code for 'a' plus `digit`, producing 'a' (97) for digit 0 through 'z' (122) for digit 25.

When `digit >= 26` (digits): `digit + 22 + 0 = digit + 22`. For digit 26 this gives 48 ('0'); for digit 35 this gives 57 ('9'). This correctly maps digits 26–35 to ASCII '0'–'9'.

#### Term 2: Letter selector `75 * (digit < 26)`

The boolean expression `digit < 26` evaluates to 1 (true) or 0 (false). Multiplied by 75, this adds 75 to the result only when the digit is in the letter range (0–25). Combined with Term 1: `digit + 22 + 75` = `digit + 97` = ASCII for 'a' + offset.

#### Term 3: Case selector `(flag != 0) << 5`

If `flag` is non-zero, `(flag != 0)` is 1, and left-shifting by 5 gives 32. Subtracting 32 from an ASCII lowercase letter converts it to uppercase (e.g., 'a' = 97; 'a' - 32 = 65 = 'A'). This term is subtracted only when `flag` is non-zero.

### Output Summary

| Condition | `flag = 0` (lowercase) | `flag != 0` (uppercase) |
|---|---|---|
| `digit` in 0–25 | `digit + 97` → 'a'–'z' (ASCII 97–122) | `digit + 65` → 'A'–'Z' (ASCII 65–90) |
| `digit` in 26–35 | `digit + 22` → '0'–'9' (ASCII 48–57) | `digit + 22` → '0'–'9' (same; digits have no case) |

### Uppercase Branch Reachability

The `encode` function calls `digitToBasic` at two sites (punycode.js:359 and punycode.js:364), always passing `0` as the `flag` argument. Therefore the uppercase branch (`flag != 0`) is never reached from the public API. The flag parameter exists for completeness per RFC 3492, but the output of `encode` is always lowercase.

In a Go port, it is safe to implement only the `flag = 0` path and document that uppercase output is not supported by the public encoder. The decoder does not call `digitToBasic` at all; it calls `basicToDigit` instead.

### Call Sites

- `encode` (punycode.js:359): emits intermediate digits in the variable-length integer encoding of delta. Always called with `flag = 0`.
- `encode` (punycode.js:364): emits the final digit of each variable-length integer. Always called with `flag = 0`.

## Cross-References

- [punycode-module.md](./punycode-module.md) — `base` constant (36) used as the sentinel return value and in division calculations
- [punycode-decode.md](./punycode-decode.md) — uses `basicToDigit` to parse variable-length integers from encoded input
- [punycode-encode.md](./punycode-encode.md) — uses `digitToBasic` to produce ASCII characters for variable-length integer output
- [punycode-bias-adapt.md](./punycode-bias-adapt.md) — `adapt` shares the same `base`, `tMin`, and `tMax` constants used in threshold calculations
