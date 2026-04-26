package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

const documentURLExpiration = 15 * time.Minute

type documentSupport struct {
	storage ports.DocumentStorage
}

func newDocumentSupport(storage ports.DocumentStorage) documentSupport {
	return documentSupport{storage: storage}
}

func (d documentSupport) ensureStorage() error {
	if d.storage == nil {
		return fmt.Errorf("document storage is not configured")
	}
	return nil
}

func (d documentSupport) hydrateClientDocuments(documents []domain.ClientDocument) ([]domain.ClientDocument, error) {
	if err := d.ensureStorage(); err != nil {
		return nil, err
	}

	hydrated := make([]domain.ClientDocument, 0, len(documents))
	for _, document := range documents {
		viewURL, err := d.storage.GetViewURL(context.Background(), document.StorageKey, document.FileName, document.ContentType, documentURLExpiration)
		if err != nil {
			return nil, err
		}
		downloadURL, err := d.storage.GetDownloadURL(context.Background(), document.StorageKey, document.FileName, documentURLExpiration)
		if err != nil {
			return nil, err
		}
		document.ViewURL = viewURL
		document.DownloadURL = downloadURL
		hydrated = append(hydrated, document)
	}
	return hydrated, nil
}

func (d documentSupport) hydrateDiaryDocuments(documents []domain.DiaryDocument) ([]domain.DiaryDocument, error) {
	if err := d.ensureStorage(); err != nil {
		return nil, err
	}

	hydrated := make([]domain.DiaryDocument, 0, len(documents))
	for _, document := range documents {
		viewURL, err := d.storage.GetViewURL(context.Background(), document.StorageKey, document.FileName, document.ContentType, documentURLExpiration)
		if err != nil {
			return nil, err
		}
		downloadURL, err := d.storage.GetDownloadURL(context.Background(), document.StorageKey, document.FileName, documentURLExpiration)
		if err != nil {
			return nil, err
		}
		document.ViewURL = viewURL
		document.DownloadURL = downloadURL
		hydrated = append(hydrated, document)
	}
	return hydrated, nil
}

func buildClientDocumentStorageKey(companyID, clientID, fileName string) string {
	return fmt.Sprintf("companies/%s/clients/%s/%s%s", companyID, clientID, uuid.NewString(), strings.ToLower(filepath.Ext(fileName)))
}

func buildDiaryDocumentStorageKey(companyID, projectID, entryID, fileName string) string {
	return fmt.Sprintf("companies/%s/projects/%s/diary/%s/%s%s", companyID, projectID, entryID, uuid.NewString(), strings.ToLower(filepath.Ext(fileName)))
}

func sanitizeDocumentFileName(fileName string) string {
	name := strings.TrimSpace(filepath.Base(fileName))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "document"
	}

	ext := strings.ToLower(filepath.Ext(name))
	baseName := strings.TrimSpace(strings.TrimSuffix(name, filepath.Ext(name)))
	if baseName == "" {
		baseName = "document"
	}

	var builder strings.Builder
	lastWasSeparator := false
	for _, r := range baseName {
		switch {
		case r > unicode.MaxASCII:
			if !lastWasSeparator {
				builder.WriteByte('-')
				lastWasSeparator = true
			}
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastWasSeparator = false
		case r == '-' || r == '_':
			builder.WriteRune(r)
			lastWasSeparator = false
		case unicode.IsSpace(r) || r == '.':
			if !lastWasSeparator {
				builder.WriteByte('-')
				lastWasSeparator = true
			}
		default:
			if !lastWasSeparator {
				builder.WriteByte('-')
				lastWasSeparator = true
			}
		}
	}

	sanitized := strings.Trim(builder.String(), "-_.")
	if sanitized == "" {
		sanitized = "document"
	}

	if ext == "" {
		return sanitized
	}

	return sanitized + ext
}

func copyReader(body io.Reader) io.Reader {
	return body
}
