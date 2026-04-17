# Port Target

- **Language**: Go 1.22+
- **Output directory**: `port/`
- **Module path**: `punycode-port` (local module — `cd port && go mod init punycode-port`)
- **Package**: `punycode`
- **Test command**: `cd port && go test ./...`
- **Style**: idiomatic Go — functions over classes, explicit error returns, table-driven tests.
- **Scope**: port every exported function from `punycode.js` (`decode`, `encode`, `toUnicode`, `toASCII`, `ucs2.decode`, `ucs2.encode`).
- **Fidelity**: every RFC test vector from `tests/tests.js` must pass in Go.
