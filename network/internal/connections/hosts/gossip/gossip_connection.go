package gossip

import (
	"github.com/titosilva/drmchain-pos/internal/patterns/tunnel"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip/gossiptunnel"
)

type GossipConnection struct {
	peer   network.Peer
	tunnel *gossiptunnel.GossipTunnel
}

func NewGossipConnection(peer network.Peer, tunnel *gossiptunnel.GossipTunnel) *GossipConnection {
	return &GossipConnection{
		peer:   peer,
		tunnel: tunnel,
	}
}

// GetPeer implements network.Connection.
func (c *GossipConnection) GetPeer() network.Peer {
	return c.peer
}

// GetTunnel implements network.Connection.
func (c *GossipConnection) GetTunnel() tunnel.DuplexTunnel {
	return c.tunnel
}

// GossipConnection implements network.Connection.
var _ network.Connection = (*GossipConnection)(nil)
