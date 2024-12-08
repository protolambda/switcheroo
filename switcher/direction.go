package switcher

import "fmt"

// Direction to match the effect on.
type Direction uint32

const (
	DirectionSourceRequest  = 0b0001                                            // requests from source to target (i.e. regular usage, before server handles the request)
	DirectionSourceResponse = 0b0010                                            // responses by source to target (i.e. bidirectional RPC, server talking to client)
	DirectionTargetRequest  = 0b0100                                            // requests from target to source (i.e. bidirectional RPC, incl. subscription events)
	DirectionTargetResponse = 0b1000                                            // responses by target to source (effect is applied after reaching the server)
	DirectionSourceAny      = DirectionSourceRequest | DirectionSourceResponse  // match any kind from source
	DirectionTargetAny      = DirectionTargetRequest | DirectionTargetResponse  // match any kind from target
	DirectionBiRequest      = DirectionSourceRequest | DirectionTargetRequest   // match requests, both directions.
	DirectionBiResponse     = DirectionSourceResponse | DirectionTargetResponse // match responses, both directions.
	DirectionAny            = DirectionBiRequest | DirectionBiResponse          // match any kind of request/response (default).
)

func (d Direction) MatchRequest() bool {
	return d&DirectionBiRequest != 0
}

func (d Direction) MatchResponse() bool {
	return d&DirectionBiResponse != 0
}

func (d Direction) MatchTarget() bool {
	return d&DirectionTargetAny != 0
}

func (d Direction) MatchSource() bool {
	return d&DirectionSourceAny != 0
}

func (d Direction) String() string {
	switch d {
	case DirectionAny:
		return "any"
	case DirectionBiRequest:
		return "bi-request"
	case DirectionBiResponse:
		return "bi-response"
	case DirectionSourceRequest:
		return "source-request"
	case DirectionSourceResponse:
		return "source-response"
	case DirectionTargetRequest:
		return "target-request"
	case DirectionTargetResponse:
		return "target-response"
	default:
		return fmt.Sprintf("unknown-%d", uint32(d))
	}
}

func (d Direction) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d Direction) Valid() bool {
	switch d {
	case DirectionAny,
		DirectionBiRequest,
		DirectionBiResponse,
		DirectionSourceRequest,
		DirectionSourceResponse,
		DirectionTargetRequest,
		DirectionTargetResponse:
		return true
	default:
		return false
	}
}

func (d *Direction) UnmarshalText(data []byte) error {
	str := string(data)
	for _, v := range []Direction{
		DirectionSourceRequest,
		DirectionSourceResponse,
		DirectionTargetRequest,
		DirectionTargetResponse,
		DirectionBiRequest,
		DirectionBiResponse,
		DirectionAny,
	} {
		if str == v.String() {
			*d = v
			return nil
		}
	}
	return fmt.Errorf("invalid direction: %q", str)
}
