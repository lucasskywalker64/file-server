package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStartServer_AndRangeRequest(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "sample.txt")
	testContent := "Hello, World! This is a test file for HTTP range requests in file-server TUI."
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	logChan := make(chan LogEvent, 10)
	srv, err := StartServer(testFile, 18080, logChan)
	if err != nil {
		t.Fatalf("StartServer failed: %v", err)
	}
	defer srv.Stop()

	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d/", srv.Port)

	// 1. Full GET Request
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}

	// Verify CORS and Content-Disposition headers
	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got '%s'", origin)
	}
	if cd := resp.Header.Get("Content-Disposition"); cd != `inline; filename="sample.txt"` {
		t.Errorf("expected Content-Disposition 'inline; filename=\"sample.txt\"', got '%s'", cd)
	}
	if ar := resp.Header.Get("Accept-Ranges"); ar != "bytes" {
		t.Errorf("expected Accept-Ranges 'bytes', got '%s'", ar)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != testContent {
		t.Errorf("expected body '%s', got '%s'", testContent, string(body))
	}

	// 2. HTTP Range Request (Partial Content)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("failed to create Range request: %v", err)
	}
	req.Header.Set("Range", "bytes=0-4") // Should return "Hello"

	respRange, err := client.Do(req)
	if err != nil {
		t.Fatalf("Range request failed: %v", err)
	}
	defer respRange.Body.Close()

	if respRange.StatusCode != http.StatusPartialContent {
		t.Errorf("expected status 206 Partial Content, got %d", respRange.StatusCode)
	}

	rangeBody, err := io.ReadAll(respRange.Body)
	if err != nil {
		t.Fatalf("failed to read range response body: %v", err)
	}
	if string(rangeBody) != "Hello" {
		t.Errorf("expected range body 'Hello', got '%s'", string(rangeBody))
	}

	// 3. HTTP OPTIONS Preflight Request
	reqOpt, err := http.NewRequest("OPTIONS", url, nil)
	if err != nil {
		t.Fatalf("failed to create OPTIONS request: %v", err)
	}
	respOpt, err := client.Do(reqOpt)
	if err != nil {
		t.Fatalf("OPTIONS request failed: %v", err)
	}
	defer respOpt.Body.Close()
	if respOpt.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204 No Content for OPTIONS, got %d", respOpt.StatusCode)
	}

	// Check log channel receives events
	select {
	case evt := <-logChan:
		if evt.Status != http.StatusOK && evt.Status != http.StatusPartialContent && evt.Status != http.StatusNoContent {
			t.Errorf("unexpected log status: %d", evt.Status)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timed out waiting for log event")
	}

	// 4. Test Stop closes log channel cleanly without hanging
	err = srv.Stop()
	if err != nil {
		t.Errorf("srv.Stop() failed: %v", err)
	}

	// Verify channel is closed
	for range logChan {
	}
}

func TestResponseWriterWrapper_Unwrap(t *testing.T) {
	rec := &responseWriterWrapper{}
	if rec.Unwrap() != nil {
		t.Errorf("expected nil underlying writer")
	}
}
