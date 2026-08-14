package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// Server encapsulates the HTTP server lifecycle and metadata.
type Server struct {
	httpServer *http.Server
	listener   net.Listener
	Port       int
	FilePath   string
	LogChan    chan LogEvent
	closeOnce  sync.Once
}

// StartServer attempts to listen on preferredPort (or subsequent available ports up to preferredPort+100),
// serving the file at filePath via http.ServeFile with CORS and range request support.
func StartServer(filePath string, preferredPort int, logChan chan LogEvent) (*Server, error) {
	var listener net.Listener
	var err error
	actualPort := preferredPort

	for p := preferredPort; p <= preferredPort+100; p++ {
		l, lErr := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if lErr == nil {
			listener = l
			actualPort = p
			break
		}
		err = lErr
	}

	if listener == nil {
		return nil, fmt.Errorf("failed to bind port starting at %d: %w", preferredPort, err)
	}

	fileName := filepath.Base(filePath)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriterWrapper{ResponseWriter: w}

		// Set CORS and range headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Range, Content-Type, Accept, Origin")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Range, Content-Length, Accept-Ranges, Content-Disposition")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, fileName))

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.ServeFile(ww, r, filePath)

		clientIP := r.RemoteAddr
		if host, _, err := net.SplitHostPort(clientIP); err == nil {
			clientIP = host
		}

		status := ww.statusCode
		if status == 0 {
			status = http.StatusOK
		}

		event := LogEvent{
			Timestamp: start,
			ClientIP:  clientIP,
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    status,
			Bytes:     ww.bytesWritten,
			Duration:  time.Since(start),
		}

		if logChan != nil {
			func() {
				defer func() {
					_ = recover()
				}()
				select {
				case logChan <- event:
				default:
					// Prevent blocking if log channel is full
				}
			}()
		}
	})

	srv := &http.Server{
		Handler:      handler,
		ReadTimeout:  30 * time.Minute, // Allow long range/streaming connections
		WriteTimeout: 0,                // Zero timeout for streaming video/audio
	}

	s := &Server{
		httpServer: srv,
		listener:   listener,
		Port:       actualPort,
		FilePath:   filePath,
		LogChan:    logChan,
	}

	go func() {
		_ = srv.Serve(listener)
	}()

	return s, nil
}

// Stop shuts down the HTTP server gracefully and closes the log channel.
func (s *Server) Stop() error {
	if s == nil {
		return nil
	}
	var err error
	s.closeOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		if s.httpServer != nil {
			err = s.httpServer.Shutdown(ctx)
		}
		if s.LogChan != nil {
			close(s.LogChan)
		}
	})
	return err
}
