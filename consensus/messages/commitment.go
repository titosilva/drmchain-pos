package messages

import (
	"github.com/titosilva/drmchain-pos/network/encodings"
)

type CommitmentMessage struct {
	Tag       string
	Signature []byte

	Commitment []byte
	Stakes     uint64
	BlockIndex uint64
	PrevHash   []byte
}

func (cm CommitmentMessage) Serialize() []byte {
	bs, _ := encodings.Encode(cm)
	return bs
}

type RevealingMessage struct {
	Signature []byte

	CommitmentHash string
	Commited       []byte
}
