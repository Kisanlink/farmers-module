package services

import (
	"github.com/Kisanlink/farmers-module/repositories"
)

type UserServiceInterface interface {
	VerifyUserAndType(user_id string) (exists bool, is_kisansathi bool, err error)
}

type UserService struct {
	UserRepo repositories.UserRepositoryInterface
}

func NewUserService(repo repositories.UserRepositoryInterface) UserServiceInterface {
	return &UserService{UserRepo: repo}
}

func (s *UserService) VerifyUserAndType(user_id string) (bool, bool, error) {
	exists, err := s.UserRepo.UserExists(user_id)
	if err != nil || !exists {
		return exists, false, err
	}

	is_kisansathi := s.UserRepo.IsKisansathi(user_id)
	return true, is_kisansathi, nil
}
