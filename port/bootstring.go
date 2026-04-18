package punycode

// basicToDigit converts a basic code point to a Bootstring digit.
// punycode.js:135-155; specs/src-punycode.md:114-123.
func basicToDigit(codePoint int32) int32 {
	if codePoint >= 0x30 && codePoint < 0x3A {
		return 26 + (codePoint - 0x30)
	}
	if codePoint >= 0x41 && codePoint < 0x5B {
		return codePoint - 0x41
	}
	if codePoint >= 0x61 && codePoint < 0x7B {
		return codePoint - 0x61
	}
	return base
}

// digitToBasic converts a Bootstring digit (0..35) to a lowercase ASCII byte.
// flag parameter dropped: both call sites in punycode.js:358,364 pass 0.
// punycode.js:157-172; specs/src-punycode.md:135-138.
func digitToBasic(digit int32) byte {
	// 0..25 → 'a'..'z'  (digit + 22 + 75 = digit + 97)
	// 26..35 → '0'..'9' (digit + 22)
	var add int32 = 22
	if digit < 26 {
		add += 75
	}
	return byte(digit + add)
}

// adapt implements the bias adaptation function from RFC 3492 §3.4.
// punycode.js:174-187; specs/src-punycode.md:142-150.
func adapt(delta, numPoints int32, firstTime bool) int32 {
	const threshold = (baseMinusTMin * tMax) >> 1 // 455

	if firstTime {
		delta /= damp
	} else {
		delta >>= 1
	}
	delta += delta / numPoints

	k := int32(0)
	for delta > threshold {
		delta /= baseMinusTMin
		k += base
	}
	return k + (baseMinusTMin+1)*delta/(delta+skew)
}
