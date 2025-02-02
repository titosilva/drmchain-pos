package gossiptunnel

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/internal/patterns/tunnel"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cqueue"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"

	"github.com/titosilva/drmchain-pos/network/encodings"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip/gossiptunnel/internal/gossipseal"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip/internal/messages"
	"github.com/titosilva/drmchain-pos/network/internal/connections/sessions"
	"golang.org/x/crypto/hkdf"
)

type GossipTunnel struct {
	selfSealer           *gossipseal.GossipSealer
	peerSealer           *gossipseal.GossipSealer
	defaultAnswerTimeout time.Duration

	session *sessions.Session
	conn    net.Conn
	connMux *sync.RWMutex

	messagesToSend     *cqueue.CQueue[[]byte]
	receivedObservable *observable.Observable[[]byte]
	controlObservable  *observable.Observable[[]byte]
	closedObservable   *observable.Observable[struct{}]
	isClosed           bool
}

// New creates a new GossipTunnel.
func New(conn net.Conn, session *sessions.Session, selfId identity.PrivateIdentity) (*GossipTunnel, error) {
	selfSealer, err := createSealer(session, selfId.GetTag())
	if err != nil {
		return nil, err
	}

	peerSealer, err := createSealer(session, session.Peer.Id.GetTag())
	if err != nil {
		return nil, err
	}

	gt := &GossipTunnel{
		selfSealer:           selfSealer,
		peerSealer:           peerSealer,
		defaultAnswerTimeout: 5 * time.Second,
		connMux:              &sync.RWMutex{},
		isClosed:             false,

		session:            session,
		conn:               conn,
		messagesToSend:     cqueue.New[[]byte](),
		receivedObservable: observable.New[[]byte](),
		controlObservable:  observable.New[[]byte](),
		closedObservable:   observable.New[struct{}](),
	}

	return gt, nil
}

func (g *GossipTunnel) SendInit() error {
	g.conn.SetWriteDeadline(time.Now().Add(g.defaultAnswerTimeout * time.Second))
	err := g.write([]byte(g.session.Id))
	if err != nil {
		return err
	}

	g.conn.SetReadDeadline(time.Now().Add(g.defaultAnswerTimeout * time.Second))
	bs, err := g.read()
	if err != nil {
		return err
	}

	var shell messages.MessageSeal
	if err = encodings.Decode(bs, &shell); err != nil {
		return err
	}

	if shell.SessionId != g.session.Id {
		return errors.New("wrong session id in seal")
	}

	data, err := g.peerSealer.Unseal(&shell)
	if err != nil {
		return err
	}

	if string(data) != g.session.Id {
		return errors.New("wrong session id")
	}

	g.peerSealer.Update()
	g.session.MarkConnected()
	log.Println("connection established with session ", g.session.Id)

	return nil
}

func (g *GossipTunnel) AnswerInit() error {
	// TODO: this should also challenge the peer
	sealed, err := g.selfSealer.Seal([]byte(g.session.Id), messages.SealTypeData)
	if err != nil {
		return err
	}

	data, err := encodings.Encode(sealed)
	if err != nil {
		return err
	}

	g.conn.SetWriteDeadline(time.Now().Add(g.defaultAnswerTimeout * time.Second))
	g.write(data)

	return nil
}

func (g *GossipTunnel) Start() {
	go g.listenConnectionLoop()
	go g.sendMessagesLoop()
}

func createSealer(session *sessions.Session, senderId string) (*gossipseal.GossipSealer, error) {
	keyGen := hkdf.Expand(sha256.New, session.KeySeed, []byte(senderId))

	sealer, err := gossipseal.New(session.Id, keyGen)
	if err != nil {
		return nil, err
	}

	return sealer, nil
}

// Close implements tunnel.WritableTunnel.
func (g *GossipTunnel) Close() error {
	g.connMux.Lock()
	defer g.connMux.Unlock()
	if g.isClosed {
		return nil
	}

	if err := g.conn.Close(); err != nil {
		return err
	}

	g.closedObservable.Notify(struct{}{})

	g.receivedObservable.Close()
	g.controlObservable.Close()
	g.closedObservable.Close()

	g.isClosed = true
	return nil
}

// Notify implements tunnel.WritableTunnel.
func (g *GossipTunnel) Notify(data []byte) error {
	g.receivedObservable.Notify(data)
	return nil
}

// Subscribe implements tunnel.WritableTunnel.
func (g *GossipTunnel) Subscribe() *observable.Subscription[[]byte] {
	return g.receivedObservable.Subscribe()
}

// Send implements tunnel.WritableTunnel.
func (g *GossipTunnel) Send(data []byte) error {
	g.messagesToSend.Enqueue(data)
	return nil
}

// WaitClose implements tunnel.WritableTunnel.
func (g *GossipTunnel) WaitClose() <-chan struct{} {
	return g.closedObservable.Subscribe().Channel()
}

func (g *GossipTunnel) listenConnectionLoop() {
	for {
		g.conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		buf, err := g.read()
		if err != nil {
			if g.isClosed {
				return
			}

			if err == io.EOF {
				g.Close()
				return
			}

			log.Println("failed to read from connection: ", err)
			continue
		}

		if buf == nil {
			continue
		}

		var sealed messages.MessageSeal
		if err = encodings.Decode(buf, &sealed); err != nil {
			log.Println("failed to decode message seal: ", err)
			continue
		}

		sealer := g.peerSealer
		if sealed.Type == messages.SealTypeControl {
			sealer = g.selfSealer
		}

		if sealed.Sequence < sealer.GetCurrentSeq() {
			log.Println("invalid sequence - unrecoverable. Sending error and closing connection")
			g.sendError(messages.ErrorTypeWrongSeq, sealed.Sequence)
			g.Close()
			continue
		}

		if sealed.SessionId != g.session.Id {
			log.Println("wrong session id")
			g.sendError(messages.ErrorTypeWrongSession, sealed.Sequence)
			continue
		}

		// TODO: this allows an attacker to flood the connection with messages, because seq is not verified in the MAC yet. Rethink this.
		if sealed.Sequence > sealer.GetCurrentSeq() {
			log.Println("invalid sequence - recoverable. Trying to recover")

			if err := g.peerSealer.UpdateToSeq(sealed.Sequence); err != nil {
				log.Println("error during recovery", err)
				g.sendError(messages.ErrorTypeInternalError, sealed.Sequence)
				continue
			}

			log.Println("recovery successful")
			continue
		}

		data, err := sealer.Unseal(&sealed)
		if err != nil {
			log.Println("failed to unseal message: ", err)
			g.sendError(messages.ErrorTypeInvalidSeal, sealed.Sequence)
			continue
		}

		sealer.Update()
		if sealed.Type == messages.SealTypeData {
			g.sendSuccess(sealed.Sequence)
			g.receivedObservable.Notify(data)
		} else {
			g.controlObservable.Notify(data)
		}
	}
}

func (g *GossipTunnel) sendMessagesLoop() {
	for {
		if g.isClosed {
			return
		}

		data, ok := g.messagesToSend.Dequeue()
		if !ok {
			continue
		}

		sub := g.controlObservable.Subscribe()
		err := g.immediateSend(data, g.selfSealer, messages.SealTypeData)

		if err != nil {
			log.Println("failed to send message: ", err)
			g.messagesToSend.Enqueue(data)
			sub.Unsubscribe()
			continue
		}

		control, err := g.waitControlMessageForSeq(sub, g.selfSealer.GetCurrentSeq()-1)
		if err != nil {
			log.Println("failed to wait for answer message: ", err)
			g.messagesToSend.Enqueue(data)
			sub.Unsubscribe()
			continue
		}

		if !control.Succeed {
			log.Println("message not accepted. Retrying. Received error ", control.ErrorType)
			g.messagesToSend.Enqueue(data)
			sub.Unsubscribe()
			continue
		}
	}
}

func (g *GossipTunnel) waitControlMessageForSeq(sub *observable.Subscription[[]byte], seq int) (messages.ControlMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.defaultAnswerTimeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return messages.ControlMessage{}, errors.New("timeout waiting for response")
		case msg := <-sub.Channel():
			var control messages.ControlMessage
			if err := encodings.Decode(msg, &control); err != nil {
				continue
			}

			if control.MessageSeq == seq {
				return control, nil
			}
		}
	}
}

func (g *GossipTunnel) sendSuccess(messageSeq int) {
	controlMessage := messages.ControlMessage{
		Succeed:    true,
		MessageSeq: messageSeq,
	}

	encoded, err := encodings.Encode(controlMessage)
	if err != nil {
		log.Println("failed to encode success message: ", err)
	}

	err = g.immediateSend(encoded, g.peerSealer, messages.SealTypeControl)
	if err != nil {
		log.Println("failed to send success message: ", err)
	}
}

func (g *GossipTunnel) sendError(errorType string, messageSeq int) {
	controlMessage := messages.ControlMessage{
		Succeed:    false,
		ErrorType:  errorType,
		MessageSeq: messageSeq,
	}

	encoded, err := encodings.Encode(controlMessage)
	if err != nil {
		log.Println("failed to encode success message: ", err)
	}

	err = g.immediateSend(encoded, g.peerSealer, messages.SealTypeControl)
	if err != nil {
		log.Println("failed to send error message: ", err)
	}
}

func (g *GossipTunnel) immediateSend(data []byte, sealer *gossipseal.GossipSealer, sealType string) error {
	sealed, err := sealer.Seal(data, sealType)
	if err != nil {
		return err
	}

	encoded, err := encodings.Encode(*sealed)
	if err != nil {
		return err
	}

	err = g.write(encoded)
	if err != nil {
		if g.isClosed {
			return nil
		}

		if err == io.EOF {
			g.Close()
		}

		return err
	}

	return nil
}

func (g *GossipTunnel) write(data []byte) error {
	g.connMux.RLock()
	defer g.connMux.RUnlock()
	_, err := g.conn.Write(data)
	return err
}

func (g *GossipTunnel) read() ([]byte, error) {
	g.connMux.RLock()
	// TODO: handle messages bigger than 1024 bytes
	buf := make([]byte, 1024)
	n, err := g.conn.Read(buf)
	g.connMux.RUnlock()

	if err != nil {
		netErr, ok := err.(net.Error)

		if ok && netErr.Timeout() {
			return nil, nil
		}

		return nil, err
	}

	return buf[:n], nil
}

// GossipTunnel implements tunnel.Tunnel
// Static impl check
var _ tunnel.WritableTunnel = &GossipTunnel{}
