package blocksdi

import (
	"github.com/titosilva/drmchain-pos/blocks/history"
	"github.com/titosilva/drmchain-pos/internal/di"
)

func AddBlocksServices(diCtx *di.DIContext) *di.DIContext {
	di.AddSingleton(diCtx, history.Factory)

	return diCtx
}
