-- UCS-2 conversion helpers for the Punycode codec.
-- Ports punycode.js:101-133 over Array UInt16 per D2.
-- Lean String is UTF-8 and Char is a valid Unicode scalar (excludes surrogates),
-- so a JS UCS-2 string is modelled as Array UInt16.
-- Behaviour spec: specs/impl-ucs2.md.

import Punycode.Constants

namespace Punycode

-- Internal tail-recursive helper for ucs2decode.
-- Walks the UInt16 array by index, combining surrogate pairs (punycode.js:101-122).
-- Declared `partial` because the termination measure (input.size - i) requires Nat.sub
-- reasoning that the elaborator cannot discharge automatically; the function trivially
-- terminates since `i` strictly increases on every branch.
private partial def ucs2decodeAux
    (input : Array UInt16) (i : Nat) (acc : Array UInt32) : Array UInt32 :=
  if h : i < input.size then
    let value : UInt32 := (input[i]'h).toUInt32
    -- High surrogate range 0xD800–0xDBFF (punycode.js:107).
    if value >= 0xD800 && value <= 0xDBFF then
      if h' : i + 1 < input.size then
        let extra : UInt32 := (input[i + 1]'h').toUInt32
        -- Low surrogate: top six bits must be 0b110111 (0xDC00–0xDFFF) (punycode.js:110).
        if (extra &&& 0xFC00) == 0xDC00 then
          -- Combine pair into astral code point (punycode.js:111).
          let cp := ((value &&& 0x3FF) <<< 10) + (extra &&& 0x3FF) + 0x10000
          ucs2decodeAux input (i + 2) (acc.push cp)
        else
          -- Unmatched high surrogate: push value, reprocess extra next (punycode.js:115-116).
          ucs2decodeAux input (i + 1) (acc.push value)
      else
        -- High surrogate at end of input: no pair possible, push as lone value.
        ucs2decodeAux input (i + 1) (acc.push value)
    else
      -- Non-surrogate or lone low surrogate: push directly (punycode.js:118-120).
      ucs2decodeAux input (i + 1) (acc.push value)
  else
    acc

/-- Convert a sequence of UTF-16 code units into an array of Unicode code points.
    Surrogate pairs are combined; lone surrogates pass through as raw values.
    Ports `punycode.js:101-123` (`ucs2decode`). -/
def ucs2decode (input : Array UInt16) : Array UInt32 :=
  ucs2decodeAux input 0 #[]

/-- Convert an array of Unicode code points into a sequence of UTF-16 code units.
    Code points ≥ U+10000 are split into surrogate pairs; BMP values (< U+10000)
    including lone surrogates are emitted as single code units.
    Ports `punycode.js:133` (`codePoints => String.fromCodePoint(...codePoints)`). -/
def ucs2encode (codePoints : Array UInt32) : Array UInt16 :=
  codePoints.foldl (fun acc (cp : UInt32) =>
    if cp < 0x10000 then
      -- BMP code point (or lone surrogate): emit as single UInt16.
      acc.push cp.toUInt16
    else
      -- Supplementary plane: split into surrogate pair (inverse of punycode.js:111).
      let cp' : UInt32 := cp - 0x10000
      let high : UInt16 := ((0xD800 : UInt32) + (cp' >>> 10)).toUInt16
      let low  : UInt16 := ((0xDC00 : UInt32) + (cp' &&& 0x3FF)).toUInt16
      (acc.push high).push low
  ) #[]

/-- Convert a Lean `String` (valid Unicode scalars) to a sequence of UTF-16 code units.
    Code points ≥ U+10000 are split into surrogate pairs. -/
def stringToUCS2 (s : String) : Array UInt16 :=
  s.foldl (fun acc c =>
    let cp : UInt32 := c.val
    if cp < 0x10000 then
      acc.push cp.toUInt16
    else
      let cp' : UInt32 := cp - 0x10000
      let high : UInt16 := ((0xD800 : UInt32) + (cp' >>> 10)).toUInt16
      let low  : UInt16 := ((0xDC00 : UInt32) + (cp' &&& 0x3FF)).toUInt16
      (acc.push high).push low
  ) #[]

/-- Convert a sequence of UTF-16 code units to a Lean `String`.
    Decodes surrogate pairs; lone surrogates (U+D800–U+DFFF) are dropped since
    Lean `Char` must be a valid Unicode scalar value (documented limitation, D2). -/
def ucs2ToString (units : Array UInt16) : String :=
  (ucs2decode units).foldl (fun s (cp : UInt32) =>
    let n := cp.toNat
    -- Valid Unicode scalar: not a surrogate and within Unicode range.
    if (n < 0xD800) || (0xE000 ≤ n && n < 0x110000) then
      s.push (Char.ofNat n)
    else
      s  -- lone surrogate: skip (cannot be a Lean Char)
  ) ""

end Punycode
