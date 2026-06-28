-- Punycode test suite — ports tests/tests.js behaviour
import Punycode.Constants
import Punycode.Helpers

open Punycode

-- BEq for Except is not in Lean core; define it for test comparisons.
instance {ε α : Type} [BEq ε] [BEq α] : BEq (Except ε α) where
  beq
    | .ok a,    .ok b    => a == b
    | .error e, .error f => e == f
    | _,        _        => false

-- Minimal test runner: count failures and exit non-zero if any.
private def runTests (tests : Array (String × Bool)) : IO Unit := do
  let mut failed := 0
  for (name, ok) in tests do
    if ok then
      IO.println s!"  ✓ {name}"
    else
      IO.println s!"  ✗ {name}"
      failed := failed + 1
  if failed > 0 then
    IO.println s!"FAILED: {failed} test(s)"
    IO.Process.exit 1
  else
    IO.println s!"All {tests.size} test(s) passed."

def main : IO Unit := do
  IO.println "=== Punycode.Constants ==="
  runTests #[
    ("maxInt = 2147483647",           maxInt == 2147483647),
    ("base = 36",                     base == 36),
    ("tMin = 1",                      tMin == 1),
    ("tMax = 26",                     tMax == 26),
    ("skew = 38",                     skew == 38),
    ("damp = 700",                    damp == 700),
    ("initialBias = 72",              initialBias == 72),
    ("initialN = 128",                initialN == 128),
    ("delimiter = '-'",               delimiter == '-'),
    ("baseMinusTMin = 35",            baseMinusTMin == 35),
    ("separators has 4 entries",      separatorCodePoints.size == 4),
    ("separator[0] = 0x002E",         separatorCodePoints[0]! == 0x002E),
    ("separator[1] = 0x3002",         separatorCodePoints[1]! == 0x3002),
    ("separator[2] = 0xFF0E",         separatorCodePoints[2]! == 0xFF0E),
    ("separator[3] = 0xFF61",         separatorCodePoints[3]! == 0xFF61),
    ("overflow message exact",
      PunyError.message .overflow ==
        "Overflow: input needs wider integers to process"),
    ("notBasic message exact",
      PunyError.message .notBasic ==
        "Illegal input >= 0x80 (not a basic code point)"),
    ("invalidInput message exact",
      PunyError.message .invalidInput == "Invalid input"),
    ("PunyError BEq: overflow == overflow",
      (PunyError.overflow == PunyError.overflow) == true),
    ("PunyError BEq: overflow != notBasic",
      (PunyError.overflow == PunyError.notBasic) == false),
  ]

  IO.println "=== Punycode.Helpers — throwPuny ==="
  runTests #[
    ("throwPuny overflow is .error",
      (throwPuny (α := String) .overflow) == .error .overflow),
    ("throwPuny notBasic is .error",
      (throwPuny (α := Nat) .notBasic) == .error .notBasic),
    ("throwPuny invalidInput is .error",
      (throwPuny (α := Unit) .invalidInput) == .error .invalidInput),
  ]

  IO.println "=== Punycode.Helpers — mapDomain ==="
  runTests #[
    -- Identity callback: domain is passed through unchanged.
    ("mapDomain identity simple domain",
      mapDomain "example.com" .ok == .ok "example.com"),

    -- Email: local-part before '@' is preserved verbatim (punycode.js:76-79).
    ("mapDomain preserves email local-part",
      mapDomain "user@example.com" .ok == .ok "user@example.com"),

    -- Separator normalisation: U+3002 IDEOGRAPHIC FULL STOP → '.'
    ("mapDomain normalises U+3002 separator",
      mapDomain "example。com" .ok == .ok "example.com"),

    -- Separator normalisation: U+FF0E FULLWIDTH FULL STOP → '.'
    ("mapDomain normalises U+FF0E separator",
      mapDomain "example．com" .ok == .ok "example.com"),

    -- Separator normalisation: U+FF61 HALFWIDTH IDEOGRAPHIC FULL STOP → '.'
    ("mapDomain normalises U+FF61 separator",
      mapDomain "example｡com" .ok == .ok "example.com"),

    -- Error from callback must propagate (D1: first error short-circuits).
    ("mapDomain propagates callback error",
      mapDomain "a.b.c" (fun _ => .error .overflow) == .error .overflow),

    -- Callback transforms each label independently.
    ("mapDomain transforms labels via callback",
      mapDomain "foo.bar" (fun s => .ok (s ++ "!")) == .ok "foo!.bar!"),

    -- Email + separator normalisation together.
    ("mapDomain email with U+3002 separator",
      mapDomain "user@foo。bar" .ok == .ok "user@foo.bar"),

    -- Multiple '@' signs: JS takes parts[0] as local, parts[1] as domain.
    -- "a@b@c.com".split('@') → ["a","b","c.com"]; parts[1]="b" is the domain.
    ("mapDomain multiple @: parts[1] is domain",
      mapDomain "a@b@c.com" .ok == .ok "a@b"),
  ]
