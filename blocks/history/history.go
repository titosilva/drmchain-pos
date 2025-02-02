package history

import (
	"errors"

	"github.com/titosilva/drmchain-pos/blocks"
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/clru"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cmap"
)

var ErrBlockIndex = errors.New("block index is not the expected")

type BlockHistory struct {
	lastBlocks    *clru.Cache[uint64, *blocks.Block]
	currentStakes *cmap.CMap[string, uint64]
	lastIndex     uint64
}

func Factory(diCtx *di.DIContext) *BlockHistory {
	// TODO: recover data from disk
	return &BlockHistory{
		lastBlocks:    clru.New[uint64, *blocks.Block](100),
		currentStakes: cmap.New[string, uint64](),
		lastIndex:     0,
	}
}

func GetFromDI(diCtx *di.DIContext) *BlockHistory {
	return di.GetService[BlockHistory](diCtx)
}

func (bh *BlockHistory) GetLastBlock() *blocks.Block {
	block, _ := bh.lastBlocks.Get(bh.lastIndex)
	return block
}

func (bh *BlockHistory) GetStakes(tag string) uint64 {
	stakes, found := bh.currentStakes.Get(tag)

	if !found {
		return 0
	}

	return stakes
}

func (bh *BlockHistory) Append(block *blocks.Block) error {
	if bh.lastIndex != 0 && block.Index != bh.lastIndex+1 {
		return ErrBlockIndex
	}

	bh.lastBlocks.Put(block.Index, block)
	bh.lastIndex = block.Index
	block.Previous = bh.GetLastBlock()

	for _, tx := range block.Transations {
		bh.IncrementStakes(tx.GetSource().GetTag(), 1)
	}

	return nil
}

func (bh *BlockHistory) IncrementStakes(tag string, amount uint64) {
	stakes, found := bh.currentStakes.Get(tag)

	if !found {
		stakes = 0
	}

	bh.currentStakes.Set(tag, stakes+amount)
}
