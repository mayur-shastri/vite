// in: internal/user/repository.go
package user

import (
	"github.com/bhavyajaix/BalkanID-filevault/internal/database" // Your GORM models package
	"gorm.io/gorm"
)

// Repository is the interface for database operations.
type Repository interface {
	CreateUser(user *database.User) error
	GetUserByID(id uint) (*database.User, error)
	GetUserByEmail(email string) (*database.User, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// CreateUser saves a new user to the database.
func (r *repository) CreateUser(user *database.User) error {
	return r.db.Create(user).Error
}

// GetUserByEmail finds a user by their email address.
func (r *repository) GetUserByEmail(email string) (*database.User, error) {
	var user database.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err // GORM returns gorm.ErrRecordNotFound if not found
	}
	return &user, nil
}

func (r *repository) GetUserByID(id uint) (*database.User, error) {
	var user database.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
