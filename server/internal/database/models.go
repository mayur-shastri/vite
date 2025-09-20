package database

import (
	"time"

	"gorm.io/gorm"
)

// User defines the user model
type User struct {
	gorm.Model            // Includes ID, CreatedAt, UpdatedAt, DeletedAt
	Username       string `gorm:"size:255;uniqueIndex;not null"`
	Email          string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash   string `gorm:"not null"`
	StorageQuotaMB int    `gorm:"default:10;not null"`
	Role           string `gorm:"size:50;default:'user';not null"`
}

// PhysicalFile represents the actual file on disk for deduplication
type PhysicalFile struct {
	gorm.Model
	FileHash       string `gorm:"size:64;uniqueIndex;not null"`
	FilePath       string `gorm:"type:text;unique;not null"`
	SizeBytes      int64  `gorm:"not null"`
	MimeType       string `gorm:"size:255;not null"`
	ReferenceCount int    `gorm:"default:1;not null"`
}

// Folder represents a directory that can contain files and other folders
type Folder struct {
	gorm.Model
	OwnerID        uint
	User           User    `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;"`
	ParentFolderID *uint   // Pointer allows for NULL (root folders)
	ParentFolder   *Folder `gorm:"foreignKey:ParentFolderID;constraint:OnDelete:CASCADE;"`
	Name           string  `gorm:"size:255;not null"`
	// Has Many relationships
	SubFolders []Folder `gorm:"foreignKey:ParentFolderID"`
	Files      []UserFile
}

// UserFile is the user-facing metadata for a file
type UserFile struct {
	gorm.Model
	OwnerID          uint
	User             User `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;"`
	PhysicalFileID   uint
	PhysicalFile     PhysicalFile `gorm:"constraint:OnDelete:RESTRICT;"` // Don't delete physical file if user file exists
	FolderID         *uint        // Pointer allows for NULL (root files)
	Folder           Folder       `gorm:"constraint:OnDelete:SET NULL;"` // If folder is deleted, file moves to root
	Filename         string       `gorm:"size:255;not null"`
	IsPublic         bool         `gorm:"default:false"`
	PublicShareToken string       `gorm:"size:32;uniqueIndex"`
	DownloadCount    int          `gorm:"default:0"`
	UploadedAt       time.Time    `gorm:"default:CURRENT_TIMESTAMP"`
}

// FileShare defines a many-to-many relationship for sharing files
type FileShare struct {
	// Composite Primary Key defined via struct tag
	FileID           uint `gorm:"primaryKey"`
	SharedWithUserID uint `gorm:"primaryKey"`
	// Relationships
	File           UserFile `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE;"`
	SharedWithUser User     `gorm:"foreignKey:SharedWithUserID;constraint:OnDelete:CASCADE;"`
}
