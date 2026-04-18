# punycode.encode

## Purpose

Converts a Unicode string into the Bootstring/Punycode ASCII-only representation of its code points, per RFC 3492. This function performs the character-by-character encoding of Unicode text into a limited ASCII alphabet suitable for domain names and other identifier contexts. `punycode.encode` is the inverse of `punycode.decode`.

## Signature

**Input:** A Unicode string (UTF-16/UCS-2 representation, with surrogate pairs as needed).

**Output:** An ASCII-only Bootstring payload containing only code points in the range U+0030 to U+007A (digits, letters, and hyphen). No `xn--` prefix is included; that is added by the higher-level `toASCII` function (see cross-references).

## Algorithm

The encoding follows the Bootstring algorithm as specified in RFC 3492 Section 5. The implementation converts the input to code points, processes basic characters separately, then iteratively encodes the remaining non-basic characters in increasing order of code point value.

### Phase 1: Convert Input to Code Points

The input string is converted from its UTF-16/UCS-2 internal representation into an array of Unicode code points via `ucs2decode` (punycode.js:294). This handles surrogate pairs correctly, combining high and low surrogates into single code points ≥ U+10000.

### Phase 2: Initialize Bootstring State

The following Bootstring parameters are initialized (punycode.js:6-13):
- Base: 36
- Minimum threshold: 1 (tMin)
- Maximum threshold: 26 (tMax)
- Skew: 38
- Damping factor: 700 (damp)
- Initial bias: 72 (initialBias)
- Initial code point threshold: 128 (initialN, i.e., U+0080, the first non-ASCII code point)
- Delimiter: hyphen-minus U+002D (`-`)

The encoder state is initialized to:
- n = initialN (128)
- delta = 0
- bias = initialBias (72)
- output = empty array

### Phase 3: Extract and Output Basic Code Points

Iterate through each code point in the input array. For every code point less than 0x80 (i.e., basic ASCII), append its character representation directly to the output array (punycode.js:305-309). Record the count of basic code points as `basicLength` (punycode.js:311).

If `basicLength > 0`, append the delimiter character (`-`) to the output (punycode.js:318-320).

Initialize `handledCPCount` to `basicLength`, representing the number of code points processed so far (punycode.js:312).

### Phase 4: Main Encoding Loop

While `handledCPCount < inputLength` (punycode.js:323):

1. **Find next code point to encode:** Scan the input array to find the smallest code point that is ≥ current n and has not yet been fully processed. Call this value m (punycode.js:327-332). If no such code point exists, m = maxInt.

2. **Advance delta:** Calculate the advance needed as `delta += (m - n) × (handledCPCount + 1)` (punycode.js:341). Before performing this addition, check for overflow: if `(m - n) > floor((maxInt - delta) / (handledCPCount + 1))`, throw an error (punycode.js:337-338). The maximum value for delta and all intermediate calculations is maxInt = 2147483647 (punycode.js:4).

3. **Update threshold:** Set `n = m` (punycode.js:342).

4. **Encode all code points in input order:**
   - Iterate through the input array in order (punycode.js:344).
   - For each code point c:
     - If `c < n`, increment delta. If delta would exceed maxInt, throw an overflow error (punycode.js:345-346).
     - If `c === n`, encode delta using the generalized variable-length integer algorithm:
       - Initialize q = delta (punycode.js:350).
       - For each bias threshold k starting at base (36), incrementing by base each iteration (punycode.js:351):
         - Compute the threshold t: if k ≤ bias, t = tMin (1); else if k ≥ bias + tMax (26), t = tMax (26); else t = k - bias (punycode.js:352).
         - While q ≥ t, emit a digit by computing `digitToBasic(t + (q - t) mod (base - t), 0)` and appending its character to output (punycode.js:358-360; digitToBasic at punycode.js:168-172 maps 0–25 to a–z and 26–35 to 0–9).
         - Reduce q: q = floor((q - t) / (base - t)) (punycode.js:361).
         - Break when q < t (punycode.js:353-354).
       - Emit the final digit: `digitToBasic(q, 0)` and append to output (punycode.js:364).
     - After encoding delta, update bias by calling `adapt(delta, handledCPCount + 1, handledCPCount === basicLength)` (punycode.js:365). The adapt function (punycode.js:179-187) implements RFC 3492 Section 3.4 bias adaptation, taking into account whether this is the first delta encoded (firstTime is true iff handledCPCount === basicLength).
     - Reset delta to 0 (punycode.js:366).
     - Increment handledCPCount (punycode.js:367).

5. **Prepare for next iteration:** Increment both delta and n (punycode.js:371-372). These represent the baseline for the next round of code point scanning.

### Phase 5: Final Output

Join the output array into a single string and return it (punycode.js:375).

## Error Conditions

The encoder throws exactly one type of error during overflow detection:

**Overflow: input needs wider integers to process** — thrown if any delta calculation would exceed maxInt = 2147483647 (punycode.js:22, 337-338, 345-346). Specifically:
- During the main advance calculation: if `(m - n) > floor((maxInt - delta) / (handledCPCount + 1))`
- During per-codepoint delta increments: if `delta > maxInt`

No test cases in the test suite (tests.js:312-321) exercise error paths; the encoder's correctness is validated by round-trip encoding and decoding.

## Round-Trip Guarantee

Every entry in `testData.strings` (tests/tests.js:7-136) is expected to satisfy:
```
encode(decoded) === encoded
```
exactly (tests.js:313-320).

Combined with `punycode.decode`, the codec forms a bijective pair on valid Punycode strings. Given any valid Punycode output, `decode(encode(input)) === input` for all strings in the test suite.

## Examples

The test suite includes 21 fixtures from `testData.strings`:

### Basic Cases (tests.js:7-31)

1. **Single basic code point** (tests.js:8-10): Input `'Bach'` (U+0042 U+0061 U+0063 U+0068) → `'Bach-'`
2. **Single non-ASCII character** (tests.js:12-15): Input `'\xFC'` (U+00FC) → `'tda'`
3. **Multiple non-ASCII characters** (tests.js:17-20): Input `'\xFC\xEB\xE4\xF6\u2665'` (U+00FC U+00EB U+00E4 U+00F6 U+2665) → `'4can8av2009b'`
4. **Mix of ASCII and non-ASCII** (tests.js:22-25): Input `'b\xFCcher'` (U+0062 U+00FC U+0063 U+0068 U+0065 U+0072) → `'bcher-kva'`
5. **Long string with both** (tests.js:28-30): Input `'Willst du die Bl\xFCthe des fr\xFChen, die Fr\xFCchte des sp\xE4teren Jahres'` → `'Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal'`

### RFC 3492 Section 7.1 Exemplars (tests.js:33-96)

6. **Arabic (Egyptian)** (tests.js:33-36): Input U+0644 U+064A U+0647 U+0645 U+0627 U+0628 U+062A U+0643 U+0644 U+0645 U+0648 U+0634 U+0639 U+0631 U+0628 U+064A U+061F → `'egbpdaj6bu4bxfgehfvwxn'`
7. **Chinese (simplified)** (tests.js:39-41): Input U+4ED6 U+4EEC U+4E3A U+4EC0 U+4E48 U+4E0D U+8BF4 U+4E2D U+6587 → `'ihqwcrb4cv8a8dqg056pqjye'`
8. **Chinese (traditional)** (tests.js:44-46): Input U+4ED6 U+5011 U+7232 U+4EC0 U+9EBD U+4E0D U+8AAA U+4E2D U+6587 → `'ihqwctvzc91f659drss3x8bo0yb'`
9. **Czech** (tests.js:49-51): Input `'Pro\u010Dprost\u011Bnemluv\xED\u010Desky'` → `'Proprostnemluvesky-uyb24dma41a'`
10. **Hebrew** (tests.js:54-56): Input U+05DC U+05DE U+05D4 U+05D4 U+05DD U+05E4 U+05E9 U+05D5 U+05D8 U+05DC U+05D0 U+05DE U+05D3 U+05D1 U+05E8 U+05D9 U+05DD U+05E2 U+05D1 U+05E8 U+05D9 U+05EA → `'4dbcagdahymbxekheh6e0a7fei0b'`
11. **Hindi (Devanagari)** (tests.js:59-61): Input U+092F U+0939 U+0932 U+094B U+0917 U+0939 U+093F U+0928 U+094D U+0926 U+0940 U+0915 U+094D U+092F U+094B U+0902 U+0928 U+0939 U+0940 U+0902 U+092C U+094B U+0932 U+0938 U+0915 U+0924 U+0947 U+0939 U+0948 U+0902 → `'i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd'`
12. **Japanese (kanji and hiragana)** (tests.js:64-66): Input U+306A U+305C U+307F U+3093 U+306A U+65E5 U+672C U+8A9E U+3092 U+8A71 U+3057 U+3066 U+304F U+308C U+306A U+3044 U+306E U+304B → `'n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa'`
13. **Korean (Hangul)** (tests.js:69-71): Input U+C138 U+ACC4 U+C758 U+BAA8 U+B4E0 U+C0AC U+B78C U+B4E4 U+C774 U+D55C U+AD6D U+C5B4 U+B97C U+C774 U+D574 U+D55C U+B2E4 U+BA74 U+C5BC U+B9C8 U+B098 U+C88B U+C744 U+AE4C → `'989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c'`
14. **Russian (Cyrillic)** (tests.js:83-85): Input U+043F U+043E U+0447 U+0435 U+043C U+0443 U+0436 U+0435 U+043E U+043D U+0438 U+043D U+0435 U+0433 U+043E U+0432 U+043E U+0440 U+044F U+0442 U+043F U+043E U+0440 U+0443 U+0441 U+0441 U+043A U+0438 → `'b1abfaaepdrnnbgefbadotcwatmq2g4l'`
15. **Spanish** (tests.js:88-90): Input `'Porqu\xE9nopuedensimplementehablarenEspa\xF1ol'` → `'PorqunopuedensimplementehablarenEspaol-fmd56a'`
16. **Vietnamese** (tests.js:93-95): Input `'T\u1EA1isaoh\u1ECDkh\xF4ngth\u1EC3ch\u1EC9n\xF3iti\u1EBFngVi\u1EC7t'` → `'TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g'`

### Additional RFC 3492 Examples (tests.js:98-124)

17. **Mixed script with digits** (tests.js:98-99): Input `'3\u5E74B\u7D44\u91D1\u516B\u5148\u751F'` → `'3B-ww4c5e180e575a65lsy2b'`
18. **Hyphens in ASCII prefix** (tests.js:102-103): Input `'\u5B89\u5BA4\u5948\u7F8E\u6075-with-SUPER-MONKEYS'` → `'-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n'`
19. **ASCII prefixes with hyphens** (tests.js:106-107): Input `'Hello-Another-Way-\u305D\u308C\u305E\u308C\u306E\u5834\u6240'` → `'Hello-Another-Way--fc4qua05auwb3674vfr0b'`
20. **Digits with non-ASCII** (tests.js:110-111): Input `'\u3072\u3068\u3064\u5C4B\u6839\u306E\u4E0B2'` → `'2-u9tlzr9756bt3uc0v'`
21. **Complex mixed script** (tests.js:114-115): Input `'Maji\u3067Koi\u3059\u308B5\u79D2\u524D'` → `'MajiKoi5-783gue6qz075azm5e'`
22. **Japanese syllables** (tests.js:118-119): Input `'\u30D1\u30D5\u30A3\u30FCde\u30EB\u30F3\u30D0'` → `'de-jg4avhby1noc0d'`
23. **Japanese hiragana** (tests.js:122-123): Input `'\u305D\u306E\u30B9\u30D4\u30FC\u30C9\u3067'` → `'d9juau41awczczp'`

### Edge Case: Pure ASCII with Delimiters (tests.js:131-134)

24. **ASCII that breaks hostname rules** (tests.js:131-133): Input `'-> $1.00 <-'` (all ASCII: U+002D U+003E U+0020 U+0024 U+0031 U+002E U+0030 U+0030 U+0020 U+003C U+002D) → `'-> $1.00 <--'`. This case demonstrates that when the input contains only ASCII characters, they are output as-is followed by a single delimiter hyphen.

## Cross-References

- [test-fixtures.md](./test-fixtures.md) — Overview of `testData.strings` and test structure.
- [punycode-decode.md](./punycode-decode.md) — Inverse function for round-trip decoding.
- [punycode-to-ascii.md](./punycode-to-ascii.md) — Higher-level function that adds `xn--` prefix and processes domain labels.
- [punycode-ucs2-decode.md](./punycode-ucs2-decode.md) — Code point array conversion from UTF-16/UCS-2 strings.
