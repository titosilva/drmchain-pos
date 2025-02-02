package identity

import (
	"crypto/ecdsa"
)

type PublicIdentity interface {
	GetPublicKey() *ecdsa.PublicKey
	GetTag() string
}

type PrivateIdentity interface {
	PublicIdentity
	GetPrivateKey() *ecdsa.PrivateKey
	GetPrivateKeyBytes() []byte
}
