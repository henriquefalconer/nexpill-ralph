# punycode.decode

## Purpose

Convert a Punycode (Bootstring) ASCII-only string into the Unicode string it encodes, per RFC 3492.

## Signature

**Input:** An ASCII-only string containing the Bootstring payload (without the `xn--` prefix).

**Output:** A Unicode string representing the decoded sequence of code points.

## Algorithm

The decode process reverses the Bootstring algorithm defined in RFC 3492. It uses the following Bootstring parameters (punycode.js:7-14):

- Base: 36
- tMin: 1
- tMax: 26
- Skew: 38
- Damp: 700
- initialBias: 72
- initialN: 128 (0x80, the first non-ASCII code point)
- Delimiter: '-' (0x2D)

### Phase 1: Process Basic Code Points

Find the position of the last '-' delimiter in the input. All characters before this delimiter are basic (ASCII) code points and are copied directly to the output (punycode.js:208-219). If no delimiter exists, begin with an empty output.

For each basic code point copied to the output, verify that its numeric value is less than 0x80. If any code point has value >= 0x80, throw an error with message `Illegal input >= 0x80 (not a basic code point)` (punycode.js:215-217).

### Phase 2: Decode Variable-Length Integers

Starting after the last delimiter (or at position 0 if no delimiter exists), process the remaining input characters as a sequence of generalized variable-length integers. Each integer encodes a delta value that specifies positions and code points to insert into the output.

**Variable-Length Integer Decoding:**

For each position in the encoded sequence:

1. Initialize delta to 0 and weight w to 1. Process input characters from left to right.

2. For each character at the current position:
   - Convert the character to a digit using the `basicToDigit` function (punycode.js:144-155):
     - ASCII digits '0'–'9' (code points 0x30–0x39) map to 26–35
     - Uppercase letters 'A'–'Z' (code points 0x41–0x5A) map to 0–25
     - Lowercase letters 'a'–'z' (code points 0x61–0x7A) map to 0–25
     - All other characters map to 36 (invalid)
   
   - If the digit is >= 36, throw an error with message `Invalid input` (punycode.js:240-242).
   
   - If the digit value, when multiplied by the current weight and added to delta, would exceed the maximum 32-bit signed integer (2147483647, defined as maxInt in punycode.js:4), throw an error with message `Overflow: input needs wider integers to process` (punycode.js:243-245).
   
   - Add digit × w to delta.
   
   - Calculate the threshold t using the current bias and position in the sequence (punycode.js:248):
     - t = clamp(k - bias, tMin, tMax), where k increases by base (36) on each iteration
   
   - If digit < t, the integer sequence is complete; break the loop.
   
   - Otherwise, continue: if weight × (base - t) would exceed maxInt, throw an overflow error (punycode.js:255-257); else set w ← w × (base - t) and continue to the next character.

3. If the input ends before a complete variable-length integer is formed (i.e., before a digit < t is encountered), throw an error with message `Invalid input` (punycode.js:234-236).

### Phase 3: Update State and Insert Code Point

After decoding each variable-length integer:

1. Call the adapt function (punycode.js:179-187) with parameters:
   - delta ← i - oldi (where oldi is the delta value before decoding the current integer, and i accumulates the decoded delta)
   - numPoints ← output.length + 1 (the size of the output after this insertion)
   - firstTime ← true only for the first variable-length integer; false thereafter
   
   The adapt function refines the bias for the next iteration.

2. If adding floor(i / output.length + 1) to n would exceed maxInt, throw an overflow error (punycode.js:268-270).

3. Increment n by floor(i / output.length + 1).

4. Set i to i modulo (output.length + 1).

5. Insert the code point n at position i in the output array and increment i.

Repeat steps 1–5 for each variable-length integer in the input until all input characters are consumed.

### Phase 4: Encode as String

Convert the output array of code points to a Unicode string using UTF-16 encoding (i.e., String.fromCodePoint applied to the final code point array) (punycode.js:280).

## Error Conditions

### `Illegal input >= 0x80 (not a basic code point)`

**Message:** `Illegal input >= 0x80 (not a basic code point)` (punycode.js:22-26)

**Trigger:** The basic (pre-delimiter) section of the input contains one or more bytes with values >= 0x80 (punycode.js:215-217).

**Example:** `punycode.decode('\x81-')` throws this error (tests.js:255-262). Note: this is tested indirectly through the ucs2.decode test suite, which shares implementation with decode.

### `Invalid input`

**Message:** `Invalid input` (punycode.js:22-26)

**Trigger:** Occurs in two scenarios (punycode.js:234-236, 240-242):
1. The input sequence ends before a complete variable-length integer is decoded (i.e., before a digit < threshold is encountered).
2. A character in the encoded sequence is not a valid Bootstring digit (i.e., its code point, when passed to basicToDigit, yields a value >= base/36).

**Example:** `punycode.decode('ls8h=')` throws this error because `=` (code point 0x3D) is outside all digit ranges in basicToDigit (tests.js:302-309).

### `Overflow: input needs wider integers to process`

**Message:** `Overflow: input needs wider integers to process` (punycode.js:22-26)

**Trigger:** Any arithmetic computation would exceed 32-bit signed integer bounds (punycode.js:243-245, 255-257, 268-270). The maximum safe value is maxInt = 2147483647 (punycode.js:4).

**Example:** `punycode.decode('\x81')` triggers this error indirectly (tests.js:263-270). Note: this is tested indirectly through the ucs2.decode test suite.

## Case Insensitivity

The Bootstring decoder treats uppercase and lowercase ASCII letters identically in the variable-length-integer sequence. Both 'A'–'Z' (code points 0x41–0x5A) and 'a'–'z' (code points 0x61–0x7A) map to the same 0–25 digit range via basicToDigit (punycode.js:148-153).

**Example:** `punycode.decode('ZZZ')` decodes to the same result as `punycode.decode('zzz')`, yielding Unicode character U+7BA5 (tests.js:299-301).

## Examples

The following test cases are derived from testData.strings (tests.js:7-136). Each entry shows the Bootstring input and its Unicode output.

1. **Single basic code point**
   - Input: `Bach-`
   - Output: `Bach`

2. **Single non-ASCII character**
   - Input: `tda`
   - Output: `ü` (U+00FC)

3. **Multiple non-ASCII characters**
   - Input: `4can8av2009b`
   - Output: `üëäö♥` (U+00FC, U+00EB, U+00E4, U+00F6, U+2665)

4. **Mix of ASCII and non-ASCII**
   - Input: `bcher-kva`
   - Output: `bücher`

5. **Long string with ASCII and non-ASCII**
   - Input: `Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal`
   - Output: `Willst du die Blüthe des frühen, die Früchte des späteren Jahres`

6. **Arabic (Egyptian)**
   - Input: `egbpdaj6bu4bxfgehfvwxn`
   - Output: `ليهمابتكلموشعربي؟` (U+0644, U+064A, U+0647, U+0645, U+0627, U+0628, U+062A, U+0643, U+0644, U+0645, U+0648, U+0634, U+0639, U+0631, U+0628, U+064A, U+061F)

7. **Chinese (simplified)**
   - Input: `ihqwcrb4cv8a8dqg056pqjye`
   - Output: `他们为什么不说中文` (U+4ED6, U+4EEC, U+4E3A, U+4EC0, U+4E48, U+4E0D, U+8BF4, U+4E2D, U+6587)

8. **Chinese (traditional)**
   - Input: `ihqwctvzc91f659drss3x8bo0yb`
   - Output: `他們為什麼不說中文` (U+4ED6, U+5011, U+7232, U+4EC0, U+9EBD, U+4E0D, U+8AAA, U+4E2D, U+6587)

9. **Czech**
   - Input: `Proprostnemluvesky-uyb24dma41a`
   - Output: `Pročprostřenemiuvičesky`

10. **Hebrew**
    - Input: `4dbcagdahymbxekheh6e0a7fei0b`
    - Output: `למהההםפשוטלאמדברים` (U+05DC, U+05DE, U+05D4, U+05D4, U+05DD, U+05E4, U+05E9, U+05D5, U+05D8, U+05DC, U+05D0, U+05DE, U+05D3, U+05D1, U+05E8, U+05D9, U+05DD, U+05E2, U+05D1, U+05E8, U+05D9, U+05EA)

11. **Hindi (Devanagari)**
    - Input: `i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd`
    - Output: `यहलोगहिन्दीक्योंनहींबोलसकतेहैं` (U+092F, U+0939, U+0932, U+094B, U+0917, U+0939, U+093F, U+0928, U+094D, U+0926, U+0940, U+0915, U+094D, U+092F, U+094B, U+0902, U+0928, U+0939, U+0940, U+0902, U+092C, U+094B, U+0932, U+0938, U+0915, U+0924, U+0947, U+0939, U+0948, U+0902)

12. **Japanese (kanji and hiragana)**
    - Input: `n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa`
    - Output: `なぜみんな日本語を話してくれないのか` (U+306A, U+305C, U+307F, U+3093, U+306A, U+65E5, U+672C, U+8A9E, U+3092, U+8A71, U+3057, U+3066, U+304F, U+308C, U+306A, U+3044, U+306E, U+304B)

13. **Korean (Hangul syllables)**
    - Input: `989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c`
    - Output: `세계의모든사람들이한국어를이해한다면얼마나좋을까` (U+C138, U+ACC4, U+C758, U+BAA8, U+B4E0, U+C0AC, U+B78C, U+B4E4, U+C774, U+D55C, U+AD6D, U+C5B4, U+B97C, U+C774, U+D574, U+D55C, U+B2E4, U+BA74, U+C5BC, U+B9C8, U+B098, U+C88B, U+C744, U+AE4C)

14. **Russian (Cyrillic)**
    - Input: `b1abfaaepdrnnbgefbadotcwatmq2g4l`
    - Output: `почемужеонинегоговятпорусски` (U+043F, U+043E, U+0447, U+0435, U+043C, U+0443, U+0436, U+0435, U+043E, U+043D, U+0438, U+043D, U+0435, U+0433, U+043E, U+0432, U+043E, U+0440, U+044F, U+0442, U+043F, U+043E, U+0440, U+0443, U+0441, U+0441, U+043A, U+0438)

15. **Spanish**
    - Input: `PorqunopuedensimplementehablarenEspaol-fmd56a`
    - Output: `PorquénopuedensimplementehablarenEspañol`

16. **Vietnamese**
    - Input: `TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g`
    - Output: `TạisaohkhôngthểchỉnótitếngViệt`

17. **Numeric and CJK**
    - Input: `3B-ww4c5e180e575a65lsy2b`
    - Output: `3年B組金八先生`

18. **Hankaku and ASCII with hyphens**
    - Input: `-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n`
    - Output: `安室奈美恵-with-SUPER-MONKEYS`

19. **Multiple hyphens and CJK**
    - Input: `Hello-Another-Way--fc4qua05auwb3674vfr0b`
    - Output: `Hello-Another-Way-それぞれの場所`

20. **CJK and number**
    - Input: `2-u9tlzr9756bt3uc0v`
    - Output: `ひとつ屋根の下2`

21. **ASCII and CJK mix**
    - Input: `MajiKoi5-783gue6qz075azm5e`
    - Output: `Majiで Koi する 5 秒前`

22. **Katakana and ASCII**
    - Input: `de-jg4avhby1noc0d`
    - Output: `パフィー de ルンバ`

23. **Katakana**
    - Input: `d9juau41awczczp`
    - Output: `そのスピードで`

24. **ASCII with special characters**
    - Input: `-> $1.00 <--`
    - Output: `-> $1.00 <-`

## Cross-References

- [test-fixtures.md](./test-fixtures.md) — Complete description of testData.strings fixture structure
- [punycode-encode.md](./punycode-encode.md) — Inverse operation; encodes Unicode strings to Punycode
- [punycode-to-unicode.md](./punycode-to-unicode.md) — Domain-level wrapper that handles the `xn--` prefix and applies decode to each label
- [punycode-ucs2-decode.md](./punycode-ucs2-decode.md) — Utility function that converts UCS-2 strings to code point arrays; related to the overflow and illegal-input error cases tested at tests.js:255-270
