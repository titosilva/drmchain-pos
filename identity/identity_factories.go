package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"errors"

	"github.com/titosilva/drmchain-pos/identity/keys"
)

func FromPublicKey(pubKey *ecdsa.PublicKey) PublicIdentity {
	return &peerIdentity{
		publicKey: pubKey,
	}
}

func FromTag(tag string) (PublicIdentity, error) {
	pubKeyBs, err := base64.RawStdEncoding.DecodeString(tag)

	if err != nil {
		return nil, err
	}

	curve := keys.GetCurve()
	x, y := elliptic.UnmarshalCompressed(curve, pubKeyBs)
	if x == nil {
		return nil, errors.New("failed public key unmarshalling")
	}

	pubKey := new(ecdsa.PublicKey)
	pubKey.Curve = curve
	pubKey.X = x
	pubKey.Y = y

	return FromPublicKey(pubKey), nil
}

func FromPrivateKey(privKey *ecdsa.PrivateKey) PrivateIdentity {
	return &selfIdentity{
		privateKey: privKey,
	}
}

func FromPrivateKeyBytes(privKeyBs []byte) (PrivateIdentity, error) {
	privKey, err := keys.BytesToPrivateKey(privKeyBs)

	if err != nil {
		return nil, err
	}

	return FromPrivateKey(privKey), nil
}

func Generate() (PrivateIdentity, error) {
	privKey, err := keys.GeneratePrivateKey()

	if err != nil {
		return nil, err
	}

	return FromPrivateKey(privKey), nil
}
