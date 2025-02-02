package messages

import (
	"github.com/titosilva/drmchain-pos/consensus"
	"github.com/titosilva/drmchain-pos/network/encodings"
)

type ConsensusShell struct {
	Type    string
	Context consensus.ConsensusContext
	Content []byte
}

func (cs ConsensusShell) GetRaw() []byte {
	data, err := encodings.Encode(cs)

	if err != nil {
		panic("failed to encode Consensus Shell")
	}

	return data
}
