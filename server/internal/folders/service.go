package folders

import (
	"context"
	"errors"
	"fmt"

	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/internal/middleware"
	"github.com/bhavyajaix/BalkanID-filevault/pkg/utils"
)

type Service interface {
	CreateFolder(ctx context.Context, name string, parentID *uint) (*database.Resource, error)
	GetResources(ctx context.Context, folderID *uint) ([]database.Resource, error)
	RenameResource(ctx context.Context, resourceID uint, newName string) (*database.Resource, error)
	MoveResource(ctx context.Context, resourceID uint, newParentID *uint) (*database.Resource, error)
	DeleteResource(ctx context.Context, resourceID uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Helper to get user ID from context
func getUserIDFromContext(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(middleware.UserContextKey).(uint)
	if !ok {
		return 0, errors.New("unauthorized")
	}
	return userID, nil
}

func (s *service) CreateFolder(ctx context.Context, name string, parentID *uint) (*database.Resource, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if parentID != nil {
		parent, err := s.repo.GetByID(*parentID)
		if err != nil || parent.OwnerID != userID || parent.Type != database.Folder {
			return nil, errors.New("access denied or invalid parent folder")
		}
	}
	token, err := utils.GenerateUUIDToken()
	if err != nil {
		return nil, fmt.Errorf("could not generate share token: %w", err)
	}
	folder := &database.Resource{
		Name:       name,
		OwnerID:    userID,
		ParentID:   parentID,
		Type:       database.Folder,
		ShareToken: &token,
	}

	err = s.repo.Create(folder)
	return folder, err
}

func (s *service) GetResources(ctx context.Context, folderID *uint) ([]database.Resource, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if folderID == nil {
		return s.repo.GetRoot(userID)
	}

	parent, err := s.repo.GetByID(*folderID)
	if err != nil {
		return nil, errors.New("folder not found")
	}
	if parent.OwnerID != userID {
		return nil, errors.New("access denied")
	}

	return s.repo.GetChildren(*folderID)
}

func (s *service) RenameResource(ctx context.Context, resourceID uint, newName string) (*database.Resource, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	resource, err := s.repo.GetByID(resourceID)
	if err != nil {
		return nil, errors.New("resource not found")
	}
	if resource.OwnerID != userID {
		return nil, errors.New("access denied")
	}

	resource.Name = newName
	err = s.repo.Update(resource)
	return resource, err
}

func (s *service) MoveResource(ctx context.Context, resourceID uint, newParentID *uint) (*database.Resource, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	resource, err := s.repo.GetByID(resourceID)
	if err != nil {
		return nil, errors.New("resource not found")
	}
	if resource.OwnerID != userID {
		return nil, errors.New("access denied to the resource you are moving")
	}

	// If moving to a new parent folder, check its validity and permissions.
	if newParentID != nil {
		newParent, err := s.repo.GetByID(*newParentID) // Dereference the pointer here
		if err != nil {
			return nil, errors.New("destination folder not found")
		}
		if newParent.OwnerID != userID {
			return nil, errors.New("access denied to the destination folder")
		}
		if newParent.Type != database.Folder {
			return nil, errors.New("can only move resources into a folder")
		}
	}
	// If newParentID is nil, we are moving to the root, so no parent checks are needed.

	resource.ParentID = newParentID // Assign the pointer directly
	if err := s.repo.Update(resource); err != nil {
		return nil, err
	}
	return resource, nil
}

func (s *service) DeleteResource(ctx context.Context, resourceID uint) error {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	resource, err := s.repo.GetByID(resourceID)
	if err != nil {
		return errors.New("resource not found")
	}
	if resource.OwnerID != userID {
		return errors.New("access denied")
	}

	return s.repo.Delete(resourceID)
}
