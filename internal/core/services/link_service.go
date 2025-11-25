package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"time"

	"github.com/google/uuid"
)

type LinkService struct {
	linkRepo ports.LinkRepository
}

func NewLinkService(linkRepo ports.LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}

func (s *LinkService) CreateLink(userID, url, description string) (*domain.Link, error) {
	link := &domain.Link{
		ID:          uuid.New().String(),
		URL:         url,
		Description: description,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}

	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) UpdateLink(userID, url, description string) (*domain.Link, error) {
	link := &domain.Link{
		ID:          userID,
		URL:         url,
		Description: description,
		UserID:      userID,
		UpdatedAt:   time.Now(),
	}

	err := s.linkRepo.UpdateLink(link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) ListLinks(userID string) ([]domain.Link, error) {
	return s.linkRepo.GetAllLinks(userID)
}

func (s *LinkService) DeleteLink(id string) error {
	return s.linkRepo.DeleteLink(id)
}
