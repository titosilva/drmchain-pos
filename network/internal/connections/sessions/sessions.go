package sessions

import (
	"sync"

	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cmap"
	"github.com/titosilva/drmchain-pos/internal/utils/uuid"
	"github.com/titosilva/drmchain-pos/network"
)

type Session struct {
	Id      string
	KeySeed []byte
	Peer    network.Peer

	previouslyConnectedMux *sync.Mutex
	previouslyConnected    bool
}

func NewSession(id string, keySeed []byte, peer network.Peer) *Session {
	return &Session{
		Id:                     id,
		KeySeed:                keySeed,
		Peer:                   peer,
		previouslyConnectedMux: &sync.Mutex{},
	}
}

type Memory struct {
	sessions *cmap.CMap[string, *Session]
}

func Factory(di *di.DIContext) *Memory {
	return NewMemory()
}

func GetFromDI(diCtx *di.DIContext) *Memory {
	return di.GetService[Memory](diCtx)
}

func (s *Session) WasConnected() bool {
	s.previouslyConnectedMux.Lock()
	defer s.previouslyConnectedMux.Unlock()
	return s.previouslyConnected
}

func (s *Session) MarkConnected() {
	s.previouslyConnectedMux.Lock()
	s.previouslyConnected = true
	s.previouslyConnectedMux.Unlock()
}

func NewMemory() *Memory {
	return &Memory{
		sessions: cmap.New[string, *Session](),
	}
}

func (m *Memory) GenerateSession(peer network.Peer, keysSeed []byte) *Session {
	session := NewSession(uuid.NewUuid(), keysSeed, peer)

	m.RegisterSession(session)
	return session
}

func (m *Memory) RegisterSession(session *Session) {
	m.sessions.Set(session.Id, session)
}

func (m *Memory) GetSession(id string) *Session {
	session, ok := m.sessions.Get(id)

	if !ok {
		return nil
	}

	return session
}
