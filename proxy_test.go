package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsWebSocket(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if isWebSocket(req) {
		t.Fatal("plain request should not be websocket")
	}
	req.Header.Set("Upgrade", "websocket")
	if !isWebSocket(req) {
		t.Fatal("request with Upgrade: websocket should be websocket")
	}
	req.Header.Set("Upgrade", "WebSocket")
	if !isWebSocket(req) {
		t.Fatal("Upgrade header check should be case-insensitive")
	}
}

func TestProxyHTTP(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "hello from backend")
	}))
	defer backend.Close()

	port := backend.Listener.Addr().(*net.TCPAddr).Port
	proxy := newProxy(port)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello from backend" {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}
