package cryptutil

import (
	"crypto/sha256"
	"encoding/hex"
)

func Hash(data []byte) []byte {
	sha := sha256.New()
	sha.Write(data)
	return sha.Sum(nil)
}

func HashToString(data []byte) string {
	return hex.EncodeToString(Hash(data))
}
