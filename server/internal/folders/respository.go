package folders

import (
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"gorm.io/gorm"
)

type Repository interface {
	Create(resource *database.Resource) error
	GetByID(id uint) (*database.Resource, error)
	GetChildren(parentID uint) ([]database.Resource, error)
	GetRoot(ownerID uint) ([]database.Resource, error)
	Update(resource *database.Resource) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(resource *database.Resource) error {
	return r.db.Create(resource).Error
}

func (r *repository) GetByID(id uint) (*database.Resource, error) {
	var resource database.Resource
	// Preload the owner's basic info to avoid extra queries later
	if err := r.db.Preload("User").First(&resource, id).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetChildren finds all direct children of a parent resource.
func (r *repository) GetChildren(parentID uint) ([]database.Resource, error) {
	var children []database.Resource
	// Add .Preload("PhysicalFile") to fetch the associated physical file for any children that are files.
	err := r.db.Preload("PhysicalFile").Where("parent_id = ?", parentID).Find(&children).Error
	return children, err
}

// GetRoot finds all resources with no parent for a specific user.
func (r *repository) GetRoot(ownerID uint) ([]database.Resource, error) {
	var resources []database.Resource
	err := r.db.Preload("PhysicalFile").Where("parent_id IS NULL AND owner_id = ?", ownerID).Find(&resources).Error
	return resources, err
}

func (r *repository) Update(resource *database.Resource) error {
	return r.db.Save(resource).Error
}

// Delete will perform a soft delete because of gorm.Model
func (r *repository) Delete(id uint) error {
	return r.db.Delete(&database.Resource{}, id).Error
}
