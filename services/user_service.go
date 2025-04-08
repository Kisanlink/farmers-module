package services

import (
	"github.com/Kisanlink/farmers-module/repositories"
)

type UserServiceInterface interface {
	VerifyUserAndType(userId string) (exists bool, isKisansathi bool, err error)
}

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

func NewUserService(repo repositories.UserRepositoryInterface) UserServiceInterface {
	return &UserService{userRepo: repo}
}

func (s *UserService) VerifyUserAndType(userId string) (bool, bool, error) {
	exists, err := s.userRepo.UserExists(userId)
	if err != nil || !exists {
		return exists, false, err
	}

	isKisansathi := s.userRepo.IsKisansathi(userId)
	return true, isKisansathi, nil
}
