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

	cfg *Source

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
	setupClientLoops(u.log.New("user", u.Meta.RemoteAddr), u.Conn, u.RPC, u.Meta.Context, u.inwards, u.outwards)
	return u
}

// Close stops a user by closing its underlying connection.
// The connection may be closed externally; the user will shutdown itself accordingly.
func (u *User) Close() error {
	return u.Conn.Close()
}

type Messenger interface {
	websocket.Messenger
	CloseWithCause(cause error)
	CloseCtx() context.Context
}

func setupClientLoops(log log.Logger, conn Messenger, rpc ws.JSONRPCConnection,
	msgCtx context.Context, inwards, outwards chan *Envelope) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				conn.CloseWithCause(fmt.Errorf("writer panic: %v", err))
			} else {
				conn.CloseWithCause(context.Canceled)
			}
			log.Info("Closed write-loop")
		}()
		log.Info("Opened write-loop")
		for {
			select {
			case <-conn.CloseCtx().Done():
				return
			case envelope, ok := <-outwards:
				if !ok {
					return
				}
				log.Info("writing message", "msg", envelope.JSON())
				if err := rpc.Write(&envelope.Msg); err != nil {
					if conn.Err() != nil {
						log.Warn("cannot write to broken connection",
							"err", err, "connectionErr", conn.Err())
						return
					}
					log.Error("failed to write message", "err", err)
					continue
				}
			}
		}
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				conn.CloseWithCause(fmt.Errorf("reader panic: %v", err))
			} else {
				conn.CloseWithCause(context.Canceled)
			}
			log.Info("Closed read-loop")
		}()
		log.Info("Opened read-loop")
		for {
			var dest jsonrpc.Message
			if err := rpc.Read(&dest); err != nil {
				if conn.Err() != nil {
					// connection issue / close
					return
				}
				log.Error("failed to decode message", "err", err)
				continue
			}
			e := &Envelope{
				Ctx: msgCtx,
				Msg: dest,
			}
			select {
			case <-conn.CloseCtx().Done():
				return
			case inwards <- e:
			}
			log.Info("reading message", "msg", e.JSON())
		}
	}()
}
