package handshake

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"log"
	"net"
	"slices"
	"time"

	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/identity/keyexchange"
	"github.com/titosilva/drmchain-pos/identity/signatures"
	"github.com/titosilva/drmchain-pos/internal/di"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/observable"
	"github.com/titosilva/drmchain-pos/internal/utils/errorutil"
	"github.com/titosilva/drmchain-pos/network"
	"golang.org/x/crypto/hkdf"

	"github.com/titosilva/drmchain-pos/network/encodings"
	"github.com/titosilva/drmchain-pos/network/internal/connections"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/handshake/internal/messages"
	"github.com/titosilva/drmchain-pos/network/internal/connections/sessions"
	config "github.com/titosilva/drmchain-pos/network/networkconfig"
)

type HandshakeHost struct {
	selfId             identity.PrivateIdentity
	cancellation       *context.Context
	cancel             context.CancelFunc
	received           *observable.Observable[UdpMessage]
	sessions           *sessions.Memory
	configuration      *config.NetworkConfig
	defaultTimeoutSecs int

	address   string
	tcpAddr   string
	udpServer *net.UDPConn
}

type HandshakeData struct {
	PeerId       identity.PublicIdentity
	PeerAddr     *net.UDPAddr
	Subscription *observable.Subscription[UdpMessage]
}

type HandshakePeerData struct {
	Id   identity.PublicIdentity
	Addr *net.UDPAddr
}

type UdpMessage struct {
	Data []byte
	Addr *net.UDPAddr
}

func Factory(diCtx *di.DIContext) connections.Handshaker {
	idProvider := identityprovider.GetFromDI(diCtx)
	selfId, err := idProvider.GetIdentity()
	session := sessions.GetFromDI(diCtx)
	config := config.GetFromDI(diCtx)
	if err != nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	h := &HandshakeHost{
		selfId:             selfId,
		cancellation:       &ctx,
		cancel:             cancel,
		received:           observable.New[UdpMessage](),
		sessions:           session,
		configuration:      config,
		tcpAddr:            config.GossipHost,
		defaultTimeoutSecs: 5,
	}

	return h
}

func GetFromDI(diCtx *di.DIContext) connections.Handshaker {
	return di.GetInterfaceService[connections.Handshaker](diCtx)
}

func (h *HandshakeHost) ConnectTo(peer network.Peer) (*sessions.Session, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", peer.Addr)
	if err != nil {
		return nil, errorutil.WithInner("failed to resolve udp address: ", err)
	}

	data := HandshakeData{
		PeerId:       peer.Id,
		PeerAddr:     udpAddr,
		Subscription: h.received.Subscribe(),
	}
	defer data.Subscription.Unsubscribe()

	// Hello -> Peer
	nonce := generateNonce()
	msg := messages.HelloMessage{
		SrcTag:  h.selfId.GetTag(),
		DstTag:  data.PeerId.GetTag(),
		SrcAddr: h.address,
		Nonce:   nonce,
	}

	if err := h.send("hello", msg, data.PeerAddr); err != nil {
		return nil, errorutil.WithInner("failed to send hello message: ", err)
	}

	// Challenge <- Peer
	var challengeMsg messages.ChallengeMessage
	if err := h.wait(&challengeMsg, data); err != nil {
		return nil, errorutil.WithInner("failed to receive challenge message: ", err)
	}

	if !slices.Equal(challengeMsg.Nonce, nonce) {
		return nil, errors.New("first nonce mismatch")
	}

	// Answer -> Peer
	ephKey, err := keyexchange.GenerateEphemeralKey()
	if err != nil {
		return nil, errorutil.WithInner("failed to generate ephemeral key: ", err)
	}

	acceptNonce := generateNonce()
	ansMsg := messages.AnswerMessage{
		EphKey:         ephKey.PublicKey().Bytes(),
		ChallengeNonce: challengeMsg.ChallengeNonce,
		AcceptNonce:    acceptNonce,
	}

	if err = h.send("answer", ansMsg, data.PeerAddr); err != nil {
		return nil, errorutil.WithInner("failed to send answer message: ", err)
	}

	// Accepted <- Peer
	var acceptedMsg messages.AcceptedMessage
	if err = h.wait(&acceptedMsg, data); err != nil {
		return nil, errorutil.WithInner("failed to receive accepted message: ", err)
	}

	if !slices.Equal(acceptedMsg.AcceptNonce, acceptNonce) {
		return nil, errors.New("accept nonce mismatch")
	}

	secret, err := keyexchange.DeriveFromPublicIdentity(peer.Id, ephKey)
	if err != nil {
		return nil, errorutil.WithInner("failed to derive secret: ", err)
	}

	if !signatures.Verify(peer.Id, secret, acceptedMsg.SecretSignature) {
		return nil, errors.New("failed to verify secret signature")
	}

	salt := append(challengeMsg.Nonce, challengeMsg.ChallengeNonce...)
	keySeed := hkdf.Extract(sha256.New, secret, salt)

	sessionPeer := network.Peer{
		Id:   data.PeerId,
		Addr: acceptedMsg.TcpAddr,
	}
	session := sessions.NewSession(acceptedMsg.SessionId, keySeed, sessionPeer)
	h.sessions.RegisterSession(session)

	// Derive secret
	return session, nil
}

func (h *HandshakeHost) receiveHandshake(helloMsg messages.HelloMessage, addr *net.UDPAddr) {
	// Hello <- Source
	srcId, err := identity.FromTag(helloMsg.SrcTag)
	if err != nil {
		log.Println("Failed to parse peer identity: ", err)
	}

	data := HandshakeData{
		PeerId:       srcId,
		PeerAddr:     addr,
		Subscription: h.received.Subscribe(),
	}
	defer data.Subscription.Unsubscribe()

	log.Println("Received hello message from ", data.PeerId.GetTag(), " at ", data.PeerAddr.String())

	// Challenge -> Source
	challengeNonce := generateNonce()
	challengeMsg := messages.ChallengeMessage{
		Nonce:          helloMsg.Nonce,
		ChallengeNonce: challengeNonce,
	}

	if err = h.send("challenge", challengeMsg, data.PeerAddr); err != nil {
		log.Println("Failed to send challenge message: ", err)
	}

	// Answer <- Source
	var answerMsg messages.AnswerMessage
	if err = h.wait(&answerMsg, data); err != nil {
		log.Println("Failed to receive answer message: ", err)
	}

	// Derive secret
	ephKey, err := keyexchange.BytesToKey(answerMsg.EphKey)
	if err != nil {
		log.Println("Failed to parse ephemeral key: ", err)
	}

	secret, err := keyexchange.DeriveFromPrivateIdentity(h.selfId, ephKey)
	if err != nil {
		log.Println("Failed to derive secret: ", err)
	}

	peer := network.Peer{
		Id:   data.PeerId,
		Addr: "", // Will be filled in later
	}

	extractSalt := append(helloMsg.Nonce, challengeNonce...)
	keySeed := hkdf.Extract(sha256.New, secret, extractSalt)
	session := h.sessions.GenerateSession(peer, keySeed)

	// Accepted -> Source
	secretSignature, err := signatures.Sign(h.selfId, secret)
	if err != nil {
		log.Println("Failed to sign secret: ", err)
	}

	acceptedMsg := messages.AcceptedMessage{
		AcceptNonce:     answerMsg.AcceptNonce,
		SessionId:       session.Id,
		SecretSignature: secretSignature,
		TcpAddr:         h.tcpAddr,
	}

	if err = h.send("accept", acceptedMsg, data.PeerAddr); err != nil {
		log.Println("Failed to send accepted message: ", err)
	}

	log.Println("Handshake completed with ", data.PeerId.GetTag(), " Session ID: ", session.Id)
}

func (h *HandshakeHost) Listen(address string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	h.udpServer = udpConn
	h.address = address
	go h.listenLoop()
	return nil
}

func (h *HandshakeHost) listenLoop() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-(*h.cancellation).Done():
			return
		default:
			n, addr, err := h.udpServer.ReadFromUDP(buf)
			if err != nil {
				log.Println("Failed to read from udp: ", err)
				return
			}

			udpMsg := UdpMessage{
				Data: buf[:n],
				Addr: addr,
			}

			var shellMsg messages.MessageShell
			if err = encodings.Decode(udpMsg.Data, &shellMsg); err != nil {
				log.Println("Failed to decode message shell: ", err)
				continue
			}

			var helloMsg messages.HelloMessage
			if encodings.Decode(shellMsg.Data, &helloMsg) == nil {
				go h.receiveHandshake(helloMsg, addr)
				continue
			}

			h.received.Notify(udpMsg)
		}
	}
}

func (h *HandshakeHost) Close() error {
	h.cancel()
	h.received.Close()
	return nil
}

// Helper methods
func (h *HandshakeHost) send(cmd string, data any, addr *net.UDPAddr) error {
	dataBytes, err := encodings.Encode(data)
	if err != nil {
		return err
	}

	signature, err := signatures.Sign(h.selfId, dataBytes)
	if err != nil {
		return err
	}

	shell := messages.MessageShell{
		Cmd:       cmd,
		Data:      dataBytes,
		Signature: signature,
	}

	shellData, err := encodings.Encode(shell)
	if err != nil {
		return err
	}

	n, err := h.udpServer.WriteToUDP(shellData, addr)
	if err != nil {
		return err
	}

	if n != len(shellData) {
		return errors.New("failed to send all data")
	}

	return nil
}

func (h *HandshakeHost) wait(out any, data HandshakeData) error {
	sub := data.Subscription

	ctx, cancel := context.WithTimeout(*h.cancellation, time.Duration(h.defaultTimeoutSecs)*time.Second)
	defer cancel()

	var shell messages.MessageShell
	for {
		select {
		case udpMsg := <-sub.Channel():
			if udpMsg.Addr.String() != data.PeerAddr.String() {
				continue
			}

			if encodings.Decode(udpMsg.Data, &shell) != nil {
				continue
			}

			if !signatures.Verify(data.PeerId, shell.Data, shell.Signature) {
				continue
			}

			if err := encodings.Decode(shell.Data, out); err != nil {
				log.Println("Received a valid shell and signature but failed to decode message: ", err)
				continue
			}

			return nil
		case <-(*h.cancellation).Done():
		case <-ctx.Done():
			return errors.New("timed out while waiting for a message")
		}
	}
}

func generateNonce() []byte {
	// TODO: check if nonce is unique
	nonce := make([]byte, 32)
	_, _ = rand.Read(nonce)
	return nonce
}

// HandshakeHost implements connections.Handshaker
var _ connections.Handshaker = (*HandshakeHost)(nil)
