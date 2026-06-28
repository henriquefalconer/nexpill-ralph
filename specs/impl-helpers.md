# Generic Private Helpers — `error`, `map`, `mapDomain` (`punycode.js:41-86`)

Three `@private` utilities used throughout the file. None are exported.

## `error(type)` (`punycode.js:41-43`)

- **Signature:** `error(type)` where `type` (String) is a key into the `errors`
  table (`punycode.js:38`).
- **Behavior:** Throws `new RangeError(errors[type])` (`punycode.js:42`) — it
  never returns. The thrown class is always `RangeError`; the message text comes
  from `errors` (`punycode.js:22-26`). Valid keys: `'overflow'`, `'not-basic'`,
  `'invalid-input'`. See [impl-constants.md](impl-constants.md) for the messages.
- **Callers:** `decode` (`punycode.js:216`, `235`, `241`, `244`, `256`, `269`)
  and `encode` (`punycode.js:338`, `346`).

## `map(array, callback)` (`punycode.js:53-60`)

- **Signature:** `map(array, callback)` → a new array of `callback` results
  (`punycode.js:48-51`).
- **Behavior:** Allocates `result = []` (`punycode.js:54`), caches
  `length = array.length` (`punycode.js:55`), then iterates **in reverse** with
  `while (length--)` (`punycode.js:56`), assigning
  `result[length] = callback(array[length])` (`punycode.js:57`). Because the
  write is indexed by the same decremented `length`, output order is preserved
  despite the reverse traversal. Returns `result` (`punycode.js:59`).
- **Caller:** `mapDomain` (`punycode.js:84`).

## `mapDomain(domain, callback)` (`punycode.js:72-86`)

A label-wise wrapper that applies `callback` to each domain label while leaving
an email local part intact. Returns the reassembled string (`punycode.js:69-70`).

Steps:

1. **Email split** (`punycode.js:73-80`): `parts = domain.split('@')`. If
   `parts.length > 1` (an email address), set `result = parts[0] + '@'`
   (`punycode.js:78`) and reduce `domain` to `parts[1]` (`punycode.js:79`). The
   inline comment (`punycode.js:76-77`) explains only the domain part is
   punycoded; the local part is left untouched.
2. **Separator normalization** (`punycode.js:82`): replace every match of
   `regexSeparators` (`punycode.js:19`) with ASCII `'\x2E'`. The comment at
   `punycode.js:81` notes `replace` is used instead of `split(regex)` for IE8
   compatibility (issue #17).
3. **Label split & map** (`punycode.js:83-84`): `labels = domain.split('.')`,
   then `map(labels, callback).join('.')`.
4. **Reassemble** (`punycode.js:85`): `return result + encoded` — re-prepending
   the email local part (empty for non-email input).

- **Callers:** `toUnicode` (`punycode.js:390`, callback decodes `xn--` labels)
  and `toASCII` (`punycode.js:409`, callback encodes non-ASCII labels). See
  [impl-domain-api.md](impl-domain-api.md).
