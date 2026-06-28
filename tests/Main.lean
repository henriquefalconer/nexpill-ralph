-- Punycode test suite — ports tests/tests.js behaviour
import Punycode.Constants

open Punycode

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
