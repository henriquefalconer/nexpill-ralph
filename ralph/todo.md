# Punycode.js → Go Port — TODO

Port `punycode.js` (443 lines) to idiomatic Go under `port/` per `TARGET.md`. Each bullet below is scoped to roughly one ralph build iteration (one commit) and is intended to be dropped into a fresh Claude Code session as-is. Every bullet cites:

- **Spec sections** as `specs/<file>.md:<line>-<line>` so the porter can `Read` the exact prose.
- **Source lines** as `punycode.js:<line>-<line>` (the reference implementation, 443 lines).
- **Test fixtures** as `tests/tests.js:<line>-<line>` (the ground-truth vectors, 371 lines).

Fidelity bar (from `TARGET.md:10`): every RFC test vector in `tests/tests.js` must pass in Go. Specs document *why* a vector exists; `tests/tests.js` is the ground truth for *what* the vector is.

## Baseline — DONE

- Test-driven specs in `specs/` (flat): `specs/punycode-ucs2-decode.md` (`tests/tests.js:245-271`), `specs/punycode-ucs2-encode.md` (`tests/tests.js:273-288`), `specs/punycode-decode.md` (`tests/tests.js:290-310`), `specs/punycode-encode.md` (`tests/tests.js:312-321`), `specs/punycode-to-unicode.md` (`tests/tests.js:323-344`), `specs/punycode-to-ascii.md` (`tests/tests.js:346-371`), and `specs/test-data-fixtures.md` for the shared vector tables at `tests/tests.js:6-243`. See `specs/README.md:7-19`.
- Source-driven specs: `specs/src-punycode.md` (covers `punycode.js:1-443` section by section) and `specs/src-scripts-prepublish.md` (the 17-line JS build step — out of port scope).
- `port/` directory created with full Go implementation (all items 1-10 complete).

## Next iterations (in dependency order) — ALL DONE

1. **Bootstrap the Go module + test scaffold.** — DONE
   - Create `port/` directory; run `cd port && go mod init punycode-port` (per `TARGET.md:5`). Target Go 1.22+ (`TARGET.md:3`), package name `punycode` (`TARGET.md:6`).
   - Add `port/doc.go` with the package-level comment: "Package punycode implements RFC 3492 Bootstring + IDNA2003 separator normalisation, ported from punycode.js".
   - Add `port/punycode_test.go` containing the shared vector tables as Go literals, mirroring the four buckets catalogued in `specs/test-data-fixtures.md:9-71`: `strings` (`tests/tests.js:7-136`), `ucs2` (`tests/tests.js:137-175`), `domains` (`tests/tests.js:176-220`), `separators` (`tests/tests.js:221-242`). Each row needs `description`, `decoded`, `encoded` fields (plus `ucs2` for the bucket of the same name).
   - Do **not** implement any functions yet. Stub all public symbols (`Decode`, `Encode`, `ToASCII`, `ToUnicode`, `UCS2Decode`, `UCS2Encode`) to return zero values so tests compile and uniformly fail.
   - Acceptance: `cd port && go build ./...` succeeds; `cd port && go test ./...` runs and fails (not errors out).

2. **Port the Bootstring constants + `error` mapping.**
   - Per `specs/src-punycode.md:22-60` (§2 "Module-level constants") and `punycode.js:3-26`.
   - Land typed Go constants: `maxInt int32 = 2147483647` (`punycode.js:4`), `base = 36`, `tMin = 1`, `tMax = 26`, `skew = 38`, `damp = 700`, `initialBias = 72`, `initialN = 128`, `delimiter = '-'` (all at `punycode.js:7-14`). Plus `baseMinusTMin = base - tMin` (`punycode.js:29`).
   - Define three sentinel errors via `errors.New` matching the message strings at `punycode.js:23-25` verbatim, so callers can match via `errors.Is`: `ErrOverflow` = "Overflow: input needs wider integers to process", `ErrNotBasic` = "Illegal input >= 0x80 (not a basic code point)", `ErrInvalidInput` = "Invalid input". The messages are load-bearing — `tests/tests.js:258, 266, 306` assert them via `RegExp`.
   - Do **not** port `regexPunycode`, `regexNonASCII`, `regexSeparators` yet — those come with their call-sites (items 6-8).

3. **Port `ucs2.decode` and `ucs2.encode`.**
   - Per `specs/punycode-ucs2-decode.md:1-37` and `specs/punycode-ucs2-encode.md:1-42`. Source: `ucs2decode` at `punycode.js:101-123`, `ucs2encode` at `punycode.js:133`. Exposed via `punycode.js:433-436`.
   - **Signature choice (critical)**: Go strings are UTF-8, not UTF-16. The lone-surrogate fixtures at `tests/tests.js:145-174` (seven `ucs2` bucket entries) cannot round-trip through a UTF-8 `string`. Accept `[]uint16` for `UCS2Decode` and emit `[]uint16` from `UCS2Encode`, or wrap in a small helper that converts a JS-style UTF-16 code-unit sequence. The returned code-point slice must preserve lone surrogates `0xD800..0xDFFF` as raw integers (per `specs/punycode-ucs2-decode.md` "Behavior rules" §lines 13-21 and `specs/punycode-ucs2-encode.md` "Behavior rules" §lines 12-29).
   - Implement the high-surrogate lookahead + rewind at `punycode.js:107-117`: combine formula `((high & 0x3FF) << 10) + (low & 0x3FF) + 0x10000`.
   - `UCS2Encode([]int32) []uint16`: for each code point ≤ `0xFFFF` emit one unit; for > `0xFFFF` emit the high/low surrogate pair (inverse of the decode formula). Lone surrogates passed in are emitted verbatim.
   - Acceptance: all seven `ucs2` bucket vectors in `port/punycode_test.go` pass. The "does not mutate argument array" invariant (`tests/tests.js:282-287`) is automatic in Go if inputs are treated as read-only slices — just don't mutate the caller's slice.

4. **Port `basicToDigit`, `digitToBasic`, `adapt` (Bootstring primitives).**
   - Per `specs/src-punycode.md:110-150` (§5 "Bootstring primitives"). Source: `basicToDigit` at `punycode.js:144-155`, `digitToBasic` at `punycode.js:168-172`, `adapt` at `punycode.js:179-187`.
   - `basicToDigit(cp int32) int32`: three ASCII ranges — `0x30..0x39` → `26..35`, `0x41..0x5A` → `0..25`, `0x61..0x7A` → `0..25`; else return `base` (36) as the "not a digit" sentinel. Case-insensitive (`specs/src-punycode.md:123` — both `'A'` and `'a'` map to digit 0).
   - `digitToBasic(digit int32, flag bool) int32`: port the one-line formula at `punycode.js:171` literally — `digit + 22 + 75*bool2int(digit < 26) - bool2int(flag)<<5`. Every call-site in `encode` passes `flag=false` (`punycode.js:359, 364`), so lowercase-only emission is the only exercised path; still implement the flag for parity.
   - `adapt(delta, numPoints int32, firstTime bool) int32`: steps 1-5 per `specs/src-punycode.md:141-149` and `punycode.js:180-186`. Use `>> 1` for integer division by 2 at `punycode.js:181` (Go `int32 >> 1`). Use Go integer division (`/`) for the three `floor()` sites — both operands positive in practice.
   - No public tests of these primitives, but the six `testData.strings` entries at `tests/tests.js:33-97` (RFC 3492 §7.1 vectors) exercise them transitively once `decode`/`encode` are wired. Add targeted Go unit tests with hand-computed values (e.g. `basicToDigit('a') == 0`, `basicToDigit('9') == 35`, `digitToBasic(0, false) == 'a'`, `adapt(0, 1, true) == 0`) so a regression surfaces at the primitive layer.

5. **Port label-level `decode`.**
   - Per `specs/punycode-decode.md:1-83` (all sections) and `specs/src-punycode.md:152-208` (§6). Source: `punycode.js:196-281`.
   - Signature: `Decode(input string) (string, error)`. Input is ASCII; output is a `string` built from decoded code points.
   - Phases (per `specs/punycode-decode.md:32-47` "Algorithm"):
     1. Basic-prefix copy (`punycode.js:208-219`). Find last `-` via a reverse scan. For each byte before it: if `≥ 0x80` return `ErrNotBasic` (`punycode.js:215-216`). Else append as rune.
     2. Main loop (`punycode.js:224-278`). Generalised variable-length integer per `specs/punycode-decode.md:32-47`.
     3. Threshold `t` (`punycode.js:248`): `k<=bias ? tMin : (k>=bias+tMax ? tMax : k-bias)`.
     4. Three overflow guards — guard A at `punycode.js:243-245`, guard B at `punycode.js:255-257`, guard C at `punycode.js:268-270`. All three must return `ErrOverflow` (inventory at `specs/src-punycode.md:311-323`). Use `int32` arithmetic throughout.
     5. Insert `n` at position `i` (`punycode.js:276`) — Go equivalent of JS `output.splice(i, 0, n)` is a `slices.Insert(output, i, n)` or manual slice manipulation.
   - Acceptance: all 25 `strings` bucket rows (`tests/tests.js:7-136`) pass in the decode direction (`tests/tests.js:290-298`); the uppercase-Z test at `tests/tests.js:299-301` decodes `ZZZ` → U+7BA5 (case-insensitivity via `basicToDigit`); the three `RangeError` tests map to the three sentinel errors — `tests/tests.js:258` (`'\x81-'` → `ErrNotBasic`), `tests/tests.js:266` (`'\x81'` → `ErrOverflow`), `tests/tests.js:306` (`'ls8h='` → `ErrInvalidInput`).

6. **Port label-level `encode`.**
   - Per `specs/punycode-encode.md:1-70` (all sections) and `specs/src-punycode.md:209-267` (§7). Source: `punycode.js:290-376`.
   - Signature: `Encode(input string) (string, error)` — error only because overflow guards D/E can fire in principle; no test triggers them (`specs/src-punycode.md:311-323`).
   - Phases (per `specs/punycode-encode.md:24-40` "Algorithm"):
     1. `ucs2decode(input)` into an `[]int32` of code points (`punycode.js:294`). For a Go `string` input, `[]rune(input)` already yields scalar code points — **use `[]rune(input)` when input is a valid UTF-8 Go string**, but the `encode` call-site in `toASCII` always receives Go strings, so this is the right choice.
     2. Basic-code-point copy (`punycode.js:304-309`): emit each `cp < 0x80`.
     3. Emit delimiter if any basic copied (`punycode.js:318-320`).
     4. Main loop (`punycode.js:323-374`): find next `m = min{cp | cp ≥ n}`; advance delta (guard D at `punycode.js:337-339`); for each cp in input order either `++delta` (guard E at `punycode.js:345-347`) or emit cp as generalised variable-length integer (`punycode.js:349-367`) and call `adapt`.
     5. Always pass `flag=false` to `digitToBasic` — the Russian vector `b1abfaaepdrnnbgefbadotcwatmq2g4l` at `tests/tests.js:83-87` (comment explaining the rationale at `tests/tests.js:74-82`) locks lowercase emission into the contract.
   - Acceptance: all 25 `strings` bucket rows pass in the encode direction (`tests/tests.js:312-321`). Confirm mixed-case basic prefix is preserved verbatim (`Bach` → `Bach-` at `tests/tests.js:7-12`; `Pročprostěnemluvíčesky` → `Proprostnemluvesky-uyb24dma41a` at `tests/tests.js:49-53`).

7. **Port `mapDomain` + separator normalisation helper.**
   - Per `specs/src-punycode.md:62-80` (§3.3 `mapDomain`) and `specs/src-punycode.md:22-60` (§2.2 `regexSeparators`). Source: `mapDomain` at `punycode.js:72-86`, separator regex at `punycode.js:19`.
   - Signature: `mapDomain(input string, callback func(string) (string, error)) (string, error)`. Keep unexported; it is an internal helper.
   - Algorithm (`specs/src-punycode.md:73-80`):
     1. `parts := strings.SplitN(input, "@", 2)`. If `len(parts) == 2`, preserve `parts[0] + "@"` as prefix and operate on `parts[1]`; else operate on `input` (`punycode.js:73-80`). Multi-`@` inputs are silently truncated — not tested, matches JS behaviour (`specs/src-punycode.md:325-335` §11.3).
     2. Normalise four separator code points to `.` (U+002E): U+002E, U+3002, U+FF0E, U+FF61. In Go, `strings.Map(func(r rune) rune { ... })` or a `strings.NewReplacer` covering the four codepoints replaces the JS global regex.
     3. `labels := strings.Split(domain, ".")`, apply callback to each (collect errors), then `strings.Join(labels, ".")` and prepend the preserved `@` prefix.
   - No direct public tests, but every `domains` bucket vector (`tests/tests.js:176-220`) and the entire `separators` bucket (`tests/tests.js:221-242`) exercise it transitively through items 8-9.

8. **Port `toUnicode`.**
   - Per `specs/punycode-to-unicode.md:1-49` (all sections) and `specs/src-punycode.md:273-281` (§8.1). Source: `punycode.js:389-395`. Public API wiring at `punycode.js:440`.
   - Signature: `ToUnicode(input string) (string, error)`. Use `mapDomain` from item 7 with a callback that:
     1. Tests `strings.HasPrefix(label, "xn--")` — **case-sensitive** (`punycode.js:17, 391`; `specs/src-punycode.md:325-335` §11.1 flags that uppercase `XN--` deliberately does not match). Do **not** use `strings.EqualFold`.
     2. If matched: `Decode(strings.ToLower(label[4:]))`. The `.toLowerCase()` at `punycode.js:392` handles the mixed-case fixture `xn--ZCKzah` at `tests/tests.js:140-149` (ucs2 bucket) — verify this stays working.
     3. Otherwise return the label unchanged.
   - Acceptance:
     - All ten `domains` bucket rows (`tests/tests.js:176-220`) round-trip in the toUnicode direction (`tests/tests.js:323-330`).
     - Identity pass-through test at `tests/tests.js:333-343`: for every `strings` bucket entry, `ToUnicode(encoded-without-xn--)` returns input unchanged (checks idempotence over non-`xn--` labels).
     - Trailing-dot test `example.com.` (`tests/tests.js:181-184`) preserves the empty final label.
     - Email test `джумла@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq` at `tests/tests.js:211-215` preserves the Unicode local part and decodes only the domain.

9. **Port `toASCII`.**
   - Per `specs/punycode-to-ascii.md:1-70` (all sections) and `specs/src-punycode.md:283-291` (§8.2). Source: `punycode.js:408-414`. Public API wiring at `punycode.js:439`.
   - Signature: `ToASCII(input string) (string, error)`. Use `mapDomain` from item 7 with a callback that:
     1. Tests whether the label contains any non-ASCII rune (`r > 0x7F`). Per `punycode.js:18` (`regexNonASCII` = `/[^\0-\x7F]/`), U+007F DEL counts as ASCII — the `foo\x7F.example` fixture at `tests/tests.js:216-219` must NOT be encoded.
     2. If any non-ASCII rune found: return `"xn--" + Encode(label)`.
     3. Otherwise return the label unchanged.
   - Acceptance:
     - All ten `domains` bucket rows pass in the toASCII direction (`tests/tests.js:346-354`).
     - Identity pass-through at `tests/tests.js:356-362`: every `encoded` string from the `strings` bucket survives round-trip through `ToASCII`.
     - **All four `separators` bucket rows** (`tests/tests.js:221-242`) normalise to `xn--maana-pta.com` via the test at `tests/tests.js:364-370` — this is the only place the separator normalisation from item 7 is directly asserted.
     - Email test at `tests/tests.js:211-215` preserves the Cyrillic local part `джумла@` verbatim and encodes only the domain.

10. **Public API surface + CI sweep.**
    - Expose the exported API matching the shape at `punycode.js:419-441` and `specs/src-punycode.md:293-309`:
      - `func Decode(string) (string, error)` (`punycode.js:437`)
      - `func Encode(string) (string, error)` (`punycode.js:438`)
      - `func ToASCII(string) (string, error)` (`punycode.js:439`)
      - `func ToUnicode(string) (string, error)` (`punycode.js:440`)
      - `func UCS2Decode([]uint16) []int32` (`punycode.js:434`)
      - `func UCS2Encode([]int32) []uint16` (`punycode.js:435`)
      - Or a nested `UCS2` struct/sub-package that mirrors the JS `punycode.ucs2.*` shape — porter's choice, document either way in `port/doc.go`.
    - `Version` constant is **out of scope** — `punycode.js:425` sets `'2.3.1'` but it is not under test (per `specs/src-punycode.md:293-309` table + `ralph/todo.md` "Out of scope" below).
    - Convert every test assertion to a Go sub-test (`t.Run(description, ...)`) so individual failures name the source `tests/tests.js:<line>` of the vector, making debugging direct.
    - Run `cd port && go test ./...`, `go vet ./...`, and `gofmt -l .` (expect empty output). All four buckets must pass: `strings` (25 rows × 2 directions = 50 assertions), `ucs2` (7 rows × 2 directions = 14), `domains` (10 rows × 2 directions = 20), `separators` (4 rows × 1 direction = 4).
    - Keep the tree dependency-free — no third-party modules; `unicode/utf16`, `strings`, `errors` are the only stdlib packages needed (per `README.md:14-19`).

## Out of scope (per `specs/src-punycode.md:325-335` and `TARGET.md`)

- **IDNA2008 / UTS-46 processing.** Punycode.js only implements Bootstring + IDNA2003 separator normalisation (`punycode.js:19`); the port matches.
- **`version` string.** Not tested (`punycode.js:425`); not a port requirement.
- **Multi-`@` email handling** beyond `parts[1]` (`punycode.js:72-80`) — not tested, not ported behaviour.
- **Uppercase `XN--` recognition** (`punycode.js:17`, `specs/src-punycode.md:325-335` §11.1) — intentionally case-sensitive.
- **`scripts/prepublish.js`** (JS-only build step per `specs/src-scripts-prepublish.md`) — no Go analogue needed.

## Fidelity anchor

> Every RFC test vector from `tests/tests.js` must pass in Go. — `TARGET.md:10`

Specs document *why* a vector exists; `tests/tests.js` is the ground truth for *what* the vector is. If a spec and a test disagree, the test wins.
