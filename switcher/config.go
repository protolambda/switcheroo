package switcher

import (
	"regexp"
	"time"

	jsonrpc "github.com/protolambda/jsonrpc2"
)

type Config struct {
	// Targets are outgoing connections towards external RPCs
	Targets map[string]*Target `yaml:"targets"`
	// Sources are expected incoming connections
	Sources map[string]*Source `yaml:"sources"`
}

type Target struct {
	Endpoint string `yaml:"endpoint"`
	// Effects applied to every
	Effects []*Effect `yaml:"effects"`
}

type Source struct {
	Effects []*Effect `yaml:"effects"`
}

type Effect struct {
	// Direction to match the effect on.
	// 'any': match any kind of request/response (default).
	// 'bi-request': match requests, both directions.
	// 'bi-response': match responses, both directions.
	// 'source-request': requests from source to target (i.e. regular usage, before server handles the request)
	// 'source-response': responses by source to target (i.e. bidirectional RPC, server talking to client)
	// 'target-request': requests from target to source (i.e. bidirectional RPC, incl. subscription events)
	// 'target-response': responses by target to source (effect is applied after reaching the server)
	Direction string `yaml:"direction,omitempty"`

	// RegexMatcher matches the method with a regex.
	// Effects are only applied to matching RPC methods.
	RegexMatcher *regexp.Regexp `yaml:"matcher,omitempty"`
	// FuncMatcher matches the message (request or response)
	FuncMatcher func(msg *jsonrpc.Message) `yaml:"-"`

	Delay      *DelayEffect      `yaml:"delay,omitempty"`
	Drop       *DropEffect       `yaml:"drop,omitempty"`
	Error      *ErrorEffect      `yaml:"error,omitempty"`
	RateLimit  *RateLimitEffect  `yaml:"rateLimit,omitempty"`
	Parallel   *ParallelEffect   `yaml:"parallel,omitempty"`
	Substitute *SubstituteEffect `yaml:"substitute,omitempty"`
}

type DelayEffect struct {
	// MaxJitter is the maximum extra delay that is added.
	// The jitter distribution is uniform. Set to 0 to disable.
	MaxJitter time.Duration `yaml:"maxJitter,omitempty"`
	// Delay is a flat extra delay added to the propagation of requests.
	// Set to 0 to disable. Negative delay has no effect.
	Delay time.Duration `yaml:"delay,omitempty"`
}

type DropEffect struct {
	// Chance drops messages with the given probability.
	// Set to 0 to disable. Negative probability has no effect.
	Chance float64 `yaml:"chance,omitempty"`
}

// ErrorEffect only applies to request messages.
type ErrorEffect struct {
	// Chance responds to the request with an error with the given probability.
	// Set to 0 to disable. Negative probability has no effect.
	Chance float64 `yaml:"chance,omitempty"`
	// Code to reply with if Chance is hit.
	Code int64 `yaml:"code,omitempty"`
	// Message to reply with if Chance is hit. Optional.
	Message string `yaml:"message,omitempty"`
	// Data to reply with if Chance is hit.
	// Can be a structured object in YAMl config, will be JSON-encoded in the response.
	// Optional.
	Data any `yaml:"data,omitempty"`
}

type RateLimitEffect struct {
	// Rate is the maximum rate of events, i.e. how fast reservations become available again.
	// Negative rates are invalid.
	Rate float64 `yaml:"rate"`
	// Burst is the number of reservations that may be filled at one time.
	Burst uint `yaml:"burst"`
}

type ParallelEffect struct {
	// Max is the number of requests that may be open at any time, awaiting a response.
	// Setting this to 0 blocks all requests.
	// Setting this to 1 makes the RPC synchronous.
	Max int `yaml:"max"`
}

type SubstituteEffect struct {
	// Result is provided as alternative response.
	// Also see Effect.Direction configuration:
	// to override the effect before it reaches the target, or only replaces the response after the server responds.
	// Can be a structured object in YAMl config, will be JSON-encoded in the response.
	Result any `yaml:"result"`
}
