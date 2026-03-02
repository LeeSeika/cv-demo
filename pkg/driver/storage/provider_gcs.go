package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/leeseika/cv-demo/pkg/config"
	"github.com/leeseika/cv-demo/pkg/threading"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

/**
 * Implementation for StorageProvider in Google Cloud Storage
 */

type gcsClient struct {
	client         *storage.Client
	bucket         string
	googleAccessID string
	privateKey     []byte
}

// GCS Client
func NewGCSClient(cfg config.GCSConfig) (*gcsClient, error) {
	gcs := gcsClient{}
	ctx := context.Background()

	opts := []option.ClientOption{}

	if cfg.IsAuth {
		buildCredentialsJSON, err := json.Marshal(cfg.Credentials)
		if err != nil {
			return nil, err
		}

		opts = append(opts,
			option.WithCredentialsJSON(buildCredentialsJSON),
		)
	} else {
		apiURL := fmt.Sprintf("%s://%s:%s/storage/v1/",
			cfg.URLScheme,
			cfg.Host,
			cfg.Port,
		)

		opts = append(opts,
			option.WithEndpoint(apiURL),
			option.WithoutAuthentication(),
			option.WithHTTPClient(&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}),
		)
	}

	var err error
	gcs.client, err = storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("gcs client init err: %w", err)
	}

	gcs.bucket = cfg.Bucket
	gcs.googleAccessID = cfg.Credentials.ClientEmail
	if len(cfg.Credentials.PrivateKey) > 0 {
		normalizedPrivateKey := strings.ReplaceAll(cfg.Credentials.PrivateKey, `\n`, "\n")
		gcs.privateKey = []byte(normalizedPrivateKey)
	}

	return &gcs, nil
}

// Bucket return bucket name
func (g *gcsClient) Bucket() string {
	return g.bucket
}

// Upload upload file object to cloud (Google Cloud Storage)
func (g *gcsClient) Upload(ctx context.Context, file multipart.File, fileKey, contentType, cacheControl string) error {
	obj := g.client.Bucket(g.bucket).Object(fileKey)
	writer := obj.NewWriter(ctx)

	writer.ContentType = contentType
	writer.CacheControl = cacheControl
	writer.ACL = []storage.ACLRule{
		{
			Entity: storage.AllUsers,
			Role:   storage.RoleReader,
		},
	}

	writer.SendCRC32C = false
	writer.ChunkSize = 0

	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("content copy failed: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("writer close failed: %w", err)
	}

	return nil
}

// Delete delete file object from cloud (Google Cloud Storage)
func (g *gcsClient) Delete(ctx context.Context, fileKey string) error {
	obj := g.client.Bucket(g.bucket).Object(fileKey)

	if err := obj.Delete(context.Background()); err != nil {
		return fmt.Errorf("file: %s bucket: %s del err: %w", fileKey, g.bucket, err)
	}

	return nil
}

// BatchUpload upload file objects to cloud in batch (Google Cloud Storage)
func (g *gcsClient) BatchUpload(ctx context.Context, req *UploadRequest, cacheControl string) error {
	var (
		wg      sync.WaitGroup        // concurrency sync
		errCh   = make(chan error, 1) // err chan
		workers = runtime.NumCPU() * 2
	)

	if req == nil || len(req.Files) == 0 {
		return errors.New("empty upload request")
	}

	// add cancel functionality to the context
	// after the first err appears, it is used to notify the subsequent coroutines to terminate their work
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pool := threading.NewPoolAndRun(workers, false)
	defer pool.Close()

	for _, f := range req.Files {
		if ctx.Err() != nil {
			break
		}

		file := f
		wg.Add(1)

		pool.Submit(func() {
			defer wg.Done()

			// add context check inside coroutine
			select {
			case <-ctx.Done():
				return
			default:
			}

			obj := g.client.Bucket(g.bucket).Object(file.RemoteKey)
			writer := obj.NewWriter(ctx)

			writer.ContentType = file.ContentType
			writer.CacheControl = cacheControl
			writer.ACL = []storage.ACLRule{
				{
					Entity: storage.AllUsers,
					Role:   storage.RoleReader,
				},
			}

			errorFunc := func(err error, errMsg string) {
				// send first err
				select {
				case errCh <- fmt.Errorf("%s: %w", errMsg, err):
					// cancel ctx
					cancel()
				default:
				}
			}

			if _, err := io.Copy(writer, bytes.NewReader(file.Content)); err != nil {
				errorFunc(err, fmt.Sprintf("file content write failed for %s", file.RemoteKey))
				return
			}

			// Close go to upload object and close writer
			if err := writer.Close(); err != nil {
				errorFunc(err, fmt.Sprintf("upload failed for %s", file.RemoteKey))
				return
			}
		})
	}

	// wait all jobs
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// listen for the first error or all completions
	select {
	case err := <-errCh:
		<-done
		return err
	case <-done:
		return nil
	}
}

// ListBucketContents get object list with a prefix in cloud (Google Cloud Storage)
func (g *gcsClient) ListBucketContents(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	bkt := g.client.Bucket(g.bucket)
	it := bkt.Objects(ctx, &storage.Query{
		Prefix: prefix,
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("get gcs list: %w", err)
		}

		objects = append(objects, ObjectInfo{
			Key:          attrs.Name,
			Size:         attrs.Size,
			LastModified: attrs.Updated,
			ETag:         fmt.Sprintf("%x", attrs.MD5),
			StorageClass: attrs.StorageClass,
		})
	}

	return objects, nil
}

// DetectContentType detect content type
func (g *gcsClient) DetectContentType(filename string, content []byte) string {
	return detectContentType(filename, content)
}

func (g *gcsClient) Download(ctx context.Context, file io.Writer, fileKey string) error {
	rc, err := g.client.Bucket(g.bucket).Object(fileKey).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to create reader for file %s: %w", fileKey, err)
	}
	defer rc.Close()

	if _, err := io.Copy(file, rc); err != nil {
		return fmt.Errorf("failed to copy file content for %s: %w", fileKey, err)
	}

	return nil
}

func (g *gcsClient) GeneratePresignedURL(ctx context.Context, fileKey string, expireDuration time.Duration) (string, error) {
	if len(fileKey) == 0 {
		return "", errors.New("file key is required")
	}
	if expireDuration <= 0 {
		return "", errors.New("expire duration must be greater than 0")
	}
	if len(g.googleAccessID) == 0 || len(g.privateKey) == 0 {
		return "", errors.New("gcs signing credentials are required")
	}

	url, err := storage.SignedURL(g.bucket, fileKey, &storage.SignedURLOptions{
		GoogleAccessID: g.googleAccessID,
		PrivateKey:     g.privateKey,
		Method:         http.MethodGet,
		Expires:        time.Now().Add(expireDuration),
	})
	if err != nil {
		return "", fmt.Errorf("generate presigned url failed: %w", err)
	}

	return url, nil
}
