package merkle_test

import (
	"crypto/rand"
	"testing"

	"github.com/titosilva/drmchain-pos/blocks/merkle"
)

func Test__MerkleVerify__ShouldAcceptPath__WhenCorrect(t *testing.T) {
	tree := merkle.NewTree()
	count := 10000
	data := make([][]byte, count)

	for i := 0; i < count; i++ {
		bs := randomBytes(32)
		tree.Add(bs)
		data[i] = bs
	}

	for i := 0; i < count; i++ {
		path := tree.GetPathTo(data[i])
		if !tree.Verify(path) {
			t.Errorf("Merkle tree verification failed")
		}
	}
}

func randomBytes(length int) []byte {
	buf := make([]byte, length)
	rand.Read(buf)
	return buf
}
