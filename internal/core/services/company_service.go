package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyService struct {
	companyRepo ports.CompanyRepository
	linkRepo    ports.LinkRepository
}

func NewCompanyService(companyRepo ports.CompanyRepository, linkRepo ports.LinkRepository) *CompanyService {
	return &CompanyService{
		companyRepo: companyRepo,
		linkRepo:    linkRepo,
	}
}

func (s *CompanyService) CreateCompany(name, cnpj, email, phone, address string) (*domain.Company, error) {
	defaultSlug, err := GenerateDefaultCompanySlug(s.companyRepo, name)
	if err != nil {
		return nil, err
	}

	company := &domain.Company{
		ID:         uuid.New().String(),
		Name:       name,
		CNPJ:       cnpj,
		Email:      email,
		Phone:      phone,
		Address:    address,
		Slug:       defaultSlug,
		PublicName: name,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.companyRepo.CreateCompany(company); err != nil {
		if isCompanySlugUniqueViolation(err) {
			return nil, errors.New("slug already in use")
		}
		return nil, err
	}

	return company, nil
}

func (s *CompanyService) GetCompany(id string) (*domain.Company, error) {
	return s.companyRepo.GetCompanyByID(id)
}

func (s *CompanyService) UpdateCompany(id, name, email, phone, address string) (*domain.Company, error) {
	company, err := s.companyRepo.GetCompanyByID(id)
	if err != nil {
		return nil, err
	}

	company.Name = name
	company.Email = email
	company.Phone = phone
	company.Address = address
	company.UpdatedAt = time.Now()

	if err := s.companyRepo.UpdateCompany(company); err != nil {
		if isCompanySlugUniqueViolation(err) {
			return nil, errors.New("slug already in use")
		}
		return nil, err
	}

	return company, nil
}

func (s *CompanyService) UpdatePublicPage(companyID, slug, publicName, bio string) (*domain.Company, error) {
	company, err := s.companyRepo.GetCompanyByID(companyID)
	if err != nil {
		return nil, err
	}

	normalizedSlug := Slugify(slug)
	if normalizedSlug == "" {
		return nil, errors.New("slug is required")
	}

	existingCompany, err := s.companyRepo.GetCompanyBySlug(normalizedSlug)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingCompany != nil && existingCompany.ID != companyID {
		return nil, errors.New("slug already in use")
	}

	company.Slug = normalizedSlug
	company.PublicName = publicName
	company.PublicBio = bio
	company.UpdatedAt = time.Now()

	if err := s.companyRepo.UpdateCompany(company); err != nil {
		if isCompanySlugUniqueViolation(err) {
			return nil, errors.New("slug already in use")
		}
		return nil, err
	}

	return company, nil
}

func (s *CompanyService) GetPublicPageBySlug(slug string) (*domain.PublicCompanyProfile, error) {
	company, err := s.companyRepo.GetCompanyBySlug(Slugify(slug))
	if err != nil {
		return nil, err
	}

	links, err := s.linkRepo.GetAllLinks(company.ID)
	if err != nil {
		return nil, err
	}
	if links == nil {
		links = []domain.Link{}
	}

	publicName := company.PublicName
	if publicName == "" {
		publicName = company.Name
	}

	return &domain.PublicCompanyProfile{
		CompanyID:  company.ID,
		Slug:       company.Slug,
		PublicName: publicName,
		Bio:        company.PublicBio,
		Avatar:     company.PublicAvatar,
		Links:      links,
	}, nil
}

func isCompanySlugUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "idx_companies_slug") ||
		strings.Contains(errMsg, "companies_slug_key")
}
