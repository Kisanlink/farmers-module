package repositories

import (
	"log"
  "fmt"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}
type UserRepositoryInterface interface {
	UserExists(userID string) (bool, error)
	IsKisansathi(userID string) bool
}
func (r *UserRepository) UserExists(userID string) (bool, error) {
    var exists bool
    
    // Using Table() similar to IsKisansathi
    err := r.db.Table("farmers").
        Select("1").
        Where("user_id = ? OR kisansathi_user_id = ?", userID, userID).
        Limit(1).
        Scan(&exists).Error
    
    if err != nil {
        log.Printf("UserExists query failed: %v", err)
        return false, fmt.Errorf("database verification failed")
    }
    return exists, nil
}

func (r *UserRepository) IsKisansathi(userID string) bool {
	var exists bool
	err := r.db.Table("farmers").
		Select("1").
		Where("kisansathi_user_id = ?", userID).
		Scan(&exists).Error
	
	if err != nil {
		log.Printf("IsKisansathi query failed: %v", err)
		return false
	}
	return exists
}