package services

import (
	"construct-backend/internal/core/ports"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"gorm.io/gorm"
)

func GenerateDefaultCompanySlug(companyRepo ports.CompanyRepository, companyName string) (string, error) {
	baseSlug := Slugify(companyName)
	if baseSlug == "" {
		baseSlug = "empresa"
	}

	for i := 0; i < 20; i++ {
		randomSuffix, err := generateRandomDigits(4)
		if err != nil {
			return "", err
		}

		candidate := fmt.Sprintf("%s_%s", baseSlug, randomSuffix)
		existingCompany, err := companyRepo.GetCompanyBySlug(candidate)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
		if existingCompany == nil {
			return candidate, nil
		}
	}

	return "", errors.New("could not generate unique company slug")
}

func generateRandomDigits(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than zero")
	}

	digits := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + n.Int64())
	}

	return string(digits), nil
}
