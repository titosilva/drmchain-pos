package connections

import (
	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/structures"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cmap"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"
	"github.com/titosilva/drmchain-pos/internal/structures/kv"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/internal/connections/sessions"
	"github.com/titosilva/drmchain-pos/network/networkconfig"
)

type Gossiper interface {
	Listen(addr string, onConnect func(network.Connection)) error
	ConnectTo(session *sessions.Session) (network.Connection, error)
	Close() error
}

type Handshaker interface {
	Listen(addr string) error // TODO: pass onSession func, similar to Gossiper.Listen
	ConnectTo(peer network.Peer) (*sessions.Session, error)
	Close() error
}

type ConnectionsImpl struct {
	connections           *cmap.CMap[identity.PublicIdentity, network.Connection]
	handshake             Handshaker
	gossip                Gossiper
	connectionsObservable *observable.Observable[network.Connection]
	configuration         *networkconfig.NetworkConfig
}

func Factory(diCtx *di.DIContext) network.ConfigurableConnections {
	handshakeHost := di.GetInterfaceService[Handshaker](diCtx)
	gossipHost := di.GetInterfaceService[Gossiper](diCtx)
	config := di.GetService[networkconfig.NetworkConfig](diCtx)

	return &ConnectionsImpl{
		connections:           cmap.New[identity.PublicIdentity, network.Connection](),
		handshake:             handshakeHost,
		gossip:                gossipHost,
		connectionsObservable: observable.New[network.Connection](),
		configuration:         config,
	}
}

func GetFromDI(diCtx *di.DIContext) *ConnectionsImpl {
	return di.GetService[ConnectionsImpl](diCtx)
}

// Init implements network.Connections.
func (c *ConnectionsImpl) Init() error {
	err := c.handshake.Listen(c.configuration.HandshakeHost)

	if err != nil {
		return err
	}

	err = c.gossip.Listen(c.configuration.GossipHost, func(conn network.Connection) {
		c.RegisterConnection(conn)
	})

	if err != nil {
		return err
	}

	return nil
}

// ConnectTo implements network.Connections.
func (c *ConnectionsImpl) ConnectTo(id identity.PublicIdentity, addr network.Address) error {
	peer := network.Peer{
		Id:   id,
		Addr: addr.AsUdp().String(),
	}

	session, err := c.handshake.ConnectTo(peer)

	if err != nil {
		return err
	}

	conn, err := c.gossip.ConnectTo(session)
	if err != nil {
		return err
	}

	return c.RegisterConnection(conn)
}

// RegisterConnection implements network.Connections.
func (c *ConnectionsImpl) RegisterConnection(conn network.Connection) error {
	c.connections.Set(conn.GetPeer().Id, conn)
	c.connectionsObservable.Notify(conn)
	return nil
}

// Subscribe implements network.Connections.
func (c *ConnectionsImpl) Subscribe() *observable.Subscription[network.Connection] {
	return c.connectionsObservable.Subscribe()
}

// Current implements network.Connections.
func (c *ConnectionsImpl) Current() structures.Enumerable[network.Connection] {
	return structures.Map(kv.GetValue, c.connections)
}

// Finish implements network.Connections.
func (c *ConnectionsImpl) Finish() error {
	err := c.handshake.Close()

	if err != nil {
		return err
	}

	c.gossip.Close()
	c.handshake.Close()
	c.connectionsObservable.Close()

	return nil
}

// Static interface impl check
var _ network.ConfigurableConnections = (*ConnectionsImpl)(nil)
