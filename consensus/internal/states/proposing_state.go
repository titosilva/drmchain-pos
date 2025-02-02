package states

import (
	"log"

	"github.com/titosilva/drmchain-pos/consensus/internal/forgery"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
)

type ProposingState struct {
	Context *MachineContext
}

func NewProposingState(ctx *MachineContext) *ProposingState {
	return &ProposingState{
		Context: ctx,
	}
}

// Run implements ConsensusMachineState.
func (p *ProposingState) Run() ConsensusMachineState {
	forgery.Elect(p.Context.Forgery)

	idProvider := identityprovider.GetFromDI(p.Context.DiCtx)
	id, err := idProvider.GetIdentity()
	if err != nil {
		log.Println("Error getting identity")
		return nil
	}

	if p.Context.Forgery.ElectedTag != id.GetTag() {
		return NewVotingState(p.Context)
	}

	return nil // TODO: Implement ProposingState
}

// static impl check
var _ ConsensusMachineState = &ProposingState{}
