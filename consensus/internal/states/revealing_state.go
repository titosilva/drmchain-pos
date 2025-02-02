package states

import (
	"context"

	"github.com/titosilva/drmchain-pos/consensus/messages"
	"github.com/titosilva/drmchain-pos/network/encodings"
)

type RevealingState struct {
	Context    *MachineContext
	timeoutCtx context.Context
}

func NewRevealingState() *RevealingState {
	return &RevealingState{}
}

// Run implements ConsensusMachineState.
func (r *RevealingState) Run() ConsensusMachineState {
	for {
		select {
		case msg := <-r.Context.MessagesChan:
			revealing, err := encodings.DecodeAs[messages.RevealingMessage](msg)
			if err != nil {
				continue
			}

			// TODO: check if revealing is valid

			r.Context.Forgery.Reveal(&revealing)
			continue
		case <-r.timeoutCtx.Done():
			return NewRevealingState()
		case <-r.Context.KillCtx.Done():
			return nil
		}
	}
}

// Static implementation check
var _ ConsensusMachineState = &RevealingState{}
