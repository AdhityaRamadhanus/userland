package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Gzip Compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

//Gzip return response as gzip encoded
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(res, req)
			return
		}

		res.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(res)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: res}
		next.ServeHTTP(gzw, req)
	})
}
