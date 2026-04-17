package s3

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileStorage interface {
	GetFile(fileName string) (*os.File, error)
}

type Minio struct {
	client *minio.Client
	bucket string
}

func NewMinio(bucket, endpoint, accessKey, secretKey string) *Minio {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Fatalf("minio init error: %v", err)
	}

	return &Minio{
		client: minioClient,
		bucket: bucket,
	}
}

func (m *Minio) GetFile(fileName string) (*os.File, error) {
	ctx := context.Background()

	object, err := m.client.GetObject(ctx, m.bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	if _, err := file.ReadFrom(object); err != nil {
		file.Close()
		return nil, err
	}

	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}
