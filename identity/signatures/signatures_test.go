package signatures_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/identity/signatures"
)

func Test__Verify__ShouldReturnTrue__WhenSignatureWasCreatedWithCorrespondingKey(t *testing.T) {
	// Arrange
	id, err := identity.Generate()

	if err != nil {
		t.Error(err)
	}

	data := []byte("data")
	signature, err := signatures.Sign(id, data)

	if err != nil {
		t.Error(err)
	}

	// Act
	accepted := signatures.Verify(id, data, signature)

	// Assert
	if !accepted {
		t.Fail()
	}
}

func Test__Verify__ShouldReturnFalse__WhenSignatureWasCreatedWithDifferentKey(t *testing.T) {
	// Arrange
	id1, err := identity.Generate()

	if err != nil {
		t.Error(err)
	}

	id2, err := identity.Generate()

	if err != nil {
		t.Error(err)
	}

	data := []byte("data")
	signature, err := signatures.Sign(id1, data)

	if err != nil {
		t.Error(err)
	}

	// Act
	accepted := signatures.Verify(id2, data, signature)

	// Assert
	if accepted {
		t.Fail()
	}
}
