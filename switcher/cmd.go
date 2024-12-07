package switcher

import (
	"context"
	"net"
	"strconv"

	"github.com/protolambda/asklog"
)

type MainCmd struct {
	LogConfig asklog.Config `ask:"."`

	ListenAddr string `ask:"--listen.addr" help:"Address to bind server to"`
	ListenPort uint16 `ask:"--listen.port" help:"Port to bind server to"`

	Root string `ask:"--root" help:"Root directory of config system"`

	srv *Server `ask:"-"`
}

func (m *MainCmd) Default() {
	m.ListenAddr = "127.0.0.1"
	m.ListenPort = 8080
	m.LogConfig.Default()
}

func (m *MainCmd) Run(ctx context.Context, args ...string) error {
	logger := m.LogConfig.New()
	addr := net.JoinHostPort(m.ListenAddr, strconv.FormatUint(uint64(m.ListenPort), 10))

	srv := NewServer(logger, addr)
	m.srv = srv
	return srv.Start()
}

func (m *MainCmd) Close() error {
	if m.srv != nil {
		return m.srv.Close()
	}
	return nil
}
