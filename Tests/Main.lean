-- Punycode test suite — ports tests/tests.js behaviour
import Punycode.Constants
import Punycode.Helpers
import Punycode.UCS2
import Punycode.Bootstring
import Punycode.Decode
import Punycode.Encode
import Punycode.Domain

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

  -- Encode tests: mirror punycode.encode(object.decoded) === object.encoded
  -- over all testData.strings fixtures (tests/tests.js:312-321).
  IO.println "=== Punycode.Encode — encode ==="
  runTests #[
    -- tests/tests.js strings[0]: a single basic code point
    ("encode \"Bach\" = \"Bach-\"",
     encode "Bach" == .ok "Bach-"),
    -- tests/tests.js strings[1]: a single non-ASCII character (U+00FC)
    ("encode \"ü\" = \"tda\"",
     encode "ü" == .ok "tda"),
    -- tests/tests.js strings[2]: multiple non-ASCII characters
    ("encode \"üëäö♥\" = \"4can8av2009b\"",
     encode "üëäö♥" == .ok "4can8av2009b"),
    -- tests/tests.js strings[3]: mix of ASCII and non-ASCII
    ("encode \"bücher\" = \"bcher-kva\"",
     encode "bücher" == .ok "bcher-kva"),
    -- tests/tests.js strings[4]: long string with both ASCII and non-ASCII
    ("encode long German string",
     encode "Willst du die Blüthe des frühen, die Früchte des späteren Jahres"
       == .ok "Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal"),
    -- tests/tests.js strings[5]: Arabic (Egyptian) — RFC 3492 §7.1
    ("encode Arabic (Egyptian)",
     encode "ليهمابتكلموشعربي؟"
       == .ok "egbpdaj6bu4bxfgehfvwxn"),
    -- tests/tests.js strings[6]: Chinese (simplified) — RFC 3492 §7.1
    ("encode Chinese (simplified)",
     encode "他们为什么不说中文" == .ok "ihqwcrb4cv8a8dqg056pqjye"),
    -- tests/tests.js strings[7]: Chinese (traditional) — RFC 3492 §7.1
    ("encode Chinese (traditional)",
     encode "他們爲什麽不說中文" == .ok "ihqwctvzc91f659drss3x8bo0yb"),
    -- tests/tests.js strings[8]: Czech — RFC 3492 §7.1
    ("encode Czech",
     encode "Pročprostěnemluvíčesky" == .ok "Proprostnemluvesky-uyb24dma41a"),
    -- tests/tests.js strings[9]: Hebrew — RFC 3492 §7.1
    ("encode Hebrew",
     encode "למההםפשוטלאמדבריםעברית"
       == .ok "4dbcagdahymbxekheh6e0a7fei0b"),
    -- tests/tests.js strings[10]: Hindi (Devanagari) — RFC 3492 §7.1
    ("encode Hindi (Devanagari)",
     encode "यहलोगहिन्दीक्योंनहींबोलसकतेहैं"
       == .ok "i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd"),
    -- tests/tests.js strings[11]: Japanese (kanji and hiragana) — RFC 3492 §7.1
    ("encode Japanese (kanji and hiragana)",
     encode "なぜみんな日本語を話してくれないのか"
       == .ok "n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa"),
    -- tests/tests.js strings[12]: Korean (Hangul syllables) — RFC 3492 §7.1
    ("encode Korean (Hangul syllables)",
     encode "세계의모든사람들이한국어를이해한다면얼마나좋을까"
       == .ok "989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c"),
    -- tests/tests.js strings[13]: Russian (Cyrillic) — RFC 3492 §7.1
    ("encode Russian (Cyrillic)",
     encode "почемужеонинеговорятпорусски"
       == .ok "b1abfaaepdrnnbgefbadotcwatmq2g4l"),
    -- tests/tests.js strings[14]: Spanish — RFC 3492 §7.1
    ("encode Spanish",
     encode "PorquénopuedensimplementehablarenEspañol"
       == .ok "PorqunopuedensimplementehablarenEspaol-fmd56a"),
    -- tests/tests.js strings[15]: Vietnamese — RFC 3492 §7.1
    ("encode Vietnamese",
     encode "TạisaohọkhôngthểchỉnóitiếngViệt"
       == .ok "TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g"),
    -- tests/tests.js strings[16]
    ("encode \"3年B組金八先生\"",
     encode "3年B組金八先生" == .ok "3B-ww4c5e180e575a65lsy2b"),
    -- tests/tests.js strings[17]
    ("encode \"安室奈美恵-with-SUPER-MONKEYS\"",
     encode "安室奈美恵-with-SUPER-MONKEYS" == .ok "-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n"),
    -- tests/tests.js strings[18]
    ("encode \"Hello-Another-Way-それぞれの場所\"",
     encode "Hello-Another-Way-それぞれの場所"
       == .ok "Hello-Another-Way--fc4qua05auwb3674vfr0b"),
    -- tests/tests.js strings[19]
    ("encode \"ひとつ屋根の下2\"",
     encode "ひとつ屋根の下2" == .ok "2-u9tlzr9756bt3uc0v"),
    -- tests/tests.js strings[20]
    ("encode \"MajiでKoiする5秒前\"",
     encode "MajiでKoiする5秒前" == .ok "MajiKoi5-783gue6qz075azm5e"),
    -- tests/tests.js strings[21]
    ("encode \"パフィーdeルンバ\"",
     encode "パフィーdeルンバ" == .ok "de-jg4avhby1noc0d"),
    -- tests/tests.js strings[22]
    ("encode \"そのスピードで\"",
     encode "そのスピードで" == .ok "d9juau41awczczp"),
    -- tests/tests.js strings[23]: ASCII string that breaks host-name label rules
    ("encode \"-> $1.00 <-\" = \"-> $1.00 <--\"",
     encode "-> $1.00 <-" == .ok "-> $1.00 <--"),
  ]

  IO.println "=== Punycode.Domain — version ==="
  runTests #[
    ("version = \"2.3.1\"", version == "2.3.1"),
  ]

  -- toUnicode tests: mirrors tests/tests.js:323-344.
  -- Fixtures from testData.domains (tests/tests.js:176-220).
  IO.println "=== Punycode.Domain — toUnicode (domain fixtures) ==="
  runTests #[
    ("toUnicode xn--maana-pta.com = mañana.com",
     toUnicode "xn--maana-pta.com" == .ok "mañana.com"),
    ("toUnicode example.com. = example.com. (trailing dot pass-through)",
     toUnicode "example.com." == .ok "example.com."),
    ("toUnicode xn--bcher-kva.com = bücher.com",
     toUnicode "xn--bcher-kva.com" == .ok "bücher.com"),
    ("toUnicode xn--caf-dma.com = café.com",
     toUnicode "xn--caf-dma.com" == .ok "café.com"),
    ("toUnicode xn----dqo34k.com = ☃-⌘.com",
     toUnicode "xn----dqo34k.com" == .ok "☃-⌘.com"),
    ("toUnicode xn----dqo34kn65z.com = U+D400+☃-⌘.com",
     toUnicode "xn----dqo34kn65z.com" == .ok "퐀☃-⌘.com"),
    ("toUnicode xn--ls8h.la = 💩.la (emoji)",
     toUnicode "xn--ls8h.la" == .ok "💩.la"),
    ("toUnicode \\x00\\x01\\x02foo.bar pass-through (no xn-- labels)",
     toUnicode "\x00\x01\x02foo.bar" == .ok "\x00\x01\x02foo.bar"),
    ("toUnicode email: Cyrillic@xn--... = Cyrillic@Cyrillic...",
     toUnicode "джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq"
       == .ok "джумла@джpумлатест.bрфa"),
    ("toUnicode foo\\x7F.example pass-through (DEL char < 0x80)",
     toUnicode "foo\x7F.example" == .ok "foo\x7F.example"),
  ]

  -- Passthrough: strings that don't start with xn-- are returned unchanged.
  -- Mirrors tests/tests.js:332-344 (both encoded and decoded forms pass through).
  IO.println "=== Punycode.Domain — toUnicode (passthrough for non-xn-- strings) ==="
  runTests #[
    ("toUnicode passes through encoded ASCII-only string",
     toUnicode "Bach-" == .ok "Bach-"),
    ("toUnicode passes through decoded Unicode string without xn-- prefix",
     toUnicode "Bach" == .ok "Bach"),
    ("toUnicode passes through non-ASCII string that does not start with xn--",
     toUnicode "ü" == .ok "ü"),
    ("toUnicode passes through Arabic encoded (no xn-- prefix)",
     toUnicode "egbpdaj6bu4bxfgehfvwxn" == .ok "egbpdaj6bu4bxfgehfvwxn"),
  ]

  -- toASCII tests: mirrors tests/tests.js:346-371.
  -- Fixtures from testData.domains (tests/tests.js:176-220).
  IO.println "=== Punycode.Domain — toASCII (domain fixtures) ==="
  runTests #[
    ("toASCII mañana.com = xn--maana-pta.com",
     toASCII "mañana.com" == .ok "xn--maana-pta.com"),
    ("toASCII example.com. = example.com. (trailing dot pass-through)",
     toASCII "example.com." == .ok "example.com."),
    ("toASCII bücher.com = xn--bcher-kva.com",
     toASCII "bücher.com" == .ok "xn--bcher-kva.com"),
    ("toASCII café.com = xn--caf-dma.com",
     toASCII "café.com" == .ok "xn--caf-dma.com"),
    ("toASCII ☃-⌘.com = xn----dqo34k.com",
     toASCII "☃-⌘.com" == .ok "xn----dqo34k.com"),
    ("toASCII U+D400+☃-⌘.com = xn----dqo34kn65z.com",
     toASCII "퐀☃-⌘.com" == .ok "xn----dqo34kn65z.com"),
    ("toASCII 💩.la = xn--ls8h.la (emoji)",
     toASCII "💩.la" == .ok "xn--ls8h.la"),
    ("toASCII \\x00\\x01\\x02foo.bar pass-through (all chars < 0x80)",
     toASCII "\x00\x01\x02foo.bar" == .ok "\x00\x01\x02foo.bar"),
    ("toASCII email: Cyrillic@Cyrillic... = Cyrillic@xn--...",
     toASCII "джумла@джpумлатест.bрфa"
       == .ok "джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq"),
    ("toASCII foo\\x7F.example pass-through (DEL char is < 0x80)",
     toASCII "foo\x7F.example" == .ok "foo\x7F.example"),
  ]

  -- Passthrough: strings already in ASCII are returned unchanged.
  -- Mirrors tests/tests.js:355-362 (encoded strings have no chars >= 0x80).
  IO.println "=== Punycode.Domain — toASCII (passthrough for ASCII strings) ==="
  runTests #[
    ("toASCII passes through ASCII encoded string Bach-",
     toASCII "Bach-" == .ok "Bach-"),
    ("toASCII passes through ASCII encoded string tda",
     toASCII "tda" == .ok "tda"),
    ("toASCII passes through ASCII encoded string egbpdaj6bu4bxfgehfvwxn",
     toASCII "egbpdaj6bu4bxfgehfvwxn" == .ok "egbpdaj6bu4bxfgehfvwxn"),
  ]

  -- Separator normalisation: U+002E/U+3002/U+FF0E/U+FF61 all become '.'.
  -- Mirrors tests/tests.js:363-370 (testData.separators).
  IO.println "=== Punycode.Domain — toASCII (separator normalisation) ==="
  runTests #[
    ("toASCII U+002E separator (standard period)",
     toASCII "mañana\x2Ecom" == .ok "xn--maana-pta.com"),
    ("toASCII U+3002 separator (ideographic full stop)",
     toASCII "mañana。com" == .ok "xn--maana-pta.com"),
    ("toASCII U+FF0E separator (fullwidth full stop)",
     toASCII "mañana．com" == .ok "xn--maana-pta.com"),
    ("toASCII U+FF61 separator (halfwidth ideographic full stop)",
     toASCII "mañana｡com" == .ok "xn--maana-pta.com"),
  ]
