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

// ErrOverflow is returned when arithmetic overflows the integer type.
// Mirrors punycode.js:23 ("overflow" error).
var ErrOverflow = errors.New("Overflow: input needs wider integers to process")

// ErrNotBasic is returned when a non-basic code point (≥ 0x80) appears where
// only basic ASCII is permitted. Mirrors punycode.js:24 ("not-basic" error).
var ErrNotBasic = errors.New("Illegal input >= 0x80 (not a basic code point)")

// ErrInvalidInput is returned when the input string is structurally malformed.
// Mirrors punycode.js:25 ("invalid-input" error).
var ErrInvalidInput = errors.New("Invalid input")
