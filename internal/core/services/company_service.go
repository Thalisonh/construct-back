package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyService struct {
	companyRepo ports.CompanyRepository
	linkRepo    ports.LinkRepository
	storage     ports.DocumentStorage
}

func NewCompanyService(companyRepo ports.CompanyRepository, linkRepo ports.LinkRepository, storage ports.DocumentStorage) *CompanyService {
	return &CompanyService{
		companyRepo: companyRepo,
		linkRepo:    linkRepo,
		storage:     storage,
	}
}

func (s *CompanyService) CreateCompany(name, cnpj, email, phone, address string) (*domain.Company, error) {
	defaultSlug, err := GenerateDefaultCompanySlug(s.companyRepo, name)
	if err != nil {
		return nil, err
	}

	company := &domain.Company{
		ID:          uuid.New().String(),
		Name:        name,
		CNPJ:        cnpj,
		Email:       email,
		Phone:       phone,
		Address:     address,
		Slug:        defaultSlug,
		PublicName:  name,
		PublicTheme: DefaultPublicTheme,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
	company, err := s.companyRepo.GetCompanyByID(id)
	if err != nil {
		return nil, err
	}
	return s.hydrateCompanyPublicAssets(company)
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

	return s.hydrateCompanyPublicAssets(company)
}

const DefaultPublicTheme = "minimal"

var validPublicThemes = map[string]bool{
	"minimal":      true,
	"professional": true,
	"premium":      true,
	"vibrant":      true,
}

func normalizePublicTheme(theme string) (string, error) {
	normalizedTheme := strings.TrimSpace(strings.ToLower(theme))
	if normalizedTheme == "" {
		return DefaultPublicTheme, nil
	}
	if !validPublicThemes[normalizedTheme] {
		return "", errors.New("invalid public theme")
	}
	return normalizedTheme, nil
}

func (s *CompanyService) UpdatePublicPage(companyID, slug, publicName, bio, theme string) (*domain.Company, error) {
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

	normalizedTheme, err := normalizePublicTheme(company.PublicTheme)
	if err != nil {
		normalizedTheme = DefaultPublicTheme
	}
	if strings.TrimSpace(theme) != "" {
		normalizedTheme, err = normalizePublicTheme(theme)
		if err != nil {
			return nil, err
		}
	}

	company.Slug = normalizedSlug
	company.PublicName = publicName
	company.PublicBio = bio
	company.PublicTheme = normalizedTheme
	company.UpdatedAt = time.Now()

	if err := s.companyRepo.UpdateCompany(company); err != nil {
		if isCompanySlugUniqueViolation(err) {
			return nil, errors.New("slug already in use")
		}
		return nil, err
	}

	return s.hydrateCompanyPublicAssets(company)
}

const maxCompanyLogoSize = 2 * 1024 * 1024

var validCompanyLogoContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func (s *CompanyService) UploadCompanyLogo(companyID, fileName, contentType string, fileSize int64, body io.Reader) (*domain.Company, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("document storage is not configured")
	}

	if fileSize <= 0 {
		return nil, fmt.Errorf("file is empty")
	}
	if fileSize > maxCompanyLogoSize {
		return nil, fmt.Errorf("logo must be up to 2MB")
	}
	if !validCompanyLogoContentTypes[contentType] {
		return nil, fmt.Errorf("logo must be JPG, PNG or WebP")
	}

	company, err := s.companyRepo.GetCompanyByID(companyID)
	if err != nil {
		return nil, err
	}

	sanitizedFileName := sanitizeDocumentFileName(fileName)
	storageKey := buildCompanyLogoStorageKey(companyID, sanitizedFileName)
	oldStorageKey := company.PublicAvatar

	if err := s.storage.Upload(context.Background(), storageKey, copyReader(body), contentType, fileSize); err != nil {
		return nil, err
	}

	company.PublicAvatar = storageKey
	company.UpdatedAt = time.Now()

	if err := s.companyRepo.UpdateCompany(company); err != nil {
		_ = s.storage.Delete(context.Background(), storageKey)
		return nil, err
	}

	if oldStorageKey != "" && !isExternalURL(oldStorageKey) && oldStorageKey != storageKey {
		_ = s.storage.Delete(context.Background(), oldStorageKey)
	}

	return s.hydrateCompanyPublicAssets(company)
}

func buildCompanyLogoStorageKey(companyID, fileName string) string {
	return fmt.Sprintf("companies/%s/public/logo/%s%s", companyID, uuid.NewString(), strings.ToLower(filepath.Ext(fileName)))
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
	publicTheme, err := normalizePublicTheme(company.PublicTheme)
	if err != nil {
		publicTheme = DefaultPublicTheme
	}
	publicAvatar := company.PublicAvatar
	if publicAvatar != "" {
		hydratedAvatar, err := s.hydratePublicAsset(publicAvatar, "company-logo")
		if err == nil {
			publicAvatar = hydratedAvatar
		}
	}

	return &domain.PublicCompanyProfile{
		CompanyID:  company.ID,
		Slug:       company.Slug,
		PublicName: publicName,
		Bio:        company.PublicBio,
		Avatar:     publicAvatar,
		Theme:      publicTheme,
		Links:      links,
	}, nil
}

func (s *CompanyService) hydrateCompanyPublicAssets(company *domain.Company) (*domain.Company, error) {
	if company == nil {
		return nil, nil
	}

	hydrated := *company
	if hydrated.PublicAvatar != "" {
		avatarURL, err := s.hydratePublicAsset(hydrated.PublicAvatar, "company-logo")
		if err == nil {
			hydrated.PublicAvatar = avatarURL
		}
	}

	if _, err := normalizePublicTheme(hydrated.PublicTheme); err != nil {
		hydrated.PublicTheme = DefaultPublicTheme
	}

	return &hydrated, nil
}

func (s *CompanyService) hydratePublicAsset(storageKey, fileName string) (string, error) {
	if isExternalURL(storageKey) || s.storage == nil {
		return storageKey, nil
	}

	ext := strings.ToLower(filepath.Ext(storageKey))
	contentType := "image/png"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".webp":
		contentType = "image/webp"
	}

	return s.storage.GetViewURL(context.Background(), storageKey, fileName+ext, contentType, documentURLExpiration)
}

func isExternalURL(value string) bool {
	lowerValue := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(lowerValue, "http://") || strings.HasPrefix(lowerValue, "https://")
}

func isCompanySlugUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "idx_companies_slug") ||
		strings.Contains(errMsg, "companies_slug_key")
}
