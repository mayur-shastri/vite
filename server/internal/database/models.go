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
type ResourceType string

const (
	File   ResourceType = "file"
	Folder ResourceType = "folder"
)

// Resource unifies files and folders into a single table.
// This is the core of the hierarchy.
type Resource struct {
	gorm.Model
	OwnerID  uint
	User     User         `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;"`
	ParentID *uint        // Pointer for nullable root resources
	Parent   *Resource    `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;"`
	Name     string       `gorm:"size:255;not null"`
	Type     ResourceType `gorm:"not null"`

	// File-specific fields (will be NULL for folders)
	PhysicalFileID *uint
	PhysicalFile   *PhysicalFile `gorm:"foreignKey:PhysicalFileID;constraint:OnDelete:RESTRICT;"`
}

// RoleType defines the permission levels.
type RoleType string

const (
	Viewer RoleType = "viewer"
	Editor RoleType = "editor"
)

// Permission stores ONLY direct permissions granted to a user for a resource.
// This is our Access Control List (ACL) table and replaces the old `FileShare`.
type Permission struct {
	ResourceID uint     `gorm:"primaryKey"`
	UserID     uint     `gorm:"primaryKey"`
	Resource   Resource `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE;"`
	User       User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Role       RoleType `gorm:"not null"`
	CreatedAt  time.Time
}

// ResourceAncestor is the Closure Table.
// It stores all parent-child relationships to any depth, allowing for
// extremely fast hierarchical lookups without recursive queries.
type ResourceAncestor struct {
	AncestorID   uint     `gorm:"primaryKey"`
	Ancestor     Resource `gorm:"foreignKey:AncestorID;constraint:OnDelete:CASCADE;"`
	DescendantID uint     `gorm:"primaryKey"`
	Descendant   Resource `gorm:"foreignKey:DescendantID;constraint:OnDelete:CASCADE;"`
	Depth        int      `gorm:"not null"`
}
