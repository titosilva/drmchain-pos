package pool

import (
	"context"
	"log"

	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cmap"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cqueue"
	"github.com/titosilva/drmchain-pos/transactions"
)

type PoolEntry struct {
	Transaction        transactions.Transaction
	Accepted           bool
	RegisteredInABlock bool
}

const PoolSize = 8192

type TransactionPool struct {
	transactionsToProcess chan transactions.Transaction
	accepted              *cqueue.CQueue[*PoolEntry]
	cancellation          context.Context
	cancel                context.CancelFunc
	processors            int

	entries *cmap.CMap[string, *PoolEntry]
}

func Factory(diCtx *di.DIContext) *TransactionPool {
	transactionsToProcess := make(chan transactions.Transaction, PoolSize)
	validatedTransactions := cqueue.New[*PoolEntry]()

	entries := cmap.New[string, *PoolEntry]()

	cancellation, cancel := context.WithCancel(context.Background()) // TODO: single cancellation for app lifecycle
	tp := &TransactionPool{
		transactionsToProcess: transactionsToProcess,
		accepted:              validatedTransactions,
		entries:               entries,
		cancellation:          cancellation,
		cancel:                cancel,
		processors:            4, // TODO: get from config
	}

	for i := 0; i < tp.processors; i++ {
		go tp.processTransactions()
	}

	return tp
}

func GetFromDI(diCtx *di.DIContext) *TransactionPool {
	return di.GetService[TransactionPool](diCtx)
}

func (tp *TransactionPool) Add(transaction transactions.Transaction) {
	tp.transactionsToProcess <- transaction
}

func (tp *TransactionPool) processTransactions() {
	log.Println("Starting transaction pool processor")
	for {
		select {
		case <-tp.cancellation.Done():
			log.Println("Stopping transaction pool processor")
			return
		case transaction := <-tp.transactionsToProcess:
			tp.processTransaction(transaction)
		}
	}
}

func (tp *TransactionPool) processTransaction(transaction transactions.Transaction) {
	entry, found := tp.entries.Get(transaction.GetHash())

	if !found {
		entry = &PoolEntry{
			Transaction: transaction,
		}
	}
	tp.entries.Set(transaction.GetHash(), entry)

	if !transaction.IsValidSignature() {
		entry.Accepted = false
		return
	}

	// TODO: fully validate transaction

	tp.accepted.Enqueue(entry)
}
