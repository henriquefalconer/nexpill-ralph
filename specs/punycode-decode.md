# punycode.decode — Specification

## Purpose

This primitive decodes a Punycode (ASCII-only) label into a string of Unicode symbols per the Bootstring algorithm of RFC 3492. It operates at the label level: given an already-stripped label body, it returns the corresponding Unicode characters. It does NOT recognize or strip the `xn--` ACE prefix — that responsibility belongs to the higher-level `toUnicode` wrapper which splits a domain into labels and only invokes this primitive on labels starting with `xn--`. Callers that feed it a raw domain will get back a literal decoding of whatever label body they provided. See `punycode.js:189-196` for the function header, and the public export wiring at `punycode.js:437` which surfaces this as `punycode.decode`.

## Contract

- Input: a Punycode label as an ASCII string. A single trailing `-` is valid and is interpreted as the Bootstring delimiter terminating the basic-code-point prefix (see `punycode.js:208-211`).
- Input case: uppercase A-Z is accepted and treated equivalently to lowercase a-z during base-36 digit decoding (case folds to digit value 0-25 for either range), as demonstrated by the dedicated test `'ZZZ' -> U+7BA5` at `tests/tests.js:299-301` and implemented by the two parallel ranges at `punycode.js:148-150` (A-Z) and `punycode.js:151-153` (a-z).
- Output: a Unicode string (a sequence of Unicode code points re-assembled from the decoded code-point array, per `punycode.js:280`).
- Error: throws `RangeError` on failure. Exactly three messages are possible: `"Overflow: input needs wider integers to process"`, `"Illegal input >= 0x80 (not a basic code point)"`, and `"Invalid input"` (defined at `punycode.js:22-26`, thrown via the `error(type)` helper at `punycode.js:41-43`).

## Bootstring parameters (RFC 3492)

The following numeric constants govern the algorithm; they are defined at `punycode.js:6-14`:

- `base = 36` (`punycode.js:7`).
- `tMin = 1` (`punycode.js:8`).
- `tMax = 26` (`punycode.js:9`).
- `skew = 38` (`punycode.js:10`).
- `damp = 700` (`punycode.js:11`).
- `initialBias = 72` (`punycode.js:12`).
- `initialN = 128`, i.e. U+0080 (`punycode.js:13`).
- `delimiter = '-'`, i.e. U+002D (`punycode.js:14`).

Two derived values matter for the implementation:

- `baseMinusTMin = base - tMin = 35`, used by both the threshold clamp and bias-adaptation loop (`punycode.js:29`).
- `maxInt = 2^31 - 1 = 2147483647` is the 32-bit overflow guard used throughout delta accumulation and the `w` (weight) update (`punycode.js:4`).

## Algorithm (high level)

The decoder follows RFC 3492 section 6.2. The steps below map to the JavaScript source for traceability; the Go port should implement the same data flow in prose terms, not the literal JS.

1. Initialize the output code-point array, the running insertion index `i = 0`, the running code-point value `n = initialN`, and `bias = initialBias` (`punycode.js:197-202`).
2. Locate the last `-` delimiter. Everything before it is the basic (ASCII) prefix and is copied verbatim to the output. If there is no `-`, the basic prefix is empty (`punycode.js:208-219`).
3. While copying the basic prefix, every character's code unit must be `< 0x80`; any character at or above that threshold raises the `not-basic` error (`punycode.js:215-216`).
4. Position the cursor just past the delimiter (if any basic-prefix bytes were copied), otherwise at the start of the input, and enter the main decoding loop (`punycode.js:224`).
5. Inside the main loop, decode one generalized variable-length integer per iteration. Record the pre-loop value `oldi`, then iterate `k` by steps of `base` starting at `base`, reading one base-36 digit per step until a digit below the per-step threshold `t` terminates the integer (`punycode.js:231-261`).
6. Each base-36 digit is obtained from a single ASCII character via the mapping: `A-Z` and `a-z` -> `0..25`, `0-9` -> `26..35`; any other byte maps to `base` and raises `invalid-input` (`punycode.js:144-155`, enforced at `punycode.js:240-242`).
7. If the cursor runs past the end of the input while still mid-integer, raise `invalid-input` (`punycode.js:234-236`).
8. On each digit accumulation, enforce the overflow guard `digit > floor((maxInt - i) / w)` before updating `i += digit * w` (`punycode.js:243-247`). Compute `t = clamp(k - bias, tMin, tMax)` (`punycode.js:248`). If `digit < t`, the integer is complete — break. Otherwise enforce the weight overflow guard `w > floor(maxInt / (base - t))` before `w *= (base - t)` (`punycode.js:250-259`).
9. After each completed delta, compute `out = len(output) + 1` and adapt the bias: `bias = adapt(i - oldi, out, oldi == 0)` (`punycode.js:263-264`). The adapt function (`punycode.js:179-187`) first divides delta by `damp` on the very first call and by 2 thereafter, then divides by `numPoints`, then repeatedly divides by `baseMinusTMin` while `delta` exceeds `(baseMinusTMin * tMax) / 2`, counting `k += base` per step; the returned bias is `k + ((baseMinusTMin + 1) * delta) / (delta + skew)` with integer division throughout.
10. Translate the accumulated `i` into a `(n, insertion-index)` pair. Guard `floor(i / out) > maxInt - n` as `overflow` (`punycode.js:268-270`), then increment `n += floor(i / out)` and set `i = i % out` (`punycode.js:272-273`).
11. Insert the code point `n` at position `i` in the output array, then increment `i` (`punycode.js:276`).
12. When the input is exhausted, assemble the output code-point array into a Unicode string via UTF-16 code-point emission (`punycode.js:280`).

## Behavior rules

Each rule below is anchored to a concrete test fixture in the `testData.strings` array at `tests/tests.js:6-136` (iterated by the describe block at `tests/tests.js:290-297`) and to the implementation line that enforces it.

- A label with a single basic code point followed by a bare `-` yields the basic prefix verbatim and enters the main loop with no further work because the input is already exhausted: `'Bach-' -> 'Bach'` (`tests/tests.js:7-12`, enforced by the basic-prefix copy at `punycode.js:208-219` and the loop guard `index < inputLength` at `punycode.js:224`).
- A label with no `-` at all has an empty basic prefix (because `lastIndexOf` returns `-1`, which is clamped to `0`) and the entire input is interpreted as encoded variable-length integers. The first character of such a label will be a base-36 digit, not literal content. Example: `'tda' -> '\u00FC'` (`tests/tests.js:13-17`, enforced at `punycode.js:208-211`).
- Multi-character non-ASCII sequences follow the same rule — there is no `-` in the encoded form, and successive deltas walk `n` through the target code points: `'4can8av2009b' -> '\u00FC\u00EB\u00E4\u00F6\u2665'` (`tests/tests.js:18-22`, full algorithm at `punycode.js:224-278`).
- Mixed ASCII and non-ASCII: the delimiter separates the ASCII prefix from the integer stream, with the prefix preserved in input order and the non-ASCII characters inserted into the output at computed positions. Example: `'bcher-kva' -> 'b\u00FCcher'` — `bcher` is copied, then a single code point U+00FC is inserted at index 1 (`tests/tests.js:23-27`, basic copy at `punycode.js:213-219`, insertion at `punycode.js:276`).
- Long mixed strings exercise the same mechanism at scale: `'Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal' -> 'Willst du die Bl\u00FCthe des fr\u00FChen, die Fr\u00FCchte des sp\u00E4teren Jahres'` (`tests/tests.js:28-32`).
- RFC 3492 section 7.1 canonical vectors must round-trip exactly. Each appears in `testData.strings`: Arabic/Egyptian (`tests/tests.js:34-38`), Chinese simplified (`tests/tests.js:39-43`), Chinese traditional (`tests/tests.js:44-48`), Czech (`tests/tests.js:49-53`), Hebrew (`tests/tests.js:54-58`), Hindi/Devanagari (`tests/tests.js:59-63`), Japanese kanji+hiragana (`tests/tests.js:64-68`), Korean Hangul (`tests/tests.js:69-73`), Russian Cyrillic (`tests/tests.js:83-87` — note the preceding block comment at `tests/tests.js:74-82` explaining that mixed-case annotation is intentionally unsupported), Spanish (`tests/tests.js:88-92`), Vietnamese (`tests/tests.js:93-97`).
- Anonymous fixtures (those without a `description` field) must also decode correctly; the describe block falls back to `object.encoded` as the test name at `tests/tests.js:292`. These cover Japanese/Latin mixtures: `'3B-ww4c5e180e575a65lsy2b'` (`tests/tests.js:98-101`), `'-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n'` (`tests/tests.js:102-105`), `'Hello-Another-Way--fc4qua05auwb3674vfr0b'` (`tests/tests.js:106-109`), `'2-u9tlzr9756bt3uc0v'` (`tests/tests.js:110-113`), `'MajiKoi5-783gue6qz075azm5e'` (`tests/tests.js:114-117`), `'de-jg4avhby1noc0d'` (`tests/tests.js:118-121`), `'d9juau41awczczp'` (`tests/tests.js:122-125`).
- ASCII strings that happen to include literal `-` characters must still decode correctly because `lastIndexOf(delimiter)` finds the FINAL `-`, so any earlier `-` characters are part of the basic prefix. The fixture `'-> $1.00 <--' -> '-> $1.00 <-'` demonstrates this: the trailing pair consists of one literal `-` (part of the basic prefix) and one Bootstring delimiter (`tests/tests.js:126-135`, rule at `punycode.js:208`). The implementation correctly handles the case where the delimiter is the last character and there is no encoded-integer remainder.
- Uppercase A-Z digits case-fold to the same values as lowercase. The extra test at `tests/tests.js:299-301` asserts `punycode.decode('ZZZ') -> '\u7BA5'`; the three `Z` bytes all decode to digit value 25 (since `'Z' - 'A' = 25`) via the A-Z branch at `punycode.js:148-150`.
- Any input byte that is not a letter or digit in the base-36 alphabet raises `invalid-input`. The extra test at `tests/tests.js:302-309` asserts `punycode.decode('ls8h=')` throws `RangeError` because `=` (U+003D) falls outside all three ranges in `basicToDigit` and therefore returns `base`, tripping the guard at `punycode.js:240-242`.

## Error conditions

The function throws `RangeError` with one of three exact messages from the `errors` table at `punycode.js:22-26`:

- `"Illegal input >= 0x80 (not a basic code point)"` (`not-basic`). Raised when any character in the basic prefix (everything before the last `-`) has a code unit `>= 0x80`. Enforcement: `punycode.js:215-216`.
- `"Overflow: input needs wider integers to process"` (`overflow`). Raised by three distinct 32-bit guards: the `digit * w` accumulation guard at `punycode.js:243-245`, the weight multiplication guard at `punycode.js:255-257`, and the `n` advancement guard at `punycode.js:268-270`.
- `"Invalid input"` (`invalid-input`). Raised either when a non-base-36 byte is encountered (`punycode.js:240-242`) or when the input stream is exhausted mid-integer before a terminating digit below `t` has been seen (`punycode.js:234-236`).

No test in the `punycode.decode` describe block at `tests/tests.js:290-310` directly exercises `not-basic` or `overflow`; those two error paths are instead asserted inside the `punycode.ucs2.decode` describe block at `tests/tests.js:255-270`, where the assertions still target `punycode.decode` (see `tests/tests.js:258` calling `punycode.decode('\x81-')` for `not-basic`, and `tests/tests.js:266` calling `punycode.decode('\x81')` for `overflow`). The Go port must preserve both error paths and their exact messages.

## Port notes (Go)

- Use `int32` (or `int` with explicit range checks) for the Bootstring arithmetic variables `i`, `w`, `delta`, `bias`, `n`, and `oldi`. The JS implementation's `maxInt = 2^31-1` guards at `punycode.js:4`, `243`, `255`, `268` are specifically calibrated for 32-bit signed overflow and must be replicated verbatim; do not widen to `int64` and skip the checks, because callers depend on the overflow error being raised on the exact same inputs.
- The output is a sequence of Unicode code points. Use `[]rune` for the accumulator and convert to `string` once at the end; rune insertion via `append` + `copy` replicates the JS `output.splice(i, 0, n)` at `punycode.js:276`. For code points above U+FFFF, Go's `rune` is already a single value, simpler than JS which requires `String.fromCodePoint` at `punycode.js:280` to emit a surrogate pair.
- Preserve case-insensitivity for A-Z vs a-z digit decoding by implementing both ASCII ranges explicitly as in `punycode.js:144-155`. Do not lowercase the whole input up-front — the basic prefix must be returned with its original case.
- See `TARGET.md` at the repo root for the expected Go module layout under `port/` and the requirement that every test vector in `tests/tests.js` is preserved in the Go test suite.

## Test-vector source

All vectors listed in `specs/test-data-fixtures.md` under the `strings` section; originals at `tests/tests.js:6-136`.
