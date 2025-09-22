// share/service.go
package share

import (
	"context"
	// Assume you have an auth package
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/internal/file"
	"github.com/bhavyajaix/BalkanID-filevault/internal/folders"
	"gorm.io/gorm"
	// "example/graph/model" // No longer needed here
)

// Service now returns the database model, not the GraphQL model.
type Service interface {
	ResolveShareLink(ctx context.Context, token string) (*database.Resource, error)
}

type service struct {
	repo             Repository
	folderRepository folders.Repository
	fileRepository   file.Repository
	db               *gorm.DB
}

func NewService(repo Repository, folderRepo folders.Repository, fileRepo file.Repository, db *gorm.DB) Service {
	return &service{
		repo:             repo,
		folderRepository: folderRepo,
		fileRepository:   fileRepo,
		db:               db,
	}
}

// ResolveShareLink's signature is updated to return *database.Resource.
func (s *service) ResolveShareLink(ctx context.Context, token string) (*database.Resource, error) {
	dbResource, err := s.repo.FindResourceByTokenAndUserAccess(ctx, token)
	if err != nil {
		return nil, err
	}

	switch dbResource.Type {
	case database.File:
		download, err := s.fileRepository.GetResourceByID(s.db, dbResource.ID)
		if err != nil {
			return nil, err
		}
		return download, nil
	case database.Folder:
		children, err := s.folderRepository.GetChildren(dbResource.ID)
		if err != nil {
			return nil, err
		}
		dbResource.Children = children
	}

	// **KEY CHANGE**: The conversion logic is removed.
	// We now return the raw, populated database resource directly.
	// The resolver will handle the conversion to the GraphQL model.
	return dbResource, nil
}

// The `toGraphQLResource` helper function is completely REMOVED from this file.
