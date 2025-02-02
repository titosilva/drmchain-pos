package blocks

import (
	"github.com/titosilva/drmchain-pos/blocks/merkle"
	"github.com/titosilva/drmchain-pos/transactions"
)

type Block struct {
	ForgerTag       string
	ForgerSignature []byte
	Index           uint64
	Hash            []byte
	Merkle          *merkle.MerkleTree
	Transations     []transactions.Transaction

	Previous *Block
}
