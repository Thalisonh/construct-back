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

func (s *LinkService) CreateLink(companyID, userID, url, description string) (*domain.Link, error) {
	link := &domain.Link{
		ID:          uuid.New().String(),
		URL:         url,
		Description: description,
		UserID:      userID,
		CompanyID:   companyID,
		CreatedAt:   time.Now(),
	}

	if err := s.linkRepo.CreateLink(link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) UpdateLink(companyID, url, description, id string) (*domain.Link, error) {
	link := &domain.Link{
		ID:          id,
		URL:         url,
		Description: description,
		CompanyID:   companyID,
		UpdatedAt:   time.Now(),
	}

	err := s.linkRepo.UpdateLink(link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) ListLinks(companyID string) ([]domain.Link, error) {
	return s.linkRepo.GetAllLinks(companyID)
}

func (s *LinkService) DeleteLink(id, companyID string) error {
	return s.linkRepo.DeleteLink(id, companyID)
}

func (s *LinkService) TrackLinkClick(id string) error {
	return s.linkRepo.RegisterClick(id)
}
