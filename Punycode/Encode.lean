-- Punycode encoder: converts a Unicode string to a bare Punycode ASCII label.
-- Ports punycode.js:290-376; behaviour spec: specs/impl-encode.md.

import Punycode.Constants
import Punycode.Bootstring
import Punycode.UCS2

namespace Punycode

-- Emit the generalized variable-length integer for `q` (punycode.js:351-362).
-- `k` advances by `base` each step until `q < t` (the most-significant digit).
-- Declared `partial`: `q` strictly decreases each step when `q ≥ t` (because
-- `base - t ≥ 1`), but Lean cannot prove this from `Nat.div` automatically.
private partial def encodeVLI (output : Array Char) (q : Nat) (bias : Nat) (k : Nat)
    : Array Char :=
  let t := if k <= bias then tMin
            else if k >= bias + tMax then tMax
            else k - bias
  if q < t then
    -- Most-significant (final) digit (punycode.js:353-355, 364).
    output.push (Char.ofNat (digitToBasic q 0).toNat)
  else
    -- Emit one digit and recurse (punycode.js:356-361).
    let qMinusT    := q - t
    let baseMinusT := base - t
    let ch := Char.ofNat (digitToBasic (t + qMinusT % baseMinusT) 0).toNat
    encodeVLI (output.push ch) (qMinusT / baseMinusT) bias (k + base)

-- Per-code-point pass for one outer iteration (punycode.js:344-369).
-- For each code point: increment delta if cp < n; emit VLI + adapt bias if cp = n.
-- Returns (output', delta', handledCPCount', bias').
private def encodeInnerPass
    (codePoints : Array UInt32) (n : Nat) (basicLength : Nat)
    (delta : Nat) (bias : Nat) (handledCPCount : Nat) (output : Array Char)
    : Except PunyError (Array Char × Nat × Nat × Nat) := do
  let mut out := output
  let mut del := delta
  let mut hcp := handledCPCount
  let mut bi  := bias
  for cp in codePoints do
    let cpNat := cp.toNat
    if cpNat < n then
      -- Increment delta; overflow guard (punycode.js:345-347).
      if del + 1 > maxInt then throw .overflow
      del := del + 1
    else if cpNat == n then
      -- Emit VLI for q = delta, then adapt bias (punycode.js:348-367).
      out := encodeVLI out del bi base
      bi  := adapt del (hcp + 1) (hcp == basicLength)
      del := 0
      hcp := hcp + 1
  return (out, del, hcp, bi)

-- Outer main encoding loop (punycode.js:323-374).
-- Each iteration finds the next minimum code point m ≥ n and encodes all occurrences of m.
-- Declared `partial`: terminates when handledCPCount reaches inputLength (all non-basic
-- code points emitted), but Lean cannot prove this structurally.
private partial def encodeLoop
    (codePoints : Array UInt32) (inputLength : Nat) (basicLength : Nat)
    (n : Nat) (delta : Nat) (bias : Nat) (handledCPCount : Nat) (output : Array Char)
    : Except PunyError (Array Char) := do
  if handledCPCount >= inputLength then return output
  -- Find next minimum code point ≥ n (punycode.js:327-332).
  let m : Nat := codePoints.foldl (fun acc cp =>
    let cpNat := cp.toNat
    if cpNat >= n && cpNat < acc then cpNat else acc) maxInt
  -- Advance delta to <m, 0>; overflow guard (punycode.js:336-342).
  let handledCPCountPlusOne := handledCPCount + 1
  if m - n > (maxInt - delta) / handledCPCountPlusOne then throw .overflow
  let delta' := delta + (m - n) * handledCPCountPlusOne
  -- Per-code-point pass using the new n = m (punycode.js:344-369).
  let (output', delta'', handledCPCount', bias') ←
    encodeInnerPass codePoints m basicLength delta' bias handledCPCount output
  -- Loop tail: advance delta and n (punycode.js:371-372).
  encodeLoop codePoints inputLength basicLength (m + 1) (delta'' + 1) bias' handledCPCount' output'

/-- Encode a Unicode string to a bare Punycode ASCII label (without `xn--` prefix).
    Ports `punycode.js:290-376` (`encode`). `toASCII` prepends `xn--` when needed.
    Returns `Except.error .overflow` if the delta arithmetic would exceed `maxInt`. -/
def encode (input : String) : Except PunyError String := do
  -- Convert to Unicode code points via UCS-2 bridge (punycode.js:294).
  let codePoints := ucs2decode (stringToUCS2 input)
  let inputLength := codePoints.size
  -- Copy basic code points < 0x80 preserving input order (punycode.js:305-309).
  let basicChars : Array Char := codePoints.filterMap fun cp =>
    if cp.toNat < 0x80 then some (Char.ofNat cp.toNat) else none
  let basicLength := basicChars.size
  -- Append delimiter if any basic code points exist (punycode.js:318-320).
  let output : Array Char :=
    if basicLength > 0 then basicChars.push '-' else basicChars
  -- Main encoding loop; state: n=initialN, delta=0, bias=initialBias (punycode.js:300-302).
  let output' ← encodeLoop codePoints inputLength basicLength
      initialN 0 initialBias basicLength output
  return String.ofList output'.toList

end Punycode
