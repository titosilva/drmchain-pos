package consensusdi

import (
	"github.com/titosilva/drmchain-pos/consensus/internal/services"
	"github.com/titosilva/drmchain-pos/internal/di"
)

func AddConsensusServices(diCtx *di.DIContext) *di.DIContext {
	di.AddSingleton(diCtx, services.CommitmentFactoryFactory)

	return diCtx
}
