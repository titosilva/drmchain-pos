package forgery

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/titosilva/drmchain-pos/consensus/messages"
	"github.com/titosilva/drmchain-pos/internal/utils/cryptutil"
)

const CommitedLength = 8

type Participation struct {
	Commitment *messages.CommitmentMessage
	Revealing  *messages.RevealingMessage
}

func NewParticipation(commitment *messages.CommitmentMessage) Participation {
	return Participation{
		Commitment: commitment,
	}
}

type BlockForgery struct {
	Participations []Participation
	ElectedTag     string
}

func NewBlockForgery() *BlockForgery {
	return &BlockForgery{
		Participations: make([]Participation, 0),
	}
}

func (bf *BlockForgery) AddParticipation(p Participation) {
	bf.Participations = append(bf.Participations, p)
}

func (bf *BlockForgery) Reveal(revealing *messages.RevealingMessage) {
	for i, p := range bf.Participations {
		if revealing.CommitmentHash == cryptutil.HashToString(p.Commitment.Serialize()) {
			bf.Participations[i].Revealing = revealing
			return
		}
	}
}

func hash(data []byte) string {
	sha := sha256.New()
	sha.Write(data)
	return hex.EncodeToString(sha.Sum(nil))
}
