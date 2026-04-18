package punycode_test

// stringsData mirrors testData.strings from tests/tests.js:7-136.
// Note: Go \xNN escapes are raw bytes; Unicode code points use \uNNNN.
var stringsData = []struct {
	description string
	decoded     string
	encoded     string
}{
	{"a single basic code point", "Bach", "Bach-"},
	{"a single non-ASCII character", "\u00FC", "tda"},
	{"multiple non-ASCII characters", "\u00FC\u00EB\u00E4\u00F6\u2665", "4can8av2009b"},
	{"mix of ASCII and non-ASCII characters", "b\u00FCcher", "bcher-kva"},
	{"long string with both ASCII and non-ASCII characters",
		"Willst du die Bl\u00FCthe des fr\u00FChen, die Fr\u00FCchte des sp\u00E4teren Jahres",
		"Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal"},
	{"Arabic (Egyptian)",
		"\u0644\u064A\u0647\u0645\u0627\u0628\u062A\u0643\u0644\u0645\u0648\u0634\u0639\u0631\u0628\u064A\u061F",
		"egbpdaj6bu4bxfgehfvwxn"},
	{"Chinese (simplified)",
		"\u4ED6\u4EEC\u4E3A\u4EC0\u4E48\u4E0D\u8BF4\u4E2D\u6587",
		"ihqwcrb4cv8a8dqg056pqjye"},
	{"Chinese (traditional)",
		"\u4ED6\u5011\u7232\u4EC0\u9EBD\u4E0D\u8AAA\u4E2D\u6587",
		"ihqwctvzc91f659drss3x8bo0yb"},
	{"Czech",
		"Pro\u010Dprost\u011Bnemluv\u00ED\u010Desky",
		"Proprostnemluvesky-uyb24dma41a"},
	{"Hebrew",
		"\u05DC\u05DE\u05D4\u05D4\u05DD\u05E4\u05E9\u05D5\u05D8\u05DC\u05D0\u05DE\u05D3\u05D1\u05E8\u05D9\u05DD\u05E2\u05D1\u05E8\u05D9\u05EA",
		"4dbcagdahymbxekheh6e0a7fei0b"},
	{"Hindi (Devanagari)",
		"\u092F\u0939\u0932\u094B\u0917\u0939\u093F\u0928\u094D\u0926\u0940\u0915\u094D\u092F\u094B\u0902\u0928\u0939\u0940\u0902\u092C\u094B\u0932\u0938\u0915\u0924\u0947\u0939\u0948\u0902",
		"i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd"},
	{"Japanese (kanji and hiragana)",
		"\u306A\u305C\u307F\u3093\u306A\u65E5\u672C\u8A9E\u3092\u8A71\u3057\u3066\u304F\u308C\u306A\u3044\u306E\u304B",
		"n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa"},
	{"Korean (Hangul syllables)",
		"\uC138\uACC4\uC758\uBAA8\uB4E0\uC0AC\uB78C\uB4E4\uC774\uD55C\uAD6D\uC5B4\uB97C\uC774\uD574\uD55C\uB2E4\uBA74\uC5BC\uB9C8\uB098\uC88B\uC744\uAE4C",
		"989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c"},
	{"Russian (Cyrillic)",
		"\u043F\u043E\u0447\u0435\u043C\u0443\u0436\u0435\u043E\u043D\u0438\u043D\u0435\u0433\u043E\u0432\u043E\u0440\u044F\u0442\u043F\u043E\u0440\u0443\u0441\u0441\u043A\u0438",
		"b1abfaaepdrnnbgefbadotcwatmq2g4l"},
	{"Spanish",
		"Porqu\u00E9nopuedensimplementehablarenEspa\u00F1ol",
		"PorqunopuedensimplementehablarenEspaol-fmd56a"},
	{"Vietnamese",
		"T\u1EA1isaoh\u1ECDkh\u00F4ngth\u1EC3ch\u1EC9n\u00F3iti\u1EBFngVi\u1EC7t",
		"TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g"},
	{"", "3\u5E74B\u7D44\u91D1\u516B\u5148\u751F", "3B-ww4c5e180e575a65lsy2b"},
	{"", "\u5B89\u5BA4\u5948\u7F8E\u6075-with-SUPER-MONKEYS", "-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n"},
	{"", "Hello-Another-Way-\u305D\u308C\u305E\u308C\u306E\u5834\u6240", "Hello-Another-Way--fc4qua05auwb3674vfr0b"},
	{"", "\u3072\u3068\u3064\u5C4B\u6839\u306E\u4E0B2", "2-u9tlzr9756bt3uc0v"},
	{"", "Maji\u3067Koi\u3059\u308B5\u79D2\u524D", "MajiKoi5-783gue6qz075azm5e"},
	{"", "\u30D1\u30D5\u30A3\u30FCde\u30EB\u30F3\u30D0", "de-jg4avhby1noc0d"},
	{"", "\u305D\u306E\u30B9\u30D4\u30FC\u30C9\u3067", "d9juau41awczczp"},
	{"ASCII string that breaks the existing rules for host-name labels", "-> $1.00 <-", "-> $1.00 <--"},
}

// ucs2Data mirrors testData.ucs2 from tests/tests.js:137-175.
// Lone-surrogate cases (tests.js:146-173) are omitted: Go strings are UTF-8
// and lone surrogates (U+D800-U+DFFF) cannot be faithfully round-tripped.
var ucs2Data = []struct {
	description string
	decoded     []int32
	encoded     string
}{
	{
		"Consecutive astral symbols",
		[]int32{127829, 119808, 119558, 119638},
		// JS '\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56'
		// = U+1F355 U+1D400 U+1D306 U+1D356
		"\U0001F355\U0001D400\U0001D306\U0001D356",
	},
}

// domainsData mirrors testData.domains from tests/tests.js:176-220.
var domainsData = []struct {
	description string
	decoded     string
	encoded     string
}{
	{"", "ma\u00F1ana.com", "xn--maana-pta.com"},
	{"", "example.com.", "example.com."},
	{"", "b\u00FCcher.com", "xn--bcher-kva.com"},
	{"", "caf\u00E9.com", "xn--caf-dma.com"},
	{"", "\u2603-\u2318.com", "xn----dqo34k.com"},
	{"", "\uD400\u2603-\u2318.com", "xn----dqo34kn65z.com"},
	{"Emoji", "\U0001F4A9.la", "xn--ls8h.la"},
	{"Non-printable ASCII", "\x00\x01\x02foo.bar", "\x00\x01\x02foo.bar"},
	{"Email address",
		"\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a",
		"\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq"},
	{"U+007F DEL is treated as ASCII", "foo\x7F.example", "foo\x7F.example"},
}

// separatorsData mirrors testData.separators from tests/tests.js:221-242.
var separatorsData = []struct {
	description string
	decoded     string
	encoded     string
}{
	{"Using U+002E as separator", "ma\u00F1ana\x2Ecom", "xn--maana-pta.com"},
	{"Using U+3002 as separator", "ma\u00F1ana\u3002com", "xn--maana-pta.com"},
	{"Using U+FF0E as separator", "ma\u00F1ana\uFF0Ecom", "xn--maana-pta.com"},
	{"Using U+FF61 as separator", "ma\u00F1ana\uFF61com", "xn--maana-pta.com"},
}
