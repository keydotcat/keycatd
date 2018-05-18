package api

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func logHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := newLoggingResponseWriter(w)
		start = time.Now()
		handler.ServeHTTP(lw, r)
		log.Printf("- [%s] %s %s (%d) %.3fs\n", r.RemoteAddr, r.Method, r.RequestURI, lw.statusCode, float64(time.Now().Sub(start)*time.Second))
	})
}
