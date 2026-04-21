package punycode

// testVectors mirrors testData from tests/tests.js:6-243.
// Fields match the JS testData keys (Strings, UCS2, Domains, Separators).
var testVectors = struct {
	Strings    []stringVector
	UCS2       []ucs2Vector
	Domains    []domainVector
	Separators []separatorVector
}{
	// 24 vectors from tests/tests.js:7-136 (testData.strings).
	Strings: []stringVector{
		{
			description: "a single basic code point",
			decoded:     "Bach",
			encoded:     "Bach-",
		},
		{
			description: "a single non-ASCII character",
			decoded:     "\u00FC",
			encoded:     "tda",
		},
		{
			description: "multiple non-ASCII characters",
			decoded:     "\u00FC\u00EB\u00E4\u00F6\u2665",
			encoded:     "4can8av2009b",
		},
		{
			description: "mix of ASCII and non-ASCII characters",
			decoded:     "b\u00FCcher",
			encoded:     "bcher-kva",
		},
		{
			description: "long string with both ASCII and non-ASCII characters",
			decoded:     "Willst du die Bl\u00FCthe des fr\u00FChen, die Fr\u00FCchte des sp\u00E4teren Jahres",
			encoded:     "Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal",
		},
		{
			description: "Arabic (Egyptian)",
			decoded:     "\u0644\u064A\u0647\u0645\u0627\u0628\u062A\u0643\u0644\u0645\u0648\u0634\u0639\u0631\u0628\u064A\u061F",
			encoded:     "egbpdaj6bu4bxfgehfvwxn",
		},
		{
			description: "Chinese (simplified)",
			decoded:     "\u4ED6\u4EEC\u4E3A\u4EC0\u4E48\u4E0D\u8BF4\u4E2D\u6587",
			encoded:     "ihqwcrb4cv8a8dqg056pqjye",
		},
		{
			description: "Chinese (traditional)",
			decoded:     "\u4ED6\u5011\u7232\u4EC0\u9EBD\u4E0D\u8AAA\u4E2D\u6587",
			encoded:     "ihqwctvzc91f659drss3x8bo0yb",
		},
		{
			description: "Czech",
			decoded:     "Pro\u010Dprost\u011Bnemluv\u00ED\u010Desky",
			encoded:     "Proprostnemluvesky-uyb24dma41a",
		},
		{
			description: "Hebrew",
			decoded:     "\u05DC\u05DE\u05D4\u05D4\u05DD\u05E4\u05E9\u05D5\u05D8\u05DC\u05D0\u05DE\u05D3\u05D1\u05E8\u05D9\u05DD\u05E2\u05D1\u05E8\u05D9\u05EA",
			encoded:     "4dbcagdahymbxekheh6e0a7fei0b",
		},
		{
			description: "Hindi (Devanagari)",
			decoded:     "\u092F\u0939\u0932\u094B\u0917\u0939\u093F\u0928\u094D\u0926\u0940\u0915\u094D\u092F\u094B\u0902\u0928\u0939\u0940\u0902\u092C\u094B\u0932\u0938\u0915\u0924\u0947\u0939\u0948\u0902",
			encoded:     "i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd",
		},
		{
			description: "Japanese (kanji and hiragana)",
			decoded:     "\u306A\u305C\u307F\u3093\u306A\u65E5\u672C\u8A9E\u3092\u8A71\u3057\u3066\u304F\u308C\u306A\u3044\u306E\u304B",
			encoded:     "n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa",
		},
		{
			description: "Korean (Hangul syllables)",
			decoded:     "\uC138\uACC4\uC758\uBAA8\uB4E0\uC0AC\uB78C\uB4E4\uC774\uD55C\uAD6D\uC5B4\uB97C\uC774\uD574\uD55C\uB2E4\uBA74\uC5BC\uB9C8\uB098\uC88B\uC744\uAE4C",
			encoded:     "989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c",
		},
		{
			description: "Russian (Cyrillic)",
			decoded:     "\u043F\u043E\u0447\u0435\u043C\u0443\u0436\u0435\u043E\u043D\u0438\u043D\u0435\u0433\u043E\u0432\u043E\u0440\u044F\u0442\u043F\u043E\u0440\u0443\u0441\u0441\u043A\u0438",
			encoded:     "b1abfaaepdrnnbgefbadotcwatmq2g4l",
		},
		{
			description: "Spanish",
			decoded:     "Porqu\u00E9nopuedensimplementehablarenEspa\u00F1ol",
			encoded:     "PorqunopuedensimplementehablarenEspaol-fmd56a",
		},
		{
			description: "Vietnamese",
			decoded:     "T\u1EA1isaoh\u1ECDkh\u00F4ngth\u1EC3ch\u1EC9n\u00F3iti\u1EBFngVi\u1EC7t",
			encoded:     "TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g",
		},
		{
			decoded: "3\u5E74B\u7D44\u91D1\u516B\u5148\u751F",
			encoded: "3B-ww4c5e180e575a65lsy2b",
		},
		{
			decoded: "\u5B89\u5BA4\u5948\u7F8E\u6075-with-SUPER-MONKEYS",
			encoded: "-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n",
		},
		{
			decoded: "Hello-Another-Way-\u305D\u308C\u305E\u308C\u306E\u5834\u6240",
			encoded: "Hello-Another-Way--fc4qua05auwb3674vfr0b",
		},
		{
			decoded: "\u3072\u3068\u3064\u5C4B\u6839\u306E\u4E0B2",
			encoded: "2-u9tlzr9756bt3uc0v",
		},
		{
			decoded: "Maji\u3067Koi\u3059\u308B5\u79D2\u524D",
			encoded: "MajiKoi5-783gue6qz075azm5e",
		},
		{
			decoded: "\u30D1\u30D5\u30A3\u30FCde\u30EB\u30F3\u30D0",
			encoded: "de-jg4avhby1noc0d",
		},
		{
			decoded: "\u305D\u306E\u30B9\u30D4\u30FC\u30C9\u3067",
			encoded: "d9juau41awczczp",
		},
		{
			description: "ASCII string that breaks the existing rules for host-name labels",
			decoded:     "-> $1.00 <-",
			encoded:     "-> $1.00 <--",
		},
	},

	// 7 vectors from tests/tests.js:137-175 (testData.ucs2).
	UCS2: []ucs2Vector{
		{
			"consecutive astral symbols",
			[]uint16{0xD83C, 0xDF55, 0xD835, 0xDC00, 0xD834, 0xDF06, 0xD834, 0xDF56},
			[]rune{127829, 119808, 119558, 119638},
		},
		{
			"high surrogate followed by non-surrogates",
			[]uint16{0xD800, 0x61, 0x62},
			[]rune{0xD800, 97, 98},
		},
		{
			"low surrogate followed by non-surrogates",
			[]uint16{0xDC00, 0x61, 0x62},
			[]rune{0xDC00, 97, 98},
		},
		{
			"high surrogate followed by another high surrogate",
			[]uint16{0xD800, 0xD800},
			[]rune{0xD800, 0xD800},
		},
		{
			"unmatched high, surrogate pair, unmatched high",
			[]uint16{0xD800, 0xD834, 0xDF06, 0xD800},
			[]rune{0xD800, 0x1D306, 0xD800},
		},
		{
			"low surrogate followed by another low surrogate",
			[]uint16{0xDC00, 0xDC00},
			[]rune{0xDC00, 0xDC00},
		},
		{
			"unmatched low, surrogate pair, unmatched low",
			[]uint16{0xDC00, 0xD834, 0xDF06, 0xDC00},
			[]rune{0xDC00, 0x1D306, 0xDC00},
		},
	},

	// 10 vectors from tests/tests.js:176-220 (testData.domains).
	Domains: []domainVector{
		{
			decoded: "ma\u00F1ana.com",
			encoded: "xn--maana-pta.com",
		},
		{
			decoded: "example.com.",
			encoded: "example.com.",
		},
		{
			decoded: "b\u00FCcher.com",
			encoded: "xn--bcher-kva.com",
		},
		{
			decoded: "caf\u00E9.com",
			encoded: "xn--caf-dma.com",
		},
		{
			decoded: "\u2603-\u2318.com",
			encoded: "xn----dqo34k.com",
		},
		{
			// Lone high surrogate U+D400 — WTF-8 raw bytes.
			// tests/tests.js:197-200
			description: "lone surrogate (WTF-8 U+D400)",
			decoded:     "\xed\x90\x80\xe2\x98\x83-\xe2\x8c\x98.com",
			encoded:     "xn----dqo34kn65z.com",
			skipToASCII: true,
		},
		{
			// U+1F4A9 stored in JS as surrogate pair \uD83D\uDCA9.
			// tests/tests.js:201-205
			description: "Emoji",
			decoded:     "\U0001F4A9.la",
			encoded:     "xn--ls8h.la",
		},
		{
			// Non-printable ASCII: all bytes < 0x80 so no encoding occurs.
			// tests/tests.js:206-210
			description: "Non-printable ASCII",
			decoded:     "\x00\x01\x02foo.bar",
			encoded:     "\x00\x01\x02foo.bar",
		},
		{
			// Cyrillic local part preserved verbatim across '@'.
			// tests/tests.js:211-215
			description: "Email address",
			decoded:     "\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a",
			encoded:     "\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq",
		},
		{
			// DEL (U+007F) is ≤ 0x7F so it is not treated as non-ASCII.
			// tests/tests.js:216-219
			decoded: "foo\x7F.example",
			encoded: "foo\x7F.example",
		},
	},

	// 4 vectors from tests/tests.js:221-242 (testData.separators).
	Separators: []separatorVector{
		{
			description: "Using U+002E as separator",
			decoded:     "ma\u00F1ana.com",
			encoded:     "xn--maana-pta.com",
		},
		{
			description: "Using U+3002 as separator",
			decoded:     "ma\u00F1ana\u3002com",
			encoded:     "xn--maana-pta.com",
		},
		{
			description: "Using U+FF0E as separator",
			decoded:     "ma\u00F1ana\uFF0Ecom",
			encoded:     "xn--maana-pta.com",
		},
		{
			description: "Using U+FF61 as separator",
			decoded:     "ma\u00F1ana\uFF61com",
			encoded:     "xn--maana-pta.com",
		},
	},
}

// ucs2EncodeExpected holds the expected []byte output for each testVectors.UCS2 entry.
// Surrogates use WTF-8 (3-byte); valid astral code points use standard 4-byte UTF-8.
var ucs2EncodeExpected = [][]byte{
	// vector 1: valid astral symbols → standard UTF-8
	[]byte(string([]rune{127829, 119808, 119558, 119638})),
	// vector 2: U+D800 'a' 'b' — WTF-8 for U+D800: ED A0 80
	{0xED, 0xA0, 0x80, 0x61, 0x62},
	// vector 3: U+DC00 'a' 'b' — WTF-8 for U+DC00: ED B0 80
	{0xED, 0xB0, 0x80, 0x61, 0x62},
	// vector 4: U+D800 U+D800
	{0xED, 0xA0, 0x80, 0xED, 0xA0, 0x80},
	// vector 5: U+D800 U+1D306 U+D800 — U+1D306 UTF-8: F0 9D 8C 86
	{0xED, 0xA0, 0x80, 0xF0, 0x9D, 0x8C, 0x86, 0xED, 0xA0, 0x80},
	// vector 6: U+DC00 U+DC00
	{0xED, 0xB0, 0x80, 0xED, 0xB0, 0x80},
	// vector 7: U+DC00 U+1D306 U+DC00
	{0xED, 0xB0, 0x80, 0xF0, 0x9D, 0x8C, 0x86, 0xED, 0xB0, 0x80},
}

type stringVector struct {
	description string
	decoded     string
	encoded     string
}

type ucs2Vector struct {
	desc       string
	units      []uint16
	codePoints []rune
}

type domainVector struct {
	description string
	decoded     string
	encoded     string
	skipToASCII bool
}

type separatorVector struct {
	description string
	decoded     string
	encoded     string
}
