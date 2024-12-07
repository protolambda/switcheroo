package switcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	jsonrpc "github.com/protolambda/jsonrpc2"
	"github.com/protolambda/switcheroo/ws"
	"github.com/protolambda/websocket"
)

type Envelope struct {
	Ctx context.Context
	Msg jsonrpc.Message
}

func (en *Envelope) JSON() string {
	out, err := json.Marshal(en.Msg)
	if err != nil {
		return fmt.Sprintf("invalid: %v", err)
	}
	return string(out)
}

type User struct {
	name string

	Conn *websocket.Connection
	Meta *websocket.ConnectionMetadata
	RPC  ws.JSONRPCConnection

	log log.Logger

	inwards  chan *Envelope
	outwards chan *Envelope
}

func NewUser(log log.Logger, conn *websocket.Connection, meta *websocket.ConnectionMetadata, name string) *User {
	u := &User{
		Conn:     conn,
		Meta:     meta,
		RPC:      ws.NewJSONRPC(conn),
		log:      log,
		inwards:  make(chan *Envelope, 100),
		outwards: make(chan *Envelope, 100),
	}
	setupClientLoops(u)
	return u
}

// Close stops a user by closing its underlying connection.
// The connection may be closed externally; the user will shutdown itself accordingly.
func (u *User) Close() error {
	return u.Conn.Close()
}

func setupClientLoops(c *User) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.Conn.CloseWithCause(fmt.Errorf("writer panic: %v", err))
			} else {
				c.Conn.CloseWithCause(context.Canceled)
			}
			c.log.Info("Closed user write-loop")
		}()
		c.log.Info("Opened user write-loop")
		for {
			select {
			case <-c.Conn.CloseCtx().Done():
				return
			case envelope, ok := <-c.outwards:
				if !ok {
					return
				}
				c.log.Info("writing message to user", "user", c.Meta.RemoteAddr, "msg", envelope.JSON())
				if err := c.RPC.Write(&envelope.Msg); err != nil {
					if c.Conn.Err() != nil {
						c.log.Warn("cannot write to broken connection",
							"err", err, "connectionErr", c.Conn.Err())
						return
					}
					c.log.Error("failed to write message", "err", err)
					continue
				}
			}
		}
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.Conn.CloseWithCause(fmt.Errorf("reader panic: %v", err))
			} else {
				c.Conn.CloseWithCause(context.Canceled)
			}
			c.log.Info("Closed user read-loop")
		}()
		c.log.Info("Opened user read-loop")
		for {
			var dest jsonrpc.Message
			if err := c.RPC.Read(&dest); err != nil {
				if c.Conn.Err() != nil {
					// connection issue / close
					return
				}
				c.log.Error("failed to decode message", "err", err)
				continue
			}
			e := &Envelope{
				Ctx: c.Meta.Context,
				Msg: dest,
			}
			select {
			case <-c.Conn.CloseCtx().Done():
				return
			case c.inwards <- e:
			}
			c.log.Info("reading message from user", "user", c.Meta.RemoteAddr, "msg", e.JSON())
		}
	}()
}
