package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"

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

func (s *LinkService) CreateLink(projectID, url, description string) (*domain.Link, error) {
	link := &domain.Link{
		ID:          uuid.New().String(),
		URL:         url,
		Description: description,
		ProjectID:   projectID,
	}

	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) ListLinks(projectID string) ([]domain.Link, error) {
	return s.linkRepo.GetAllLinks(projectID)
}

func (s *LinkService) DeleteLink(id string) error {
	return s.linkRepo.DeleteLink(id)
}
