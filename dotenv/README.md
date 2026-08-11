# dotenv — parses .env style configuration content and loads it into the process environment

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/dotenv.svg)](https://pkg.go.dev/github.com/malcolmston/express/dotenv)

Package dotenv parses .env style configuration content and loads it into the
process environment, mirroring the behavior of the npm "dotenv" library
(motdotla/dotenv). It exposes Parse and Bytes for turning raw .env content
into a map of key/value pairs, and Load and LoadOverride for reading a file
from disk and applying it to the running process via os.Setenv.

The classic use case is keeping configuration out of source code: secrets,
connection strings, feature flags, and per-environment tunables live in a
local .env file that is not committed, and the application reads them at
startup. Parse and Bytes are also handy on their own when you have .env
formatted text from some other source (an embedded asset, a network response,
a test fixture) and want the parsed values without touching the environment.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/dotenv"
```

## Usage

This is the package's own `ExampleBytes`, so it compiles and its output is
asserted on every `go test ./dotenv/`.

```go
content := []byte("# app config\n" +
		"export HOST=localhost\n" +
		"PORT=8080\n" +
		"GREETING=\"hi\\nthere\"\n")

	env, err := dotenv.Bytes(content)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("%q\n", env["HOST"])
	fmt.Printf("%q\n", env["PORT"])
	fmt.Printf("%q\n", env["GREETING"])
```

```
"localhost"
"8080"
"hi\nthere"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Bytes(b []byte) (map[string]string, error)` | Bytes parses .env formatted content from a byte slice. |
| `func Load(path string) error` | Load reads the .env file at path and sets each variable into the process environment. |
| `func LoadOverride(path string) error` | LoadOverride reads the .env file at path and sets each variable into the process environment, overriding any variables that are already set. |
| `func Parse(r io.Reader) (map[string]string, error)` | Parse reads .env formatted content from r and returns the parsed key/value pairs. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/dotenv).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
