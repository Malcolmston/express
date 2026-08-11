# API deviations

Every package in this repository is a port of an existing library, and the
cross-language parity harness in the aggregator
(`parity/express/`, `parity/express/nested/<pkg>/`) scores each one against its
real upstream. A handful of differences are deliberate: Go's type system, its
error convention, or a bounded integer type makes exact parity either impossible
or a worse API. Those are recorded here, one section per package, and the
corresponding parity case is tagged `"deviation"` so the harness counts it
separately from a bug.

Nothing may be listed here merely because it is inconvenient to fix. If upstream
is right and the port is wrong, the port gets fixed.

## `ms` (npm `ms`, oracle `ms@4.0.0-nightly.202508271359`)

Parity: 178/178 compared cases (100%), 2 deviations.
Harness: `parity/express/nested/ms/`.

| deviation | parity case | why |
| --- | --- | --- |
| `Parse` returns a range error for a duration beyond about ±292 years, where upstream returns an ordinary (unbounded) JavaScript number. | `p-err-overflow-1e30`, `p-err-overflow-years` | The port's return type is `time.Duration`, an `int64` count of nanoseconds. `ms.parse("100000y")` has no representable answer. Go leaves an out-of-range `float64 → int64` conversion implementation-defined, so the alternative to an error is a silently wrapped duration — the worse of the two behaviours. |
| There is no single dual-dispatch entry point equivalent to upstream's default `ms(value, options?)`, which switches on `typeof value`. | `x-ms-string-*`, `x-ms-number-*`, `p-err-ms-empty` | Runtime dispatch on argument type is not a thing a statically typed API should emulate. The two directions are `Parse` and `Format`/`FormatLong`; both arms are covered by cases and behave identically to upstream. Likewise `parseStrict`, which upstream distinguishes from `parse` only by a TypeScript template-literal argument type, maps onto the same `Parse`. |
| The `{ long: true }` format option is a separate function, `FormatLong`, rather than an options argument. | every `fl-*` case | Same reason: one boolean option does not earn an options struct in Go, and the split is checkable at compile time. |

## `bytes` (npm `bytes`, oracle `bytes@3.1.2`)

Parity: 218/218 compared cases (100%), 12 deviations.
Harness: `parity/express/nested/bytes/`.

| deviation | parity case | why |
| --- | --- | --- |
| `Parse` returns an overflow error where upstream returns an oversized JavaScript number: past `int64` (`"1000000000000000000000"`, `"99999999999999999999pb"`) and at the very top of the range, where the decimal literal rounds up to the double 2^63 (`"9223372036854775807"`). | `pb-1e21`, `pb-1e23`, `pb-pb-overflow`, `pb-int64max` | The port's return type is `int64`. Go leaves an out-of-range `float64 → int64` conversion implementation-defined, so the alternative to an error is a silently wrapped byte count — and this value is routinely used as a request-body size limit, where a wrapped number is a security bug, not a rounding artefact. |
| `Parse` has no overload for a number argument; upstream's `bytes.parse(1024)` returns its argument unchanged. | `pn-int`, `pn-zero`, `pn-neg`, `pn-frac` | Go has no function overloading, and a `Parse(string)` that also accepted `any` to implement an identity function would be a worse API. `pn-frac` is doubly unrepresentable: upstream returns the fraction unfloored. |
| `FormatOpts` clamps `DecimalPlaces` into 0..100 instead of failing; `Number.prototype.toFixed` throws a `RangeError` outside that range. | `fo-dp-negative`, `fo-dp-101` | `FormatOpts` returns a bare `string` and has no error channel, and a formatter is not worth panicking from. Clamping is the only remaining option that still produces the documented output for every in-range value. |
| An explicit `decimalPlaces: null` keeps the port's default of 2, where upstream treats it as `toFixed(0)`. | `fo-dp-null` | Upstream checks only `options.decimalPlaces !== undefined`, so JavaScript's `null` slips through and `ToIntegerOrInfinity(null)` is 0. A Go `*int` has exactly one "absent" state; JSON `null` and an omitted key both decode to `nil`. |
| Formatting a count above 2^53 prints the exact `int64`, where upstream can only print the nearest double. | `fo-unit-b-2pow53` | `bytes.format` takes a double, so 2^53+1 has already become 2^53 before the function is entered. The port is strictly more accurate here and there is nothing to reproduce. |

Everything else about this port is upstream's behaviour, quirks included, and
several of those quirks are load-bearing surprises rather than bugs in the port:
`Parse` falls back to JavaScript's `parseInt(value, 10)` for any string the unit
grammar rejects (so `"1zb"`, `"1024 please"` and `" 1kb"` are 1, 1024 and **1
byte** rather than errors), `Format` breaks a rounding tie away from zero like
`toFixed` rather than to even like `strconv.FormatFloat`, and its trailing-zero
trimmer is upstream's regexp, which cannot touch a fractional part that begins
with a zero (`"1.010KB"` stays).

## `qs` (npm `qs`, oracle `qs@6.15.3`)

Parity: 268/268 compared cases (100%), 3 deviations.
Harness: `parity/express/nested/qs/`. Inventory: that harness's `COVERAGE.md`.

| deviation | parity case | why |
| --- | --- | --- |
| `ParseOptions.Depth` cannot express upstream's `depth: 0`, which disables bracket parsing entirely so that `a[b]=c` yields the single literal key `a[b]`. | `o-depth-zero` | The port's int options use 0 to mean "the documented default" and a negative value to mean "no limit", because Go cannot tell an unset field from an explicit zero. Spending a third sentinel — or a `*int` — on a mode whose whole effect is "do not parse the syntax this package exists to parse" is a worse API than pointing at `net/url`, which already does exactly that. |
| An option value upstream rejects with a `TypeError` is treated here as the documented default: an unrecognised `Duplicates` becomes `combine`, and an unrecognised `Format` becomes `FormatRFC3986`. | `d-bad-duplicates`, `t-fmt-bad` | `ParseWith` returns a `map[string]any` and `StringifyWith` returns a `string`; neither has an error channel, and adding one would change both signatures for the sake of a programmer error a Go caller expresses with the package's own exported constants (`qs.DuplicatesLast`, `qs.FormatRFC1738`) rather than a bare string. Note that this is *not* the port accepting hostile input upstream refuses — both of these are misconfigurations by the calling program, never something a request can trigger. |
| Keys are emitted in sorted order at every level of `StringifyWith`, where upstream follows JavaScript object insertion order unless given a `sort` comparator. | every `stringify` case | A Go map has no insertion order to follow, so the only alternatives are "sorted" and "random". The harness closes the gap from the other side, handing upstream a code-unit comparator that reproduces `sort.Strings`. |

Everything else is upstream's behaviour, including the parts that surprise:
the prototype guard drops the *whole* pair for `a[__proto__][b]=c` rather than
just the offending level, and it also drops `toString=x` and
`constructor[prototype][x]=y`; `arrayLimit` is tested with a strict `<`, so at
the default of 20 the index `a[20]` is already an object key; `%5B`/`%5D` are
rewritten to literal brackets across the *whole* query string before the
key/value split point is chosen, so an encoded `]` inside a value can move that
boundary (`a=1%5D=2` has the key `a=1]`); a leading `?` is only special when
`IgnoreQueryPrefix` is set; and a value that fails to percent-decode is returned
with its `+` already turned into a space rather than reported as an error.

## `semver` (npm `semver`, oracle `semver@7.8.5`)

Parity: 313/313 compared cases (100%), 7 deviations.
Harness: `parity/express/nested/semver/`.

| deviation | parity case | why |
| --- | --- | --- |
| `Version.String` — and `Clean`, which is defined in terms of it — keeps build metadata, where upstream's `SemVer#version`/`toString()` and `clean()` drop it (`"1.2.3+build.5"` renders as itself, not as `"1.2.3"`). | `p-build`, `p-build-only-dash`, `p-build-leadzero`, `p-pre-and-build`, `cl-build` | node-semver can afford to drop it because it keeps the original text in `SemVer#raw`; the Go type has no `raw` field, so a `String` that dropped build metadata would make `Parse` → `String` lossy and leave the `Build` slice unrenderable. Build metadata is part of the version per the 2.0.0 spec — it is only excluded from *precedence*, which `Compare` gets right. |
| `Range.String` returns the expression the range was parsed from; upstream's `Range#range` returns the normalised comparator set (`"^1.2.3"` → `">=1.2.3 <2.0.0-0"`, `"*"` → `""`). | `rs-caret`, `rs-star` | The port expands a range into comparators internally but does not model the `-0` prerelease sentinel upstream uses to write those bounds down; it gets the same answers from an explicit prerelease rule instead. Rendering a comparator set it does not literally hold would be a fiction, and echoing the source expression is the more useful `String` for a Go value that is otherwise opaque. `Range.Test` — the part callers act on — matches upstream exactly. |

Not deviations, and fixed rather than excused during the 7.8.5 measurement:
`Parse` no longer accepts the loose-mode-only `=`/`V` prefixes (`Clean` still
strips them, as upstream's `clean()` does), `ParseRange` rejects a range string
that does not fit npm's grammar instead of treating it as a wildcard, `Coerce`
rejects a component with a leading zero, `IncWith` validates its identifier, and
`Sort` preserves the input strings and breaks precedence ties with build
metadata like upstream's `compareBuild`.

## `jsonwebtoken` (npm `jsonwebtoken`, oracle `jsonwebtoken@9.0.3`)

Parity: 120/120 compared cases (100%), 9 deviations.
Harness: `parity/express/nested/jsonwebtoken/`.

| deviation | parity case | why |
| --- | --- | --- |
| `Decode` and `DecodeComplete` return `ErrInvalidToken` for a malformed token, where `jwt.decode` returns `null`. | `dec-two-segments`, `dec-empty-token`, `dec-malformed-base64` | Go's convention for "I could not do that" is an error, not a nil result the caller may forget to check. The distinction matters here precisely because the value is untrusted. |
| A payload that is valid JSON but not an object is an error, where upstream returns the bare value. | `dec-non-json-payload` | `Decode` returns `Claims`, i.e. `map[string]interface{}`. A JWT payload that is a bare string or number is not something a claims-reading caller can use. |
| `Sign` refuses the unsecured `none` algorithm outright; upstream can mint an unsigned token. | `sign-alg-none` | The package implements the HMAC family only. An unsigned token is not a token, and there is no legitimate reason for a Go program to be able to mint one. |
| `Sign` refuses asymmetric key material (a PEM key or certificate) as an HMAC secret; upstream refuses it only on `verify`. | `sign-hs256-with-public-key-pem` | Upstream's asymmetry is the bug: `jwt.sign(payload, publicKeyPem)` happily produces exactly the forgery `jwt.verify` then rejects, which is how the algorithm-confusion attack is built. Refusing it in both directions is strictly safer and has no legitimate use. |
| `SignOptions.Audience` is a `[]string`, so a single entry is written as a JSON string; Node's `audience: ["x"]` — an array of exactly one — has no distinct Go spelling. | `sign-audience-single-element-array` | The common form, `audience: "x"`, is the one worth expressing. Making `Audience` an `any` to reproduce the difference would push the type check to runtime for every caller. |
| Only a positive `ExpiresIn`/`NotBefore` counts as set; upstream honours zero and negative values. | `sign-zero-expires-in`, `sign-negative-expires-in` | `time.Duration` has no "absent" value, so the zero value must mean "option not supplied". A caller who wants an already-expired token can set the `exp` claim directly. |
| A caller-supplied `header` cannot override `alg` or `typ`; upstream's `Object.assign` lets it. | `sign-extra-header-params` (covered, not divergent, because the case supplies neither) | A token whose advertised algorithm disagrees with the MAC that was actually computed is a bug generator. Extra header parameters are merged as upstream merges them. |

Everything else is upstream's behaviour, including the parts that are easy to get
wrong: `exp` is inclusive and `nbf` is not, `clockTolerance` widens both, `maxAge`
is measured from `iat` and fails when `iat` is absent, `expiresIn`/`notBefore` are
measured from the payload's own `iat` when it has one, `noTimestamp` deletes an
`iat` the caller supplied, an option that would overwrite an existing claim is
refused rather than silently winning, and the header and payload are emitted in an
order that reproduces upstream's token bytes exactly.

Three behaviours in this package were **fixed** rather than declared, because
upstream was right and the port was unsafe: an HMAC token is no longer verified
against asymmetric key material (algorithm confusion), an empty secret is refused
by both `Sign` and `Verify`, and a non-nil but empty `VerifyOptions.Algorithms`
(or `Issuer`) now permits nothing instead of everything. See
`parity/express/nested/jsonwebtoken/security.json`.

## `jwtdecode` (npm `jwt-decode`, oracle `jwt-decode@4.0.0`)

Parity: 45/45 compared cases (100%), 4 deviations.
Harness: `parity/express/nested/jwtdecode/`.

| deviation | parity case | why |
| --- | --- | --- |
| A payload that is valid JSON but not an object is an error, where `jwtDecode` returns the bare value. | `reject-json-number-payload`, `reject-json-string-payload`, `reject-json-array-payload`, `reject-json-true-payload` | `Decode` and `DecodeHeader` return `map[string]interface{}`. Returning `any` so that a number could come back would make every caller type-assert. `null` is the one non-object payload both sides agree on, and it is covered by `accept-json-null-payload`. |

`ErrInvalidToken` replaces upstream's `InvalidTokenError`, and one sentinel covers
the four messages upstream distinguishes; per `parity/HARNESS.md` the harness
compares whether the call failed, and the message difference is recorded in that
harness's `COVERAGE.md`. Segment selection and base64 decoding were **fixed** to
match upstream exactly rather than declared: a two-segment string still yields a
payload, a fourth segment is ignored, and both base64 alphabets are accepted.

## `otpauth` (npm `otpauth`, oracle `otpauth@9.5.1`)

Parity: 158/158 compared cases (100%), 6 deviations.
Harness: `parity/express/nested/otpauth/`.

| deviation | parity case | why |
| --- | --- | --- |
| `DecodeSecret` rejects an empty secret; upstream HMACs a zero-length key. | `secret-empty`, `gen-empty-secret` | A one-time password derived from an empty key is the same for everyone. Refusing it turns a silent authentication bypass into an error at enrollment. |
| `DecodeSecret` rejects a base32 string whose length leaves a partial byte; upstream discards the partial bits. | `secret-single-char`, `secret-nine-chars`, `gen-single-char-secret` | A truncated secret means the enrollment QR code was mis-copied. Generating codes from a silently shortened key would make the failure appear later, as an unexplained mismatch. |
| The `SHA3-224/256/384/512` algorithms an otpauth URI may name can be parsed but not computed. | `gen-sha3-algorithm` | SHA-3 is not in the Go standard library and this package has no dependencies. `Parse` still accepts such a URI, as upstream does, so a caller can detect and reject it explicitly. |

The rest of the URI layer was **fixed** to match upstream byte for byte rather
than declared: `URL` escapes components exactly as `encodeURIComponent` does,
always writes the effective `algorithm`, `digits` and `period`/`counter`, and
canonicalises the secret; `Parse` rejects an unknown type, a missing or non-base32
secret, an HOTP URI with no counter, a non-positive `digits`/`period`, an unknown
algorithm and an empty label, fills in the scheme's defaults, and does not treat
`+` in the query as a space. `Config.OmitIssuerFromLabel` mirrors upstream's
`issuerInLabel` with the sense inverted so that the Go zero value is the common
case. Codes longer than ten digits are zero padded as upstream's `padStart` does
rather than clamped.

## `keygrip` (npm `keygrip`, oracle `keygrip@1.1.0`)

Parity: 63/63 compared cases (100%), 2 deviations.
Harness: `parity/express/nested/keygrip/`.

| deviation | parity case | why |
| --- | --- | --- |
| `New(keys)` defaults to HMAC-SHA256; `keygrip(keys)` defaults to HMAC-SHA1. | `sign-default-scheme`, `verify-default-scheme` | Adopting upstream's default would weaken every caller and, worse, silently invalidate every digest already issued by this package. `NewWithOptions(keys, Options{Algorithm: "sha1"})` reproduces upstream's default exactly — case `sign-sha1` proves the digests are identical — so interoperating with a Node service costs one argument. |

`NewWithOptions` was **added** rather than the difference declared, so that the
`algorithm` and `encoding` arguments of the Node constructor are expressible at
all: `sha1`, `sha256`, `sha384`, `sha512` and `md5`, with `base64` (the Node
default), `base64url` and `hex`. An unrecognised algorithm or encoding is an error
on both sides. `New` still panics on an empty key list, as the Node constructor
throws, and still copies the caller's slice — upstream closes over the array, so a
later mutation changes its keys, which is not a behaviour worth reproducing.

## `statuses` (npm `statuses`, oracle `statuses@2.0.1`)

Parity: 141/141 compared cases (100%), 1 deviation.
Harness: `parity/express/nested/statuses/`.

| deviation | parity case | why |
| --- | --- | --- |
| There is no dual-dispatch entry point equivalent to upstream's default `statuses(codeOrMessage)`, and in particular `Code("404")` is a lookup miss where `statuses('404')` returns `"Not Found"`. | `status-numeric-string` | Upstream's callable `parseInt`s a string argument and, if it parses, treats it as a status *code* rather than a reason phrase — a coercion that exists only because one JavaScript function is overloaded on argument type. Go splits the operation into `Message(int)` and `Code(string)`, and `"404"` is not a registered reason phrase: no HTTP status is named "404". Upstream is the outlier. Both arms of the dispatch are covered by cases and agree with upstream on every non-coerced input. |

Everything else in this port is upstream's behaviour, data included: the same 63
codes, the same phrases (`"I'm a Teapot"`, `"Unprocessable Entity"` — both of
which differ between `statuses` major versions), and the same three
classification sets, with 304 deliberately excluded from `IsRedirect` because it
is cache validation rather than a redirect.

One behaviour was **fixed** rather than declared: `Code` used to trim surrounding
whitespace before looking a phrase up, so `Code(" Not Found ")` returned 404
where upstream reports a miss. Accepting input upstream rejects is a divergence
however convenient it reads, so the trim was removed.

## The HTTP content-negotiation cluster

`accepts`, `negotiator`, `type-is`, `mime-types` and `content-type` all measure
**100% of compared cases with no deviations at all** — see
`parity/express/nested/{accepts,negotiator,typeis,mimetypes,contenttype}/`. Each
of those ports had real divergences when first measured, and every one of them
was fixed rather than declared; the per-harness `COVERAGE.md` lists what changed.
Two notes worth carrying here, because they are the kind of thing that looks like
a bug when read in isolation:

- `Accepts.Types`/`Languages`/`Charsets`/`Encodings` return the whole ranked
  `[]string` even when offers are supplied, where the Node package overloads the
  plural name onto the singular function and returns one value. Taking the first
  element reproduces upstream exactly, which is how `Type`, `Language`, `Charset`
  and `Encoding` are defined. This is a superset, not a difference in answer.
- An **absent** `Accept`/`Accept-Language`/`Accept-Charset` header means "anything
  is acceptable" while a header that is *present but empty* accepts **nothing**.
  Both packages now distinguish the two through `http.Header.Values`, because
  upstream distinguishes `undefined` from `''` and collapsing them made the port
  serve content the client had explicitly refused.

## The string / format / id cluster

`slugify`, `pluralize`, `camelcase`, `kebabcase`, `deburr`, `filesize`,
`prettybytes` and `nanoid` are measured by
`parity/express/nested/{slugify,pluralize,camelcase,kebabcase,deburr,filesize,prettybytes,nanoid}/`.
Four of the eight reach **100% of compared cases with no deviations at all**
(`camelcase` 260/260, `kebabcase` 102/102, `deburr` 81/81, `filesize` 188/188);
each had real divergences when first measured and every one of them was fixed
rather than declared. The remaining four carry the deviations below.

### `pluralize` (npm `pluralize`, oracle `pluralize@8.0.0`)

Parity: 450/450 compared cases (100%), 14 deviations, 464 cases total.
Harness: `parity/express/nested/pluralize/`.

Every difference is one thing: **the port was transcribed from
plurals/pluralize `master`, which is ahead of the last published release.**
8.0.0 shipped in 2019 and npm has no publish of master, so the oracle has to be
8.0.0 — an unpinnable oracle makes the score unreproducible. The port's own
vector file `pluralize/pluralize_parity_test.go` is extracted from master's
`test.js`, so "fixing" the port to 8.0.0 would break a suite that is closer to
upstream's intent than the release is.

| deviation | parity case | why |
| --- | --- | --- |
| `canvas`/`canvases` is an irregular pair. | `pl-057-canvas`, `sg-057-canvases` | master adds the pair; 8.0.0 falls through to the rules and yields `canvas` (unchanged) and `canvase`. |
| The `-ies` stem rules are wider: `(…\|ok\|…)ies → $1ie` and a longer `\b(…\|hipp\|…\|calor\|…)ies` alternation. | `sg-103-cookies`, `sg-105-calories`, `sg-106-hippies` | master widened them; 8.0.0 falls through to `ies$ → y` and produces `cooky`, `calory`, `hippy`. |
| `species` is not uncountable, so it singularises to `specie`. | `unc-005-species-sg` | master drops it from the uncountable list; 8.0.0 keeps it. |
| The possessive pronouns `my`/`our`, `his`/`their`, `her`/`their`, `its`/`their` are irregular pairs. | `devi-my-*`, `devi-his-*`, `devi-her-*`, `devi-its-*` | master adds them; 8.0.0 has none, so it applies the regular rules (`my` → `mies`, `his` → `hi`). |

### `slugify` (npm `slugify`, oracle `slugify@1.6.9`)

Parity: 70/70 compared cases (100%), 5 deviations, 75 cases total.
Harness: `parity/express/nested/slugify/`.

| deviation | parity case | why |
| --- | --- | --- |
| `Options.Separator: ""` falls back to `"-"`, so an explicitly empty replacement cannot be expressed. | `opt-replacement-empty` | `Separator` is a plain `string` whose zero value is indistinguishable from "unset", and the documented no-`Options` default is `"-"`. Upstream defaults only when the key is `undefined`, so `slugify("Hello Big World", {replacement: ""})` is `"HelloBigWorld"` there. Making the field a `*string` would break every existing `Options{Separator: "_"}` literal for one rarely used input. |
| The charmap spells `Ę`, `Ī`, `Ķ` and `Ū` in **upper** case. | `charmap-dev-e-ogonek`, `charmap-dev-i-macron`, `charmap-dev-k-cedilla`, `charmap-dev-u-macron` | The port's table is a verbatim copy of simov/slugify `master`'s `charmap.json`, where these four are `"E"`, `"I"`, `"K"`, `"U"`. The pinned 1.6.9 release still carries the lower-case values, which upstream's *own* Polish and Latvian test vectors contradict — those vectors are what `slugify/slugify_parity_test.go` asserts. Upstream's release is the outlier against upstream's own tests. |

Three real divergences were **fixed** rather than declared while measuring this
package: JavaScript's `\s` (and `String.prototype.trim`) match the Unicode space
separators, U+2028/U+2029 and U+FEFF, where Go's regexp `\s` is ASCII-only and
`strings.TrimSpace` disagrees at both ends. `slugify.go` now spells the class out
(`jsSpaceChars` / `jsSpaceClass`), so `"a b"`, `"a　b"` and
`"a b c"` slug the way upstream slugs them instead of losing the
separator entirely.

### `prettybytes` (npm `pretty-bytes`, oracle `pretty-bytes@7.1.1`)

Parity: 187/187 compared cases (100%), 3 deviations, 190 cases total.
Harness: `parity/express/nested/prettybytes/`.

| deviation | parity case | why |
| --- | --- | --- |
| A non-finite input is formatted rather than rejected: `PrettyBytes(math.NaN())` returns a string where `prettyBytes(NaN)` throws `TypeError`. | `err-nan`, `err-infinity`, `err-neg-infinity` | `PrettyBytes(float64) string` has no error return. The alternatives are panicking from a formatting helper, which is not idiomatic, or widening the signature of a shipped API for an input a caller can check with `math.IsNaN` in one line. |

Two real divergences were **fixed** here: the port always used fixed notation
where upstream concatenates a JavaScript `Number`, which switches to exponential
form outside `[1e-6, 1e21)` — `1e-9 B` rather than `0.000000001 B`, and `1e+76 YB`
rather than a 77-digit integer. `jsNumberString` now reproduces `String(Number)`,
unpadded exponent included, and `formatPrecision` follows upstream's
`minPrecision = max(3, String(Math.floor(n)).length)` rule.

### `nanoid` (npm `nanoid`, oracle `nanoid@6.0.1`)

Parity: 123/123 compared cases (100%), 3 deviations, 126 cases total.
Harness: `parity/express/nested/nanoid/`.

| deviation | parity case | why |
| --- | --- | --- |
| With an alphabet of astral-plane characters (`"🌮🌯🥑🌶"`), `Custom` returns a well-formed id of exactly the requested length; `customAlphabet` returns lone surrogates and a shorter id. | `custom-nonascii-03-len`, `custom-nonascii-03-inalpha`, `custom-nonascii-03-charset` | **Upstream is the outlier.** `nanoid` indexes its alphabet with `alphabet[byte & mask]`, i.e. by UTF-16 code unit, so a four-emoji alphabet has it splicing *halves* of surrogate pairs — 13 characters for `size: 16`, several of them unpaired surrogates. Matching that would mean emitting invalid UTF-8 from a Go string. |

One real divergence was **fixed**: `Custom` used to measure and index the
alphabet in **bytes**, so any multi-byte alphabet both mis-applied the documented
1..256 limit and spliced individual UTF-8 bytes into the id. It now works in code
points, which is what upstream's `.length` means for every alphabet inside the
BMP. Two behaviours are deliberately *stricter* than upstream rather than
declared as deviations, because upstream's behaviour is a hang rather than an
answer: an empty alphabet and an alphabet longer than 256 characters both return
an error here, while `customAlphabet` loops forever.

Two upstream surfaces are **untested** rather than matched, and the reason is
recorded in `parity/express/nested/nanoid/COVERAGE.md`: `customRandom` and
`random` give upstream an injectable byte source, so
`customRandom("abcdef", 10, n => new Uint8Array(n).fill(3))()` is deterministically
`"dddddddddd"`. `nanoid.Custom` always reads `crypto/rand` and accepts no
`io.Reader`, so there is no way to feed both implementations the same bytes and
compare the resulting id. Until the port grows such a seam, the strongest
cross-language claim available is the property set the harness does check —
length, alphabet membership, reachability of every alphabet character,
distinctness, and rejection of an invalid size.
