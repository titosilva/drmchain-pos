package storage

import "github.com/titosilva/drmchain-pos/internal/di"

// Represents a storage that can store, retrieve and delete data
type BlobStorage interface {
	Store(key string, data []byte) error
	Exists(key string) (bool, error)
	Retrieve(key string) ([]byte, error)
	Delete(key string) error
}

func GetFromDI(diCtx *di.DIContext) BlobStorage {
	return di.GetInterfaceService[BlobStorage](diCtx)
}
