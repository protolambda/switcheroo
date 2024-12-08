package switcher

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
	"github.com/protolambda/switcheroo/ws"
	"github.com/protolambda/websocket"
)

type Remote struct {
	cfg *Target

	log log.Logger

	// TODO RPC connection
}

// TODO figure out how we want to model live / lazy connections to remote endpoints

func (r *Remote) Setup() {
	cl := websocket.NewClient(r.cfg.Endpoint)
	rpc := ws.NewJSONRPC(cl)
	ctx := context.Background()
	inwards := make(chan *Envelope)
	outwards := make(chan *Envelope)

	// TODO fix ws library to make websocket.Client compatible with wider Messenger interface
	setupClientLoops(r.log, nil, rpc, ctx, inwards, outwards)
}
