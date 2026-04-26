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
	// detect subcommands before flag parsing
	var subcommand string
	if len(os.Args) > 1 && (os.Args[1] == "install" || os.Args[1] == "uninstall") {
		subcommand = os.Args[1]
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	authKey := flag.String("authkey", os.Getenv("TS_AUTHKEY"), "Tailscale auth key for headless registration")
	hostname := flag.String("hostname", envOr("TSOPENCODE_HOSTNAME", "opencode"), "Tailscale node name")
	stateDir := flag.String("state-dir", envOr("TSOPENCODE_STATE_DIR", defaultStateDir()), "base dir for tsnet state")
	opencodeBin := flag.String("opencode-bin", "opencode", "path to opencode binary")
	flag.Parse()

	switch subcommand {
	case "uninstall":
		if err := uninstallService(); err != nil {
			log.Fatalf("uninstall: %v", err)
		}
		return
	case "install":
		if err := installService(*authKey, *hostname, *stateDir, *opencodeBin); err != nil {
			log.Fatalf("install: %v", err)
		}
		return
	}

	port, err := freePort()
	if err != nil {
		log.Fatalf("find free port: %v", err)
	}

	cmd, err := spawnOpencode(*opencodeBin, port)
	if err != nil {
		log.Fatalf("spawn opencode: %v", err)
	}
	defer func() {
		cmd.Process.Signal(syscall.SIGTERM)
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	if err := waitReady(port, 15*time.Second); err != nil {
		log.Fatalf("opencode not ready: %v", err)
	}

	ts, err := newTSNet(*hostname, *stateDir, *authKey)
	if err != nil {
		log.Fatalf("tsnet init: %v", err)
	}
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	ln, dnsName, err := ts.Listen(ctx)
	cancel() // startup context no longer needed
	if err != nil {
		log.Fatalf("tsnet listen: %v", err)
	}

	httpLn, err := ts.ListenHTTP()
	if err != nil {
		log.Fatalf("tsnet listen http: %v", err)
	}

	proxy := newProxy(port)
	srv := &http.Server{Handler: proxy}
	httpSrv := &http.Server{Handler: proxy}
	fmt.Printf("tsopencode %s\n  https://%s\n  http://%s\n", Version, dnsName, dnsName)

	serveErr := make(chan error, 2)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()
	go func() {
		if err := httpSrv.Serve(httpLn); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-sig:
		log.Printf("signal %v, shutting down", s)
	case err := <-serveErr:
		log.Fatalf("serve: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("https shutdown: %v", err)
	}
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
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
