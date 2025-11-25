package services

import (
	"construct-backend/internal/core/ports"
	"errors"
)

type UserService struct {
	userRepo ports.UserRepository
}

func NewUserService(userRepo ports.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
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
