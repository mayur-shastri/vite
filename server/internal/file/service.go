package file

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"gorm.io/gorm"
)

// UploadParams contains all necessary data for a file upload.
type UploadParams struct {
	Upload   graphql.Upload
	OwnerID  uint
	ParentID *uint
}

// Service is the interface for file-related business logic.
type Service interface {
	UploadFile(params UploadParams) (*database.Resource, error)
	DeleteFile(resourceID uint, userID uint) error
	RenameFile(resourceID uint, userID uint, newName string) (*database.Resource, error)
	MoveFile(resourceID uint, userID uint, newParentID *uint) (*database.Resource, error)
}

type service struct {
	repo        Repository
	db          *gorm.DB
	storagePath string
}

// NewService creates a new file service.
func NewService(repo Repository, db *gorm.DB, storagePath string) Service {
	return &service{repo: repo, db: db, storagePath: storagePath}
}

func (s *service) UploadFile(params UploadParams) (*database.Resource, error) {
	// 1. Open the uploaded file stream from the graphql.Upload object
	src := params.Upload.File
	// 2. Calculate the file's SHA-256 hash
	// We need to read the file to hash it, then rewind it to save it.
	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return nil, fmt.Errorf("failed to hash file content: %w", err)
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Begin a database transaction for atomic operations
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	// Defer a rollback in case of any error. It will be ignored if we commit.
	defer tx.Rollback()

	var physicalFileID uint

	// 3. Check for existing physical file (deduplication)
	existingPF, err := s.repo.GetPhysicalFileByHash(tx, hash)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("error checking for existing file: %w", err)
	}

	if existingPF != nil {
		// --- CASE A: FILE IS A DUPLICATE ---
		// The physical file already exists, so we just increment its reference count.
		if err := s.repo.IncrementReferenceCount(tx, existingPF.ID); err != nil {
			return nil, fmt.Errorf("failed to increment reference count: %w", err)
		}
		physicalFileID = existingPF.ID
	} else {
		// --- CASE B: FILE IS UNIQUE ---
		// Derive path from hash and ensure the nested directories exist.
		filePath := s.getStoragePath(hash)
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create storage directories: %w", err)
		}

		// Create the new file on the disk.
		dst, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination file: %w", err)
		}
		defer dst.Close()

		// Rewind the source file reader to the beginning.
		if _, err := src.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("failed to rewind file reader: %w", err)
		}

		// Copy the file content to its new permanent location on the disk.
		if _, err := io.Copy(dst, src); err != nil {
			return nil, fmt.Errorf("failed to save file to disk: %w", err)
		}

		// Create the physical file record in the database.
		newPF := &database.PhysicalFile{
			FileHash:       hash,
			FilePath:       filePath,
			SizeBytes:      params.Upload.Size,
			MimeType:       params.Upload.ContentType,
			ReferenceCount: 1,
		}
		if err := s.repo.CreatePhysicalFile(tx, newPF); err != nil {
			// If DB write fails, try to clean up the orphaned file on disk.
			os.Remove(filePath)
			return nil, fmt.Errorf("failed to create physical file record: %w", err)
		}
		physicalFileID = newPF.ID
	}

	// 4. Create the logical resource record for the user.
	// This acts as the user's "pointer" to the physical file.
	newResource := &database.Resource{
		OwnerID:        params.OwnerID,
		ParentID:       params.ParentID,
		Name:           params.Upload.Filename,
		Type:           database.File,
		PhysicalFileID: &physicalFileID,
	}
	if err := s.repo.CreateResource(tx, newResource); err != nil {
		return nil, fmt.Errorf("failed to create resource record: %w", err)
	}

	// 5. Create the owner's permission record (ACL entry).
	ownerPermission := &database.Permission{
		ResourceID: newResource.ID,
		UserID:     params.OwnerID,
		Role:       database.Editor, // Or a specific "owner" role
		CreatedAt:  time.Now(),
	}
	if err := s.repo.CreatePermission(tx, ownerPermission); err != nil {
		return nil, fmt.Errorf("failed to create owner permission: %w", err)
	}

	// 6. If all steps succeeded, commit the transaction.
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.repo.GetResourceByID(s.db, newResource.ID) // Re-fetch to populate associations
}

// getStoragePath derives the nested storage path from a hash.
func (s *service) getStoragePath(hash string) string {
	return filepath.Join(s.storagePath, hash[:2], hash[2:4], hash)
}

// DeleteFile handles the logic for deleting a file resource.
func (s *service) DeleteFile(resourceID uint, userID uint) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// 1. Fetch the resource to be deleted.
	resource, err := s.repo.GetResourceByID(tx, resourceID)
	if err != nil {
		return fmt.Errorf("resource not found: %w", err)
	}

	// 2. Authorization: Check if the user is the owner.
	if resource.OwnerID != userID {
		return errors.New("unauthorized: only the owner can delete this file")
	}
	if resource.PhysicalFileID == nil {
		return errors.New("invalid resource: resource has no physical file to delete")
	}

	// 3. Decrement the reference count of the physical file.
	physicalFileID := *resource.PhysicalFileID
	if err := s.repo.DecrementReferenceCount(tx, physicalFileID); err != nil {
		return fmt.Errorf("failed to update reference count: %w", err)
	}

	// 4. Delete the logical resource. DB cascade will handle permissions.
	if err := s.repo.DeleteResource(tx, resourceID); err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	// 5. Check if the physical file is now orphaned (garbage collection).
	pf, err := s.repo.GetPhysicalFileByID(tx, physicalFileID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check physical file status: %w", err)
	}

	if pf != nil && pf.ReferenceCount <= 0 {
		// Delete the physical file record from the database.
		if err := s.repo.DeletePhysicalFile(tx, physicalFileID); err != nil {
			return fmt.Errorf("failed to delete physical file record: %w", err)
		}
		// Delete the actual file from the disk.
		if err := os.Remove(pf.FilePath); err != nil {
			// Log this error but don't fail the transaction, as the DB state is consistent.
			fmt.Printf("warning: failed to delete physical file from disk: %s, error: %v\n", pf.FilePath, err)
		}
	}

	return tx.Commit().Error
}

// RenameFile handles the logic for renaming a file resource.
func (s *service) RenameFile(resourceID uint, userID uint, newName string) (*database.Resource, error) {
	// For simple updates, GORM can handle the transaction implicitly.
	// For consistency, we can also wrap it explicitly.
	resource, err := s.repo.GetResourceByID(s.db, resourceID)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	// Authorization check
	if resource.OwnerID != userID {
		return nil, errors.New("unauthorized: only the owner can rename this file")
	}

	resource.Name = newName
	if err := s.repo.UpdateResource(s.db, resource); err != nil {
		return nil, fmt.Errorf("failed to update resource name: %w", err)
	}

	return resource, nil
}

// MoveFile handles the logic for moving a file to a different folder.
func (s *service) MoveFile(resourceID uint, userID uint, newParentID *uint) (*database.Resource, error) {
	resource, err := s.repo.GetResourceByID(s.db, resourceID)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	// Authorization check
	if resource.OwnerID != userID {
		return nil, errors.New("unauthorized: only the owner can move this file")
	}

	// Optional: Check if the new parent folder exists and if the user has write access to it.
	// For simplicity, we are skipping that here, but in a real system, you would add that check.
	// if newParentID != nil { ... check permissions on newParentID ... }

	resource.ParentID = newParentID
	if err := s.repo.UpdateResource(s.db, resource); err != nil {
		return nil, fmt.Errorf("failed to update resource parent: %w", err)
	}

	return resource, nil
}
