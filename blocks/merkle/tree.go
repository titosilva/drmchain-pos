package merkle

import "slices"

type MerkleTree struct {
	root *MerkleNode
}

func NewTree() *MerkleTree {
	return &MerkleTree{
		root: nil,
	}
}

func (t *MerkleTree) Add(data []byte) {
	if t.root == nil {
		t.root = newNode(data)
		return
	}

	t.root.add(data)
}

func (t *MerkleTree) GetRoot() []byte {
	if t.root == nil {
		return nil
	}

	return t.root.hash
}

func (t *MerkleTree) GetPathTo(data []byte) [][]byte {
	path := t.root.getPath(hash(data))

	if path == nil {
		return nil
	}

	slices.Reverse(path)
	return path
}

func (t *MerkleTree) Verify(path [][]byte) bool {
	if t.root == nil {
		return false
	}

	return t.root.verify(path)
}
