package driver

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"sync"
	"text/template"
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
	_storageProvider       StorageProvider
	_storageProviderOnce   sync.Once
	_storageURLBuilder     StorageURLBuilder
	_storageURLBuilderOnce sync.Once
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
		ObjectExists(ctx context.Context, fileKey string) (bool, error)

		GeneratePresignedUploadURL(ctx context.Context, fileKey string, contentType string, expireDuration time.Duration) (string, error)
	}

	StorageURLBuilder interface {
		// BuildURL build the public URL for the given file key
		BuildURL(bucket string, fileKey string) string
		BuildURLWithDefaultBucket(fileKey string) string
		BuildFileKey(shopID string, uuid string, filename string) string
	}

	defaultStorageURLBuilder struct {
		defaultBucket string
		domain        string
		template      *template.Template
	}
)

func (d *defaultStorageURLBuilder) BuildURL(bucket string, fileKey string) string {
	props := map[string]string{
		"Domain": d.domain,
		"Bucket": bucket,
		"Source": fileKey,
	}

	var result bytes.Buffer
	err := d.template.Execute(&result, props)
	if err != nil {
		return ""
	}

	return result.String()
}

func (d *defaultStorageURLBuilder) BuildURLWithDefaultBucket(fileKey string) string {
	return d.BuildURL(d.defaultBucket, fileKey)
}

func (d *defaultStorageURLBuilder) BuildFileKey(shopID string, uuid string, filename string) string {
	fileExt := path.Ext(filename)
	fileID := fmt.Sprintf("%s%s", uuid, fileExt)
	shopHash := fmt.Sprintf("%x", md5.Sum([]byte(shopID)))
	fileKey := fmt.Sprintf("i/%s/%s/%s/%s", safeSubstr(shopHash, 0, 1), safeSubstr(shopHash, 0, 2), safeSubstr(shopHash, 0, 8), fileID)
	return fileKey
}

// InitStorageProvider initialize storage provider.
func InitStorageProvider(conf config.StorageConfig) {
	_storageProviderOnce.Do(func() {
		initStorageProviderOnce(conf)
	})
}

func initStorageProviderOnce(conf config.StorageConfig) {
	switch conf.OSSType {
	case S3Client:
		var err error
		s3Config := conf.S3
		_storageProvider, err = storage.NewS3Client(s3Config)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize s3 client")
			return
		}
	case GCSClient:
		var err error
		gcsConfig := conf.GCS
		_storageProvider, err = storage.NewGCSClient(gcsConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize gcs client")
			return
		}
	default:
		log.Info().Msgf("unknown oss-type: should be s3 or gcs ")
		return
	}
	log.Info().Msgf(" %s client is initialized", conf.OSSType)
}

func GetStorageProvider() StorageProvider {
	return _storageProvider
}

func GetStorageURLBuilder() StorageURLBuilder {
	return _storageURLBuilder
}

func InitStorageURLBuilder(conf config.StorageConfig) {
	_storageURLBuilderOnce.Do(func() {
		initStorageURLBuilderOnce(conf)
	})
}

func initStorageURLBuilderOnce(conf config.StorageConfig) {
	domain := conf.AssetsBuilderConfig.AssetsDomain
	format := conf.AssetsBuilderConfig.AssetsURLFormat
	defaultBucket := ""
	switch conf.OSSType {
	case S3Client:
		defaultBucket = conf.S3.Bucket
	case GCSClient:
		defaultBucket = conf.GCS.Bucket
	}

	tpl, err := template.New("assets_url").Parse(format)
	if err != nil {
		log.Err(err).Msg("failed to parse assets URL format")
	}

	_storageURLBuilder = &defaultStorageURLBuilder{
		defaultBucket: defaultBucket,
		domain:        domain,
		template:      tpl,
	}
}

func safeSubstr(s string, start, end int) string {
	if start > len(s) {
		return ""
	}
	if end > len(s) {
		end = len(s)
	}
	if start < 0 {
		start = 0
	}
	return s[start:end]
}
