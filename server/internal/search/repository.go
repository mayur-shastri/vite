package search

import (
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"gorm.io/gorm"
)

type searchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) Repository {
	return &searchRepository{db: db}
}

func (r *searchRepository) SearchResources(filters SearchFilters, pagination Pagination) ([]database.Resource, error) {
	var resources []database.Resource

	// Start with the base query, joining with tables we need for filtering
	query := r.db.Model(&database.Resource{}).
		Joins("LEFT JOIN physical_files ON physical_files.id = resources.physical_file_id")

	// --- Dynamically apply filters ---

	// Always scope to the owner
	query = query.Where("resources.owner_id = ?", filters.OwnerID)

	if filters.Name != nil {
		query = query.Where("resources.name LIKE ?", "%"+*filters.Name+"%")
	}

	if len(filters.Types) > 0 {
		query = query.Where("resources.type IN ?", filters.Types)
	}

	if len(filters.MimeTypes) > 0 {
		query = query.Where("physical_files.mime_type IN ?", filters.MimeTypes)
	}

	if filters.MinSizeBytes != nil {
		query = query.Where("physical_files.size_bytes >= ?", *filters.MinSizeBytes)
	}

	if filters.MaxSizeBytes != nil {
		query = query.Where("physical_files.size_bytes <= ?", *filters.MaxSizeBytes)
	}

	if filters.AfterDate != nil {
		query = query.Where("resources.created_at >= ?", *filters.AfterDate)
	}

	if filters.BeforeDate != nil {
		query = query.Where("resources.created_at <= ?", *filters.BeforeDate)
	}

	// Tag filtering is the most complex due to the many-to-many relationship
	if len(filters.Tags) > 0 {
		query = query.Joins("JOIN resource_tags ON resource_tags.resource_id = resources.id").
			Joins("JOIN tags ON tags.id = resource_tags.tag_id").
			Where("tags.name IN ?", filters.Tags).
			Group("resources.id").
			Having("COUNT(DISTINCT tags.name) = ?", len(filters.Tags)) // Ensure all tags match
	}

	// --- Apply pagination and execute ---
	err := query.
		Order("resources.created_at DESC").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		// Preload associations to avoid N+1 queries in the GraphQL resolver
		Preload("User").
		Preload("PhysicalFile").
		Preload("Tags").
		Find(&resources).Error

	return resources, err
}
