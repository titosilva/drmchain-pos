package states

import (
	"context"

	"github.com/titosilva/drmchain-pos/consensus/internal/forgery"
	"github.com/titosilva/drmchain-pos/consensus/messages"
	"github.com/titosilva/drmchain-pos/network/encodings"
)

type CommitmentState struct {
	Context    *MachineContext
	timeoutCtx context.Context
}

func (cs *CommitmentState) Run() ConsensusMachineState {
	for {
		select {
		case msg := <-cs.Context.MessagesChan:
			var commitment messages.CommitmentMessage
			if err := encodings.Decode(msg, &commitment); err != nil {
				continue
			}

			// TODO: check if commitment is valid

			p := forgery.NewParticipation(&commitment)
			cs.Context.Forgery.AddParticipation(p)
			continue
		case <-cs.timeoutCtx.Done():
			return NewRevealingState()
		case <-cs.Context.KillCtx.Done():
			return nil
		}
	}
}

// Static implementation check
var _ ConsensusMachineState = &CommitmentState{}
