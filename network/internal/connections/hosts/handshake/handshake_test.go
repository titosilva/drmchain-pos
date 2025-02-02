package handshake_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/internal/di/defaultdi"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/handshake"
	"github.com/titosilva/drmchain-pos/network/networkdi"
)

func Test__TwoHosts__Connecting(t *testing.T) {
	diCtx := defaultdi.ConfigureDefaultDI()
	diCtx = networkdi.AddNetworkServices(diCtx)
	h1 := handshake.GetFromDI(diCtx)
	h2 := handshake.GetFromDI(diCtx)

	if h1 == nil || h2 == nil {
		t.Error("Error getting handshake from DI")
	}

	if err := h1.Listen("localhost:52001"); err != nil {
		t.Error("Error listening: ", err)
	}
	defer h1.Close()

	if err := h2.Listen("localhost:52002"); err != nil {
		t.Error("Error listening: ", err)
	}
	defer h2.Close()

	idProvider := identityprovider.GetFromDI(diCtx)
	id, err := idProvider.GetIdentity()
	if err != nil {
		t.Error("Error getting identity: ", err)
	}

	peer := network.Peer{
		Id:   id,
		Addr: "localhost:52002",
	}

	if _, err := h1.ConnectTo(peer); err != nil {
		t.Error("Error connecting: ", err)
	}
}
