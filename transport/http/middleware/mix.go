package middleware

import (
	"net/http"
)

type interceptingWriter struct {
	http.ResponseWriter
	code int
}

// WriteHeader may not be explicitly called, so care must be taken to
// initialize w.code to its default value of http.StatusOK.
func (w *interceptingWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *interceptingWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseWriter.Write(p)
	return n, err
}

func Mix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		iw := &interceptingWriter{w, http.StatusOK}
		next.ServeHTTP(iw, r)
	})
}
