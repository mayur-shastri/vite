package tag

import (
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

type TagRepository interface {
	FindResourceByID(id uint) (*database.Resource, error)
	FindTagByID(id uint) (*database.Tag, error)
	FindOrCreateTagByName(name string) (*database.Tag, error)
	AddTagToResource(resource *database.Resource, tag *database.Tag) error
	RemoveTagFromResource(resource *database.Resource, tag *database.Tag) error
}

// NewTagRepository creates a new instance of the tag repository.
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) FindResourceByID(id uint) (*database.Resource, error) {
	var resource database.Resource
	if err := r.db.Preload("PhysicalFile").Preload("Tags").First(&resource, id).Error; err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r *tagRepository) FindTagByID(id uint) (*database.Tag, error) {
	var tag database.Tag
	if err := r.db.First(&tag, id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) FindOrCreateTagByName(name string) (*database.Tag, error) {
	var tag database.Tag
	if err := r.db.Where(database.Tag{Name: name}).FirstOrCreate(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) AddTagToResource(resource *database.Resource, tag *database.Tag) error {
	return r.db.Model(resource).Association("Tags").Append(tag)
}

func (r *tagRepository) RemoveTagFromResource(resource *database.Resource, tag *database.Tag) error {
	return r.db.Model(resource).Association("Tags").Delete(tag)
}
