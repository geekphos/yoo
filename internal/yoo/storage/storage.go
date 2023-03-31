package storage

import (
	"sync"

	"github.com/minio/minio-go/v7"
)

var (
	once sync.Once
	S    *minio.Client
)

func InitStorage(oss *minio.Client) {
	once.Do(func() {
		S = oss
	})
}
