package switcher

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/protolambda/asklog"
)

type MainCmd struct {
	LogConfig asklog.Config `ask:"."`

	ListenAddr string `ask:"--listen.addr" help:"Address to bind server to"`
	ListenPort uint16 `ask:"--listen.port" help:"Port to bind server to"`

	Config string `ask:"--config" help:"File path to YAML config"`

	srv *Server `ask:"-"`
}

func (m *MainCmd) Default() {
	m.ListenAddr = "127.0.0.1"
	m.ListenPort = 8080
	m.Config = "config.yaml"
	m.LogConfig.Default()
}

func (m *MainCmd) Run(ctx context.Context, args ...string) error {
	logger := m.LogConfig.New()
	addr := net.JoinHostPort(m.ListenAddr, strconv.FormatUint(uint64(m.ListenPort), 10))

	cfg, err := LoadConfig(m.Config)
	if err != nil {
		return fmt.Errorf("failed to load config %q: %w", m.Config, err)
	}

	srv := NewServer(logger, addr, cfg)
	m.srv = srv
	return srv.Start()
}

func (m *MainCmd) Close() error {
	if m.srv != nil {
		return m.srv.Close()
	}
	return nil
}
