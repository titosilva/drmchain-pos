package states

type VotingState struct {
	Context *MachineContext
}

func NewVotingState(ctx *MachineContext) *VotingState {
	return &VotingState{
		Context: ctx,
	}
}

// Run implements ConsensusMachineState.
func (v *VotingState) Run() ConsensusMachineState {
	panic("not implemented") // TODO: Implement
}

// static impl check
var _ ConsensusMachineState = &VotingState{}
