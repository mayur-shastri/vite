package tag

import "github.com/bhavyajaix/BalkanID-filevault/internal/database"

type TagService interface {
	AddTag(resourceID uint, tagName string) (*database.Resource, error)
	RemoveTag(resourceID uint, tagID uint) (*database.Resource, error)
}

type tagService struct {
	repo TagRepository
}

// NewTagService creates a new instance of the tag service.
func NewTagService(repo TagRepository) TagService {
	return &tagService{repo: repo}
}

func (s *tagService) AddTag(resourceID uint, tagName string) (*database.Resource, error) {
	// 1. Get the resource
	resource, err := s.repo.FindResourceByID(resourceID)
	if err != nil {
		return nil, err // e.g., resource not found
	}

	// 2. Find or create the tag
	tag, err := s.repo.FindOrCreateTagByName(tagName)
	if err != nil {
		return nil, err
	}

	// 3. Associate them
	if err := s.repo.AddTagToResource(resource, tag); err != nil {
		return nil, err
	}

	// The resource object from FindResourceByID already has the preloaded tags,
	// so we just need to add the new one for the return value.
	resource.Tags = append(resource.Tags, tag)

	return resource, nil
}

func (s *tagService) RemoveTag(resourceID uint, tagID uint) (*database.Resource, error) {
	// 1. Get the resource
	resource, err := s.repo.FindResourceByID(resourceID)
	if err != nil {
		return nil, err
	}

	// 2. Get the tag
	tag, err := s.repo.FindTagByID(tagID)
	if err != nil {
		return nil, err // e.g., tag not found
	}

	// 3. Disassociate them
	if err := s.repo.RemoveTagFromResource(resource, tag); err != nil {
		return nil, err
	}

	// We can refetch the resource to ensure the tag list is accurate
	return s.repo.FindResourceByID(resourceID)
}
