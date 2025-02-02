package consensus

import "slices"

type ConsensusContext struct {
	BlockIndex uint64
	Phase      string
	PrevHash   []byte
}

func (ctx *ConsensusContext) IsSame(otherCtx ConsensusContext) bool {
	return ctx.BlockIndex == otherCtx.BlockIndex &&
		slices.Equal(ctx.PrevHash, otherCtx.PrevHash) &&
		ctx.Phase == otherCtx.Phase
}
