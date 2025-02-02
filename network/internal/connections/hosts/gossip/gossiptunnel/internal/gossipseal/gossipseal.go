package gossipseal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"log"

	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip/internal/messages"
)

var counter int = 0

type GossipSealer struct {
	sessionId    string
	currentSeq   int
	keyGenerator io.Reader
	currentKeys  []byte
	currentIv    []byte
	sealerCount  int
}

func New(sessionId string, keyGenerator io.Reader) (*GossipSealer, error) {
	iv := make([]byte, 32)
	if _, err := io.ReadFull(keyGenerator, iv); err != nil {
		return nil, errors.New("failed to generate IV")
	}

	gs := &GossipSealer{
		sessionId:    sessionId,
		currentSeq:   -1,
		keyGenerator: keyGenerator,
		currentKeys:  make([]byte, 48),
		currentIv:    iv,
		sealerCount:  counter,
	}
	counter++

	if err := gs.Update(); err != nil {
		return nil, err
	}

	return gs, nil
}

func (g *GossipSealer) GetCurrentSeq() int {
	return g.currentSeq
}

func (g *GossipSealer) Seal(data []byte, sealType string) (*messages.MessageSeal, error) {
	log.Print("Sealing with seq ", g.currentSeq, " on sealer ", g.sealerCount)
	aes, err := aes.NewCipher(g.currentKeys[16:48])
	if err != nil {
		return nil, errors.New("failed to create AES cipher")
	}

	aesofb := cipher.NewOFB(aes, g.currentIv[:aes.BlockSize()])
	encrypted := make([]byte, len(data))
	aesofb.XORKeyStream(encrypted, data)

	mac := hmac.New(sha256.New, g.currentKeys[:16])
	dataToMac := getDataToMac(encrypted, g.currentSeq)
	mac.Write(dataToMac)
	macSum := mac.Sum(nil)

	msg := &messages.MessageSeal{
		Type:      sealType,
		SessionId: g.sessionId,
		Mac:       macSum,
		Sequence:  g.currentSeq,
		Encrypted: encrypted,
	}

	g.Update()

	return msg, nil
}

func (g *GossipSealer) Unseal(msg *messages.MessageSeal) ([]byte, error) {
	if msg.SessionId != g.sessionId {
		return nil, errors.New("wrong session ID")
	}

	if msg.Sequence != g.currentSeq {
		return nil, errors.New("wrong sequence")
	}

	mac := hmac.New(sha256.New, g.currentKeys[:16])
	dataToMac := getDataToMac(msg.Encrypted, msg.Sequence)
	mac.Write(dataToMac)
	expectedMac := mac.Sum(nil)

	if !hmac.Equal(msg.Mac, expectedMac) {
		return nil, errors.New("wrong MAC")
	}

	aes, err := aes.NewCipher(g.currentKeys[16:48])
	if err != nil {
		return nil, errors.New("failed to create AES cipher")
	}

	// TODO: which mode to use?
	aesofb := cipher.NewOFB(aes, g.currentIv[:aes.BlockSize()])
	decrypted := make([]byte, len(msg.Encrypted))
	aesofb.XORKeyStream(decrypted, msg.Encrypted)

	return decrypted, nil
}

func getDataToMac(data []byte, seq int) []byte {
	dataToMac := make([]byte, len(data))
	copy(dataToMac, data)
	intBuff := make([]byte, 4)
	binary.BigEndian.PutUint32(intBuff, uint32(seq))
	dataToMac = append(dataToMac, intBuff...)
	return dataToMac
}

func (g *GossipSealer) Update() error {
	n, err := g.keyGenerator.Read(g.currentKeys)

	if err != nil || n != len(g.currentKeys) {
		return errors.New("failed to generate next keys")
	}

	g.currentSeq++
	sha256 := sha256.New()
	sha256.Write(g.currentIv)
	g.currentIv = sha256.Sum(nil)
	log.Println("Updated sealer ", g.sealerCount, " to seq ", g.currentSeq)
	return nil
}

func (g *GossipSealer) UpdateToSeq(seq int) error {
	if seq < g.currentSeq {
		return errors.New("cannot update to a previous sequence")
	}

	for g.currentSeq < seq {
		if err := g.Update(); err != nil {
			return err
		}
	}

	return nil
}
