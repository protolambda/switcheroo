package switcher

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"regexp"
	"time"

	"golang.org/x/time/rate"

	jsonrpc "github.com/protolambda/jsonrpc2"
)

type Config struct {
	// Targets are outgoing connections towards external RPCs
	Targets map[string]*Target `yaml:"targets"`
	// Sources are expected incoming connections
	Sources map[string]*Source `yaml:"sources"`
}

type Target struct {
	// Endpoint to connect to
	Endpoint string `yaml:"endpoint"`
	// KeepAlive keeps the connection to the endpoint open, even if it's not being used.
	KeepAlive bool `yaml:"keepAlive,omitempty"`
	// Effects applied to every
	Effects []*Effect `yaml:"effects"`
}

type Source struct {
	Effects []*Effect `yaml:"effects"`
}

type Effect struct {
	// Direction to match the effect on.
	Direction Direction `yaml:"direction,omitempty"`

	// RegexFilter filters the method with a regex.
	// Effects are only applied to matching RPC methods.
	// If no filter is configured, the message is let through (there may still be a FuncFilter).
	RegexMatcher *regexp.Regexp `yaml:"filter,omitempty"`
	// FuncFilter matches the message (request or response).
	// If true, the message is let through. If false, the message is dropped.
	// Optional additional filter step.
	FuncFilter func(msg *jsonrpc.Message) bool `yaml:"-"`

	Delay      *DelayEffect      `yaml:"delay,omitempty"`
	Drop       *DropEffect       `yaml:"drop,omitempty"`
	Error      *ErrorEffect      `yaml:"error,omitempty"`
	RateLimit  *RateLimitEffect  `yaml:"rateLimit,omitempty"`
	Parallel   *ParallelEffect   `yaml:"parallel,omitempty"`
	Substitute *SubstituteEffect `yaml:"substitute,omitempty"`
}

func (ef *Effect) Init() {
	// TODO start effect sub-routines
}

func (ef *Effect) Run(incoming chan *Envelope, outgoing chan *Envelope) {
	// TODO pump messages through all effect sub-routines
}

func (ef *Effect) Close() error {
	// TODO close all sub-effects
	return nil
}

type DelayEffect struct {
	// MaxJitter is the maximum extra delay that is added.
	// The jitter distribution is uniform. Set to 0 to disable.
	MaxJitter time.Duration `yaml:"maxJitter,omitempty"`
	// Time is a flat extra delay added to the propagation of requests.
	// Set to 0 to disable. Negative delay has no effect.
	Time time.Duration `yaml:"time,omitempty"`
}

func (ef *DelayEffect) Run(incoming chan *Envelope, outgoing chan *Envelope) {
	// TODO read incoming message
	// TODO go routine with delay
	// TODO pass on message
}

type DropEffect struct {
	// Chance drops messages with the given probability.
	// Set to 0 to disable. Negative probability has no effect.
	Chance float64 `yaml:"chance,omitempty"`
}

func (ef *DropEffect) Run(incoming chan *Envelope, outgoing chan *Envelope) {
	// TODO read incoming msg
	// check chance
	if checkChance(ef.Chance) {
		// drop, with warn log maybe
	} else {
		// pass on the message
	}
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

func (ef *ErrorEffect) Run(incoming chan *Envelope, outgoing chan *Envelope) {
	// TODO take incoming msg
	// if request, forward on a matching error
	// if response, replace with an error
}

type RateLimitEffect struct {
	// Rate is the maximum rate of events, i.e. how fast reservations become available again.
	// Negative rates are invalid.
	Rate float64 `yaml:"rate"`
	// Burst is the number of reservations that may be filled at one time.
	Burst uint `yaml:"burst"`

	// limiter that implements the actual rate-limit
	limiter *rate.Limiter `yaml:"-"`
}

func (ef *RateLimitEffect) Init() error {
	if ef.Rate < 0 {
		return fmt.Errorf("invalid rate, cannot be negative: %f", ef.Rate)
	}
	ef.limiter = rate.NewLimiter(rate.Limit(ef.Rate), int(ef.Burst))
	return nil
}

func (ef *RateLimitEffect) Run(incoming chan *Envelope, outgoing chan *Envelope) {
	// TODO take incoming request
	// await rate limiter
	// pass msg onwards
}

type ParallelEffect struct {
	// Max is the number of requests that may be open at any time, awaiting a response.
	// Setting this to 0 blocks all requests.
	// Setting this to 1 makes the RPC synchronous.
	Max int `yaml:"max"`

	tokens chan struct{} `yaml:"-"`
}

func (p *ParallelEffect) Init() error {
	p.tokens = make(chan struct{}, p.Max)
	return nil
}

func (ef *ParallelEffect) Run(incoming chan *Envelope, outgoing chan *Envelope) {
	// TODO register requests by ID and sender

	// TODO on response, remove by ID and sender

	// TODO on request, take token

	// TODO on response, give back token
}

type SubstituteEffect struct {
	// Result is provided as alternative response.
	// Also see Effect.Direction configuration:
	// to override the effect before it reaches the target, or only replaces the response after the server responds.
	// Can be a structured object in YAMl config, will be JSON-encoded in the response.
	Result any `yaml:"result"`
}

func (ef *SubstituteEffect) Run(incoming chan *Envelope, outgoing chan *Envelope) {
	// TODO take incoming requests, substitute responses, write outgoing
}

func checkChance(f float64) bool {
	if f <= 0 {
		return false
	}
	if f >= 1 {
		return true
	}
	return randUniformFloat64() < f
}

// randUniformFloat64 returns a random uniform float64 in the range 0 to 1
func randUniformFloat64() float64 {
	var x [8]uint8
	_, err := rand.Read(x[:])
	if err != nil {
		panic(fmt.Errorf("failed to get randomness: %w", err))
	}
	v := binary.LittleEndian.Uint64(x[:])
	return float64(v) / float64(math.MaxUint64)
}
