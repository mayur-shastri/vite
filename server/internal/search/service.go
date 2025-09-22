package search

import (
	"context"
	"errors"
	"fmt"

	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/internal/middleware"
)

// getUserIDFromContext is a helper to extract the user ID from the context.
// It's good practice to have this in a shared package or define it where needed.
func getUserIDFromContext(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(middleware.UserContextKey).(uint)
	if !ok {
		return 0, errors.New("unauthorized: user ID not found in context")
	}
	return userID, nil
}

type service struct {
	repo Repository // Assuming a Repository interface is defined in this package
}

// NewService creates a new instance of the search service.
func NewSearchService(repo Repository) Service {
	return &service{repo: repo}
}

// Search validates input, applies security and business rules, then calls the repository.
func (s *service) Search(ctx context.Context, filters SearchFilters, pagination Pagination) ([]database.Resource, error) {
	// 1. Get the current user ID from the context.
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Business Rule & Validation: Perform checks on the filter inputs.
	if filters.MinSizeBytes != nil && filters.MaxSizeBytes != nil {
		if *filters.MinSizeBytes > *filters.MaxSizeBytes {
			return nil, errors.New("invalid filters: minSizeBytes cannot be greater than maxSizeBytes")
		}
	}

	// Add other validations for dates, etc., as needed.

	// 3. Security Check: Enforce that the search is scoped to the current user.
	// This is a critical step to ensure users can only search for their own resources.
	filters.OwnerID = userID

	// 4. Call the repository to perform the actual data retrieval.
	resources, err := s.repo.SearchResources(filters, pagination)
	if err != nil {
		// Wrap the error to provide more context if it fails at the repository level.
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	// 5. Return the retrieved resources.
	return resources, nil
}
