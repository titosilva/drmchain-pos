package signatures

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/titosilva/drmchain-pos/identity"
)

func Verify(id identity.PublicIdentity, data []byte, signature []byte) bool {
	pubKey := id.GetPublicKey()
	return ecdsa.VerifyASN1(pubKey, data, signature)
}

func Sign(id identity.PrivateIdentity, data []byte) ([]byte, error) {
	privKey := id.GetPrivateKey()
	return ecdsa.SignASN1(rand.Reader, privKey, data)
}
