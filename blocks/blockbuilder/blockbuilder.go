package blockbuilder

import (
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/transactions/pool"
)

type BlockBuilder struct {
	pool *pool.TransactionPool
}

func Factory(diCtx *di.DIContext) *BlockBuilder {
	return &BlockBuilder{
		pool: pool.GetFromDI(diCtx),
	}
}

// func (b *BlockBuilder) Build() {
// 	transactions := b.pool.Add()
// }
