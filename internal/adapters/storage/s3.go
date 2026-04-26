package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3DocumentStorage struct {
	bucket        string
	client        *s3.Client
	presignClient *s3.PresignClient
}

func NewS3DocumentStorage(ctx context.Context, bucket string) (*S3DocumentStorage, error) {
	if bucket == "" {
		return nil, fmt.Errorf("DOCUMENTS_BUCKET is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config for S3: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	return &S3DocumentStorage{
		bucket:        bucket,
		client:        client,
		presignClient: s3.NewPresignClient(client),
	}, nil
}

func NewS3DocumentStorageFromEnv(ctx context.Context) (*S3DocumentStorage, error) {
	return NewS3DocumentStorage(ctx, os.Getenv("DOCUMENTS_BUCKET"))
}

func (s *S3DocumentStorage) Upload(ctx context.Context, storageKey string, body io.Reader, contentType string, contentLength int64) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(s.bucket),
		Key:                aws.String(storageKey),
		Body:               body,
		ContentType:        aws.String(contentType),
		ContentLength:      aws.Int64(contentLength),
		ContentDisposition: aws.String("inline"),
	})
	return err
}

func (s *S3DocumentStorage) Delete(ctx context.Context, storageKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(storageKey),
	})
	return err
}

func (s *S3DocumentStorage) GetViewURL(ctx context.Context, storageKey, fileName, contentType string, expiresIn time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(storageKey),
		ResponseContentDisposition: aws.String(inlineDisposition(fileName)),
		ResponseContentType:        aws.String(contentType),
	}

	request, err := s.presignClient.PresignGetObject(ctx, input, func(options *s3.PresignOptions) {
		options.Expires = expiresIn
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}

func (s *S3DocumentStorage) GetDownloadURL(ctx context.Context, storageKey, fileName string, expiresIn time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(storageKey),
		ResponseContentDisposition: aws.String(attachmentDisposition(fileName)),
	}

	request, err := s.presignClient.PresignGetObject(ctx, input, func(options *s3.PresignOptions) {
		options.Expires = expiresIn
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}

func inlineDisposition(fileName string) string {
	return contentDisposition("inline", fileName)
}

func attachmentDisposition(fileName string) string {
	return contentDisposition("attachment", fileName)
}

func contentDisposition(kind, fileName string) string {
	safeName := strings.ReplaceAll(fileName, "\"", "")
	encoded := url.PathEscape(fileName)
	mediaType, params, err := mime.ParseMediaType(fmt.Sprintf("%s; filename=\"%s\"", kind, safeName))
	if err != nil {
		return fmt.Sprintf("%s; filename*=UTF-8''%s", kind, encoded)
	}
	params["filename*"] = "UTF-8''" + encoded
	return mime.FormatMediaType(mediaType, params)
}
