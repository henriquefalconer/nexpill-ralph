# punycode.encode — Specification

## Purpose
Encodes a Unicode string (intended to be a single domain-name label) into its Punycode ASCII-only form per RFC 3492 Bootstring. It is a label-level primitive and does NOT prepend the `xn--` ACE prefix; adding that prefix when the input contains non-ASCII code points is the responsibility of `toASCII`. See the JSDoc and signature at `punycode.js:283-290`, and the public API wiring at `punycode.js:438`.

## Contract
- Input: a Unicode string representing a single label (no dot-separation is done here).
- Output: an ASCII-only Punycode string. Any basic (ASCII, code point < 0x80) characters are copied through verbatim in input order; non-basic code points are encoded as a sequence of lowercase ASCII letters and digits after a `-` delimiter (when a basic prefix exists). See `punycode.js:304-320`.
- Case of generated digits: always lowercase. The implementation emits variable-length-integer digits by calling the digit-to-basic conversion with the mixed-case flag forced to zero at `punycode.js:358-364`; uppercase (`A-Z`) Punycode digits are therefore never produced by this function. The behaviour of that conversion is defined at `punycode.js:168-172`.
- Error: on arithmetic overflow during delta or `n` advancement, throws a `RangeError` with the message `"Overflow: input needs wider integers to process"`. Error table at `punycode.js:22-26`; throw sites at `punycode.js:337-339` and `punycode.js:344-347`.

## Bootstring parameters
The constants that drive the algorithm are declared at `punycode.js:6-14`:
- `base = 36`
- `tMin = 1`
- `tMax = 26`
- `skew = 38`
- `damp = 700`
- `initialBias = 72`
- `initialN = 128` (i.e. `0x80`, first non-basic code point)
- `delimiter = '-'` (ASCII `0x2D`)
Additionally `maxInt = 2147483647` (the maximum signed 32-bit value) is declared at `punycode.js:4` and used as the overflow guard.

## Algorithm (high level)
The numbered steps below mirror the body of `encode` at `punycode.js:290-376`. Describe in prose, not in JS:

1. Decode the input from UCS-2/UTF-16 code units to a sequence of Unicode code points, combining surrogate pairs into single scalar values. This is done at `punycode.js:294` by invoking the helper defined at `punycode.js:101-123`.
2. Initialize encoder state: `n = initialN` (128), `delta = 0`, `bias = initialBias` (72). See `punycode.js:300-302`.
3. Walk the code-point sequence once and copy every basic code point (value < 0x80) to the output buffer in input order. See `punycode.js:304-309`. Record the number of basic code points copied as `basicLength`, and set `handledCPCount = basicLength` (`punycode.js:311-312`).
4. If `basicLength > 0`, append the literal delimiter `-` to the output. If the input had no basic code points at all, no delimiter is emitted and the extended section begins immediately. See `punycode.js:318-320`.
5. Main loop — while `handledCPCount < inputLength` (`punycode.js:323`):
   a. Find `m`, the smallest code point in the input that is `>= n`. Initialize `m = maxInt` before scanning (`punycode.js:327-332`).
   b. Overflow-guard: if `m - n > floor((maxInt - delta) / (handledCPCount + 1))`, raise the `overflow` error (`punycode.js:337-339`). Otherwise advance `delta += (m - n) * (handledCPCount + 1)` and set `n = m` (`punycode.js:341-342`).
   c. Walk the code-point sequence again. For each `currentValue`:
      - If `currentValue < n`, pre-increment `delta`; if that increment exceeds `maxInt`, raise the `overflow` error (`punycode.js:344-347`).
      - If `currentValue == n`, encode the current `delta` as a generalized variable-length integer by repeatedly computing the digit threshold `t = k <= bias ? tMin : (k >= bias + tMax ? tMax : k - bias)` for `k` starting at `base` and incrementing by `base` each iteration (`punycode.js:350-355`). While `q >= t`, emit the ASCII digit for `t + (q - t) mod (base - t)` via the digit-to-basic conversion with `flag = 0`, then set `q = floor((q - t) / (base - t))` (`punycode.js:356-362`). When `q < t`, emit the final digit for `q` (`punycode.js:364`). Then update `bias = adapt(delta, handledCPCount + 1, handledCPCount == basicLength)`, reset `delta = 0`, and increment `handledCPCount` (`punycode.js:365-367`).
   d. After the per-point walk, increment both `delta` and `n` (`punycode.js:371-372`).
6. Concatenate and return the buffered output (`punycode.js:375`).

The bias-adaptation helper used in step 5c is defined at `punycode.js:179-187`; on the first invocation (i.e. when `handledCPCount == basicLength`) the raw delta is divided by `damp = 700` before the usual adaptation, otherwise it is halved. See `punycode.js:181`.

## Behavior rules
Each rule below is exercised by at least one fixture in `tests/tests.js:6-136` (the `testData.strings` array iterated by the describe block at `tests/tests.js:312-321`).

- Single basic code point: an all-ASCII input with no non-basic code points produces the input itself followed by the trailing `-` delimiter. Example: `Bach` → `Bach-`. Fixture at `tests/tests.js:7-12`; delimiter emission at `punycode.js:318-320`.
- Pure non-ASCII input (no basic prefix): when the input contains no code points below `0x80`, no leading delimiter and no trailing delimiter appear before the extended portion — the output is exactly the variable-length-integer section. Example: `U+00FC` → `tda`. Fixture at `tests/tests.js:13-17`; the guarded delimiter branch at `punycode.js:318-320` skips emission when `basicLength == 0`.
- Multiple non-ASCII: several non-basic code points, some above the BMP multilingual subset, encode through repeated invocations of the main loop in step 5. Example: `U+00FC U+00EB U+00E4 U+00F6 U+2665` → `4can8av2009b`. Fixture at `tests/tests.js:18-22`; loop at `punycode.js:323-374`.
- Mixed ASCII + non-ASCII: basic code points copy verbatim, then the delimiter, then the extended encoding of the non-basic code points. Example: `bücher` → `bcher-kva`. Fixture at `tests/tests.js:23-27`; see `punycode.js:304-320`.
- Long mixed strings: the algorithm is input-length-independent beyond the overflow guard. Fixture at `tests/tests.js:28-32`.
- RFC 3492 §7.1 canonical vectors (Arabic, Chinese simplified, Chinese traditional, Czech, Hebrew, Hindi, Japanese, Korean, Russian, Spanish, Vietnamese): each vector is the Bootstring reference encoding and must round-trip bit-exactly. Fixtures at `tests/tests.js:33-97`.
- Mixed-case annotation is NOT supported. The RFC 3492 sample string for Russian (Cyrillic) encodes to `b1abfaaepdrnnbgefbaDotcwatmq2g4l` with a capital `D` carrying the mixed-case flag; this implementation emits the all-lowercase form `b1abfaaepdrnnbgefbadotcwatmq2g4l` because the generalized-variable-length-integer digit emission always passes `flag = 0` into the digit conversion. Explanatory JSDoc at `tests/tests.js:74-82`, fixture at `tests/tests.js:83-87`, implementation detail at `punycode.js:358-364` (and the flag semantics defined at `punycode.js:168-172`).
- Astral / supplementary-plane friendly input: the Japanese fixtures combine CJK and kana code points. `ucs2decode` runs first and combines surrogate pairs into single scalar code points before the encoder sees them, so BMP-straddling inputs work without special handling in the encoder proper. Fixtures at `tests/tests.js:98-125`; surrogate combining at `punycode.js:107-117` invoked from `punycode.js:294`.
- ASCII-only input that violates host-name rules: when the input is entirely ASCII but contains characters disallowed in DNS labels (spaces, `$`, `.`, `<`, `>`, `-`, digits adjacent to letters, etc.), every character is copied verbatim into the basic prefix and a single trailing `-` delimiter is appended. Example: `-> $1.00 <-` → `-> $1.00 <--` (note the double trailing hyphen: the last `-` of the output is the Bootstring delimiter following the input's own trailing `-`). Fixture and explanatory JSDoc at `tests/tests.js:126-135`; basic-prefix copy at `punycode.js:304-309`, delimiter append at `punycode.js:318-320`.
- Note: the `RangeError: Invalid input` / `ls8h=` / uppercase-`ZZZ` cases at `tests/tests.js:300-308` belong to the `punycode.decode` describe block and are out of scope for this specification.

## Error conditions
The `encode` function can throw exactly one error type, the `overflow` `RangeError` with message `"Overflow: input needs wider integers to process"`. Two guards produce it:
- Before bulk-advancing delta to move the decoder state from `<n, i>` to `<m, 0>`: if `m - n > floor((maxInt - delta) / (handledCPCount + 1))`, overflow is raised. See `punycode.js:337-339`.
- While scanning for code points less than `n` during the main loop, `++delta` is checked after the increment; if it exceeds `maxInt`, overflow is raised. See `punycode.js:344-347`.
No fixture in `testData.strings` triggers either overflow path, so the describe block at `tests/tests.js:312-321` does not assert error behavior for `encode`.

## Port notes (Go)
- Use a signed 32-bit integer type (`int32`) for `n`, `delta`, `bias`, `handledCPCount`, and the thresholds `t`, `q`. This mirrors the `maxInt = 2^31 - 1` guard at `punycode.js:4,337-339,344-347` and makes the overflow checks a direct translation. A Go `int` is at least 32 bits, but pinning to `int32` preserves the intent.
- Convert the input to code points up front using `[]rune(input)`. Because Go strings are UTF-8, this already produces scalar code points and no explicit surrogate-pair combining (the job of `ucs2decode` at `punycode.js:101-123`) is required. Length and indexed iteration should both use the rune slice.
- Always emit lowercase Punycode digits. Translate `digitToBasic(x, 0)` at `punycode.js:168-172` as: if `x < 26` return `byte('a' + x)`, else return `byte('0' + (x - 26))`. Do not expose a mixed-case flag.
- Use integer division (`/` on `int32`) for all `floor(...)` occurrences in `punycode.js:181-186, 337, 361`; Go's `/` on non-negative integers already truncates toward zero.
- Return `error` values rather than panicking on overflow, unless the target Go API defined in `TARGET.md` specifies otherwise.

## Test-vector source
All vectors listed in `specs/test-data-fixtures.md` under the `strings` section; originals at `tests/tests.js:6-136`.
