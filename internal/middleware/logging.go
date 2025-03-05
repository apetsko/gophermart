package middleware

import (
	"net/http"
	"time"
)

type logger interface {
	Info(message string, keysAndValues ...interface{})
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		responseData:   &responseData{status: http.StatusOK}, // Значение по умолчанию
	}
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func LoggingMiddleware(logger logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := newLoggingResponseWriter(w)
			next.ServeHTTP(lw, r)

			logger.Info(
				r.Method,
				"uri", r.RequestURI,
				"UserID", r.Header.Get("UserID"),
				"duration_ms", time.Since(start).Milliseconds(),
				"size", lw.responseData.size,
				"status", lw.responseData.status,
			)
		})
	}
}
