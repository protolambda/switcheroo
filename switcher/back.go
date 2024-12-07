package switcher

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/protolambda/websocket"
)

type Backend struct {
	log log.Logger

	wsSrv *websocket.Server[*User]

	acceptNew atomic.Bool

	mux *http.ServeMux
}

func NewBackend(log log.Logger) *Backend {
	mux := http.NewServeMux()
	backend := &Backend{
		log:   log,
		wsSrv: nil,
		mux:   mux,
	}
	mux.HandleFunc("GET /dial/{name}", backend.handleDial)
	backend.initWebsocketServer()
	backend.acceptNew.Store(true)
	return backend
}

func (ba *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ba.mux.ServeHTTP(w, r)
}

// Close closes all users, and stops the backend from accepting new users.
func (ba *Backend) Close() error {
	ba.acceptNew.Store(false)
	var result error
	ba.wsSrv.Range(func(e *User) bool {
		result = errors.Join(result, e.Close())
		return true
	})
	return result
}

func (ba *Backend) initWebsocketServer() {
	ba.wsSrv = websocket.NewServer[*User](func(c *websocket.Connection, meta *websocket.ConnectionMetadata) (*User, error) {
		if !ba.acceptNew.Load() {
			return nil, errors.New("not accepting new users")
		}
		name := GetName(meta.Context)
		if name == "" {
			return nil, fmt.Errorf("cannot upgrade user %s without name", meta.RemoteAddr)
		}

		ba.log.Info("new WS JSON RPC connection to provider",
			"remote", meta.RemoteAddr, "origin", meta.Origin, "name", name)

		out := NewUser(ba.log.With("name", name), c, meta, name)

		return out, nil
	}, websocket.WithOnDisconnect(func(e *User) {
		_ = e.Close()
		ba.log.Info("disconnected websocket",
			"remote", e.Meta.RemoteAddr, "origin", e.Meta.Origin)
	}), websocket.WithCheckOrigin[*User](func(r *http.Request) bool {
		return true
	}), websocket.WithOnUpgradeFailed[*User](func(r *http.Request, err error) {
		ba.log.Warn("failed to upgrade websocket",
			"origin", r.Header.Get("Origin"),
			"remote", r.RemoteAddr, "err", err)
	}))
}

func (ba *Backend) handleDial(w http.ResponseWriter, r *http.Request) {
	addCorsHeader(w)
	ctx := r.Context()
	name := r.PathValue("name")
	// Attach name to context,
	// so the websocket handler knows how this upgrade-request came to be.
	r = r.WithContext(WithName(ctx, name))

	ba.log.Info("Upgrading to websocket", "name", name)
	ba.wsSrv.Handle(w, r)
	ba.log.Info("websocket stopped", "name", name)
}
