package main

import (
	"context"
	"fmt"
	"net"
	"path/filepath"

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

func (t *tsnetServer) Listen() (net.Listener, error) {
	if _, err := t.s.Up(context.Background()); err != nil {
		return nil, fmt.Errorf("tsnet up: %w", err)
	}
	ln, err := t.s.ListenTLS("tcp", ":443")
	if err != nil {
		return nil, fmt.Errorf("tsnet listen: %w", err)
	}
	return ln, nil
}

func (t *tsnetServer) Close() error {
	return t.s.Close()
}
