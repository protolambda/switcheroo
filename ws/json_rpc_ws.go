package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/protolambda/jsonrpc2"
	"github.com/protolambda/websocket"
)

type JSONRPCConnection interface {
	Write(msg *jsonrpc.Message) error
	Read(dest *jsonrpc.Message) error
}

// JSONRPC represents a websocket JSON RPC connection. It can be used for both client-side and server-side
type JSONRPC struct {
	wLock sync.Mutex
	wBuf  bytes.Buffer

	rLock sync.Mutex
	batch []jsonrpc.Message

	ws websocket.Messenger
}

// NewJSONRPC creates a JSON RPC messenger,
// wrapping around a websocket connection.
// Incoming batch-requests are broken apart into sequential reads.
// No special effort is made to collect responses back into a batch (JSON-RPC 2.0 spec says "SHOULD", not "MUST").
func NewJSONRPC(ws websocket.Messenger) *JSONRPC {
	return &JSONRPC{
		ws: ws,
	}
}

func (w *JSONRPC) Close() error {
	return w.ws.Close()
}

func (w *JSONRPC) Err() error {
	return w.ws.Err()
}

// Write to the RPC, safe for concurrent use
func (w *JSONRPC) Write(msg *jsonrpc.Message) error {
	w.wLock.Lock()
	defer w.wLock.Unlock()
	w.wBuf.Reset()
	err := json.NewEncoder(&w.wBuf).Encode(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON RPC message: %w", err)
	}
	if err := w.ws.Write(websocket.TextMessage, w.wBuf.Bytes()); err != nil {
		return fmt.Errorf("failed to write JSON RPC message: %w", err)
	}
	return nil
}

// Read from the RPC, safe for concurrent use
func (w *JSONRPC) Read(dest *jsonrpc.Message) error {
	w.rLock.Lock()
	defer w.rLock.Unlock()

	for {
		// dequeue batch-element, if any is left
		if len(w.batch) > 0 {
			*dest = w.batch[0]
			w.batch = w.batch[1:]
			return nil
		}

		typ, data, err := w.ws.Read()
		if err != nil {
			return err
		}
		if typ != websocket.TextMessage {
			return fmt.Errorf("unexpected message: %s", typ)
		}
		var x []jsonrpc.Message
		if err := json.Unmarshal(data, &x); err != nil {
			// not a batch
			if err := json.Unmarshal(data, dest); err != nil {
				return fmt.Errorf("failed to decode JSON RPC message: %w", err)
			}
			return nil
		} else {
			// a batch
			w.batch = x
		}
	}
}

var _ JSONRPCConnection = (*JSONRPC)(nil)
