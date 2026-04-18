package punycode

// Bootstring constants from punycode.js:3-14 / RFC 3492 §5.
const (
	maxInt      int32 = 2147483647
	base        int32 = 36
	tMin        int32 = 1
	tMax        int32 = 26
	skew        int32 = 38
	damp        int32 = 700
	initialBias int32 = 72
	initialN    int32 = 128
	delimiter   byte  = '-'

	// baseMinusTMin is derived from base and tMin (punycode.js:29).
	baseMinusTMin int32 = base - tMin // 35
)
