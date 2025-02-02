package merkle

import (
	"crypto/sha256"
	"slices"
)

func hash(data []byte) []byte {
	sha := sha256.New()
	sha.Write(data)
	return sha.Sum(nil)
}

type MerkleNode struct {
	left     *MerkleNode
	right    *MerkleNode
	hash     []byte
	data     []byte
	children int
}

func newNode(data []byte) *MerkleNode {
	return &MerkleNode{
		left:     nil,
		right:    nil,
		hash:     hash(data),
		data:     data,
		children: 1,
	}
}

func (n *MerkleNode) add(data []byte) {
	defer n.update()
	if n.left == nil && n.right == nil {
		n.left = newNode(n.data)
		n.right = newNode(data)
		return
	}

	if n.left.children <= n.right.children {
		n.left.add(data)
	} else {
		n.right.add(data)
	}
}

func (n *MerkleNode) update() {
	if n.left == nil && n.right == nil {
		n.children = 1
		return
	}

	n.hash = hash(append(n.left.hash, n.right.hash...))
	n.data = nil
	n.children = n.left.children + n.right.children
}

func (n *MerkleNode) getPath(hash []byte) [][]byte {
	if n.left == nil && n.right == nil {
		if slices.Equal(hash, n.hash) {
			return [][]byte{n.hash}
		}

		return nil
	}

	left := n.left.getPath(hash)
	if left != nil {
		return append(left, n.hash)
	}

	right := n.right.getPath(hash)
	if right != nil {
		return append(right, n.hash)
	}

	return nil
}

func (n *MerkleNode) verify(path [][]byte) bool {
	if len(path) == 0 {
		return false
	}

	if slices.Equal(path[0], n.hash) {
		return slices.Equal(path[0], n.hash)
	}

	if n.left == nil && n.right == nil {
		return false
	}

	return n.left.verify(path[1:]) || n.right.verify(path[1:])
}
