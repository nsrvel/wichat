package minio

import (
	"context"
	"fmt"
	"io"
	"strings"

	mio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Object is a downloaded object; callers must Close the body.
type Object struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
}

// Upload is one object for UploadMany.
type Upload struct {
	Key         string
	ContentType string
	Body        io.Reader
	Size        int64
}

// Client is object storage I/O for one bucket.
type Client interface {
	Get(ctx context.Context, key string) (Object, error)
	Upload(ctx context.Context, key, contentType string, body io.Reader, size int64) error
	UploadMany(ctx context.Context, items []Upload) error
	Copy(ctx context.Context, srcKey, dstKey string) error
	Delete(ctx context.Context, key string) error
	DeletePrefix(ctx context.Context, prefix string) error
	PublicURL(key string) string
}

type minioClient struct {
	client        *mio.Client
	bucket        string
	publicBaseURL string
	log           logger.Logger
}

type disabledClient struct{}

// New returns a Client. Empty endpoint yields a disabled client (all ops return ErrDisabled except PublicURL).
func New(cfg Config, log logger.Logger) (Client, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return disabledClient{}, nil
	}
	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("minio bucket required when endpoint is set")
	}

	client, err := mio.New(endpoint, &mio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: strings.TrimSpace(cfg.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	return minioClient{
		client:        client,
		bucket:        bucket,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
		log:           log,
	}, nil
}

func (c minioClient) Get(ctx context.Context, key string) (Object, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, key, mio.GetObjectOptions{})
	if err != nil {
		c.log.Error("minio get object", "key", key, "err", err)

		return Object{}, err
	}
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		c.log.Error("minio stat object", "key", key, "err", err)

		return Object{}, err
	}

	return Object{
		Body:          obj,
		ContentType:   info.ContentType,
		ContentLength: info.Size,
	}, nil
}

func (c minioClient) Upload(ctx context.Context, key, contentType string, body io.Reader, size int64) error {
	_, err := c.client.PutObject(ctx, c.bucket, key, body, size, mio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		c.log.Error("minio put object", "key", key, "err", err)
	}

	return err
}

func (c minioClient) UploadMany(ctx context.Context, items []Upload) error {
	// ponytail: sequential uploads; no multi-object transaction — upgrade with worker pool if needed.
	for _, item := range items {
		if err := c.Upload(ctx, item.Key, item.ContentType, item.Body, item.Size); err != nil {
			return err
		}
	}

	return nil
}

func (c minioClient) Copy(ctx context.Context, srcKey, dstKey string) error {
	src := mio.CopySrcOptions{Bucket: c.bucket, Object: srcKey}
	dst := mio.CopyDestOptions{Bucket: c.bucket, Object: dstKey}
	_, err := c.client.CopyObject(ctx, dst, src)
	if err != nil {
		c.log.Error("minio copy object", "src", srcKey, "dst", dstKey, "err", err)
	}

	return err
}

func (c minioClient) Delete(ctx context.Context, key string) error {
	err := c.client.RemoveObject(ctx, c.bucket, key, mio.RemoveObjectOptions{})
	if err != nil {
		c.log.Error("minio remove object", "key", key, "err", err)
	}

	return err
}

func (c minioClient) DeletePrefix(ctx context.Context, prefix string) error {
	list := c.client.ListObjects(ctx, c.bucket, mio.ListObjectsOptions{Prefix: prefix, Recursive: true})
	for obj := range list {
		if obj.Err != nil {
			c.log.Error("minio list objects", "prefix", prefix, "err", obj.Err)

			return obj.Err
		}
		if err := c.client.RemoveObject(ctx, c.bucket, obj.Key, mio.RemoveObjectOptions{}); err != nil {
			c.log.Error("minio remove object", "key", obj.Key, "err", err)

			return err
		}
	}

	return nil
}

func (c minioClient) PublicURL(key string) string {
	key = strings.TrimPrefix(key, "/")
	if c.publicBaseURL == "" {
		return "/" + key
	}

	return c.publicBaseURL + "/" + key
}

func (disabledClient) Get(context.Context, string) (Object, error) {
	return Object{}, ErrDisabled
}

func (disabledClient) Upload(context.Context, string, string, io.Reader, int64) error {
	return ErrDisabled
}

func (disabledClient) UploadMany(context.Context, []Upload) error {
	return ErrDisabled
}

func (disabledClient) Copy(context.Context, string, string) error {
	return ErrDisabled
}

func (disabledClient) Delete(context.Context, string) error {
	return ErrDisabled
}

func (disabledClient) DeletePrefix(context.Context, string) error {
	return ErrDisabled
}

func (disabledClient) PublicURL(key string) string {
	key = strings.TrimPrefix(key, "/")

	return "/" + key
}
