package keyexchange_test

import (
	"bytes"
	"testing"

	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/identity/keyexchange"
)

func Test__DeriveFromPublicKey__ShouldReturnSameSecretAsDeriveFromPrivateKey__WhenSameEphKeyIsUsed(t *testing.T) {
	// Arrange
	self, err := identity.Generate()

	if err != nil {
		t.Error(err)
	}

	ephKey, err := keyexchange.GenerateEphemeralKey()

	if err != nil {
		t.Error(err)
	}

	// Act
	secret1, err := keyexchange.DeriveFromPrivateIdentity(self, ephKey.PublicKey())

	if err != nil {
		t.Error(err)
	}

	secret2, err := keyexchange.DeriveFromPublicIdentity(self, ephKey)

	if err != nil {
		t.Error(err)
	}

	// Assert
	if !bytes.Equal(secret1, secret2) {
		t.Fail()
	}
}

func Test__KeyToBytes__ShouldReturnSameKey__WhenBytesToKeyIsUsed(t *testing.T) {
	// Arrange
	ephKey, err := keyexchange.GenerateEphemeralKey()

	if err != nil {
		t.Error(err)
	}

	// Act
	data := keyexchange.KeyToBytes(ephKey.PublicKey())
	key, err := keyexchange.BytesToKey(data)

	if err != nil {
		t.Error(err)
	}

	// Assert
	if !ephKey.PublicKey().Equal(key) {
		t.Fail()
	}
}
