package server

import (
	"io"
	"net/http"
	"time"
)

// LogEvent contains detailed metadata about an HTTP request served.
type LogEvent struct {
	Timestamp time.Time
	ClientIP  string
	Method    string
	Path      string
	Status    int
	Bytes     int64
	Duration  time.Duration
}

// responseWriterWrapper wraps http.ResponseWriter to capture HTTP status code and response size.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

func (w *responseWriterWrapper) ReadFrom(r io.Reader) (n int64, err error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
		n, err = rf.ReadFrom(r)
		w.bytesWritten += n
		return n, err
	}
	n, err = io.Copy(w.ResponseWriter, r)
	w.bytesWritten += n
	return n, err
}

func (w *responseWriterWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap returns the underlying ResponseWriter for http.ResponseController support.
func (w *responseWriterWrapper) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
