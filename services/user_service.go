package services

import (
	"github.com/Kisanlink/farmers-module/repositories"
)

type UserServiceInterface interface {
	VerifyUserAndType(userID string) (exists bool, isKisansathi bool, err error)
}

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewUserService(repo repositories.UserRepositoryInterface) UserServiceInterface {
	return &UserService{userRepo: repo}
}

func (s *UserService) VerifyUserAndType(userID string) (bool, bool, error) {
	exists, err := s.userRepo.UserExists(userID)
	if err != nil || !exists {
		return exists, false, err
	}
	
	isKisansathi := s.userRepo.IsKisansathi(userID)
	return true, isKisansathi, nil
}