package tunnel

import "github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"

// A tunnel represents a connection between two endpoints.
// Data can be passed through the tunnel in both directions.

// DuplexTunnel is an interface that represents a tunnel.
type DuplexTunnel interface {
	Send(data []byte) error
	Subscribe() *observable.Subscription[[]byte]
	WaitClose() <-chan struct{}
	Close() error
}

// WritableTunnel is a DuplexTunnel with a method to insert something to be received by others
// A DuplexTunnel may or not implement this interface
type WritableTunnel interface {
	DuplexTunnel
	// This should insert data in each Receive channel subscribed
	Notify(data []byte) error
}

type ErrorTunnelClosed struct{}

// Error implements the error interface
func (e *ErrorTunnelClosed) Error() string {
	return "tunnel is closed"
}

// Static impl check: error
var _ error = (*ErrorTunnelClosed)(nil)
