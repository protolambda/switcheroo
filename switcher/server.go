package switcher

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
)

type Server struct {
	log log.Logger

	// Address to listen to as server (incl port)
	// Warning: may have a 0 port, if binding to system-chosen port
	addr string

	boundAddr net.Addr

	srv *http.Server

	backend *Backend

	running atomic.Bool
}

func NewServer(log log.Logger, addr string) *Server {
	backend := NewBackend(log)
	return &Server{
		log:  log,
		addr: addr,
		srv: &http.Server{
			Handler: backend,
		},
	}
}

func (s *Server) Start() error {
	if s.running.CompareAndSwap(false, true) {
		return errors.New("server was already started")
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to bind to address %q: %w", s.addr, err)
	}
	s.boundAddr = listener.Addr()

	go func() {
		err := s.srv.Serve(listener)
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return
		}
		s.log.Error("HTTP server failed", "err", err)
	}()
	return nil
}

func (s *Server) Close() error {
	if s.running.CompareAndSwap(true, false) {
		return errors.New("server was not running or already closed")
	}

	var result error
	if err := s.backend.Close(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to close backend: %w", err))
	}

	if err := s.srv.Close(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to close HTTP server: %w", err))
	}
	return result
}

func (s *Server) Address() string {
	return s.boundAddr.String()
}
