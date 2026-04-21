# Spec: `map(array, callback)` private helper

Source definition: `punycode.js:45-60`

---

## 1. Subject

**Function:** `map(array, callback)` — a reverse-iteration `Array#map`-like utility.

**Visibility:** `@private` (`punycode.js:47`)

**Purpose:** Provides array mapping semantics without relying on the native `Array.prototype.map` method.

---

## 2. Contract

Derived from JSDoc at `punycode.js:45-52`.

The function accepts two arguments:

1. **`array`** (Array): The array to iterate over.
2. **`callback`** (Function): A function that receives each array element and returns a transformed value.

**Return value:** A new array (`punycode.js:51`) whose element at index `i` is the result of `callback(array[i])`. The output array has the same length as the input array, and the index mapping is preserved: `result[i]` corresponds to the callback's return value for `array[i]`.

---

## 3. Implementation walkthrough

The implementation at `punycode.js:53-60`:

```javascript
function map(array, callback) {
	const result = [];
	let length = array.length;
	while (length--) {
		result[length] = callback(array[length]);
	}
	return result;
}
```

**Iteration strategy:** The function iterates backwards over the input array using a `while` loop with a decremented counter (`punycode.js:56`). The counter starts at `array.length` and decrements until it reaches 0 (the condition `while (length--)` is false when `length` wraps from 0 to –1 in the unsigned interpretation, but in practice the loop terminates before negative values are used due to short-circuit evaluation of falsy 0).

**Index preservation:** Although the traversal is backwards (from high indices to low), the assignment at `punycode.js:57` writes the callback result to `result[length]` using the same decremented index. This means:

- When `length` is 0 (last iteration of the while loop, for the first array element), `callback(array[0])` is computed and assigned to `result[0]`.
- When `length` is 1, the result goes to `result[1]`, etc.

Thus the output preserves the original index alignment, and the semantic order of elements is correct despite the backward traversal.

---

## 4. Why hand-rolled

The function implements mapping without calling the native `Array.prototype.map`. This is likely a legacy micro-optimization or compatibility measure. The code's own comment at `punycode.js:81` ("Avoid `split(regex)` for IE8 compatibility. See #17") hints at the broader pattern: this library targets IE8 or earlier JavaScript engines where native methods were either unavailable, slow, or subject to quirks. Rolling a simple `while`-based loop avoids method lookup overhead and the potential variability of `Array.prototype.map` behavior across older environments.

---

## 5. Callsites

**Location:** `punycode.js:84`

Within the `mapDomain` function (`punycode.js:72-86`), the `map` helper is invoked to process an array of domain labels:

```javascript
const encoded = map(labels, callback).join('.');
```

Here, `labels` is an array of domain name segments (obtained by splitting on the dot separator at `punycode.js:83`), and `callback` is the user-supplied function passed to `mapDomain`. The result of `map(labels, callback)` is an array of transformed labels, which are then joined back into a dot-separated domain string.

**No other callsites within `punycode.js`.**

---

## 6. Edge cases

**Empty array:** If `array.length` is 0, the `while (length--)` loop never executes (the condition evaluates to false immediately), and the function returns an empty array `[]` (`punycode.js:54`). This is correct behavior.

**Sparse arrays:** If the input array has missing elements (sparse array), those indices will have `undefined` passed to the callback. The callback's return value is still assigned to the corresponding index in the output array. The sparsity is not propagated; the output array will be densely populated with whatever the callback returns (including `undefined` if the callback returns it).

**Falsy callbacks:** The function does not validate that `callback` is callable. If a non-function is passed, the assignment `callback(array[length])` will throw a `TypeError` at runtime.

---

## 7. Implementation citations

| Item | Location |
|---|---|
| Function definition | `punycode.js:53-60` |
| JSDoc contract | `punycode.js:45-52` |
| Backwards iteration with decrement | `punycode.js:56` |
| Index-preserving assignment | `punycode.js:57` |
| Callsite in `mapDomain` | `punycode.js:84` |
| IE8 compatibility context | `punycode.js:81` |
