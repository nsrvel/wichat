package minio

// Config wires a single MinIO/S3-compatible bucket.
type Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	PublicBaseURL string
	UseSSL        bool
	Region        string
}
