package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const maxDocumentSize = 10 * 1024 * 1024

var allowedDocumentTypes = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"image/webp":      {},
	"application/pdf": {},
}

var allowedDocumentExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
	".pdf":  {},
}

func validateDocumentUpload(header *multipart.FileHeader, file multipart.File) (string, error) {
	if header == nil {
		return "", fmt.Errorf("file is required")
	}

	if header.Size <= 0 {
		return "", fmt.Errorf("empty file is not allowed")
	}

	if header.Size > maxDocumentSize {
		return "", fmt.Errorf("file exceeds max size of 10MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if _, ok := allowedDocumentExtensions[ext]; !ok {
		return "", fmt.Errorf("unsupported file extension")
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("read upload: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("reset upload stream: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])
	if _, ok := allowedDocumentTypes[contentType]; !ok {
		return "", fmt.Errorf("unsupported file type")
	}

	return contentType, nil
}
