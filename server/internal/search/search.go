package search

import (
	"context"
	"time"

	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
)

// SearchFilters defines the parameters for a resource search.
type SearchFilters struct {
	OwnerID      uint
	Name         *string
	Types        []string
	MimeTypes    []string
	MinSizeBytes *int64
	MaxSizeBytes *int64
	AfterDate    *time.Time
	BeforeDate   *time.Time
	Tags         []string
	UploaderName *string
}

// Pagination defines the offset and limit for search results.
type Pagination struct {
	Offset int
	Limit  int
}

// Repository defines the database operations for searching.
type Repository interface {
	SearchResources(filters SearchFilters, pagination Pagination) ([]database.Resource, error)
}

// Service defines the business logic for searching resources.
type Service interface {
	Search(ctx context.Context, filters SearchFilters, pagination Pagination) ([]database.Resource, error)
}