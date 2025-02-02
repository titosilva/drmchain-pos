package consensusnetwork

import (
	"context"
	"log"
	"time"

	"github.com/titosilva/drmchain-pos/consensus"
	"github.com/titosilva/drmchain-pos/consensus/messages"
	"github.com/titosilva/drmchain-pos/internal/patterns/longtask"
	"github.com/titosilva/drmchain-pos/internal/patterns/tunnel"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cbag"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/clru"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cmap"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"
	"github.com/titosilva/drmchain-pos/internal/utils/cryptutil"
	"github.com/titosilva/drmchain-pos/network"
	"github.com/titosilva/drmchain-pos/network/encodings"
)

type MessageRegistry = clru.Cache[string, time.Time]

func newRegistry() *MessageRegistry {
	return clru.New[string, time.Time](256)
}

type ConsensusHost struct {
	net            *network.Network
	listenTask     *longtask.LongTask[any]
	recentMessages *MessageRegistry
	recentByTunnel *cmap.CMap[tunnel.DuplexTunnel, *MessageRegistry]
	tunnelSubs     *cbag.CBag[*observable.Subscription[[]byte]]

	currentContext *consensus.ConsensusContext
}

func NewHost() *ConsensusHost {
	return &ConsensusHost{}
}

func (ch *ConsensusHost) ObserveMessages() {
	connectionsSub := ch.net.GetConnections().Subscribe()

	for conn := range ch.net.GetConnections().Current().All() {
		ch.handleConnection(conn)
	}

	task := longtask.Run(func(cancellation context.Context) any {
		log.Println("Starting message observer.")

		for {
			select {
			case conn := <-connectionsSub.Channel():
				ch.handleConnection(conn)
			case <-connectionsSub.WaitClose():
				log.Println("Network closed. Stopping message observer.")
				ch.StopObservingMessages()
				return true
			case <-cancellation.Done():
				log.Println("Message observer cancelled. Stopping message observer.")
				ch.StopObservingMessages()
				return false
			}
		}
	}).Finally(func() {
		connectionsSub.Unsubscribe()
	})

	ch.listenTask = task
	go ch.listenTask.Await()
}

func (ch *ConsensusHost) handleConnection(conn network.Connection) {
	tunnel := conn.GetTunnel()
	ch.recentByTunnel.Set(tunnel, newRegistry())

	task := longtask.Run(func(cancellation context.Context) any {
		ch.listenTunnel(tunnel)
		return true
	}).Finally(func() {
		ch.recentByTunnel.Delete(tunnel)
	})

	go task.Await()
}

func (ch *ConsensusHost) StopObservingMessages() {
	log.Println("Stopping message observer.")
	ch.listenTask.Cancel()

	for sub := range ch.tunnelSubs.All() {
		sub.Unsubscribe()
	}
}

func (ch *ConsensusHost) listenTunnel(tunnel tunnel.DuplexTunnel) {
	tunnelSub := tunnel.Subscribe()
	defer tunnelSub.Unsubscribe()

	ch.tunnelSubs.Add(tunnelSub)
	defer ch.tunnelSubs.Remove(tunnelSub)
	log.Println("Listening tunnel")

	for {
		select {
		case data := <-tunnelSub.Channel():
			if len(data) == 0 {
				continue
			}

			var reg *MessageRegistry
			var found bool
			if reg, found = ch.recentByTunnel.Get(tunnel); !found {
				reg = newRegistry()
				ch.recentByTunnel.Set(tunnel, reg)
			}

			reg.Put(cryptutil.HashToString(data), time.Now())
			log.Println("Received consensus message ", cryptutil.HashToString(data))
			go ch.handleMessage(data)
		case <-tunnelSub.WaitClose():
			return
		}
	}
}

func (ch *ConsensusHost) handleMessage(data []byte) {
	if _, seen := ch.recentMessages.Get(cryptutil.HashToString(data)); seen {
		return
	}

	var msg messages.ConsensusShell
	if err := encodings.Decode(data, &msg); err != nil {
		return
	}

	if !ch.currentContext.IsSame(msg.Context) {
		return
	}

	go ch.PropagateMessage(msg)
}

func (ch *ConsensusHost) PropagateMessage(msg messages.ConsensusShell) {
	msgHash := cryptutil.HashToString(msg.GetRaw())
	log.Println("Publishing consensus message with hash ", msgHash)

	for sub := range ch.net.GetConnections().Current().All() {
		tunnel := sub.GetTunnel()
		reg, found := ch.recentByTunnel.Get(tunnel)
		if !found {
			reg = newRegistry()
			ch.recentByTunnel.Set(tunnel, reg)
		}

		if _, seen := reg.Get(msgHash); seen {
			continue
		}

		go tunnel.Send(msg.GetRaw())
	}
}
