-- Domain-level functions: toUnicode and toASCII.
-- Ports punycode.js:389-441; behaviour spec: specs/impl-domain-api.md.

import Punycode.Helpers
import Punycode.Decode
import Punycode.Encode

namespace Punycode

/-- Library version string (punycode.js:425). -/
def version : String := "2.3.1"

/-- Convert a Punycoded domain name or email address to Unicode.
    Labels prefixed with `xn--` are decoded; all others pass through unchanged.
    Idempotent: re-running on already-Unicode input is a no-op.
    Ports `punycode.js:389-395` (`toUnicode`). -/
def toUnicode (input : String) : Except PunyError String :=
  mapDomain input fun label =>
    if label.startsWith "xn--" then
      -- String.drop returns a Substring; convert back to String before lowercasing.
      decode ((label.drop 4).toString.toLower)
    else
      .ok label

private def hasNonASCII (s : String) : Bool :=
  s.toList.any fun (c : Char) => c.val >= (0x80 : UInt32)

/-- Convert a Unicode domain name or email address to Punycode.
    Labels containing any character ≥ U+0080 are encoded with the `xn--` prefix;
    ASCII-only labels pass through unchanged.
    Idempotent: re-running on already-ASCII input is a no-op.
    Ports `punycode.js:408-414` (`toASCII`). -/
def toASCII (input : String) : Except PunyError String :=
  mapDomain input fun label => do
    if hasNonASCII label then
      return "xn--" ++ (← encode label)
    else
      return label

end Punycode
