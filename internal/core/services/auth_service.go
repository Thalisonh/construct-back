package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo    ports.UserRepository
	companyRepo ports.CompanyRepository
	secret      string
}

func NewAuthService(userRepo ports.UserRepository, companyRepo ports.CompanyRepository, secret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		companyRepo: companyRepo,
		secret:      secret,
	}
}

func (s *AuthService) Signup(email, password, name, companyName, cnpj string) (string, error) {
	existingUser, _ := s.userRepo.GetUserByEmail(email)
	if existingUser != nil {
		return "", errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Create Company first
	company := &domain.Company{
		ID:        uuid.New().String(),
		Name:      companyName,
		CNPJ:      cnpj,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     email,
	}

	if err := s.companyRepo.CreateCompany(company); err != nil {
		return "", err
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Username:  Slugify(name),
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		CompanyID: company.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"company_id": user.CompanyID,
		"role":       user.Role,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"company_id": user.CompanyID,
		"role":       user.Role,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) LoginWithGoogle(idToken string) (string, error) {
	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, idToken, os.Getenv("AUDIENCE"))
	if err != nil {
		return "", errors.New("invalid google token")
	}

	email := payload.Claims["email"].(string)
	user, err := s.userRepo.GetUserByEmail(email)
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		return "", err
	}

	if user == nil {
		// Create new user if not exists
		user = &domain.User{
			ID:        uuid.New().String(),
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.userRepo.CreateUser(user); err != nil {
			return "", err
		}
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"company_id": user.CompanyID,
		"role":       user.Role,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) VerifyToken(token string) error {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return err
	}
	return nil
}
