-- Generic private helpers used by the codec.  Ports punycode.js:41-86.
-- Behaviour spec: specs/impl-helpers.md.

import Punycode.Constants

namespace Punycode

-- Raise a PunyError in an Except-valued computation (punycode.js:41-43, D1).
-- Callers write: `return ← throwPuny .overflow` or inline `.error e`.
def throwPuny {α : Type} (e : PunyError) : Except PunyError α := .error e

-- Replace every Unicode label-separator with ASCII '.'.
-- Ports the regexSeparators replace at punycode.js:82 without a regex engine
-- (D4: replace the four code points U+002E/U+3002/U+FF0E/U+FF61 directly).
private def normalizeSeparators (s : String) : String :=
  String.ofList (s.toList.map fun c =>
    if separatorCodePoints.contains c.val then '.' else c)

/-- Apply `callback` to each DNS label in `domain`, returning the reassembled
    string.  An email local-part before `@` is passed through unchanged.
    Ports punycode.js:72-86; callback is Except-valued per D1. -/
def mapDomain (domain : String) (callback : String → Except PunyError String)
    : Except PunyError String := do
  -- Email split: preserve the local-part intact (punycode.js:73-80).
  -- JS: parts = domain.split('@'); parts[0]+'@' as emailPfx, parts[1] as domain.
  let parts := domain.splitOn "@"
  let emailParts := match parts with
    | []           => ("", domain)            -- unreachable for non-empty input
    | [single]     => ("", single)            -- no @ present
    | head :: rest => (head ++ "@", rest.head!) -- email: local@domain
  let emailPfx  := emailParts.1
  let domainStr := emailParts.2
  -- Normalise separators, then split on '.' (punycode.js:82-83).
  let labels := (normalizeSeparators domainStr).splitOn "."
  -- Map callback over labels, propagating the first error (punycode.js:84, D1).
  let encodedLabels ← labels.mapM callback
  return emailPfx ++ String.intercalate "." encodedLabels

end Punycode
