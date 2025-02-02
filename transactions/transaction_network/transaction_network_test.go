package transactionnetwork_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/identity/signatures"
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/di/defaultdi"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/networkconfig"
	"github.com/titosilva/drmchain-pos/network/networkdi"
	"github.com/titosilva/drmchain-pos/transactions"
	"github.com/titosilva/drmchain-pos/transactions/notifier"
	transactionnetwork "github.com/titosilva/drmchain-pos/transactions/transaction_network"
	"github.com/titosilva/drmchain-pos/transactions/transactionsdi"
)

func Test__TransactionPropagationWith3Hosts(t *testing.T) {
	nw1, di1, err := openNetwork("localhost:2503", "localhost:2504")
	if err != nil {
		t.Fatalf("Error opening network 1: %s", err)
	}
	defer nw1.Close()

	nw2, di2, err := openNetwork("localhost:2505", "localhost:2506")
	if err != nil {
		t.Fatalf("Error opening network 2: %s", err)
	}
	defer nw2.Close()

	nw3, di3, err := openNetwork("localhost:2507", "localhost:2508")
	if err != nil {
		t.Fatalf("Error opening network 3: %s", err)
	}
	defer nw3.Close()

	if err := nw1.GetConnections().ConnectTo(nw2.GetSelf(), network.Address{Host: "localhost", Port: 2505}); err != nil {
		t.Fatalf("Error connecting network 1 to network 2: %s", err)
	}

	if err := nw2.GetConnections().ConnectTo(nw3.GetSelf(), network.Address{Host: "localhost", Port: 2507}); err != nil {
		t.Fatalf("Error connecting network 2 to network 3: %s", err)
	}

	if nw2.GetConnections().Current().Count() != 2 {
		t.Fatalf("Expected 2 connections, got %d", nw2.GetConnections().Current().Count())
	}

	if nw3.GetConnections().Current().Count() != 1 {
		t.Fatalf("Expected 1 connection, got %d", nw3.GetConnections().Current().Count())
	}
	t.Log("All networks connected")

	obs1 := transactionnetwork.GetFromDI(di1)
	obs1.ObserveTransactions()
	defer obs1.StopObservingTransactions()

	obs2 := transactionnetwork.GetFromDI(di2)
	obs2.ObserveTransactions()
	defer obs2.StopObservingTransactions()

	obs3 := transactionnetwork.GetFromDI(di3)
	obs3.ObserveTransactions()
	defer obs3.StopObservingTransactions()

	t.Log("All networks observing transactions")

	notifier2 := notifier.GetFromDI(di2)
	sub2 := notifier2.Subscribe()
	defer sub2.Unsubscribe()

	notifier3 := notifier.GetFromDI(di3)
	sub3 := notifier3.Subscribe()
	defer sub3.Unsubscribe()

	id1, err := identityprovider.GetFromDI(di1).GetIdentity()
	if err != nil {
		t.Fatalf("Error getting identity 1: %s", err)
	}

	signature, err := signatures.Sign(id1, []byte("Hello, world!"))
	if err != nil {
		t.Fatalf("Error signing transaction: %s", err)
	}

	tranToSend := transactions.TransactionShape{
		SourceTag: id1.GetTag(),
		Signature: signature,
		Content:   []byte("Hello, world!"),
	}

	t.Log("Sending transaction")
	obs1.PublishTransaction(&tranToSend)

	t.Log("Waiting for transaction in network 2")
	tran, ok := sub2.WaitNextWithTimeoutMs(500)
	if !ok {
		t.Fatalf("Timeout waiting for transaction")
	}

	if tran == nil {
		t.Fatalf("Expected transaction, got nil")
	}

	tran, ok = sub3.WaitNextWithTimeoutMs(500)
	if !ok {
		t.Fatalf("Timeout waiting for transaction")
	}

	if tran == nil {
		t.Fatalf("Expected transaction, got nil")
	}

	if tran.GetSource().GetTag() != id1.GetTag() {
		t.Fatalf("Expected transaction from id1, got %s", tran.GetSource().GetTag())
	}

	if string(tran.GetContent()) != "Hello, world!" {
		t.Fatalf("Expected content 'Hello, world!', got %s", string(tran.GetContent()))
	}

	if !tran.IsValidSignature() {
		t.Fatalf("Invalid signature")
	}

	// Observer 2 should not receive the transaction again
	if _, ok := sub2.WaitNextWithTimeoutMs(1000); ok {
		t.Fatalf("Expected timeout, got transaction")
	}
}

func newDI(handshakeHost string, gossipHost string) *di.DIContext {
	diCtx := defaultdi.ConfigureDefaultDI()
	diCtx = networkdi.AddNetworkServices(diCtx)
	diCtx = transactionsdi.AddTransactionServices(diCtx)

	config := networkconfig.GetFromDI(diCtx)
	config.HandshakeHost = handshakeHost
	config.GossipHost = gossipHost

	return diCtx
}

func openNetwork(handshakeHost string, gossipHost string) (*network.Network, *di.DIContext, error) {
	diCtx := newDI(handshakeHost, gossipHost)
	net := di.GetService[network.Network](diCtx)

	if err := net.Open(); err != nil {
		return nil, nil, err
	}

	return net, diCtx, nil
}
