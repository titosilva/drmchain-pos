package forgery

import (
	"math/big"
	"slices"
	"sort"
)

func Elect(forgery *BlockForgery) {
	// Elect a forger
	distributedRandom := make([]byte, CommitedLength)
	ps := slices.Clone(forgery.Participations)
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Commitment.Stakes < ps[j].Commitment.Stakes
	})

	totalCoins := uint64(0)
	for _, p := range ps {
		if p.Revealing == nil {
			continue
		}

		if len(p.Revealing.Commited) != CommitedLength {
			continue
		}

		for i := 0; i < CommitedLength; i++ {
			distributedRandom[i] = distributedRandom[i] ^ p.Revealing.Commited[i]
			totalCoins += p.Commitment.Stakes
		}
	}

	currentCoin := uint64(0)
	coinIndex := reduce(distributedRandom, totalCoins)
	for _, p := range ps {
		currentCoin += p.Commitment.Stakes

		if coinIndex < currentCoin {
			forgery.ElectedTag = p.Commitment.Tag
			break
		}
	}
}

func reduce(large []byte, mod uint64) uint64 {
	b := big.NewInt(0)
	b.SetBytes(large)
	b.Mod(b, big.NewInt(int64(mod)))
	return b.Uint64()
}
