// Package punycode implements the Punycode encoding algorithm (RFC 3492)
// and the IDNA domain name processing (RFC 5891).
//
// It is a Go port of mathiasbynens/punycode.js v2.3.1.
// The JS nested namespace punycode.ucs2.{encode,decode} is flattened to
// UCS2Encode / UCS2Decode in this package.
package punycode

import "errors"

// Version mirrors the upstream punycode.js version string.
const Version = "2.3.1"

// Bootstring parameters (RFC 3492 §5.1).
const (
	maxInt      = 2147483647 // 0x7FFFFFFF, 2^31−1
	base        = 36
	tMin        = 1
	tMax        = 26
	skew        = 38
	damp        = 700
	initialBias = 72
	initialN    = 128 // 0x80
	delimiter   = '-' // U+002D

	baseMinusTMin = base - tMin // 35; pre-computed for bias calculations
)

// Sentinel errors matching punycode.js:22-25.
var (
	ErrOverflow     = errors.New("Overflow: input needs wider integers to process")
	ErrNotBasic     = errors.New("Illegal input >= 0x80 (not a basic code point)")
	ErrInvalidInput = errors.New("Invalid input")
)
