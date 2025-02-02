package identity

import (
	"crypto/ecdsa"

	"github.com/titosilva/drmchain-pos/identity/keys"
)

// peerIdentity is a struct that represents the identity of a peer in the network.
// It implements the PublicIdentity interface.
type peerIdentity struct {
	publicKey *ecdsa.PublicKey
}

// GetPublicKey implements PublicIdentity.
func (n peerIdentity) GetPublicKey() *ecdsa.PublicKey {
	return n.publicKey
}

// GetTag implements PublicIdentity.
func (n peerIdentity) GetTag() string {
	pubKey := n.GetPublicKey()
	return keys.ToTag(pubKey)
}

var _ PublicIdentity = peerIdentity{}
