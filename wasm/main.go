//go:build js && wasm

// Command express (wasm) exposes the portable, browser-safe UTILITY subpackages
// of the express module to JavaScript. Built with GOOS=js GOARCH=wasm it
// registers a `__mgo_express` object on the JS global whose methods call the
// very same Go implementations — duration/byte parsing, id generation, slugs
// and string validation — so identical code runs in Go and in the browser or
// Node. The HTTP server (Application/Router/Request/Response) is intentionally
// NOT exposed: it cannot run in a browser. See express.mjs for the idiomatic
// JS wrapper.
package main

import (
	"syscall/js"
	"time"

	"github.com/malcolmston/express/bytes"
	"github.com/malcolmston/express/ms"
	"github.com/malcolmston/express/nanoid"
	"github.com/malcolmston/express/pluralize"
	"github.com/malcolmston/express/slugify"
	"github.com/malcolmston/express/titlecase"
	"github.com/malcolmston/express/uuid"
	"github.com/malcolmston/express/validatorjs"
)

func main() {
	obj := js.Global().Get("Object").New()

	// ms — parse a human duration string ("2h") to milliseconds; format the
	// reverse. Parse errors return JS null.
	obj.Set("ms", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return nil
		}
		d, err := ms.Parse(a[0].String())
		if err != nil {
			return nil
		}
		return float64(d.Milliseconds())
	}))
	obj.Set("msFormat", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return ""
		}
		return ms.Format(time.Duration(a[0].Int()) * time.Millisecond)
	}))

	// bytes — parse a human size string ("1KB") to a byte count; format the
	// reverse. Parse errors return JS null.
	obj.Set("bytes", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return nil
		}
		n, err := bytes.Parse(a[0].String())
		if err != nil {
			return nil
		}
		return float64(n)
	}))
	obj.Set("bytesFormat", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return ""
		}
		return bytes.Format(int64(a[0].Int()))
	}))

	// uuid — random v4 id, and RFC-4122 validation.
	obj.Set("uuidV4", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		id, err := uuid.V4()
		if err != nil {
			return ""
		}
		return id
	}))
	obj.Set("uuidValidate", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return false
		}
		return uuid.Validate(a[0].String())
	}))

	// nanoid — url-safe random id.
	obj.Set("nanoid", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		id, err := nanoid.New()
		if err != nil {
			return ""
		}
		return id
	}))

	// slugify(text, lower?) — url-safe slug. lower defaults to true on the JS
	// side; it is honored here when provided.
	obj.Set("slugify", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return ""
		}
		lower := true
		if len(a) > 1 && !a[1].IsUndefined() && !a[1].IsNull() {
			lower = a[1].Bool()
		}
		return slugify.Slugify(a[0].String(), slugify.Options{Lower: lower, Trim: true})
	}))

	// validatorjs — pure predicate checks.
	obj.Set("isEmail", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return false
		}
		return validatorjs.IsEmail(a[0].String())
	}))
	obj.Set("isURL", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return false
		}
		return validatorjs.IsURL(a[0].String())
	}))

	// pluralize — English plural of a word.
	obj.Set("plural", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return ""
		}
		return pluralize.Plural(a[0].String())
	}))

	// titlecase — Title Case a string.
	obj.Set("titleCase", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 0 {
			return ""
		}
		return titlecase.TitleCase(a[0].String())
	}))

	js.Global().Set("__mgo_express", obj)

	select {} // keep the Go runtime alive so the exported funcs stay callable
}
