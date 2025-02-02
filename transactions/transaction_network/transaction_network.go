package transactionnetwork

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/patterns/longtask"
	"github.com/titosilva/drmchain-pos/internal/patterns/tunnel"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cbag"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/clru"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cmap"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/transactions"
)

type TransactionRegistry = clru.Cache[string, time.Time]

func newRegistry() *TransactionRegistry {
	return clru.New[string, time.Time](256)
}

type NetworkTransactionsHandler struct {
	net                *network.Network
	listenTask         *longtask.LongTask[any]
	tunnelSubs         *cbag.CBag[*observable.Subscription[[]byte]]
	workflow           transactions.TransactionWorkflow
	recentTransactions *TransactionRegistry
	recentByTunnel     *cmap.CMap[tunnel.DuplexTunnel, *TransactionRegistry]
}

func Factory(diCtx *di.DIContext) *NetworkTransactionsHandler {
	net := di.GetService[network.Network](diCtx)
	workflow := di.GetInterfaceService[transactions.TransactionWorkflow](diCtx)

	return &NetworkTransactionsHandler{
		net:                net,
		tunnelSubs:         cbag.New[*observable.Subscription[[]byte]](),
		workflow:           workflow,
		recentTransactions: newRegistry(),
		recentByTunnel:     cmap.New[tunnel.DuplexTunnel, *TransactionRegistry](),
	}
}

func GetFromDI(diCtx *di.DIContext) *NetworkTransactionsHandler {
	return di.GetService[NetworkTransactionsHandler](diCtx)
}

func (nth *NetworkTransactionsHandler) ObserveTransactions() {
	connectionsSub := nth.net.GetConnections().Subscribe()

	for conn := range nth.net.GetConnections().Current().All() {
		nth.handleConnection(conn)
	}

	task := longtask.Run(func(cancellation context.Context) any {
		log.Println("Starting transaction observer.")

		for {
			select {
			case conn := <-connectionsSub.Channel():
				nth.handleConnection(conn)
			case <-connectionsSub.WaitClose():
				log.Println("Network closed. Stopping transaction observer.")
				nth.StopObservingTransactions()
				return true
			case <-cancellation.Done():
				log.Println("Transaction observer cancelled. Stopping transaction observer.")
				nth.StopObservingTransactions()
				return false
			}
		}
	}).Finally(func() {
		connectionsSub.Unsubscribe()
	})

	nth.listenTask = task
	go nth.listenTask.Await()
}

func (nth *NetworkTransactionsHandler) handleConnection(conn network.Connection) {
	tunnel := conn.GetTunnel()
	nth.recentByTunnel.Set(tunnel, newRegistry())

	task := longtask.Run(func(cancellation context.Context) any {
		nth.listenTunnel(tunnel)
		return true
	}).Finally(func() {
		nth.recentByTunnel.Delete(tunnel)
	})

	go task.Await()
}

func (nth *NetworkTransactionsHandler) StopObservingTransactions() {
	log.Println("Stopping transaction observer.")
	nth.listenTask.Cancel()

	for sub := range nth.tunnelSubs.All() {
		sub.Unsubscribe()
	}
}

func (nth *NetworkTransactionsHandler) listenTunnel(tunnel tunnel.DuplexTunnel) {
	tunnelSub := tunnel.Subscribe()
	defer tunnelSub.Unsubscribe()

	nth.tunnelSubs.Add(tunnelSub)
	defer nth.tunnelSubs.Remove(tunnelSub)
	log.Println("Listening tunnel")

	for {
		select {
		case data := <-tunnelSub.Channel():
			if len(data) == 0 {
				continue
			}

			var reg *TransactionRegistry
			var found bool
			if reg, found = nth.recentByTunnel.Get(tunnel); !found {
				reg = newRegistry()
				nth.recentByTunnel.Set(tunnel, reg)
			}

			reg.Put(hashMessage(data), time.Now())
			log.Println("Received transaction ", hashMessage(data))
			go nth.handleMessage(data)
		case <-tunnelSub.WaitClose():
			return
		}
	}
}

func hashMessage(data []byte) string {
	sha := sha256.New()
	sha.Write(data)
	return hex.EncodeToString(sha.Sum(nil))
}

func (nth *NetworkTransactionsHandler) handleMessage(data []byte) {
	if _, seen := nth.recentTransactions.Get(hashMessage(data)); seen {
		return
	}

	encoder := transactions.NewTransactionAsn1Encoder()
	tran, err := encoder.DecodeTransaction(data)
	if err != nil {
		log.Println("Error decoding transaction: ", err)
		return
	}
	go nth.PublishTransaction(tran)

	err = nth.workflow.Process(tran)

	if err != nil {
		log.Println("Error processing transaction: ", err)
	}
}

func (nth *NetworkTransactionsHandler) PublishTransaction(tran transactions.Transaction) {
	log.Println("Publishing transaction with hash ", hashMessage(tran.GetRaw()))
	for sub := range nth.net.GetConnections().Current().All() {
		tunnel := sub.GetTunnel()
		reg, found := nth.recentByTunnel.Get(tunnel)
		if !found {
			reg = newRegistry()
			nth.recentByTunnel.Set(tunnel, reg)
		}

		if _, seen := reg.Get(hashMessage(tran.GetRaw())); seen {
			continue
		}

		go tunnel.Send(tran.GetRaw())
	}
}
