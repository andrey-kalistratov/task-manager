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

	"github.com/andrey-kalistratov/task-manager/worker/internal/config"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
)

var _ task.FileStorage = (*Storage)(nil)

type Storage struct {
	cfg    *config.S3Config
	client *transfermanager.Client
}

func NewStorage(cfg *config.S3Config) *Storage {
	awsCfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID, cfg.SecretAccessKey, "",
		),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
	modifier := func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	}
	return &Storage{
		client: transfermanager.New(s3.NewFromConfig(awsCfg, modifier)),
		cfg:    cfg,
	}
}

func (s Storage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, &transfermanager.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(s.cfg.Prefix + path),
	})
	if err != nil {
		return nil, fmt.Errorf("download object: %w", err)
	}
	return io.NopCloser(object.Body), nil
}

func (s Storage) Upload(ctx context.Context, path string, r io.Reader) error {
	_, err := s.client.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(s.cfg.Prefix + path),
		Body:   r,
	})
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}
	return nil
}
