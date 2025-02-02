package localstorage

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/titosilva/drmchain-pos/internal/di"
	"github.com/titosilva/drmchain-pos/storage"
)

type LocalStorage struct {
	basePath string
}

func Factory(diCtx *di.DIContext) storage.BlobStorage {
	basePath := "/tmp/drmchain-pos/storage"
	return New(basePath)
}

func New(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
	}
}

// Delete implements storage.BlobStorage.
func (l *LocalStorage) Delete(key string) error {
	exists, err := l.Exists(key)

	if err != nil {
		return err
	}

	if !exists {
		return errors.New("blob does not exist")
	}

	path := filepath.Join(l.basePath, key)
	return os.Remove(path)
}

// Exists implements storage.BlobStorage.
func (l *LocalStorage) Exists(key string) (bool, error) {
	path := filepath.Join(l.basePath, key)
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

// Retrieve implements storage.BlobStorage.
func (l *LocalStorage) Retrieve(key string) ([]byte, error) {
	path := filepath.Join(l.basePath, key)
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	bs, err := io.ReadAll(file)

	if err != nil {
		return nil, err
	}

	return bs, nil
}

// Store implements storage.BlobStorage.
func (l *LocalStorage) Store(key string, data []byte) error {
	path := filepath.Join(l.basePath, key)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// static implementation check
var _ storage.BlobStorage = &LocalStorage{}
