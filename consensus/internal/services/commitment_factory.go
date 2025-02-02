package services

import (
	"crypto/rand"

	"github.com/titosilva/drmchain-pos/blocks/history"
	"github.com/titosilva/drmchain-pos/consensus/internal/forgery"
	"github.com/titosilva/drmchain-pos/consensus/messages"
	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/identity/signatures"
	"github.com/titosilva/drmchain-pos/internal/di"
	identityprovider "github.com/titosilva/drmchain-pos/internal/shared/identity_provider"
	"github.com/titosilva/drmchain-pos/internal/utils/cryptutil"
)

type CommitmentFactory struct {
	id identity.PrivateIdentity
	bh *history.BlockHistory
}

func CommitmentFactoryFactory(diCtx *di.DIContext) *CommitmentFactory {
	idProv := identityprovider.GetFromDI(diCtx)
	bh := history.GetFromDI(diCtx)

	id, err := idProv.GetIdentity()
	if err != nil {
		return nil
	}

	return &CommitmentFactory{
		id: id,
		bh: bh,
	}
}

func (cf *CommitmentFactory) CreateCommitment() (*messages.CommitmentMessage, error) {
	value, err := generateCommitValue()
	if err != nil {
		return nil, err
	}

	stakes := cf.bh.GetStakes(cf.id.GetTag())

	commitment := messages.CommitmentMessage{
		Tag:        cf.id.GetTag(),
		Commitment: cryptutil.Hash(value),
		Stakes:     stakes, // A user could want to commit less than this
		BlockIndex: cf.bh.GetLastBlock().Index + 1,
		PrevHash:   cf.bh.GetLastBlock().Hash,
	}

	commitment.Signature, err = signatures.Sign(cf.id, commitment.Serialize())
	if err != nil {
		return nil, err
	}

	return &commitment, nil
}

func generateCommitValue() ([]byte, error) {
	v := make([]byte, forgery.CommitedLength)

	if _, err := rand.Read(v); err != nil {
		return nil, err
	}

	return v, nil
}
