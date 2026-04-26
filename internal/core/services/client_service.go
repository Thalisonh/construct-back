package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

type ClientService struct {
	clientRepo ports.ClientRepository
	documents  documentSupport
}

func NewClientService(clientRepo ports.ClientRepository, storage ports.DocumentStorage) *ClientService {
	return &ClientService{
		clientRepo: clientRepo,
		documents:  newDocumentSupport(storage),
	}
}

func (s *ClientService) CreateClient(companyID, userID, name, phone, address, summary string) (*domain.Client, error) {
	client := &domain.Client{
		ID:        uuid.New().String(),
		UserID:    userID,
		CompanyID: companyID,
		Name:      name,
		Phone:     phone,
		Address:   address,
		Summary:   summary,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.clientRepo.CreateClient(client); err != nil {
		return nil, err
	}

	return client, nil
}

func (s *ClientService) GetClient(id, companyID string) (*domain.Client, error) {
	return s.clientRepo.GetClientByID(id, companyID)
}

func (s *ClientService) ListClients(companyID string) ([]domain.Client, error) {
	return s.clientRepo.GetAllClients(companyID)
}

func (s *ClientService) UpdateClient(id, name, phone, address, summary, companyID string) (*domain.Client, error) {
	client, err := s.clientRepo.GetClientByID(id, companyID)
	if err != nil {
		return nil, err
	}

	client.Name = name
	client.Phone = phone
	client.Address = address
	client.Summary = summary
	client.UpdatedAt = time.Now()

	if err := s.clientRepo.UpdateClient(client); err != nil {
		return nil, err
	}

	return client, nil
}

func (s *ClientService) DeleteClient(id, companyID string) error {
	return s.clientRepo.DeleteClient(id, companyID)
}

func (s *ClientService) AddComment(clientID, content string) (*domain.Comment, error) {
	comment := &domain.Comment{
		ID:        uuid.New().String(),
		ClientID:  clientID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.clientRepo.AddComment(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *ClientService) UploadClientDocument(clientID, companyID, userID, fileName, contentType string, fileSize int64, body io.Reader) (*domain.ClientDocument, error) {
	if err := s.documents.ensureStorage(); err != nil {
		return nil, err
	}

	if _, err := s.clientRepo.GetClientByID(clientID, companyID); err != nil {
		return nil, err
	}

	now := time.Now()
	sanitizedFileName := sanitizeDocumentFileName(fileName)
	document := &domain.ClientDocument{
		ID:          uuid.NewString(),
		ClientID:    clientID,
		CompanyID:   companyID,
		FileName:    sanitizedFileName,
		ContentType: contentType,
		FileSize:    fileSize,
		StorageKey:  buildClientDocumentStorageKey(companyID, clientID, sanitizedFileName),
		UploadedBy:  userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.documents.storage.Upload(context.Background(), document.StorageKey, copyReader(body), contentType, fileSize); err != nil {
		return nil, err
	}

	if err := s.clientRepo.CreateClientDocument(document); err != nil {
		_ = s.documents.storage.Delete(context.Background(), document.StorageKey)
		return nil, err
	}

	documents, err := s.documents.hydrateClientDocuments([]domain.ClientDocument{*document})
	if err != nil {
		return nil, err
	}
	return &documents[0], nil
}

func (s *ClientService) ListClientDocuments(clientID, companyID string) ([]domain.ClientDocument, error) {
	documents, err := s.clientRepo.GetClientDocuments(clientID, companyID)
	if err != nil {
		return nil, err
	}
	return s.documents.hydrateClientDocuments(documents)
}

func (s *ClientService) DeleteClientDocument(clientID, documentID, companyID string) error {
	document, err := s.clientRepo.GetClientDocumentByID(documentID, clientID, companyID)
	if err != nil {
		return err
	}

	if err := s.clientRepo.DeleteClientDocument(documentID, clientID, companyID); err != nil {
		return err
	}

	if err := s.documents.ensureStorage(); err != nil {
		return err
	}
	return s.documents.storage.Delete(context.Background(), document.StorageKey)
}
