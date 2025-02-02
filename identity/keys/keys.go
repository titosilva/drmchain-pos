package keys

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

func GetCurve() elliptic.Curve {
	return elliptic.P256()
}

func GetECDHCurve() ecdh.Curve {
	return ecdh.P256()
}

func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(GetCurve(), rand.Reader)
}

func ToTag(pubKey *ecdsa.PublicKey) string {
	bs := elliptic.MarshalCompressed(pubKey.Curve, pubKey.X, pubKey.Y)
	return base64.RawStdEncoding.EncodeToString(bs)
}

func PrivateKeyToBytes(privKey *ecdsa.PrivateKey) []byte {
	return privKey.D.Bytes() // TODO: this should be encrypted
}

func BytesToPrivateKey(bytes []byte) (*ecdsa.PrivateKey, error) {
	privKey := new(ecdsa.PrivateKey)
	privKey.Curve = GetCurve()
	privKey.D = new(big.Int).SetBytes(bytes)
	// For some reason, X is not set by the previous function
	// TODO: find a better way
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.Curve.ScalarBaseMult(bytes)

	return privKey, nil
}
