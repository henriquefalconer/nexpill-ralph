# Punycode.js → Go Port — TODO

Port `punycode.js` to idiomatic Go under `port/` per `TARGET.md`. Each bullet is scoped to roughly one ralph build iteration (one commit). Citations are `specs/<file>.md` for behaviour, `punycode.js:<line>` for reference code, and `tests/tests.js:<line>` for vectors.

## Baseline — DONE

- Test-driven specs in `specs/` (flat structure), one per Mocha describe block plus a shared fixtures spec and an index. See `specs/README.md`. Every claim is cited to `punycode.js` and/or `tests/tests.js`.
- Source-driven specs in `specs/` (flat structure), one per source file: `specs/src-punycode.md` (the 443-line library) and `specs/src-scripts-prepublish.md` (the 17-line build step). Every section cites the exact source line range.

## Next iterations (in dependency order)

1. **Bootstrap the Go module.** Create `port/go.mod` via `cd port && go mod init punycode-port`; add a `port/doc.go` with the package comment; add a `port/punycode_test.go` stub holding the shared test-vector tables as Go literals, mirroring the four buckets in `specs/test-data-fixtures.md`. No functions yet — the table-driven tests compile against stubs that return zero values and are expected to fail.

2. **Port `ucs2.decode` and `ucs2.encode`.** Per `specs/punycode-ucs2-decode.md` and `specs/punycode-ucs2-encode.md`. Signature choice: accept `[]uint16` or a UTF-16 `string` wrapper so lone-surrogate fixtures (`tests/tests.js:145-174`) round-trip byte-for-byte. Verify the `ucs2` bucket passes.

3. **Port `basicToDigit`, `digitToBasic`, `adapt`.** The Bootstring helpers at `punycode.js:144-187`. Unit-test each in isolation against hand-computed values; they have no public-API tests of their own but feed every other test.

4. **Port `decode` (label-level).** Per `specs/punycode-decode.md`. Use `int32` arithmetic to match the `maxInt = 2^31-1` overflow guards. Verify all `strings` bucket vectors (decode direction) plus the uppercase-Z test and the three `RangeError` paths (map to Go `error` returns).

5. **Port `encode` (label-level).** Per `specs/punycode-encode.md`. Always emit lowercase digits (no mixed-case flag). Verify all `strings` bucket vectors (encode direction); confirm the `b1abfaaepdrnnbgefbadotcwatmq2g4l` expectation (lowercase, per `tests/tests.js:74-82`).

6. **Port `mapDomain` and `regexSeparators`.** Per `specs/punycode-to-unicode.md` §"Algorithm" and `specs/punycode-to-ascii.md` §"Algorithm". In Go, replace the JS regex with a `strings.Map` rune replacer covering U+002E / U+3002 / U+FF0E / U+FF61. Handle the `@` split via `strings.SplitN(_, "@", 2)`.

7. **Port `toUnicode`.** Per `specs/punycode-to-unicode.md`. Verify all `domains` bucket vectors plus the `strings` identity-pass-through loop (`tests/tests.js:332-343`).

8. **Port `toASCII`.** Per `specs/punycode-to-ascii.md`. Verify `domains` vectors, `strings` identity-pass-through, AND the IDNA2003 `separators` bucket (`tests/tests.js:221-242`).

9. **Public API surface.** Expose `Decode`, `Encode`, `ToASCII`, `ToUnicode`, and a `UCS2` sub-package (or exported struct) matching the JS shape at `punycode.js:419-441`. Add a `Version` constant to match `punycode.js:425` if the consumer needs it; otherwise skip.

10. **CI / final sweep.** Ensure `cd port && go test ./...` passes on every vector. Add table-driven sub-tests so individual failures name the source `tests/tests.js` line. Run `go vet` and `gofmt`.

## Out of scope

- IDNA2008 / UTS-46 processing. Punycode.js only implements Bootstring + IDNA2003 separator normalisation (`punycode.js:19`).
- The `version` string is not tested and not a port requirement (`punycode.js:425`).
- Multi-`@` email handling beyond what `parts[1]` gives (`punycode.js:72-80`) — not tested, not ported behaviour.

## Fidelity requirement (from `TARGET.md`)

> Every RFC test vector from `tests/tests.js` must pass in Go.

This is the acceptance bar. Specs document *why* a vector exists; `tests/tests.js` is the ground truth for *what* the vector is.
