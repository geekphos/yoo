package oss

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioOptions struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
}

func NewMinio(opts *MinioOptions) (*minio.Client, error) {
	return minio.New(opts.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(opts.AccessKeyID, opts.SecretAccessKey, ""),
	})
}
