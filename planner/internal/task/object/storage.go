package object

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
)

var _ task.FileStorage = (*FileStorage)(nil)

type FileStorage struct {
	client *transfermanager.Client
	bucket string
}

func NewFileStorage(cfg *config.Config) *FileStorage {
	awsCfg := aws.Config{
		Region: cfg.Storage.S3.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.Storage.S3.AccessKeyID, cfg.Storage.S3.SecretAccessKey, "",
		),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
	modifier := func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Storage.S3.Endpoint)
		o.UsePathStyle = true
	}
	return &FileStorage{
		client: transfermanager.New(s3.NewFromConfig(awsCfg, modifier)),
		bucket: cfg.Storage.S3.Bucket,
	}
}

func (s FileStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, &transfermanager.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("download object: %w", err)
	}
	return io.NopCloser(object.Body), nil
}

func (s FileStorage) Upload(ctx context.Context, path string, r io.Reader) error {
	_, err := s.client.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   r,
	})
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}
	return nil
}
