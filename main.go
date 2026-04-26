package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var Version = "dev"

func main() {
	authKey := flag.String("authkey", os.Getenv("TS_AUTHKEY"), "Tailscale auth key for headless registration")
	hostname := flag.String("hostname", envOr("TSOPENCODE_HOSTNAME", "opencode"), "Tailscale node name")
	stateDir := flag.String("state-dir", envOr("TSOPENCODE_STATE_DIR", defaultStateDir()), "base dir for tsnet state")
	opencodeBin := flag.String("opencode-bin", "opencode", "path to opencode binary")
	flag.Parse()

	port, err := freePort()
	if err != nil {
		log.Fatalf("find free port: %v", err)
	}

	cmd, err := spawnOpencode(*opencodeBin, port)
	if err != nil {
		log.Fatalf("spawn opencode: %v", err)
	}
	defer cmd.Process.Kill()

	if err := waitReady(port, 15*time.Second); err != nil {
		log.Fatalf("opencode not ready: %v", err)
	}

	ts, err := newTSNet(*hostname, *stateDir, *authKey)
	if err != nil {
		log.Fatalf("tsnet init: %v", err)
	}
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ln, err := ts.Listen(ctx)
	if err != nil {
		log.Fatalf("tsnet listen: %v", err)
	}

	srv := &http.Server{Handler: newProxy(port)}
	fmt.Printf("tsopencode %s — https://%s.<tailnet>.ts.net\n", Version, *hostname)

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("serve error: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func defaultStateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".tsopencode"
	}
	return filepath.Join(home, ".config", "tsopencode")
}
