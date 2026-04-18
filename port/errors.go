package punycode

import "errors"

// Sentinel errors mirroring punycode.js:22-26.
// Message strings match verbatim so test output is grep-compatible.
var (
	ErrOverflow     = errors.New("Overflow: input needs wider integers to process")
	ErrNotBasic     = errors.New("Illegal input >= 0x80 (not a basic code point)")
	ErrInvalidInput = errors.New("Invalid input")
)
