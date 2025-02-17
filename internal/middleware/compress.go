package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/apetsko/gophermart/internal/logging"
)

type bufferedResponseWriter struct {
	http.ResponseWriter
	buffer     *bytes.Buffer
	statusCode int
}

func (w *bufferedResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.buffer.Write(b)
}

type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{ReadCloser: r, zr: zr}, nil
}

func (cr *compressReader) Read(p []byte) (int, error) {
	return cr.zr.Read(p)
}

func (cr *compressReader) Close() error {
	return cr.zr.Close()
}

func GzipMiddleware(logger *logging.ZapLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		gzipWithLogger := func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					http.Error(w, "Failed to decompress request body", http.StatusInternalServerError)
					return
				}
				defer cr.Close()
				r.Body = cr
			}

			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			buf := new(bytes.Buffer)
			bufferedWriter := &bufferedResponseWriter{
				ResponseWriter: w,
				buffer:         buf,
			}

			next.ServeHTTP(bufferedWriter, r)

			contentType := bufferedWriter.ResponseWriter.Header().Get("Content-Type")
			if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
				bufferedWriter.ResponseWriter.Header().Add("Content-Encoding", "gzip")
				bufferedWriter.ResponseWriter.WriteHeader(bufferedWriter.statusCode)

				gz := gzip.NewWriter(bufferedWriter.ResponseWriter)
				defer gz.Close()

				if _, err := gz.Write(bufferedWriter.buffer.Bytes()); err != nil {
					logger.Error("Failed to compress response: " + err.Error())
					http.Error(bufferedWriter.ResponseWriter, "Failed to compress response", http.StatusInternalServerError)
					return
				}
			} else {
				w.WriteHeader(bufferedWriter.statusCode)
				if _, err := w.Write(bufferedWriter.buffer.Bytes()); err != nil {
					logger.Error("Failed to write response: " + err.Error())
				}
			}
		}
		return http.HandlerFunc(gzipWithLogger)
	}
}
