package minio

import (
	"context"
	"fmt"

	mio "github.com/minio/minio-go/v7"
)

// EnsurePublicReadBucket creates the bucket when missing and allows anonymous GetObject.
func EnsurePublicReadBucket(ctx context.Context, client Client) error {
	c, ok := client.(minioClient)
	if !ok {
		return nil
	}

	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("minio bucket exists: %w", err)
	}
	if !exists {
		if err := c.client.MakeBucket(ctx, c.bucket, mio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("minio make bucket: %w", err)
		}
	}

	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": ["*"]},
    "Action": ["s3:GetObject"],
    "Resource": ["arn:aws:s3:::%s/*"]
  }]
}`, c.bucket)
	if err := c.client.SetBucketPolicy(ctx, c.bucket, policy); err != nil {
		return fmt.Errorf("minio bucket policy: %w", err)
	}

	return nil
}
