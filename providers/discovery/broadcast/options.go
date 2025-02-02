package broadcast

import (
	"sync"
	"time"

	"github.com/choria-io/go-choria/inter"
	"github.com/choria-io/go-choria/protocol"
)

type dOpts struct {
	filter         *protocol.Filter
	collective     string
	msg            inter.Message
	discovered     []string
	cl             ChoriaClient
	mu             *sync.Mutex
	timeout        time.Duration
	dynamicTimeout bool
	name           string
}

// DiscoverOption configures the broadcast discovery method
type DiscoverOption func(o *dOpts)

// Name sets a NATS connection name to use, without this random names will be made.
//
// This setting is important if you make a daemon that makes many long client connections
// as each client connection makes Prometheus stats based on the name and you'll be
// leaking many stats over time
func Name(n string) DiscoverOption {
	return func(o *dOpts) {
		o.name = n
	}
}

// Filter sets the filter to use for the discovery, else a blank one is used
func Filter(f *protocol.Filter) DiscoverOption {
	return func(o *dOpts) {
		o.filter = f
	}
}

// Collective sets the collective to discover in, else main collective is used
func Collective(c string) DiscoverOption {
	return func(o *dOpts) {
		o.collective = c
	}
}

// Timeout sets the discovery timeout, else the configured default is used
func Timeout(t time.Duration) DiscoverOption {
	return func(o *dOpts) {
		o.timeout = t
	}
}

// SlidingWindow enables a sliding window for discovery timeout that
// terminates discovery after 300ms of no responses
func SlidingWindow() DiscoverOption {
	return func(o *dOpts) {
		o.dynamicTimeout = true
	}
}
func choriaClient(c ChoriaClient) DiscoverOption {
	return func(o *dOpts) {
		o.cl = c
	}
}
