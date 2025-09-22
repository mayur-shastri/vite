package permission

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/internal/folders"
	"github.com/bhavyajaix/BalkanID-filevault/internal/middleware"
	"github.com/bhavyajaix/BalkanID-filevault/internal/user"
)

func getUserIDFromContext(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(middleware.UserContextKey).(uint)
	if !ok {
		return 0, errors.New("unauthorized")
	}
	return userID, nil
}

type Service interface {
	GrantPermission(ctx context.Context, resourceID uint, targetEmail string, role database.RoleType) (*database.Resource, error)
	RevokePermission(ctx context.Context, resourceID uint, targetEmail string) (*database.Resource, error)
}

type service struct {
	permRepo     Repository
	resourceRepo folders.Repository
	userRepo     user.Repository
}

func NewService(permRepo Repository, resourceRepo folders.Repository, userRepo user.Repository) Service {
	return &service{permRepo: permRepo, resourceRepo: resourceRepo, userRepo: userRepo}
}

func (s *service) GrantPermission(ctx context.Context, resourceID uint, targetEmail string, role database.RoleType) (*database.Resource, error) {
	// 1. Get the current user (the one granting permission).
	ownerID, err := getUserIDFromContext(ctx) // Assuming you create this helper
	if err != nil {
		return nil, err
	}

	fmt.Println("1. cooked");

	// 2. Security Check: Verify the current user owns the resource.
	resourceToShare, err := s.resourceRepo.GetByID(resourceID)
	if err != nil {
		return nil, errors.New("resource not found")
	}
	if resourceToShare.OwnerID != ownerID {
		return nil, errors.New("access denied: only the owner can grant permissions")
	}

	fmt.Println("2. cooked");

	// 3. Find the user to share with by their email.
	targetUser, err := s.userRepo.GetUserByEmail(targetEmail)
	if err != nil {
		return nil, errors.New("user with the specified email not found")
	}

	fmt.Println("3. cooked");

	// 4. Business Rule: Prevent owner from sharing with themselves.
	if targetUser.ID == ownerID {
		return nil, errors.New("cannot share a resource with yourself")
	}

	fmt.Println("4. cooked");

	// 5. Create the permission and save it.
	newPermission := &database.Permission{
		ResourceID: resourceID,
		UserID:     targetUser.ID,
		Role:       role,
		CreatedAt:  time.Now(),
	}
	if err := s.permRepo.CreateOrUpdate(newPermission); err != nil {
		return nil, err
	}

	fmt.Println("5. cooked");

	// Return the updated resource (you'll need to Preload permissions)
	return s.resourceRepo.GetByID(resourceID)
}

func (s *service) RevokePermission(ctx context.Context, resourceID uint, targetEmail string) (*database.Resource, error) {
	// 1. Get the current user (the one revoking the permission).
	ownerID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Security Check: Verify the current user is the owner of the resource.
	resource, err := s.resourceRepo.GetByID(resourceID)
	if err != nil {
		return nil, errors.New("resource not found")
	}
	if resource.OwnerID != ownerID {
		return nil, errors.New("access denied: only the owner can revoke permissions")
	}

	// 3. Find the user whose permission is being revoked.
	targetUser, err := s.userRepo.GetUserByEmail(targetEmail)
	if err != nil {
		return nil, errors.New("user with the specified email not found")
	}

	// 4. Call the repository to delete the permission record.
	if err := s.permRepo.Delete(resourceID, targetUser.ID); err != nil {
		return nil, fmt.Errorf("failed to revoke permission: %w", err)
	}

	// 5. Return the updated resource. When fetched, its permissions list will be updated.
	return s.resourceRepo.GetByID(resourceID)
}
