package punycode

// basicToDigit maps an ASCII code point to its Punycode digit value [0,35],
// or returns base (36) as a sentinel for "not a valid digit".
// Mirrors punycode.js:144-155.
func basicToDigit(cp rune) int {
	switch {
	case cp >= 0x30 && cp < 0x3A: // '0'..'9' → 26..35
		return 26 + int(cp-0x30)
	case cp >= 0x41 && cp < 0x5B: // 'A'..'Z' → 0..25
		return int(cp - 0x41)
	case cp >= 0x61 && cp < 0x7B: // 'a'..'z' → 0..25
		return int(cp - 0x61)
	default:
		return base
	}
}

// digitToBasic maps a Punycode digit [0,35] back to an ASCII code point.
// flag=true requests uppercase output for alphabetic digits (0..25).
// Mirrors punycode.js:168-172. Both call sites pass flag=false.
func digitToBasic(digit int, flag bool) rune {
	// digit + 22 + 75*(digit<26) - ((flag!=0)<<5), expressed with explicit ifs.
	r := digit + 22
	if digit < 26 {
		r += 75 // digit+97 == 'a'+digit
	}
	if flag {
		r -= 32 // shift lowercase to uppercase
	}
	return rune(r)
}

// adapt updates the Bootstring bias after each code-point insertion/extraction.
// Implements RFC 3492 §3.4. Mirrors punycode.js:178-187.
func adapt(delta, numPoints int, firstTime bool) int {
	if firstTime {
		delta /= damp
	} else {
		delta >>= 1
	}
	delta += delta / numPoints
	k := 0
	// threshold: baseMinusTMin*tMax >> 1 == 35*26>>1 == 455
	for delta > baseMinusTMin*tMax>>1 {
		delta /= baseMinusTMin
		k += base
	}
	return k + (baseMinusTMin+1)*delta/(delta+skew)
}
