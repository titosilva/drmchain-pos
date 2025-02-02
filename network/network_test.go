package network_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/di/defaultdi"
	"github.com/titosilva/drmchain-pos/network"

	"github.com/titosilva/drmchain-pos/network/networkconfig"
	"github.com/titosilva/drmchain-pos/network/networkdi"
)

func Test__TwoNetworkHostsConnecting(t *testing.T) {
	nw1, err := openNetwork("localhost:2503", "localhost:2504")
	if err != nil {
		t.Fatalf("Error opening network 1: %s", err)
	}
	defer nw1.Close()

	nw2, err := openNetwork("localhost:2505", "localhost:2506")
	if err != nil {
		t.Fatalf("Error opening network 2: %s", err)
	}
	defer nw2.Close()

	if err := nw1.GetConnections().ConnectTo(nw2.GetSelf(), network.Address{Host: "localhost", Port: 2505}); err != nil {
		t.Fatalf("Error connecting network 1 to network 2: %s", err)
	}

	if nw2.GetConnections().Current().Count() != 1 {
		t.Fatalf("Expected 1 connection, got %d", nw2.GetConnections().Current().Count())
	}

	found := false
	for conn := range nw2.GetConnections().Current().All() {
		if conn.GetPeer().Id.GetTag() == nw1.GetSelf().GetTag() {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("Connection not found in network 2")
	}
}

func newDI(handshakeHost string, gossipHost string) *di.DIContext {
	diCtx := defaultdi.ConfigureDefaultDI()
	diCtx = networkdi.AddNetworkServices(diCtx)

	config := networkconfig.GetFromDI(diCtx)
	config.HandshakeHost = handshakeHost
	config.GossipHost = gossipHost

	return diCtx
}

func openNetwork(handshakeHost string, gossipHost string) (*network.Network, error) {
	diCtx := newDI(handshakeHost, gossipHost)
	net := di.GetService[network.Network](diCtx)

	if err := net.Open(); err != nil {
		return nil, err
	}

	return net, nil
}
