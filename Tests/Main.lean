-- Punycode test suite — ports tests/tests.js behaviour
import Punycode.Constants
import Punycode.Helpers
import Punycode.UCS2
import Punycode.Bootstring
import Punycode.Decode

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

  -- UCS-2 fixtures: (description, UTF-16 code units, Unicode code points).
  -- Mirrors testData.ucs2 from tests/tests.js:137-175.
  -- Each row tests both ucs2decode (encoded→decoded) and ucs2encode (decoded→encoded).
  let ucs2Fixtures : Array (String × Array UInt16 × Array UInt32) := #[
    -- Four surrogate pairs back-to-back (tests/tests.js:140-144).
    ("Consecutive astral symbols",
     #[0xD83C, 0xDF55, 0xD835, 0xDC00, 0xD834, 0xDF06, 0xD834, 0xDF56],
     #[127829, 119808, 119558, 119638]),
    -- Lone high surrogate followed by ASCII (tests/tests.js:145-149).
    ("U+D800 (high surrogate) followed by non-surrogates",
     #[0xD800, 97, 98],
     #[55296, 97, 98]),
    -- Lone low surrogate followed by ASCII (tests/tests.js:150-154).
    ("U+DC00 (low surrogate) followed by non-surrogates",
     #[0xDC00, 97, 98],
     #[56320, 97, 98]),
    -- Two consecutive lone high surrogates (tests/tests.js:155-159).
    ("High surrogate followed by another high surrogate",
     #[0xD800, 0xD800],
     #[0xD800, 0xD800]),
    -- Lone high, valid pair, lone high (tests/tests.js:160-164).
    ("Unmatched high surrogate, followed by surrogate pair, followed by unmatched high",
     #[0xD800, 0xD834, 0xDF06, 0xD800],
     #[0xD800, 0x1D306, 0xD800]),
    -- Two consecutive lone low surrogates (tests/tests.js:165-169).
    ("Low surrogate followed by another low surrogate",
     #[0xDC00, 0xDC00],
     #[0xDC00, 0xDC00]),
    -- Lone low, valid pair, lone low (tests/tests.js:170-174).
    ("Unmatched low surrogate, followed by surrogate pair, followed by unmatched low",
     #[0xDC00, 0xD834, 0xDF06, 0xDC00],
     #[0xDC00, 0x1D306, 0xDC00]),
  ]

  IO.println "=== Punycode.UCS2 — ucs2decode ==="
  runTests (ucs2Fixtures.map fun fixture =>
    let (desc, encoded, decoded) := fixture
    (s!"ucs2decode: {desc}", ucs2decode encoded == decoded))

  IO.println "=== Punycode.UCS2 — ucs2encode ==="
  runTests (ucs2Fixtures.map fun fixture =>
    let (desc, encoded, decoded) := fixture
    (s!"ucs2encode: {desc}", ucs2encode decoded == encoded))

  IO.println "=== Punycode.UCS2 — ucs2encode extra ==="
  runTests #[
    -- ASCII round-trip: no surrogate involved (mirrors tests/tests.js:282-287).
    ("ucs2encode [0x61, 0x62, 0x63] = [97, 98, 99]",
     ucs2encode #[(0x61 : UInt32), 0x62, 0x63] == #[(0x61 : UInt16), 0x62, 0x63]),
  ]

  IO.println "=== Punycode.UCS2 — bridges ==="
  runTests #[
    ("stringToUCS2 \"abc\" gives BMP code units",
     stringToUCS2 "abc" == #[(97 : UInt16), 98, 99]),
    ("stringToUCS2 with astral U+1D306",
     stringToUCS2 "𝌆" == #[(0xD834 : UInt16), 0xDF06]),
    ("ucs2ToString BMP-only round-trip",
     ucs2ToString #[(97 : UInt16), 98, 99] == "abc"),
    ("ucs2ToString surrogate pair → Char",
     ucs2ToString #[(0xD834 : UInt16), 0xDF06] == "𝌆"),
    -- Lone surrogate cannot be a Lean Char; it is dropped.
    ("ucs2ToString lone high surrogate dropped",
     ucs2ToString #[(0xD800 : UInt16), 97] == "a"),
    ("ucs2ToString lone low surrogate dropped",
     ucs2ToString #[(0xDC00 : UInt16), 97] == "a"),
  ]

  IO.println "=== Punycode.Bootstring — basicToDigit ==="
  runTests #[
    -- Digit characters '0'..'9' → 26..35 (punycode.js:145-147).
    ("basicToDigit '0' (0x30) = 26",  basicToDigit 0x30 == 26),
    ("basicToDigit '9' (0x39) = 35",  basicToDigit 0x39 == 35),
    ("basicToDigit '5' (0x35) = 31",  basicToDigit 0x35 == 31),
    -- Uppercase letters 'A'..'Z' → 0..25 (punycode.js:148-150).
    ("basicToDigit 'A' (0x41) = 0",   basicToDigit 0x41 == 0),
    ("basicToDigit 'Z' (0x5A) = 25",  basicToDigit 0x5A == 25),
    ("basicToDigit 'M' (0x4D) = 12",  basicToDigit 0x4D == 12),
    -- Lowercase letters 'a'..'z' → 0..25 (punycode.js:151-153).
    ("basicToDigit 'a' (0x61) = 0",   basicToDigit 0x61 == 0),
    ("basicToDigit 'z' (0x7A) = 25",  basicToDigit 0x7A == 25),
    ("basicToDigit 'm' (0x6D) = 12",  basicToDigit 0x6D == 12),
    -- Case-insensitive: 'A' and 'a' give the same digit.
    ("basicToDigit 'A' == basicToDigit 'a'",
      basicToDigit 0x41 == basicToDigit 0x61),
    -- Non-digit characters return `base` (36) as sentinel (punycode.js:154).
    ("basicToDigit '-' (0x2D) = base", basicToDigit 0x2D == base),
    ("basicToDigit '!' (0x21) = base", basicToDigit 0x21 == base),
    ("basicToDigit 0x00 = base",       basicToDigit 0x00 == base),
    ("basicToDigit 0x7F = base",       basicToDigit 0x7F == base),
  ]

  IO.println "=== Punycode.Bootstring — digitToBasic ==="
  runTests #[
    -- Digits 0..25 with flag=0 → lowercase letters 'a'..'z' (ascii 97..122).
    ("digitToBasic 0 0 = 'a' (97)",    digitToBasic 0  0 == 97),
    ("digitToBasic 25 0 = 'z' (122)",  digitToBasic 25 0 == 122),
    ("digitToBasic 12 0 = 'm' (109)",  digitToBasic 12 0 == 109),
    -- Digits 26..35 with any flag → ascii digits '0'..'9' (ascii 48..57).
    ("digitToBasic 26 0 = '0' (48)",   digitToBasic 26 0 == 48),
    ("digitToBasic 35 0 = '9' (57)",   digitToBasic 35 0 == 57),
    ("digitToBasic 28 0 = '2' (50)",   digitToBasic 28 0 == 50),
    -- Digits 0..25 with flag=1 → uppercase letters 'A'..'Z' (ascii 65..90).
    ("digitToBasic 0 1 = 'A' (65)",    digitToBasic 0  1 == 65),
    ("digitToBasic 25 1 = 'Z' (90)",   digitToBasic 25 1 == 90),
    ("digitToBasic 12 1 = 'M' (77)",   digitToBasic 12 1 == 77),
    -- Round-trip: digitToBasic is inverse of basicToDigit for lowercase.
    ("round-trip digit 0",  basicToDigit (digitToBasic 0  0) == 0),
    ("round-trip digit 25", basicToDigit (digitToBasic 25 0) == 25),
    ("round-trip digit 26", basicToDigit (digitToBasic 26 0) == 26),
    ("round-trip digit 35", basicToDigit (digitToBasic 35 0) == 35),
  ]

  IO.println "=== Punycode.Bootstring — adapt ==="
  runTests #[
    -- adapt with delta=0: result is 0 regardless of numPoints or firstTime.
    ("adapt 0 1 false = 0",  adapt 0 1 false == 0),
    ("adapt 0 10 true = 0",  adapt 0 10 true == 0),
    -- firstTime=true: heavy damping by damp (700).
    -- adapt 700 1 true: delta1=1, delta2=2, loop: 2≤455 → (2,0),
    --   result = 36*2/(2+38) = 72/40 = 1.
    ("adapt 700 1 true = 1",   adapt 700 1 true == 1),
    -- adapt 700 2 true: delta1=1, delta2=1+0=1, loop: 1≤455 → (1,0),
    --   result = 36*1/(1+38) = 36/39 = 0.
    ("adapt 700 2 true = 0",   adapt 700 2 true == 0),
    -- firstTime=false: halve.
    -- adapt 1400 2 false: delta1=700, delta2=700+350=1050,
    --   loop: 1050>455 → delta=30, k=36; 30≤455 → (30,36),
    --   result = 36 + 36*30/(30+38) = 36 + 1080/68 = 36 + 15 = 51.
    ("adapt 1400 2 false = 51", adapt 1400 2 false == 51),
    -- adapt 10000 100 false: delta1=5000, delta2=5050,
    --   loop: 5050>455 → delta=144, k=36; 144≤455 → (144,36),
    --   result = 36 + 36*144/(144+38) = 36 + 5184/182 = 36 + 28 = 64.
    ("adapt 10000 100 false = 64", adapt 10000 100 false == 64),
    -- adapt requires multiple loop iterations.
    -- adapt 500000 50 false: delta1=250000, delta2=255000,
    --   loop: 255000>455 → 7285>455 → 208≤455; k=72,
    --   result = 72 + 36*208/(208+38) = 72 + 7488/246 = 72 + 30 = 102.
    ("adapt 500000 50 false = 102", adapt 500000 50 false == 102),
  ]

  -- Decode tests: RFC 3492 §7.1 samples and edge cases.
  -- Mirrors tests/tests.js:290-310 (fixture-driven + special cases).
  IO.println "=== Punycode.Decode — decode ==="
  runTests #[
    -- Basic code points only (trailing delimiter stripped): tests/tests.js:8-12.
    ("decode \"Bach-\" = \"Bach\"",
     decode "Bach-" == .ok "Bach"),
    -- Single non-ASCII: ü (U+00FC): tests/tests.js:13-17.
    ("decode \"tda\" = \"ü\"",
     decode "tda" == .ok "ü"),
    -- Mix of ASCII and non-ASCII: bücher: tests/tests.js:23-27.
    ("decode \"bcher-kva\" = \"bücher\"",
     decode "bcher-kva" == .ok "bücher"),
    -- RFC 3492 §7.1 Arabic (Egyptian): tests/tests.js:35-38.
    ("decode \"egbpdaj6bu4bxfgehfvwxn\" = Arabic phrase",
     decode "egbpdaj6bu4bxfgehfvwxn" == .ok "ليهمابتكلموشعربي؟"),
    -- RFC 3492 §7.1 Chinese simplified: tests/tests.js:39-43.
    ("decode \"ihqwcrb4cv8a8dqg056pqjye\" = Chinese simplified",
     decode "ihqwcrb4cv8a8dqg056pqjye" == .ok "他们为什么不说中文"),
    -- RFC 3492 §7.1 Russian (lowercase form): tests/tests.js:83-87.
    ("decode \"b1abfaaepdrnnbgefbadotcwatmq2g4l\" = Russian",
     decode "b1abfaaepdrnnbgefbadotcwatmq2g4l" == .ok "почемужеонинеговорятпорусски"),
    -- Emoji: 💩 (U+1F4A9): tests/tests.js (domain fixture, bare label).
    ("decode \"ls8h\" = 💩",
     decode "ls8h" == .ok "💩"),
    -- Uppercase Z is treated identically to lowercase z (case-insensitive basicToDigit):
    -- tests/tests.js:299-301.
    ("decode \"ZZZ\" = 箥",
     decode "ZZZ" == .ok "箥"),
    -- not-basic error: '\x81-' has non-basic char in prefix region: tests/tests.js:255-261.
    ("decode \"\\x81-\" = .error .notBasic",
     decode "\x81-" == .error .notBasic),
    -- invalid-input: '\x81' has no delimiter; basicToDigit(0x81) = base → invalidInput.
    -- (test label in JS says "overflow" but actual error is invalidInput — see impl-decode.md:69-76)
    ("decode \"\\x81\" = .error .invalidInput",
     decode "\x81" == .error .invalidInput),
    -- invalid-input: '=' is not a valid base-36 digit: tests/tests.js:302-309.
    ("decode \"ls8h=\" = .error .invalidInput",
     decode "ls8h=" == .error .invalidInput),
    -- Empty input decodes to empty string.
    ("decode \"\" = \"\"",
     decode "" == .ok ""),
    -- ASCII string that breaks host-name rules: tests/tests.js:131-135.
    ("decode \"-> $1.00 <--\" = \"-> $1.00 <-\"",
     decode "-> $1.00 <--" == .ok "-> $1.00 <-"),
  ]
