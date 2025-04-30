package repositories

import (
	"fmt"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

type UserRepositoryInterface interface {
	UserExists(user_id string) (bool, error)
	IsKisansathi(user_id string) bool
}

func (r *UserRepository) UserExists(user_id string) (bool, error) {
	var exists bool

	// Using Table() similar to IsKisansathi
	err := r.DB.Table("farmers").
		Select("1").
		Where("user_id = ? OR kisansathi_user_id = ?", user_id, user_id).
		Limit(1).
		Scan(&exists).Error

	if err != nil {
		return false, fmt.Errorf("database verification failed")
	}
	return exists, nil
}

func (r *UserRepository) IsKisansathi(user_id string) bool {
	var exists bool
	err := r.DB.Table("farmers").
		Select("1").
		Where("kisansathi_user_id = ?", user_id).
		Scan(&exists).Error

	if err != nil {
		return false
	}
	return exists
}
