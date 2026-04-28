package ports

import (
	"context"
	"io"
	"time"
)

type DocumentStorage interface {
	Upload(ctx context.Context, storageKey string, body io.Reader, contentType string, contentLength int64) error
	Delete(ctx context.Context, storageKey string) error
	GetViewURL(ctx context.Context, storageKey, fileName, contentType string, expiresIn time.Duration) (string, error)
	GetDownloadURL(ctx context.Context, storageKey, fileName string, expiresIn time.Duration) (string, error)
}
