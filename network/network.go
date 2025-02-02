package network

import (
	"net"
	"strconv"

	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/patterns/tunnel"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/internal/structures"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"
)

type Network struct {
	connections ConfigurableConnections
	self        identity.PrivateIdentity
}

func Factory(diCtx *di.DIContext) *Network {
	connections := di.GetInterfaceService[ConfigurableConnections](diCtx)
	idProv := identityprovider.GetFromDI(diCtx)
	self, _ := idProv.GetIdentity()

	return &Network{
		connections: connections,
		self:        self,
	}
}

func (nw *Network) Open() error {
	return nw.connections.Init()
}

func (nw *Network) GetConnections() Connections {
	return nw.connections
}

func (nw *Network) GetSelf() identity.PrivateIdentity {
	return nw.self
}

func (nw *Network) Close() error {
	return nw.connections.Finish()
}

type Connections interface {
	Current() structures.Enumerable[Connection]
	ConnectTo(id identity.PublicIdentity, addr Address) error
	Subscribe() *observable.Subscription[Connection]
}

type ConfigurableConnections interface {
	Connections
	Init() error
	RegisterConnection(conn Connection) error
	Finish() error
}

type Connection interface {
	GetPeer() Peer
	GetTunnel() tunnel.DuplexTunnel
}

type Address struct {
	Host string
	Port int
}

func (a Address) AsUdp() *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", a.String())

	if err != nil {
		return nil
	}

	return addr
}

func (a Address) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

type Peer struct {
	Id   identity.PublicIdentity
	Addr string
}
