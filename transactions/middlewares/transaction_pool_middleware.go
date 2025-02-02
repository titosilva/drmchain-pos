package middlewares

import (
	"github.com/titosilva/drmchain-pos/transactions"
	"github.com/titosilva/drmchain-pos/transactions/pool"
)

type TransactionPoolMiddleware struct {
	transactionPool *pool.TransactionPool
}

// NewTransactionPoolMiddleware creates a new TransactionPoolMiddleware
func NewTransactionPoolMiddleware(transactionPool *pool.TransactionPool) *TransactionPoolMiddleware {
	return &TransactionPoolMiddleware{
		transactionPool: transactionPool,
	}
}

// Next implements transactions.TransactionMiddleware.
func (t *TransactionPoolMiddleware) Next(tran transactions.Transaction) (transactions.Transaction, error) {
	t.transactionPool.Add(tran)
	return tran, nil
}

// TransactionPoolMiddleware is a default implementation of TransactionMiddleware
// Static impl check
var _ transactions.TransactionMiddleware = &TransactionPoolMiddleware{}
