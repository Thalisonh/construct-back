package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"time"

	"github.com/google/uuid"
)

type CompanyService struct {
	companyRepo ports.CompanyRepository
}

func NewCompanyService(companyRepo ports.CompanyRepository) *CompanyService {
	return &CompanyService{
		companyRepo: companyRepo,
	}
}

func (s *CompanyService) CreateCompany(name, cnpj, email, phone, address string) (*domain.Company, error) {
	company := &domain.Company{
		ID:        uuid.New().String(),
		Name:      name,
		CNPJ:      cnpj,
		Email:     email,
		Phone:     phone,
		Address:   address,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.companyRepo.CreateCompany(company); err != nil {
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
		return nil, err
	}

	return company, nil
}
