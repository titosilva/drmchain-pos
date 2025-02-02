package gossip

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/internal/di"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/internal/utils/errorutil"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/internal/connections"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip/gossiptunnel"
	"github.com/titosilva/drmchain-pos/network/internal/connections/sessions"
)

type GossipHost struct {
	address string

	tcpServer    *net.TCPListener
	cancellation *context.Context
	cancel       context.CancelFunc
	selfId       identity.PrivateIdentity

	sessions  *sessions.Memory
	onConnect func(network.Connection)
}

func Factory(diCtx *di.DIContext) connections.Gossiper {
	cancellation, cancel := context.WithCancel(context.Background())

	sessions := sessions.GetFromDI(diCtx)
	idProvider := identityprovider.GetFromDI(diCtx)
	identity, _ := idProvider.GetIdentity()

	return &GossipHost{
		selfId:       identity,
		cancellation: &cancellation,
		cancel:       cancel,
		sessions:     sessions,
	}
}

func GetFromDI(diCtx *di.DIContext) connections.Gossiper {
	return di.GetInterfaceService[connections.Gossiper](diCtx)
}

func (g *GossipHost) GetPeer() network.Peer {
	return network.Peer{
		Id:   g.selfId,
		Addr: g.address,
	}
}

func (g *GossipHost) GetSessions() *sessions.Memory {
	return g.sessions
}

func (g *GossipHost) Listen(addr string, onConnect func(network.Connection)) error {
	g.address = addr
	g.onConnect = onConnect
	tcpAddr, err := net.ResolveTCPAddr("tcp", g.address)
	if err != nil {
		return errorutil.WithInner("failed to resolve tcp address", err)
	}

	tcpServer, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return errorutil.WithInner("failed to listen on tcp address", err)
	}

	g.tcpServer = tcpServer
	go g.acceptConnections()
	return nil
}

func (g *GossipHost) acceptConnections() {
	for {
		select {
		case <-(*g.cancellation).Done():
			return
		default:
			g.tcpServer.SetDeadline(time.Now().Add(100 * time.Millisecond))
			conn, err := g.tcpServer.Accept()

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Println("failed to accept connection", err)
				}
				continue
			}

			go g.handleConnection(conn)
		}
	}
}

func (g *GossipHost) handleConnection(conn net.Conn) {
	log.Printf("new connection accepted from %s Reading message shell\n", conn.RemoteAddr().String())
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("failed to read connection sessionId", err)
		return
	}

	sessionId := string(buf[:n])

	session := g.sessions.GetSession(sessionId)
	if session == nil {
		log.Println("session ", sessionId, " not found")
		return
	}

	if session.WasConnected() {
		log.Println("session ", sessionId, " already connected")
		return
	}

	tunnel, err := gossiptunnel.New(conn, session, g.selfId)
	if err != nil {
		log.Println("failed to create tunnel", err)
		return
	}
	log.Println("new tunnel created for session ", session.Id)

	connection := NewGossipConnection(session.Peer, tunnel)
	g.onConnect(connection)
	err = tunnel.AnswerInit()
	if err != nil {
		log.Println("failed to receive init message", err)
		return
	}

	tunnel.Start()
	log.Println("connected to ", session.Peer.Addr, " with session ", session.Id, " and tag ", session.Peer.Id.GetTag())
}

func (g *GossipHost) ConnectTo(session *sessions.Session) (network.Connection, error) {
	peer := session.Peer
	conn, err := net.Dial("tcp", peer.Addr)
	if err != nil {
		return nil, errorutil.WithInner("failed to dial peer", err)
	}

	log.Println("TCP connected to ", peer.Addr)

	tunnel, err := gossiptunnel.New(conn, session, g.selfId)
	if err != nil {
		return nil, errorutil.WithInner("failed to create tunnel", err)
	}

	connection := NewGossipConnection(peer, tunnel)
	g.onConnect(connection)
	err = tunnel.SendInit()
	if err != nil {
		return nil, errorutil.WithInner("failed to send init message", err)
	}

	session.MarkConnected()
	tunnel.Start()
	return connection, nil
}

func (g *GossipHost) Close() error {
	g.cancel()
	g.tcpServer.Close()
	return nil
}

// GossipHost implements connection.Gossiper
var _ connections.Gossiper = &GossipHost{}
