package transactions

import (
	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"
)

type Transaction interface {
	GetContent() []byte
	GetSignature() []byte
	GetSource() identity.PublicIdentity
	GetSourceTag() string
	IsValidSignature() bool
	GetRaw() []byte
	GetHash() string
}

type TransactionEncoder interface {
	EncodeTransaction(Transaction) ([]byte, error)
	DecodeTransaction([]byte) (Transaction, error)
}

type TransactionMiddleware interface {
	Next(Transaction) (Transaction, error)
}

type TransactionWorkflow interface {
	Process(Transaction) error
}

type TransactionWorkflowBuilder interface {
	AddMiddleware(TransactionMiddleware) TransactionWorkflowBuilder
	Build() TransactionWorkflow
}

type TransactionNotifier interface {
	Subscribe() *observable.Subscription[Transaction]
	Notify(Transaction)
}
