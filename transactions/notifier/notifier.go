package notifier

import (
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"
	"github.com/titosilva/drmchain-pos/transactions"
)

type DefaultTransactionNotifier struct {
	transactionObservable *observable.Observable[transactions.Transaction]
}

func Factory(diCtx *di.DIContext) *DefaultTransactionNotifier {
	return &DefaultTransactionNotifier{
		transactionObservable: observable.New[transactions.Transaction](),
	}
}

func GetFromDI(diCtx *di.DIContext) *DefaultTransactionNotifier {
	return di.GetService[DefaultTransactionNotifier](diCtx)
}

func (d *DefaultTransactionNotifier) Notify(transaction transactions.Transaction) {
	d.transactionObservable.Notify(transaction)
}

// Subscribe implements transactions.TransactionNotifier.
func (d *DefaultTransactionNotifier) Subscribe() *observable.Subscription[transactions.Transaction] {
	return d.transactionObservable.Subscribe()
}

// DefaultTransactionNotifier is a default implementation of TransactionNotifier
// Static impl check
var _ transactions.TransactionNotifier = &DefaultTransactionNotifier{}
