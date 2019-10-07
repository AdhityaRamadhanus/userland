package gcs

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/AdhityaRamadhanus/userland"
)

//KeyValueService implements userland.KeyValueService interface using redis
type ObjectStorageService struct {
	storageClient *storage.Client
	bucketName    string
}

//NewKeyValueService construct a new KeyValueService from redis client
func NewObjectStorageService(storageClient *storage.Client, bucketName string) *ObjectStorageService {
	return &ObjectStorageService{
		storageClient: storageClient,
		bucketName:    bucketName,
	}
}

func (o *ObjectStorageService) Write(reader io.Reader, metadata userland.ObjectMetaData) (string, error) {
	bucket := o.storageClient.Bucket(o.bucketName)
	object := bucket.Object(metadata.Path)
	objectWriter := object.NewWriter(context.Background())
	objectWriter.ContentType = metadata.ContentType
	objectWriter.CacheControl = metadata.CacheControl

	if _, err := io.Copy(objectWriter, reader); err != nil {
		return "", err
	}
	if err := objectWriter.Close(); err != nil {
		return "", err
	}

	attrs, _ := object.Attrs(context.Background())
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", attrs.Bucket, attrs.Name), nil
}

func (o *ObjectStorageService) Fetch(path string) ([]byte, userland.ObjectMetaData, error) {
	bucket := o.storageClient.Bucket(o.bucketName)
	object := bucket.Object(path)
	objectReader, err := object.NewReader(context.Background())
	if err != nil {
		return nil, userland.ObjectMetaData{}, err
	}

	metadata := userland.ObjectMetaData{
		ContentType:     objectReader.Attrs.ContentType,
		ContentEncoding: objectReader.Attrs.ContentEncoding,
		Size:            objectReader.Attrs.Size,
	}

	content, err := ioutil.ReadAll(objectReader)
	defer objectReader.Close()
	if err != nil {
		return nil, userland.ObjectMetaData{}, err
	}
	return content, metadata, nil
}
