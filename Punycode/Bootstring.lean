-- Bootstring digit/bias helpers for the Punycode codec.
-- Ports punycode.js:144-187; behaviour spec: specs/impl-bootstring-helpers.md.

import Punycode.Constants

namespace Punycode

/-- Maps an ASCII code point to its Bootstring digit value in [0, 35],
    or returns `base` (36) as the "not a digit" sentinel.
    Case-insensitive: uppercase and lowercase letters map identically.
    Ports `punycode.js:144-155` (`basicToDigit`). -/
def basicToDigit (codePoint : UInt32) : Nat :=
  let cp := codePoint.toNat
  if cp >= 0x30 && cp < 0x3A then      -- '0'..'9' → 26..35
    26 + (cp - 0x30)
  else if cp >= 0x41 && cp < 0x5B then  -- 'A'..'Z' → 0..25
    cp - 0x41
  else if cp >= 0x61 && cp < 0x7B then  -- 'a'..'z' → 0..25
    cp - 0x61
  else
    base  -- sentinel: not a Bootstring digit

/-- Maps a digit value in [0, 35] to an ASCII code point.
    With `flag = 0`, letters are lowercase ('a'..'z'); with `flag ≠ 0`, uppercase
    ('A'..'Z'). Undefined for digits 26–35 with `flag ≠ 0`.
    Ports `punycode.js:168-172` (`digitToBasic`) using explicit branches (D3)
    instead of the branchless bit-trick `digit + 22 + 75*(digit<26) - ((flag≠0)<<5)`. -/
def digitToBasic (digit : Nat) (flag : Nat) : UInt32 :=
  (if digit < 26 then
    if flag != 0 then digit + 65   -- A..Z (uppercase)
    else digit + 97                -- a..z (lowercase)
  else
    digit + 22                     -- 0..9 (26+22=48='0', 35+22=57='9')
  ).toUInt32

-- Scaling loop of `adapt` (punycode.js:183-185).
-- Declared `partial` because Lean cannot automatically prove termination of
-- `delta / baseMinusTMin` from a `delta > 455` guard; the loop trivially
-- terminates since baseMinusTMin = 35 > 1 guarantees strict decrease.
private partial def adaptLoop (delta : Nat) (k : Nat) : Nat × Nat :=
  if delta > baseMinusTMin * tMax / 2 then   -- 35 * 26 / 2 = 455
    adaptLoop (delta / baseMinusTMin) (k + base)
  else
    (delta, k)

/-- Bias adaptation per RFC 3492 §3.4.
    Called by both `decode` and `encode` after each inserted code point.
    Ports `punycode.js:179-187` (`adapt`). -/
def adapt (delta : Nat) (numPoints : Nat) (firstTime : Bool) : Nat :=
  -- Step 1: initial scale — heavy damping on first insertion, halve otherwise.
  let delta1 := if firstTime then delta / damp else delta / 2
  -- Step 2: per-point adjustment.
  let delta2 := delta1 + delta1 / numPoints
  -- Step 3: scaling loop (reduces delta until ≤ 455; increases k by base each round).
  let (delta3, k) := adaptLoop delta2 0
  -- Step 4: final bias formula.
  k + (baseMinusTMin + 1) * delta3 / (delta3 + skew)

end Punycode
