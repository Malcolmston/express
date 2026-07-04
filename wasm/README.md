# express utilities — JavaScript adapter (WebAssembly)

Run the **same Go implementations** of express's portable utility subpackages
from JavaScript — in the browser or Node — via WebAssembly. No reimplementation:
`main.go` exposes the pure, browser-safe utilities to JS and `express.mjs` wraps
them in an idiomatic API.

Only the **utility** subpackages are exposed. The HTTP server
(`Application`/`Router`/`Request`/`Response`) is intentionally left out — it
cannot run in a browser.

## Build

```sh
./build.sh          # produces express.wasm (+ copies the Go wasm_exec.js runtime)
```

## Use (Node or browser)

```js
import { loadExpress } from './express.mjs';
const ex = await loadExpress();

ex.ms('2h');                    // 7200000        (ms → parse duration to ms)
ex.msFormat(7200000);           // '2h'
ex.bytes('1KB');                // 1024           (binary units, 1KB = 1024)
ex.bytesFormat(1536);           // '1.5KB'
ex.uuidV4();                    // random v4 uuid string
ex.uuidValidate(id);            // true / false
ex.nanoid();                    // 21-char url-safe id
ex.slugify('Hello World');      // 'hello-world'  (lowercases by default)
ex.slugify('Hello World', false)// 'Hello-World'
ex.isEmail('a@b.com');          // true
ex.isURL('https://x.com');      // true
ex.plural('house');             // 'houses'
ex.titleCase('hello world');    // 'Hello World'
```

## Exposed functions

| JS method                 | Go source                            |
| ------------------------- | ------------------------------------ |
| `ms(str)`                 | `ms.Parse` → milliseconds            |
| `msFormat(millis)`        | `ms.Format`                          |
| `bytes(str)`              | `bytes.Parse`                        |
| `bytesFormat(n)`          | `bytes.Format`                       |
| `uuidV4()`                | `uuid.V4`                            |
| `uuidValidate(s)`         | `uuid.Validate`                      |
| `nanoid()`                | `nanoid.New`                         |
| `slugify(text, lower?)`   | `slugify.Slugify`                    |
| `isEmail(s)`              | `validatorjs.IsEmail`                |
| `isURL(s)`                | `validatorjs.IsURL`                  |
| `plural(word)`            | `pluralize.Plural`                   |
| `titleCase(s)`            | `titlecase.TitleCase`                |

## Verify

```sh
./build.sh && node test.mjs
```

The adapter is compiled with `GOOS=js GOARCH=wasm`; on normal platforms `stub.go`
keeps `go build ./...` and CI green. Build artifacts (`*.wasm`) are gitignored.
