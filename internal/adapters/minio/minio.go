package minio

import (
	"github.com/Olegsandrik/Exponenta/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Adapter struct {
	BucketName string
	Client     *minio.Client
}

func NewMinioAdapter(cfg *config.Config) (*Adapter, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioPassword, cfg.MinioUser, ""),
		Secure: false,
	})

	if err != nil {
		return nil, err
	}

	return &Adapter{
		BucketName: cfg.MinioBucket,
		Client:     client,
	}, nil
}

func NewEmptyObjectOptions() minio.GetObjectOptions {
	return minio.GetObjectOptions{}
}
