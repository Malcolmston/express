package express

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SendFile sends the file at path to the client. It uses http.ServeContent
// under the hood, so it supports Range requests, conditional GET
// (If-Modified-Since / If-None-Match), and content-type detection by extension.
// Returns an error if the file cannot be opened.
func (res *Response) SendFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		res.Status(http.StatusNotFound)
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		res.finalError(err)
		return err
	}
	if info.IsDir() {
		res.Status(http.StatusNotFound)
		return os.ErrInvalid
	}

	// Content-Type by extension unless already set.
	if res.GetHeader("Content-Type") == "" {
		if ct := mimeByExt(path); ct != "" {
			res.Set("Content-Type", ct)
		}
	}
	res.written = true
	http.ServeContent(res.Writer, res.req.Raw, filepath.Base(path), info.ModTime(), f)
	return nil
}

// Download sends the file at path as an attachment, prompting the browser to
// save it. When filename is empty the base name of path is used.
func (res *Response) Download(path string, filename ...string) error {
	name := filepath.Base(path)
	if len(filename) > 0 && filename[0] != "" {
		name = filename[0]
	}
	res.Attachment(name)
	return res.SendFile(path)
}

// Attachment sets the Content-Disposition header to attachment. With a filename
// it also advertises the download name and sets a matching Content-Type.
func (res *Response) Attachment(filename ...string) *Response {
	if len(filename) > 0 && filename[0] != "" {
		name := filepath.Base(filename[0])
		res.Set("Content-Disposition", "attachment; filename=\""+escapeQuotes(name)+"\"")
		if res.GetHeader("Content-Type") == "" {
			if ct := mimeByExt(name); ct != "" {
				res.Set("Content-Type", ct)
			}
		}
	} else {
		res.Set("Content-Disposition", "attachment")
	}
	return res
}

// mimeByExt returns a Content-Type for a file path's extension, or "".
func mimeByExt(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".xml":
		return "application/xml; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".csv":
		return "text/csv; charset=utf-8"
	case ".wasm":
		return "application/wasm"
	default:
		return "application/octet-stream"
	}
}

func escapeQuotes(s string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s)
}
