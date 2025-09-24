package permission

import (
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	CreateOrUpdate(permission *database.Permission) error
	Delete(resourceID, userID uint) error
	FindPermission(resourceID, userID uint) (*database.Permission, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// CreateOrUpdate performs an "upsert". If a permission for that user/resource
// already exists, it updates the role; otherwise, it creates a new record.
func (r *repository) CreateOrUpdate(permission *database.Permission) error {
	// On a conflict of the primary key (ResourceID, UserID), update the "role" column.
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "resource_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role"}),
	}).Create(permission).Error
}

func (r *repository) Delete(resourceID, userID uint) error {
	return r.db.Where("resource_id = ? AND user_id = ?", resourceID, userID).Delete(&database.Permission{}).Error
}

func (r *repository) FindPermission(resourceID, userID uint) (*database.Permission, error) {
	var permission database.Permission

	subQuery := r.db.Model(&database.ResourceAncestor{}).
		Select("ancestor_id").
		Where("descendant_id = ?", resourceID)

	// Check permission on resource itself OR any ancestor
	err := r.db.Where("(resource_id = ? OR resource_id IN (?)) AND user_id = ?", resourceID, subQuery, userID).
		First(&permission).Error

	return &permission, err
}

