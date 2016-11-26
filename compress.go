package compress

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/go-http-utils/compressible"
	"github.com/go-http-utils/headers"
	"github.com/go-http-utils/negotiator"
)

// Version is this package's version.
const Version = "0.1.0"

type compressWriter struct {
	rw http.ResponseWriter
	w  io.WriteCloser
}

func (cw compressWriter) Header() http.Header {
	return cw.rw.Header()
}

func (cw compressWriter) WriteHeader(status int) {
	cw.rw.Header().Del(headers.ContentLength)

	cw.rw.WriteHeader(status)
}

func (cw compressWriter) Write(b []byte) (int, error) {
	cw.rw.Header().Del(headers.ContentLength)

	return cw.w.Write(b)
}

// Handler wraps the http.Handler h with compress support.
func Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var w io.WriteCloser

		if req.Method != http.MethodHead &&
			res.Header().Get(headers.ContentEncoding) == "" &&
			compressible.Test(req.Header.Get(headers.ContentType)) {
			n := negotiator.New(req)
			encoding, matched := n.Encoding([]string{"gzip", "deflate"})

			if !matched {
				res.WriteHeader(http.StatusNotAcceptable)
				res.Write([]byte("supported encodings: gzip, deflate"))
				return
			}

			switch encoding {
			case "gzip":
				w = gzip.NewWriter(res)
			case "deflate":
				w, _ = flate.NewWriter(res, flate.DefaultCompression)
			}

			cw := compressWriter{rw: res, w: w}

			cw.Header().Set(headers.ContentEncoding, encoding)
			cw.Header().Set(headers.Vary, headers.AcceptEncoding)

			defer cw.w.Close()

			h.ServeHTTP(cw, req)
			return
		}

		h.ServeHTTP(res, req)
	})
}
