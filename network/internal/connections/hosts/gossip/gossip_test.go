package gossip_test

import (
	"crypto/rand"
	"testing"

	"github.com/titosilva/drmchain-pos/internal/di/defaultdi"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip"
	"github.com/titosilva/drmchain-pos/network/internal/connections/sessions"
	"github.com/titosilva/drmchain-pos/network/networkdi"
)

func Test__TwoGossipHostsConnecting(t *testing.T) {
	keySeed := make([]byte, 32)
	if _, err := rand.Read(keySeed); err != nil {
		t.Errorf("Error generating random key seed: %s", err)
	}

	diCtx := defaultdi.ConfigureDefaultDI()
	diCtx = networkdi.AddNetworkServices(diCtx)

	g1 := gossip.GetFromDI(diCtx).(*gossip.GossipHost)
	g2 := gossip.GetFromDI(diCtx).(*gossip.GossipHost)

	c1conns := make([]network.Connection, 0)
	onConnect1 := func(c network.Connection) {
		c1conns = append(c1conns, c)
	}
	if err := g1.Listen("localhost:55001", onConnect1); err != nil {
		t.Fatalf("Error listening on g1: %s", err)
	}
	defer g1.Close()

	c2conns := make([]network.Connection, 0)
	onConnect2 := func(c network.Connection) {
		c2conns = append(c2conns, c)
	}
	if err := g2.Listen("localhost:55002", onConnect2); err != nil {
		t.Fatalf("Error listening on g2: %s", err)
	}
	defer g2.Close()

	session1 := g1.GetSessions().GenerateSession(g2.GetPeer(), keySeed)
	session2 := sessions.NewSession(session1.Id, session1.KeySeed, g1.GetPeer())
	g2.GetSessions().RegisterSession(session2)

	conn, err := g1.ConnectTo(session1)
	if err != nil {
		t.Fatalf("Error connecting g1 to g2: %s", err)
	}

	if len(c2conns) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(c2conns))
	}

	tun1 := conn.GetTunnel()
	tun2 := c2conns[0].GetTunnel()

	sub2 := tun2.Subscribe()
	defer sub2.Close()
	tun1.Send([]byte("Hello, world!"))
	msg, ok := sub2.WaitNextWithTimeoutMs(5000)
	if !ok {
		t.Fatalf("Timeout waiting for message")
	}

	if string(msg) != "Hello, world!" {
		t.Fatalf("Expected message 'Hello, world!', got '%s'", string(msg))
	}

	tun1.Send([]byte("Goodbye, world! 2"))
	msg, ok = sub2.WaitNextWithTimeoutMs(5000)
	if !ok {
		t.Fatalf("Timeout waiting for message")
	}

	if string(msg) != "Goodbye, world! 2" {
		t.Fatalf("Expected message 'Goodbye, world! 2', got '%s'", string(msg))
	}

	sub1 := tun1.Subscribe()
	tun2.Send([]byte("Hello, world! 3"))
	msg, ok = sub1.WaitNextWithTimeoutMs(5000)
	if !ok {
		t.Fatalf("Timeout waiting for message")
	}

	if string(msg) != "Hello, world! 3" {
		t.Fatalf("Expected message 'Hello, world! 3', got '%s'", string(msg))
	}

	if err := tun1.Close(); err != nil {
		t.Fatalf("Error closing connection: %s", err)
	}

	if err := tun2.Close(); err != nil {
		t.Fatalf("Error closing connection: %s", err)
	}

	g1.Close()
	g2.Close()
}
