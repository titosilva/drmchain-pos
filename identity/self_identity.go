package identity

import (
	"crypto/ecdsa"

	"github.com/titosilva/drmchain-pos/identity/keys"
)

// selfIdentity is a struct that represents the identity of the node itself, with a private key.
// It implements the PrivateIdentity interface.
type selfIdentity struct {
	privateKey *ecdsa.PrivateKey
}

// GetPrivateKey implements PrivateIdentity.
func (s selfIdentity) GetPrivateKey() *ecdsa.PrivateKey {
	return s.privateKey
}

// GetPrivateKeyBytes implements PrivateIdentity.
func (s selfIdentity) GetPrivateKeyBytes() []byte {
	return keys.PrivateKeyToBytes(s.privateKey)
}

// GetPublicKey implements PrivateIdentity.
func (s selfIdentity) GetPublicKey() *ecdsa.PublicKey {
	return &s.privateKey.PublicKey
}

// GetTag implements PrivateIdentity.
func (s selfIdentity) GetTag() string {
	pubKey := s.GetPublicKey()
	return keys.ToTag(pubKey)
}

var _ PrivateIdentity = selfIdentity{}
var _ PublicIdentity = selfIdentity{}
