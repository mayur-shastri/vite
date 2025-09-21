package file

import (
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"gorm.io/gorm"
)

// Repository is the interface for file-related database operations.
type Repository interface {
	// --- Existing Methods ---
	GetPhysicalFileByHash(db *gorm.DB, hash string) (*database.PhysicalFile, error)
	CreatePhysicalFile(db *gorm.DB, pf *database.PhysicalFile) error
	IncrementReferenceCount(db *gorm.DB, physicalFileID uint) error
	CreateResource(db *gorm.DB, resource *database.Resource) error
	CreatePermission(db *gorm.DB, permission *database.Permission) error

	GetResourceByID(db *gorm.DB, id uint) (*database.Resource, error)
	GetPhysicalFileByID(db *gorm.DB, id uint) (*database.PhysicalFile, error)
	DecrementReferenceCount(db *gorm.DB, physicalFileID uint) error
	DeleteResource(db *gorm.DB, id uint) error
	DeletePhysicalFile(db *gorm.DB, id uint) error
	UpdateResource(db *gorm.DB, resource *database.Resource) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new file repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// --- Existing Method Implementations ---

func (r *repository) GetPhysicalFileByHash(db *gorm.DB, hash string) (*database.PhysicalFile, error) {
	var pf database.PhysicalFile
	if err := db.Where("file_hash = ?", hash).First(&pf).Error; err != nil {
		return nil, err
	}
	return &pf, nil
}

func (r *repository) CreatePhysicalFile(db *gorm.DB, pf *database.PhysicalFile) error {
	return db.Create(pf).Error
}

func (r *repository) IncrementReferenceCount(db *gorm.DB, physicalFileID uint) error {
	return db.Model(&database.PhysicalFile{}).Where("id = ?", physicalFileID).UpdateColumn("reference_count", gorm.Expr("reference_count + 1")).Error
}

func (r *repository) CreateResource(db *gorm.DB, resource *database.Resource) error {
	return db.Create(resource).Error
}

func (r *repository) CreatePermission(db *gorm.DB, permission *database.Permission) error {
	return db.Create(permission).Error
}

// GetResourceByID fetches a resource by its primary key.
func (r *repository) GetResourceByID(db *gorm.DB, id uint) (*database.Resource, error) {
	var resource database.Resource
	if err := db.First(&resource, id).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetPhysicalFileByID fetches a physical file by its primary key.
func (r *repository) GetPhysicalFileByID(db *gorm.DB, id uint) (*database.PhysicalFile, error) {
	var pf database.PhysicalFile
	if err := db.First(&pf, id).Error; err != nil {
		return nil, err
	}
	return &pf, nil
}


// DecrementReferenceCount decreases the reference counter for a physical file.
func (r *repository) DecrementReferenceCount(db *gorm.DB, physicalFileID uint) error {
	return db.Model(&database.PhysicalFile{}).Where("id = ?", physicalFileID).UpdateColumn("reference_count", gorm.Expr("reference_count - 1")).Error
}

// DeleteResource deletes a logical resource record.
func (r *repository) DeleteResource(db *gorm.DB, id uint) error {
	return db.Delete(&database.Resource{}, id).Error
}

// DeletePhysicalFile deletes a physical file record.
func (r *repository) DeletePhysicalFile(db *gorm.DB, id uint) error {
	return db.Delete(&database.PhysicalFile{}, id).Error
}

// UpdateResource saves changes to a resource record (for rename and move).
func (r *repository) UpdateResource(db *gorm.DB, resource *database.Resource) error {
	return db.Save(resource).Error
}