package keyexchange

import (
	"crypto/ecdh"
	"crypto/rand"

	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/identity/keys"
)

func GenerateEphemeralKey() (*ecdh.PrivateKey, error) {
	return keys.GetECDHCurve().GenerateKey(rand.Reader)
}

func DeriveFromPrivateIdentity(self identity.PrivateIdentity, ephKey *ecdh.PublicKey) ([]byte, error) {
	privKey := self.GetPrivateKey()
	idKey, err := privKey.ECDH()

	if err != nil {
		return nil, err
	}

	return idKey.ECDH(ephKey)
}

func DeriveFromPublicIdentity(self identity.PublicIdentity, ephKey *ecdh.PrivateKey) ([]byte, error) {
	pubKey := self.GetPublicKey()
	idKey, err := pubKey.ECDH()

	if err != nil {
		return nil, err
	}

	return ephKey.ECDH(idKey)
}

func KeyToBytes(key *ecdh.PublicKey) []byte {
	return key.Bytes()
}

func BytesToKey(data []byte) (*ecdh.PublicKey, error) {
	return ecdh.P256().NewPublicKey(data)
}
