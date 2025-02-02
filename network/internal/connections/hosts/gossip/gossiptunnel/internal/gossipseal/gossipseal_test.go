package gossipseal_test

import (
	"crypto/sha256"
	"testing"

	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip/gossiptunnel/internal/gossipseal"
	"github.com/titosilva/drmchain-pos/network/internal/connections/hosts/gossip/internal/messages"
	"golang.org/x/crypto/hkdf"
)

func Test__Unseal__ShouldUndoSeal(t *testing.T) {
	tun1, tun2 := generateTunnels(t)

	checkSealUnseal(t, tun1, tun2, "hello")
	checkSealUnseal(t, tun1, tun2, "world")
	checkSealUnseal(t, tun1, tun2, "foo")
	checkSealUnseal(t, tun1, tun2, "bar")
}

func checkSealUnseal(t *testing.T, sealer1, sealer2 *gossipseal.GossipSealer, msg string) {
	shell, err := sealer1.Seal([]byte(msg), messages.SealTypeData)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := sealer2.Unseal(shell)

	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if string(decrypted) != msg {
		t.Fatal("unexpected message", string(decrypted))
	}

	sealer2.Update()
}

func Test__Unseal__ShouldFailIfSealIsCorrupted(t *testing.T) {
	sealer1, sealer2 := generateTunnels(t)

	shell, err := sealer1.Seal([]byte("hello"), messages.SealTypeData)
	if err != nil {
		t.Fatal(err)
	}
	shell.Encrypted[0] += 1

	_, err = sealer2.Unseal(shell)

	if err == nil {
		t.Error("expected error")
	}
}

func generateTunnels(t *testing.T) (*gossipseal.GossipSealer, *gossipseal.GossipSealer) {
	sealer1 := buildSealer(t, "session", []byte("secret"))
	sealer2 := buildSealer(t, "session", []byte("secret"))

	return sealer1, sealer2
}

func buildSealer(t *testing.T, sessionId string, secret []byte) *gossipseal.GossipSealer {
	seed := hkdf.Extract(sha256.New, secret, []byte("salt"))
	keyGen := hkdf.Expand(sha256.New, seed, []byte("test"))

	sealer, err := gossipseal.New(sessionId, keyGen)
	if err != nil {
		t.Fatal(err)
	}

	return sealer
}
