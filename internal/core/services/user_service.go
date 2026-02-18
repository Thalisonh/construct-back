package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"errors"

	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo ports.UserRepository
	linkRepo ports.LinkRepository
}

func NewUserService(userRepo ports.UserRepository, linkRepo ports.LinkRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		linkRepo: linkRepo,
	}
}

func Slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	reg, _ := regexp.Compile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func (s *UserService) VerifyUserName(username string) error {
	user, _ := s.userRepo.VerifyUserName(username)
	if user == nil {
		return nil
	}

	return errors.New("username already exists")
}

func (s *UserService) UpdateUsername(userID, username string) error {
	user, _ := s.userRepo.VerifyUserName(username)
	if user != nil {
		return errors.New("username already exists")
	}

	if err := s.userRepo.UpdateUsername(userID, username); err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetUsername(userID string) (string, error) {
	username, err := s.userRepo.GetUsername(userID)
	if err != nil {
		return "", err
	}

	return username, nil
}

func (s *UserService) GetPublicProfile(username string) (*domain.PublicProfile, error) {
	profile, err := s.userRepo.GetPublicProfile(username)
	if err != nil {
		return nil, err
	}

	links, _ := s.linkRepo.GetAllLinks(profile.CompanyID)
	if links == nil {
		profile.Links = []domain.Link{}

		return profile, nil
	}

	profile.Links = links

	return profile, nil
}

func (s *UserService) UpdateBio(userID, bio string) error {
	if err := s.userRepo.UpdateBio(userID, bio); err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetProfile(userID string) (*domain.User, error) {
	return s.userRepo.GetUserByID(userID)
}

func (s *UserService) UpdateProfile(userID, name, email, phone, companyID string) error {
	user := &domain.User{
		ID:        userID,
		Name:      name,
		Email:     email,
		Phone:     phone,
		CompanyID: companyID,
	}

	return s.userRepo.UpdateProfile(user)
}

func (s *UserService) UpdatePassword(userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(userID, string(hashedPassword))
}

func (s *UserService) GetCompanyMembers(companyID string) ([]domain.User, error) {
	return s.userRepo.ListUsersByCompanyID(companyID)
}

func (s *UserService) AddCompanyMember(companyID, email, name, password, role string) (*domain.User, error) {
	existingUser, _ := s.userRepo.GetUserByEmail(email)
	if existingUser != nil {
		return nil, errors.New("usuário já existe")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Username:  Slugify(name),
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		CompanyID: companyID,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}
