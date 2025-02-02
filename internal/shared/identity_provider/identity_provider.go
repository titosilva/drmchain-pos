package identityprovider

import (
	"github.com/titosilva/drmchain-pos/identity"
	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/storage"
)

// TODO: move to the identity package
type IdentityProvider struct {
	storage storage.BlobStorage
}

func Factory(diCtx *di.DIContext) *IdentityProvider {
	blobStore := di.GetInterfaceService[storage.BlobStorage](diCtx)
	return &IdentityProvider{storage: blobStore}
}

func GetFromDI(diCtx *di.DIContext) *IdentityProvider {
	return di.GetService[IdentityProvider](diCtx)
}

func New(storage storage.BlobStorage) *IdentityProvider {
	return &IdentityProvider{storage: storage}
}

func (i *IdentityProvider) GetIdentity() (identity.PrivateIdentity, error) {
	exists, err := i.storage.Exists("identity")
	if err != nil {
		return nil, err
	}

	if !exists {
		return i.createAndSaveIdentity()
	} else {
		return i.loadIdentity()
	}
}

func (i *IdentityProvider) createAndSaveIdentity() (identity.PrivateIdentity, error) {
	identity, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	i.storage.Store("identity", identity.GetPrivateKeyBytes())
	return identity, nil
}

func (i *IdentityProvider) loadIdentity() (identity.PrivateIdentity, error) {
	data, err := i.storage.Retrieve("identity")

	if err != nil {
		return nil, err
	}

	return identity.FromPrivateKeyBytes(data)
}
