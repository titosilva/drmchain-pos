package middlewares

import (
	"errors"

	"github.com/titosilva/drmchain-pos/transactions"
)

type SignatureValidatorMiddleware struct {
}

func NewSignatureValidatorMiddleware() *SignatureValidatorMiddleware {
	return &SignatureValidatorMiddleware{}
}

// Next implements transactions.TransactionMiddleware.
func (s *SignatureValidatorMiddleware) Next(tran transactions.Transaction) (transactions.Transaction, error) {
	if !tran.IsValidSignature() {
		return nil, errors.New("invalid signature")
	}

	return tran, nil
}

// SignatureValidatorMiddleware is a simple implementation of the TransactionMiddleware interface.
// Static impl check
var _ transactions.TransactionMiddleware = &SignatureValidatorMiddleware{}
