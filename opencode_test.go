package main

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func TestFreePort(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	if port <= 0 || port > 65535 {
		t.Fatalf("unexpected port %d", port)
	}
	l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatalf("port %d not free: %v", port, err)
	}
	l.Close()
}

func TestWaitReady(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	defer l.Close()

	if err := waitReady(port, 2*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestWaitReadyTimeout(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	if err := waitReady(port, 200*time.Millisecond); err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
