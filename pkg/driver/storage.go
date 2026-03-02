package driver

import (
	"context"
	"io"
	"mime/multipart"
	"sync"
	"time"

	"github.com/leeseika/cv-demo/pkg/config"
	"github.com/leeseika/cv-demo/pkg/driver/storage"
	"github.com/rs/zerolog/log"
)

const (
	GCSClient = "gcs"
	S3Client  = "s3"
)

var (
	storageCfg           *config.StorageConfig
	_storageProvider     StorageProvider
	_storageProviderOnce sync.Once
)

type (
	StorageProvider interface {
		// Bucket return bucket name
		Bucket() string

		// Upload upload file object to cloud
		Upload(ctx context.Context, file multipart.File, fileKey string, contentType string, cacheControl string) error

		// Delete delete file object from cloud
		Delete(ctx context.Context, fileKey string) error

		// BatchUpload upload file objects to cloud in batch
		BatchUpload(ctx context.Context, req *storage.UploadRequest, cacheControl string) error

		// ListBucketContents get object list with a prefix in cloud
		ListBucketContents(ctx context.Context, prefix string) ([]storage.ObjectInfo, error)

		// DetectContentType detect content type
		DetectContentType(filename string, content []byte) string

		Download(ctx context.Context, file io.Writer, fileKey string) error

		GeneratePresignedURL(ctx context.Context, fileKey string, expireDuration time.Duration) (string, error)
	}
)

// InitStorageProvider initialize storage provider.
func InitStorageProvider(conf config.StorageConfig) {
	_storageProviderOnce.Do(func() {
		initStorageProviderOnce(conf)
	})
}

func initStorageProviderOnce(conf config.StorageConfig) {
	storageCfg = &conf

	switch storageCfg.OSSType {
	case S3Client:
		var err error
		s3Config := storageCfg.S3
		_storageProvider, err = storage.NewS3Client(s3Config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize s3 client")
			return
		}
	case GCSClient:
		var err error
		gcsConfig := storageCfg.GCS
		_storageProvider, err = storage.NewGCSClient(gcsConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize gcs client")
			return
		}
	default:
		log.Info().Msgf("unknown oss-type: shoud be s3 or gcs ")
		return
	}

	log.Info().Msgf(" %s client is initialized", storageCfg.OSSType)
}
