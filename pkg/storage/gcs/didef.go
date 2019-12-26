package gcs

import (
	"context"
	"os"

	_gcs "cloud.google.com/go/storage"
	"github.com/sarulabs/di"
	log "github.com/sirupsen/logrus"
)

var (
	ServiceBuilder = di.Def{
		Name:  "objectstorage-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			ctx := context.Background()
			gcsClient, err := _gcs.NewClient(ctx)
			if err != nil {
				log.Error(err)
			}
			objectStorageService := NewObjectStorageService(gcsClient, os.Getenv("BUCKET_NAME"))
			return objectStorageService, nil
		},
	}
)
