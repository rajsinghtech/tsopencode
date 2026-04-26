package main

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"tailscale.com/tsnet"
)

type tsnetServer struct {
	s *tsnet.Server
}

func newTSNet(hostname, stateDir, authKey string) (*tsnetServer, error) {
	s := &tsnet.Server{
		Hostname:  hostname,
		Dir:       filepath.Join(stateDir, "tsnet-state"),
		AuthKey:   authKey,
		Ephemeral: false,
	}
	return &tsnetServer{s: s}, nil
}

func (t *tsnetServer) Listen(ctx context.Context) (net.Listener, string, error) {
	status, err := t.s.Up(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("tsnet up: %w", err)
	}
	dnsName := strings.TrimSuffix(status.Self.DNSName, ".")
	ln, err := t.s.ListenTLS("tcp", ":443")
	if err != nil {
		return nil, "", fmt.Errorf("tsnet listen: %w", err)
	}
	return ln, dnsName, nil
}

func (t *tsnetServer) ListenHTTP() (net.Listener, error) {
	ln, err := t.s.Listen("tcp", ":80")
	if err != nil {
		return nil, fmt.Errorf("tsnet listen http: %w", err)
	}
	return ln, nil
}

func (t *tsnetServer) Close() error {
	return t.s.Close()
}
