package gossiptunnel_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/internal/di/defaultdi"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip/gossiptunnel"
	"github.com/titosilva/drmchain-pos/network/internal/connections/sessions"
)

func Test__TwoGossipTunnels__ShouldExchangeMessagesCorrectly(t *testing.T) {
	diCtx := defaultdi.ConfigureDefaultDI()
	idProv := identityprovider.GetFromDI(diCtx)
	selfId, err := idProv.GetIdentity()
	if err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "localhost:4321")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	conn1, conn2, err := makeSelfConnections(ln)
	if err != nil {
		t.Fatal(err)
	}

	tun1, err := gossiptunnel.New(conn1, makeSession(selfId, conn1), selfId)
	if err != nil {
		t.Fatal(err)
	}
	defer tun1.Close()

	tun2, err := gossiptunnel.New(conn2, makeSession(selfId, conn2), selfId)
	if err != nil {
		t.Fatal(err)
	}
	defer tun2.Close()

	c := make(chan []byte)
	go func(c chan []byte) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		sub := tun2.Subscribe()
		for {
			select {
			case <-ctx.Done():
				c <- nil
				return
			case <-tun2.WaitClose():
				c <- nil
				cancel()
				return
			case msg := <-sub.Channel():
				c <- msg
				return
			}
		}
	}(c)

	tun1.Start()
	tun2.Start()

	tun1.Send([]byte("hello"))

	received := <-c
	if received == nil {
		t.Fatal("timeout")
	}

	if string(received) != "hello" {
		t.Fatal("unexpected message", string(received))
	}

	tun1.Close()
	tun2.Close()
}

func makeSelfConnections(ln net.Listener) (net.Conn, net.Conn, error) {
	// makes a server and a client connection
	// with tcp. Must have buffering
	c := make(chan net.Conn)
	go func(c chan net.Conn) {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		c <- conn
	}(c)

	conn, err := net.Dial("tcp", "localhost:4321")
	if err != nil {
		return nil, nil, err
	}

	return <-c, conn, nil
}

func makeSession(selfId identity.PrivateIdentity, conn net.Conn) *sessions.Session {
	return sessions.NewSession("session", []byte("keyseed"), network.Peer{
		Id:   selfId,
		Addr: conn.RemoteAddr().String(),
	})
}
