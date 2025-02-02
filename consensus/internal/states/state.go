package states

import (
	"context"

	"github.com/titosilva/drmchain-pos/consensus/internal/forgery"
	"github.com/titosilva/drmchain-pos/internal/di"
)

type ConsensusMachineState interface {
	Run() ConsensusMachineState
}

/*
Phases:
1. Commitment (6s)
2. Revealing (6s)
4. Voting (24s)
5. Propagating (24s)
*/

type MachineContext struct {
	MessagesChan <-chan []byte
	KillCtx      context.Context
	Kill         context.CancelFunc

	Forgery      *forgery.BlockForgery
	CurrentState ConsensusMachineState
	DiCtx        *di.DIContext
}

func (mc *MachineContext) Run() {
	for mc.CurrentState != nil {
		mc.CurrentState = mc.CurrentState.Run()
	}
}
