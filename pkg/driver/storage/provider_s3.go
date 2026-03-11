package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/leeseika/cv-demo/pkg/config"
	"github.com/leeseika/cv-demo/pkg/threading"
)

/**
 * Implementation for StorageProvider in s3
 */
type s3Client struct {
	client *s3.Client // S3 Client from SDK
	bucket string     // Bucket
}

func NewS3Client(cfg config.S3Config) (*s3Client, error) {
	s3c := s3Client{}
	ctx := context.Background()

	endpoint := fmt.Sprintf("%s://%s",
		cfg.URLScheme,
		cfg.Endpoint,
	)

	var awsConfig aws.Config
	var err error

	if cfg.IsAuth {
		awsConfig, err = awsCfg.LoadDefaultConfig(ctx,
			awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"", // todo: session
			)),
			awsCfg.WithRegion(cfg.Region),
		)
	} else {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 30 * time.Second,
		}

		awsConfig, err = awsCfg.LoadDefaultConfig(ctx,
			awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"", // todo: session
			)),
			awsCfg.WithHTTPClient(httpClient),
			awsCfg.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, err
	}

	s3c.client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = cfg.UsePathStyle
	})

	s3c.bucket = cfg.Bucket

	return &s3c, nil
}

// Bucket return bucket name
func (a *s3Client) Bucket() string {
	return a.bucket
}

// Upload upload file object to cloud (S3)
func (a *s3Client) Upload(ctx context.Context, file multipart.File, fileKey string, contentType string, cacheControl string) error {
	_, err := a.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:       aws.String(a.bucket),
		Key:          aws.String(fileKey),
		Body:         file,
		ACL:          types.ObjectCannedACLPublicRead, // public read
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(cacheControl),
	})
	if err != nil {
		return fmt.Errorf("upload file err: %w", err)
	}
	return nil
}

// Delete delete file object from cloud (S3)
func (a *s3Client) Delete(ctx context.Context, fileKey string) error {
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("delete file failed: %w", err)
	}

	return nil
}

// BatchUpload upload file objects to cloud in batch (S3)
func (a *s3Client) BatchUpload(c context.Context, req *UploadRequest, cacheControl string) error {
	var (
		wg      sync.WaitGroup
		errCh   = make(chan error, 1)
		workers = runtime.NumCPU() * 2
	)

	if req == nil || len(req.Files) == 0 {
		return errors.New("empty upload request")
	}

	// add cancel functionality to the context
	// after the first err appears, it is used to notify the subsequent coroutines to terminate their work
	ctx, cancel := context.WithCancel(c)
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

			// exec upload
			_, err := a.client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:       aws.String(a.bucket),
				Key:          aws.String(file.RemoteKey),
				Body:         bytes.NewReader(file.Content),
				ACL:          types.ObjectCannedACLPublicRead, // public read
				ContentType:  aws.String(file.ContentType),
				CacheControl: aws.String(cacheControl),
			})
			if err != nil {
				// send first err
				select {
				case errCh <- fmt.Errorf("upload failed for %s: %w", file.RemoteKey, err):
					// exec cancel
					cancel()
				default:
				}
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

// ListBucketContents get object list with a prefix in cloud (S3)
func (a *s3Client) ListBucketContents(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(a.bucket),
		Prefix: aws.String(prefix),
	}

	for {
		resp, err := a.client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("list req err: %w", err)
		}

		for _, obj := range resp.Contents {
			objects = append(objects, ObjectInfo{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
				ETag:         aws.ToString(obj.ETag),
				StorageClass: string(obj.StorageClass),
			})
		}

		// Handling paging (enhanced nil pointer protection)
		if resp.IsTruncated == nil || !*resp.IsTruncated {
			break
		}
		input.ContinuationToken = resp.NextContinuationToken
	}

	return objects, nil
}

// DetectContentType detect content type
func (a *s3Client) DetectContentType(filename string, content []byte) string {
	return detectContentType(filename, content)
}

func (a *s3Client) Download(ctx context.Context, file io.Writer, fileKey string) error {
	if len(fileKey) == 0 {
		return errors.New("file key is required")
	}

	resp, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to get object %s: %w", fileKey, err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to copy object content for %s: %w", fileKey, err)
	}

	return nil
}

func (a *s3Client) ObjectExists(ctx context.Context, fileKey string) (bool, error) {
	if len(fileKey) == 0 {
		return false, errors.New("file key is required")
	}

	_, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(fileKey),
	})
	if err == nil {
		return true, nil
	}

	var responseErr *awshttp.ResponseError
	if errors.As(err, &responseErr) {
		if responseErr.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}
	}

	return false, fmt.Errorf("head object failed for %s: %w", fileKey, err)
}

func (a *s3Client) GeneratePresignedUploadURL(ctx context.Context, fileKey string, contentType string, expireDuration time.Duration) (string, error) {
	if len(fileKey) == 0 {
		return "", errors.New("file key is required")
	}
	if len(contentType) == 0 {
		return "", errors.New("content type is required")
	}
	if expireDuration <= 0 {
		return "", errors.New("expire duration must be greater than 0")
	}

	presignClient := s3.NewPresignClient(a.client)
	presignResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(a.bucket),
		Key:         aws.String(fileKey),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expireDuration
	})
	if err != nil {
		return "", fmt.Errorf("generate presigned url failed: %w", err)
	}

	return presignResult.URL, nil
}
