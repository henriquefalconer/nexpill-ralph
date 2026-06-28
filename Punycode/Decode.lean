-- Punycode decoder: converts a bare Punycode ASCII label to a Unicode string.
-- Ports punycode.js:196-281; behaviour spec: specs/impl-decode.md.

import Punycode.Constants
import Punycode.Helpers
import Punycode.Bootstring

namespace Punycode

-- Insert element x at position pos in arr (clamped to arr.size).
-- Ports `output.splice(i, 0, n)` (punycode.js:276).
private def arrayInsert (arr : Array UInt32) (pos : Nat) (x : UInt32) : Array UInt32 :=
  let pos := min pos arr.size
  arr.extract 0 pos ++ #[x] ++ arr.extract pos arr.size

-- Variable-length integer inner loop (punycode.js:232-261).
-- `index` advances each iteration; the loop breaks when `digit < t`.
-- Declared `partial`: Lean cannot prove termination from the `digit < t` break condition
-- automatically; the loop trivially terminates because the digit stream is finite and
-- `index` strictly increases each iteration.
private partial def decodeVLI
    (chars : Array Char) (inputLength : Nat)
    (index : Nat) (i : Nat) (bias : Nat) (w : Nat) (k : Nat)
    : Except PunyError (Nat × Nat) := do
  if index >= inputLength then throw .invalidInput
  let digit := basicToDigit (chars[index]!).val
  -- Not a valid base-36 digit (punycode.js:240-242).
  if digit >= base then throw .invalidInput
  -- Guard: i += digit * w would overflow (punycode.js:243-245).
  if w != 0 && digit > (maxInt - i) / w then throw .overflow
  let i' := i + digit * w
  let t := if k <= bias then tMin
            else if k >= bias + tMax then tMax
            else k - bias
  -- Break: this digit is the most-significant digit (punycode.js:250-252).
  if digit < t then return (i', index + 1)
  -- Guard: w *= baseMinusT would overflow (punycode.js:255-257).
  let baseMinusT := base - t
  if w > maxInt / baseMinusT then throw .overflow
  decodeVLI chars inputLength (index + 1) i' bias (w * baseMinusT) (k + base)

-- Outer main decoding loop (punycode.js:224-278).
-- Each iteration decodes one VLI delta and inserts one code point into output.
-- Declared `partial` for the same reason as decodeVLI.
private partial def decodeLoop
    (chars : Array Char) (inputLength : Nat)
    (index : Nat) (i : Nat) (n : Nat) (bias : Nat)
    (output : Array UInt32)
    : Except PunyError (Array UInt32) := do
  if index >= inputLength then return output
  let oldi := i
  let (i', idx') ← decodeVLI chars inputLength index i bias 1 base
  let out := output.size + 1
  let bias' := adapt (i' - oldi) out (oldi == 0)
  -- Guard: n += floor(i'/out) would overflow (punycode.js:268-270).
  if i' / out > maxInt - n then throw .overflow
  let n' := n + i' / out
  let pos := i' % out
  -- Insert code point n' at position pos; advance i to pos+1 (punycode.js:276).
  let output' := arrayInsert output pos n'.toUInt32
  decodeLoop chars inputLength idx' (pos + 1) n' bias' output'

/-- Decode a bare Punycode ASCII label to a Unicode string.
    Ports `punycode.js:196-281` (`decode`). Input must not carry an `xn--` prefix.
    Returns `Except.error .notBasic` if a non-basic character appears in the basic
    prefix region, and `Except.error .invalidInput` or `Except.error .overflow` for
    malformed extended data. -/
def decode (input : String) : Except PunyError String := do
  -- Work on a random-access Char array (input is expected to be ASCII).
  let chars := input.toList.toArray
  let inputLength := chars.size
  -- Find the last delimiter (punycode.js:208-211).
  -- JS lastIndexOf returns -1 if absent (clamped to 0); we initialise to 0.
  let basic : Nat := Id.run do
    let mut result := 0
    for i in [:inputLength] do
      if chars[i]! == delimiter then result := i
    return result
  -- Copy the basic code-point prefix to output (punycode.js:213-219).
  let mut output : Array UInt32 := #[]
  for j in [:basic] do
    let cp := (chars[j]!).val.toNat
    if cp >= 0x80 then throw .notBasic
    output := output.push cp.toUInt32
  -- Extended part starts just after the delimiter, or at 0 if none (punycode.js:224).
  let startIndex := if basic > 0 then basic + 1 else 0
  -- Main variable-length decoding loop (punycode.js:224-278).
  let output' ← decodeLoop chars inputLength startIndex 0 initialN initialBias output
  -- Convert code points to a Lean String (punycode.js:280: String.fromCodePoint).
  -- Valid Punycode output contains only valid Unicode scalars; surrogates are filtered
  -- defensively (they cannot appear in correct output).
  return String.ofList (output'.toList.filterMap fun cp =>
    let n := cp.toNat
    if (n < 0xD800 || n >= 0xE000) && n < 0x110000 then some (Char.ofNat n) else none)

end Punycode
