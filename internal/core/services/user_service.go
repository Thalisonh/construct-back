package services

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"errors"
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

	links, _ := s.linkRepo.GetAllLinks(profile.ID)
	if links == nil {
		profile.Links = []domain.Link{}

		return profile, nil
	}

	profile.Links = links

	return profile, nil
}
