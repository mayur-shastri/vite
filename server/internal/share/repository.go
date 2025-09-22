// share/repository.go
package share

import (
	"context"
	"errors"

	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/internal/middleware"
	"gorm.io/gorm"
)

func getUserIDFromContext(ctx context.Context) (uint, error) {
	userID, ok := ctx.Value(middleware.UserContextKey).(uint)
	if !ok {
		return 0, errors.New("unauthorized")
	}
	return userID, nil
}

// Repository defines the interface for share-related database operations.
type Repository interface {
	FindResourceByTokenAndUserAccess(ctx context.Context, token string) (*database.Resource, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new share repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindResourceByTokenAndUserAccess(ctx context.Context, token string) (*database.Resource, error) {
	// 1. Find the resource using the share token
	var resource database.Resource
	userID, _ := getUserIDFromContext(ctx)
	if err := r.db.WithContext(ctx).
		Preload("PhysicalFile").
		Where("share_token = ?", token).
		First(&resource).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("share link not found or invalid")
		}
		return nil, err
	}

	// 2. Check for direct permission on the resource itself
	var directPermissionCount int64
	if err := r.db.Model(&database.Permission{}).
		Where("resource_id = ? AND user_id = ?", resource.ID, userID).
		Count(&directPermissionCount).Error; err != nil {
		return nil, err
	}

	if directPermissionCount > 0 {
		return &resource, nil
	}

	// 3. If no direct permission, check for permission on any ancestor
	// This uses a subquery for efficiency:
	// SELECT COUNT(*) FROM "permissions"
	// WHERE user_id = ? AND resource_id IN (
	//   SELECT ancestor_id FROM "resource_ancestors" WHERE descendant_id = ?
	// )
	var ancestorPermissionCount int64
	subQuery := r.db.Model(&database.ResourceAncestor{}).Select("ancestor_id").Where("descendant_id = ?", resource.ID)

	if err := r.db.Model(&database.Permission{}).
		Where("user_id = ? AND resource_id IN (?)", userID, subQuery).
		Count(&ancestorPermissionCount).Error; err != nil {
		return nil, err
	}

	if ancestorPermissionCount > 0 {
		// User has inherited access from a parent folder
		return &resource, nil
	}

	// If we reach here, the user has no direct or inherited permission
	return nil, errors.New("access denied")
}
