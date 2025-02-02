package transactions

import (
	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/identity/signatures"
	"github.com/titosilva/drmchain-pos/internal/utils/cryptutil"
)

type TransactionShape struct {
	SourceTag string
	Signature []byte
	Content   []byte
}

// GetContent implements Transaction.
func (t *TransactionShape) GetContent() []byte {
	return t.Content
}

// GetSignature implements Transaction.
func (t *TransactionShape) GetSignature() []byte {
	return t.Signature
}

// GetSource implements Transaction.
func (t *TransactionShape) GetSource() identity.PublicIdentity {
	id, err := identity.FromTag(t.SourceTag)
	if err != nil {
		panic(err)
	}

	return id
}

// GetSourceTag implements Transaction.
func (t *TransactionShape) GetSourceTag() string {
	return t.SourceTag
}

// IsValidSignature implements Transaction.
func (t *TransactionShape) IsValidSignature() bool {
	return signatures.Verify(t.GetSource(), t.GetContent(), t.GetSignature())
}

func (t *TransactionShape) GetRaw() []byte {
	encoder := NewTransactionAsn1Encoder()
	raw, err := encoder.EncodeTransaction(t)
	if err != nil {
		panic(err)
	}
	return raw
}

func (t *TransactionShape) GetHash() string {
	return cryptutil.HashToString(t.GetRaw())
}

// TransactionShape is a simple implementation of the Transaction interface.
// Static impl check
var _ Transaction = &TransactionShape{}
