package express

import (
	"io"
	"mime/multipart"
	"net/url"
)

// defaultMaxMemory is the amount of a multipart body kept in memory before
// spilling to temp files, matching net/http's default (32 MiB).
const defaultMaxMemory = 32 << 20

// Multipart returns middleware that parses multipart/form-data request bodies,
// making fields available via req.FormValue and files via req.FormFile. maxMemory
// bounds the in-memory buffer (0 uses the 32 MiB default).
func Multipart(maxMemory int64) Handler {
	if maxMemory <= 0 {
		maxMemory = defaultMaxMemory
	}
	return func(req *Request, res *Response, next Next) {
		ct := req.Get("Content-Type")
		if len(ct) < 19 || ct[:19] != "multipart/form-data" {
			next()
			return
		}
		if err := req.Raw.ParseMultipartForm(maxMemory); err != nil {
			next(err)
			return
		}
		next()
	}
}

// Text returns middleware that reads a text/plain body into req.Body() as a
// string.
func Text() Handler {
	return func(req *Request, res *Response, next Next) {
		if !req.Is("text") {
			next()
			return
		}
		data, err := io.ReadAll(req.Raw.Body)
		if err != nil {
			next(err)
			return
		}
		req.Raw.Body.Close()
		req.SetBody(string(data))
		next()
	}
}

// FormFile returns the first uploaded file for the given form field, along with
// its multipart header. The multipart form must have been parsed first (via the
// Multipart middleware or by calling FormFile, which parses on demand).
func (req *Request) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	if req.Raw.MultipartForm == nil {
		if err := req.Raw.ParseMultipartForm(defaultMaxMemory); err != nil {
			return nil, nil, err
		}
	}
	return req.Raw.FormFile(name)
}

// Files returns all uploaded file headers for a form field.
func (req *Request) Files(name string) []*multipart.FileHeader {
	if req.Raw.MultipartForm == nil {
		if err := req.Raw.ParseMultipartForm(defaultMaxMemory); err != nil {
			return nil
		}
	}
	if req.Raw.MultipartForm == nil || req.Raw.MultipartForm.File == nil {
		return nil
	}
	return req.Raw.MultipartForm.File[name]
}

// Form parses and returns all form values (query + body) as url.Values.
func (req *Request) Form() url.Values {
	_ = req.Raw.ParseForm()
	return req.Raw.Form
}
