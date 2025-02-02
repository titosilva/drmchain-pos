package transactions

import (
	"github.com/titosilva/drmchain-pos/network/encodings"
)

type TransactionAsn1Encoder struct {
}

// NewTransactionAsn1Encoder creates a new instance of the TransactionAsn1Encoder.
func NewTransactionAsn1Encoder() *TransactionAsn1Encoder {
	return &TransactionAsn1Encoder{}
}

// DecodeTransaction implements TransactionEncoder.
func (t *TransactionAsn1Encoder) DecodeTransaction(raw []byte) (Transaction, error) {
	var transactionShape TransactionShape

	if err := encodings.Decode(raw, &transactionShape); err != nil {
		return nil, err
	}

	return &transactionShape, nil
}

// EncodeTransaction implements TransactionEncoder.
func (t *TransactionAsn1Encoder) EncodeTransaction(tran Transaction) ([]byte, error) {
	transaction := tran.(*TransactionShape)
	return encodings.Encode(transaction)
}

// TransactionAsn1Encoder is a simple implementation of the TransactionEncoder interface.
// Static impl check
var _ TransactionEncoder = &TransactionAsn1Encoder{}
