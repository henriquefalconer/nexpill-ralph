# `punycode.ucs2.encode` — Behaviour Specification

## Purpose

`punycode.ucs2.encode(codePoints)` converts an array of numeric Unicode code points into a JavaScript string composed of the corresponding UCS-2/UTF-16 code units. It is the exact inverse of `punycode.ucs2.decode`; see [ucs2-decode.md](ucs2-decode.md) for the complementary direction.

**Implementation** (`punycode.js:133`):

```js
const ucs2encode = codePoints => String.fromCodePoint(...codePoints);
```

The function is a single-expression arrow function that spreads the input array as arguments to the native `String.fromCodePoint` built-in. No manual iteration, no internal buffer — the entire transformation is delegated to the JavaScript engine's own Unicode handling.

---

## Signature

```
punycode.ucs2.encode(codePoints: number[]): string
```

| Parameter     | Type       | Description                                                   |
|---------------|------------|---------------------------------------------------------------|
| `codePoints`  | `number[]` | An array of integer Unicode code points (U+0000 – U+10FFFF, or lone surrogate values). |

**Returns** a new JavaScript `string` whose UTF-16 encoding represents exactly those code points in order.

---

## Relationship to `testData.ucs2` Fixtures

The test suite defines a shared fixture array `testData.ucs2` (`tests/tests.js:137-175`). Each element has:

- `description` — human-readable label for the case.
- `decoded` — the array of numeric code points.
- `encoded` — the expected JavaScript string.

The `punycode.ucs2.decode` suite (`tests/tests.js:245-271`) iterates these fixtures in the **string → array** direction. The `punycode.ucs2.encode` suite (`tests/tests.js:273-288`) iterates the **same** fixture objects in the opposite direction — feeding `decoded` into `encode` and asserting the result equals `encoded`. This shared-fixture design guarantees that both functions satisfy a round-trip property: for every fixture, `encode(decode(encoded)) === encoded` and `decode(encode(decoded))` deep-equals `decoded`.

```js
// tests/tests.js:274-280
for (const object of testData.ucs2) {
    it(object.description, function() {
        assert.deepEqual(
            punycode.ucs2.encode(object.decoded),
            object.encoded
        );
    });
}
```

---

## Fixture Cases and the Properties They Verify

### 1. Consecutive astral symbols (`tests/tests.js:140-144`)

```
decoded: [127829, 119808, 119558, 119638]
encoded: '🍕𝐀𝌆𝍖'
```

Each code point is above U+FFFF (the BMP boundary) and therefore requires a surrogate pair in UTF-16. Verifies that `String.fromCodePoint` correctly generates two UTF-16 code units per astral code point, and that the resulting string has the right sequence of surrogate pairs for four consecutive astral symbols.

### 2. U+D800 (high surrogate) followed by non-surrogates (`tests/tests.js:145-149`)

```
decoded: [55296, 97, 98]   // 0xD800, 'a', 'b'
encoded: '\uD800ab'
```

Code point U+D800 is itself a high-surrogate value. When passed individually to `String.fromCodePoint`, it is emitted as a lone surrogate code unit (U+D800). Verifies that lone high-surrogate code points round-trip to the identical lone surrogate code unit in the output string.

### 3. U+DC00 (low surrogate) followed by non-surrogates (`tests/tests.js:150-154`)

```
decoded: [56320, 97, 98]   // 0xDC00, 'a', 'b'
encoded: '\uDC00ab'
```

Mirrors case 2 for a lone low-surrogate code point. Verifies the same round-trip property for U+DC00.

### 4. High surrogate followed by another high surrogate (`tests/tests.js:155-159`)

```
decoded: [0xD800, 0xD800]
encoded: '\uD800\uD800'
```

Both code points are lone high surrogates. Verifies that two consecutive lone high-surrogate code points produce two consecutive lone high-surrogate code units without being interpreted as a valid surrogate pair.

### 5. Unmatched high surrogate, surrogate pair, unmatched high surrogate (`tests/tests.js:160-164`)

```
decoded: [0xD800, 0x1D306, 0xD800]
encoded: '\uD800𝌆\uD800'
```

Mixes a lone high surrogate, a valid astral code point (U+1D306, encoded as the surrogate pair `𝌆`), and another lone high surrogate. Verifies correct handling of interleaved lone surrogates and legitimate surrogate pairs.

### 6. Low surrogate followed by another low surrogate (`tests/tests.js:165-169`)

```
decoded: [0xDC00, 0xDC00]
encoded: '\uDC00\uDC00'
```

Mirrors case 4 for lone low surrogates.

### 7. Unmatched low surrogate, surrogate pair, unmatched low surrogate (`tests/tests.js:170-174`)

```
decoded: [0xDC00, 0x1D306, 0xDC00]
encoded: '\uDC00𝌆\uDC00'
```

Mirrors case 5 for lone low surrogates.

---

## Assertion Mechanics

Every round-trip assertion uses `assert.deepEqual` (`tests/tests.js:276`):

```js
assert.deepEqual(
    punycode.ucs2.encode(object.decoded),
    object.encoded
);
```

`assert.deepEqual` performs a structural equality check. For primitive `string` values this is equivalent to strict `===` equality.

---

## "Does Not Mutate Argument Array" Test (`tests/tests.js:282-287`)

```js
const codePoints = [0x61, 0x62, 0x63];
const result = punycode.ucs2.encode(codePoints);
it('does not mutate argument array', function() {
    assert.deepEqual(result, 'abc');
    assert.deepEqual(codePoints, [0x61, 0x62, 0x63]);
});
```

This test encodes the three-element array `[0x61, 0x62, 0x63]` (code points for `'a'`, `'b'`, `'c'`) and asserts two things:

1. The return value is the string `'abc'`.
2. The original `codePoints` array is still `[0x61, 0x62, 0x63]` after the call.

**Why this holds:** The implementation (`punycode.js:133`) uses the spread operator (`...codePoints`) to pass array elements as individual arguments to `String.fromCodePoint`. The spread operator reads values from the array but does not write back to it. `String.fromCodePoint` itself accepts only variadic numeric arguments and has no reference to the original array. Therefore the source array is guaranteed to be untouched.

Note that this fixture (`codePoints` / `result`) is constructed inline in the test body (`tests/tests.js:282-283`), outside the `testData.ucs2` loop.

---

## Cross-References

- [ucs2-decode.md](ucs2-decode.md) — `punycode.ucs2.decode` is the inverse of this function. It accepts a JavaScript string and returns an array of numeric code points (`tests/tests.js:245-271`, JSDoc at `punycode.js:125-132`). The same `testData.ucs2` fixtures (`tests/tests.js:137-175`) are used by both suites, ensuring the two functions are consistent inverses of each other.
