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
	client   *s3.Client
	transfer *transfermanager.Client
	bucket   string
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
		client: s3.NewFromConfig(awsCfg, modifier),
		bucket: cfg.Storage.S3.Bucket,
	}
}

func (s FileStorage) Download(ctx context.Context, file task.File) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(file.Path),
	})
	if err != nil {
		return nil, fmt.Errorf("download file %q: %w", file.Path, err)
	}
	return object.Body, nil
}

func (s FileStorage) Upload(ctx context.Context, path string, r io.Reader) (task.File, error) {
	_, err := s.transfer.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   r,
	})
	if err != nil {
		return task.File{}, fmt.Errorf("upload file %q: %w", path, err)
	}
	return task.File{
		Path:     path,
		Provider: task.ProviderS3,
	}, nil
}
