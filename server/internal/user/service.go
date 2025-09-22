// in: internal/user/service.go
package user

import (
	"fmt"

	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

// Service is the interface for user-related business logic.
type Service interface {
	Register(username, email, password string) (*database.User, error)
	Login(email, password string) (*database.User, string, error) // Returns user and token
	GetUserByID(id uint) (*database.User, error)
}

type service struct {
	repo Repository
}

// NewService creates a new user service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Register handles new user registration.
func (s *service) Register(username, email, password string) (*database.User, error) {
	// Check if user already exists
	if _, err := s.repo.GetUserByEmail(email); err == nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}

	newUser := &database.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	// Create user in the database
	if err := s.repo.CreateUser(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// Login handles user authentication.
func (s *service) Login(email, password string) (*database.User, string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email or password") // Generic error
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Passwords don't match
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Generate JWT
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("could not generate token: %w", err)
	}

	return user, token, nil
}

func (s *service) GetUserByID(id uint) (*database.User, error) {
	return s.repo.GetUserByID(id)
}