package networkdi

import (
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/internal/connections"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/handshake"
	"github.com/titosilva/drmchain-pos/network/internal/connections/sessions"
	"github.com/titosilva/drmchain-pos/network/networkconfig"
)

func AddNetworkServices(diCtx *di.DIContext) *di.DIContext {
	di.AddSingleton(diCtx, sessions.Factory)
	di.AddInterfaceFactory(diCtx, handshake.Factory)
	di.AddInterfaceFactory(diCtx, gossip.Factory)
	di.AddInterfaceFactory(diCtx, connections.Factory)
	di.AddSingleton(diCtx, network.Factory)
	di.AddSingleton(diCtx, networkconfig.Factory)

	return diCtx
}
