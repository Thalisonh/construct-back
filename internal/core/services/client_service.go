package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"time"

	"github.com/google/uuid"
)

type ClientService struct {
	clientRepo ports.ClientRepository
}

func NewClientService(clientRepo ports.ClientRepository) *ClientService {
	return &ClientService{
		clientRepo: clientRepo,
	}
}

func (s *ClientService) CreateClient(userID, name, phone, address, summary string) (*domain.Client, error) {
	client := &domain.Client{
		ID:        uuid.New().String(),
		UserID:    userID,
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

func (s *ClientService) GetClient(id, userID string) (*domain.Client, error) {
	return s.clientRepo.GetClientByID(id, userID)
}

func (s *ClientService) ListClients(userID string) ([]domain.Client, error) {
	return s.clientRepo.GetAllClients(userID)
}

func (s *ClientService) UpdateClient(id, name, phone, address, summary, userID string) (*domain.Client, error) {
	client, err := s.clientRepo.GetClientByID(id, userID)
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

func (s *ClientService) DeleteClient(id string) error {
	return s.clientRepo.DeleteClient(id)
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
