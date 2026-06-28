-- Constants and error type for the Punycode Bootstring codec.
-- Ports punycode.js:1-31; behaviour spec: specs/impl-constants.md.

namespace Punycode

-- Overflow sentinel: highest positive signed 32-bit integer (punycode.js:4).
-- Every overflow guard compares against this value.
def maxInt : Nat := 2147483647

-- RFC 3492 Bootstring parameters (punycode.js:6-14).
def base        : Nat := 36
def tMin        : Nat := 1
def tMax        : Nat := 26
def skew        : Nat := 38
def damp        : Nat := 700
def initialBias : Nat := 72
def initialN    : Nat := 128  -- 0x80: first non-basic code point
def delimiter   : Char := '-' -- '\x2D': basic/extended separator

-- Precomputed convenience (punycode.js:29).
def baseMinusTMin : Nat := base - tMin  -- = 35

-- RFC 3490 label separators (punycode.js:19, regexSeparators).
-- Used by mapDomain to normalise all four forms to ASCII '.'.
def separatorCodePoints : Array UInt32 :=
  #[0x002E, 0x3002, 0xFF0E, 0xFF61]

/-- Errors that the codec can raise (mirrors punycode.js:21-26 errors table). -/
inductive PunyError
  | overflow    -- "Overflow: input needs wider integers to process"
  | notBasic    -- "Illegal input >= 0x80 (not a basic code point)"
  | invalidInput -- "Invalid input"
  deriving Repr, BEq, DecidableEq

/-- Exact error message strings from punycode.js:22-25. -/
def PunyError.message : PunyError → String
  | .overflow     => "Overflow: input needs wider integers to process"
  | .notBasic     => "Illegal input >= 0x80 (not a basic code point)"
  | .invalidInput => "Invalid input"

end Punycode
