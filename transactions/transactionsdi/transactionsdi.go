package transactionsdi

import (
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/transactions/notifier"
	"github.com/titosilva/drmchain-pos/transactions/pool"
	transactionnetwork "github.com/titosilva/drmchain-pos/transactions/transaction_network"
	"github.com/titosilva/drmchain-pos/transactions/workflow"
)

func AddTransactionServices(diCtx *di.DIContext) *di.DIContext {
	di.AddSingleton(diCtx, transactionnetwork.Factory)
	di.AddInterfaceFactory(diCtx, workflow.Factory)
	di.AddSingleton(diCtx, notifier.Factory)
	di.AddSingleton(diCtx, pool.Factory)

	return diCtx
}
