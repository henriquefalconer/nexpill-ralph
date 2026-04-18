# Test Fixtures

The `testData` object in `tests/tests.js` defines four arrays of shared test-data fixtures. Each fixture is a JSON object with properties `decoded`, `encoded`, and optional `description`. These fixtures are consumed by every describe block in the test suite and serve as the ground truth for the behavior documented in sibling specification files.

## Strings Fixtures

**Purpose and shape:** The `strings` array (tests/tests.js:7-136) contains decoded Unicode strings and their corresponding Punycode-encoded representations. These fixtures test the core codec functions `punycode.encode()` and `punycode.decode()`, which convert between Unicode strings and ASCII-only Punycode labels. The `decoded` value is a Unicode string; the `encoded` value is its Punycode representation (with ASCII-only characters).

**Notable entries:**

- Single ASCII-only label "Bach" encodes to "Bach-" (basic code points require trailing delimiter if present; tests/tests.js:8-12)
- Single non-ASCII character U+00FC encodes to "tda" (tests/tests.js:13-17)
- Multiple non-ASCII characters "\xFC\xEB\xE4\xF6\u2665" encode to "4can8av2009b" (tests/tests.js:18-22)
- Mixed ASCII and non-ASCII "b\xFCcher" encodes to "bcher-kva" (tests/tests.js:23-27)
- Long mixed string "Willst du die Blüthe..." encodes to "Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal" (tests/tests.js:28-32)
- Arabic (Egyptian) "\u0644\u064A\u0647\u0645\u0627\u0628\u062A\u0643\u0644\u0645\u0648\u0634\u0639\u0631\u0628\u064A\u061F" encodes to "egbpdaj6bu4bxfgehfvwxn" (RFC 3492 section 7.1; tests/tests.js:34-38)
- Chinese (simplified) "\u4ED6\u4EEC\u4E3A\u4EC0\u4E48\u4E0D\u8BF4\u4E2d\u6587" encodes to "ihqwcrb4cv8a8dqg056pqjye" (tests/tests.js:39-43)
- Chinese (traditional) "\u4ED6\u5011\u7232\u4EC0\u9EBD\u4E0D\u8AAA\u4E2D\u6587" encodes to "ihqwctvzc91f659drss3x8bo0yb" (tests/tests.js:44-48)
- Czech "Pro\u010Dprost\u011Bnemluv\xED\u010Desky" encodes to "Proprostnemluvesky-uyb24dma41a" (tests/tests.js:49-53)
- Hebrew "\u05DC\u05DE\u05D4\u05D4\u05DD\u05E4\u05E9\u05D5\u05D8\u05DC\u05D0\u05DE\u05D3\u05D1\u05E8\u05D9\u05DD\u05E2\u05D1\u05E8\u05D9\u05EA" encodes to "4dbcagdahymbxekheh6e0a7fei0b" (tests/tests.js:54-58)
- Hindi (Devanagari) "\u092F\u0939\u0932\u094B\u0917\u0939\u093F\u0928\u094D\u0926\u0940\u0915\u094D\u092F\u094B\u0902\u0928\u0939\u0940\u0902\u092C\u094B\u0932\u0938\u0915\u0924\u0947\u0939\u0948\u0902" encodes to "i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd" (tests/tests.js:59-63)
- Japanese (kanji and hiragana) "\u306A\u305C\u307F\u3093\u306A\u65E5\u672C\u8A9E\u3092\u8A71\u3057\u3066\u304F\u308C\u306A\u3044\u306E\u304B" encodes to "n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa" (tests/tests.js:64-68)
- Korean (Hangul syllables) "\uC138\uACC4\uC758\uBAA8\uB4E0\uC0AC\uB78C\uB4E4\uC774\uD55C\uAD6D\uC5B4\uB97C\uC774\uD574\uD55C\uB2E4\uBA74\uC5BC\uB9C8\uB098\uC88B\uC744\uAE4C" encodes to "989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c" (tests/tests.js:69-73)
- Russian (Cyrillic) "\u043F\u043E\u0447\u0435\u043C\u0443\u0436\u0435\u043E\u043D\u0438\u043D\u0435\u0433\u043E\u0432\u043E\u0440\u044F\u0442\u043F\u043E\u0440\u0443\u0441\u0441\u043A\u0438" encodes to "b1abfaaepdrnnbgefbadotcwatmq2g4l" (RFC 3492 section 7.1, mixed-case annotation omitted; tests/tests.js:83-87)
- Spanish "Porqu\xE9nopuedensimplementehablarenEspa\xF1ol" encodes to "PorqunopuedensimplementehablarenEspaol-fmd56a" (tests/tests.js:88-92)
- Vietnamese "T\u1EA1isaoh\u1ECDkh\xF4ngth\u1EC3ch\u1EC9n\xF3iti\u1EBFngVi\u1EC7t" encodes to "TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g" (tests/tests.js:93-97)
- String "3\u5E74B\u7D44\u91D1\u516B\u5148\u751F" encodes to "3B-ww4c5e180e575a65lsy2b" (tests/tests.js:98-101)
- String "\u5B89\u5BA4\u5948\u7F8E\u6075-with-SUPER-MONKEYS" encodes to "-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n" (tests/tests.js:102-105)
- String "Hello-Another-Way-\u305D\u308C\u305E\u308C\u306E\u5834\u6240" encodes to "Hello-Another-Way--fc4qua05auwb3674vfr0b" (tests/tests.js:106-109)
- String "\u3072\u3068\u3064\u5C4B\u6839\u306E\u4E0B2" encodes to "2-u9tlzr9756bt3uc0v" (tests/tests.js:110-113)
- String "Maji\u3067Koi\u3059\u308B5\u79D2\u524D" encodes to "MajiKoi5-783gue6qz075azm5e" (tests/tests.js:114-117)
- String "\u30D1\u30D5\u30A3\u30FCde\u30EB\u30F3\u30D0" encodes to "de-jg4avhby1noc0d" (tests/tests.js:118-121)
- String "\u305D\u306E\u30B9\u30D4\u30FC\u30C9\u3067" encodes to "d9juau41awczczp" (tests/tests.js:122-125)
- ASCII string "-> $1.00 <-" encodes to "-> $1.00 <--" (edge case: breaks existing host-name label rules; tests/tests.js:131-135)

Used by: `punycode-encode.md`, `punycode-decode.md`

## UCS2 Fixtures

**Purpose and shape:** The `ucs2` array (tests/tests.js:137-175) contains arrays of numeric Unicode code points and their UCS-2 string representations. These fixtures test `punycode.ucs2.encode()` and `punycode.ucs2.decode()`, which convert between code-point arrays and UCS-2 strings, with special handling for astral characters (code points above U+FFFF) represented as surrogate pairs.

**Notable entries:**

- Consecutive astral symbols [127829, 119808, 119558, 119638] encode to "\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56" (tests/tests.js:140-144)
- U+D800 (high surrogate) followed by non-surrogates [55296, 97, 98] encode to "\uD800ab" (tests/tests.js:145-149)
- U+DC00 (low surrogate) followed by non-surrogates [56320, 97, 98] encode to "\uDC00ab" (tests/tests.js:150-154)
- High surrogate followed by another high surrogate [0xD800, 0xD800] encode to "\uD800\uD800" (tests/tests.js:155-159)
- Unmatched high surrogate, surrogate pair, unmatched high surrogate [0xD800, 0x1D306, 0xD800] encode to "\uD800\uD834\uDF06\uD800" (tests/tests.js:160-164)
- Low surrogate followed by another low surrogate [0xDC00, 0xDC00] encode to "\uDC00\uDC00" (tests/tests.js:165-169)
- Unmatched low surrogate, surrogate pair, unmatched low surrogate [0xDC00, 0x1D306, 0xDC00] encode to "\uDC00\uD834\uDF06\uDC00" (tests/tests.js:170-174)

Used by: `punycode-ucs2-encode.md`, `punycode-ucs2-decode.md`

## Domains Fixtures

**Purpose and shape:** The `domains` array (tests/tests.js:176-220) contains Unicode domain names (or email addresses) and their Punycode-encoded (ACE) representations. These fixtures test `punycode.toASCII()` and `punycode.toUnicode()`, which convert entire domain names or email addresses, applying label-wise Punycode and handling IDNA2003 separators. The `decoded` value is a Unicode domain or email; the `encoded` value is its ACE representation.

**Notable entries:**

- Spanish "ma\xF1ana.com" encodes to "xn--maana-pta.com" (tests/tests.js:177-180)
- Trailing dot "example.com." encodes to "example.com." (preserved; tests/tests.js:181-184)
- German "b\xFCcher.com" encodes to "xn--bcher-kva.com" (tests/tests.js:185-188)
- French "caf\xE9.com" encodes to "xn--caf-dma.com" (tests/tests.js:189-192)
- Symbols "\u2603-\u2318.com" encodes to "xn----dqo34k.com" (tests/tests.js:193-196)
- Symbol with surrogate "\uD400\u2603-\u2318.com" encodes to "xn----dqo34kn65z.com" (tests/tests.js:197-200)
- Emoji poo "\uD83D\uDCA9.la" encodes to "xn--ls8h.la" (tests/tests.js:201-205)
- Non-printable ASCII "\0\x01\x02foo.bar" encodes to "\0\x01\x02foo.bar" (preserved; tests/tests.js:206-210)
- Email address "\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a" encodes to "\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq" (local part preserved; tests/tests.js:211-215)
- DEL character "foo\x7F.example" encodes to "foo\x7F.example" (preserved; tests/tests.js:216-219)

Used by: `punycode-to-ascii.md`, `punycode-to-unicode.md`

## Separators Fixtures

**Purpose and shape:** The `separators` array (tests/tests.js:221-242) contains domain names using different IDNA2003-compatible label separators (U+002E, U+3002, U+FF0E, U+FF61). These fixtures test `punycode.toASCII()`, which normalizes all separator variants to U+002E (full stop) internally before Punycode encoding (punycode.js:19, 82). The `decoded` value uses a non-standard separator; the `encoded` value has all separators normalized to U+002E.

**Notable entries:**

- U+002E separator "ma\xF1ana.com" (literal period) encodes to "xn--maana-pta.com" (tests/tests.js:222-226)
- U+3002 separator "ma\xF1ana\u3002com" (ideographic full stop) encodes to "xn--maana-pta.com" (tests/tests.js:227-231)
- U+FF0E separator "ma\xF1ana\uFF0Ecom" (fullwidth full stop) encodes to "xn--maana-pta.com" (tests/tests.js:232-236)
- U+FF61 separator "ma\xF1ana\uFF61com" (halfwidth ideographic full stop) encodes to "xn--maana-pta.com" (tests/tests.js:237-241)

Used by: `punycode-to-ascii.md`

## Cross-references

- [Punycode encode specification](punycode-encode.md)
- [Punycode decode specification](punycode-decode.md)
- [UCS-2 encode specification](punycode-ucs2-encode.md)
- [UCS-2 decode specification](punycode-ucs2-decode.md)
- [Domain to ASCII specification](punycode-to-ascii.md)
- [Domain to Unicode specification](punycode-to-unicode.md)
