package middlewares

import "github.com/titosilva/drmchain-pos/transactions"

type TransactionNotifierMiddleware struct {
	transactionNotifier transactions.TransactionNotifier
}

func NewTransactionNotifierMiddleware(transactionNotifier transactions.TransactionNotifier) *TransactionNotifierMiddleware {
	return &TransactionNotifierMiddleware{
		transactionNotifier: transactionNotifier,
	}
}

// Next implements transactions.TransactionMiddleware.
func (t *TransactionNotifierMiddleware) Next(tran transactions.Transaction) (transactions.Transaction, error) {
	t.transactionNotifier.Notify(tran)
	return tran, nil
}

// TransactionNotifierMiddleware is a default implementation of TransactionMiddleware
// Static impl check
var _ transactions.TransactionMiddleware = &TransactionNotifierMiddleware{}
