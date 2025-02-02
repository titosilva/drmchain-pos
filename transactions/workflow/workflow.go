package workflow

import (
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/transactions"
	"github.com/titosilva/drmchain-pos/transactions/middlewares"
	"github.com/titosilva/drmchain-pos/transactions/notifier"
	"github.com/titosilva/drmchain-pos/transactions/pool"
)

type DefaultTransactionWorkflow struct {
	middleware []transactions.TransactionMiddleware
}

func Factory(diCtx *di.DIContext) transactions.TransactionWorkflow {
	builder := NewBuilder()

	notifier := notifier.GetFromDI(diCtx)
	builder.AddMiddleware(middlewares.NewSignatureValidatorMiddleware())
	builder.AddMiddleware(middlewares.NewTransactionNotifierMiddleware(notifier))

	pool := pool.GetFromDI(diCtx)
	builder.AddMiddleware(middlewares.NewTransactionPoolMiddleware(pool))

	return builder.Build()
}

// Process implements transactions.TransactionWorkflow.
func (d *DefaultTransactionWorkflow) Process(tran transactions.Transaction) error {
	currentTran := tran
	var err error
	for _, m := range d.middleware {
		currentTran, err = m.Next(currentTran)

		if err != nil {
			return err
		}

		if currentTran == nil {
			return nil
		}
	}

	return nil
}

// DefaultTransactionWorkflow is an implementation of the TransactionWorkflow interface.
// Static impl check
var _ transactions.TransactionWorkflow = &DefaultTransactionWorkflow{}

type DefaultTransactionWorkflowBuilder struct {
	workflow *DefaultTransactionWorkflow
}

func NewBuilder() *DefaultTransactionWorkflowBuilder {
	return &DefaultTransactionWorkflowBuilder{
		workflow: &DefaultTransactionWorkflow{},
	}
}

// AddMiddleware implements transactions.TransactionWorkflowBuilder.
func (d *DefaultTransactionWorkflowBuilder) AddMiddleware(middleware transactions.TransactionMiddleware) transactions.TransactionWorkflowBuilder {
	d.workflow.middleware = append(d.workflow.middleware, middleware)
	return d
}

// Build implements transactions.TransactionWorkflowBuilder.
func (d *DefaultTransactionWorkflowBuilder) Build() transactions.TransactionWorkflow {
	return d.workflow
}

// DefaultTransactionWorkflowBuilder is an implementation of the TransactionWorkflowBuilder interface.
// Static impl check
var _ transactions.TransactionWorkflowBuilder = &DefaultTransactionWorkflowBuilder{}
