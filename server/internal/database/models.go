package database

import (
	"time"

	"gorm.io/gorm"
)

// RoleAdmin defines the string identifier for an administrator user.
const RoleAdmin = "admin"

// User defines the user model
type User struct {
	gorm.Model
	Username                 string `gorm:"size:255;uniqueIndex;not null"`
	Email                    string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash             string `gorm:"not null"`
	StorageQuotaMB           int    `gorm:"default:10;not null"`
	Role                     string `gorm:"size:50;default:'user';not null"`
	StorageUsed              int    `gorm:"not null"`
	DeduplicationStorageUsed int    `gorm:"not null"`
	// "Has Many" relationships for easier preloading
	Resources   []Resource   `gorm:"foreignKey:OwnerID"`
	Permissions []Permission `gorm:"foreignKey:UserID"`
}

// PhysicalFile represents the actual file on disk for deduplication
type PhysicalFile struct {
	gorm.Model
	FileHash       string `gorm:"size:64;uniqueIndex;not null"`
	FilePath       string `gorm:"type:text;unique;not null"`
	SizeBytes      int64  `gorm:"not null;index"`
	MimeType       string `gorm:"size:255;not null;index"`
	ReferenceCount int    `gorm:"default:1;not null"`
}

// ResourceType defines if a resource is a file or folder.
type ResourceType string

const (
	File   ResourceType = "file"
	Folder ResourceType = "folder"
)

type Tag struct {
	gorm.Model
	Name string `gorm:"size:100;uniqueIndex;not null"` // Tag names should be unique
}

// Resource unifies files and folders into a single table.
type Resource struct {
	gorm.Model
	OwnerID        uint          `gorm:"index"` // Indexed for performance
	User           User          `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;"`
	ParentID       *uint         `gorm:"index"` // Indexed for performance
	Parent         *Resource     `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;"`
	Name           string        `gorm:"size:255;not null;index"`
	IsPublic       bool          `gorm:"default:false;not null;index"`
	ShareToken     *string       `gorm:"size:255;uniqueIndex"`
	Type           ResourceType  `gorm:"type:varchar(50);not null;index"` // Explicit type
	PhysicalFileID *uint         `gorm:"index"`
	Tags           []*Tag        `gorm:"many2many:resource_tags;"`
	PhysicalFile   *PhysicalFile `gorm:"foreignKey:PhysicalFileID;constraint:OnDelete:RESTRICT;"`
	// "Has Many" relationships for easier preloading
	Permissions []Permission `gorm:"foreignKey:ResourceID"`
	Children    []Resource   `gorm:"foreignKey:ParentID"`
}

// RoleType defines the permission levels.
type RoleType string

const (
	Viewer RoleType = "VIEWER"
	Editor RoleType = "EDITOR"
)

// Permission is the Access Control List (ACL) table.
type Permission struct {
	ResourceID uint     `gorm:"primaryKey"`
	UserID     uint     `gorm:"primaryKey"`
	Resource   Resource `gorm:"foreignKey:ResourceID;constraint:OnDelete:CASCADE;"`
	User       User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Role       RoleType `gorm:"type:varchar(50);not null"` // Explicit type
	CreatedAt  time.Time
}

// ResourceAncestor is the Closure Table for fast hierarchy lookups.
type ResourceAncestor struct {
	AncestorID   uint     `gorm:"primaryKey"`
	Ancestor     Resource `gorm:"foreignKey:AncestorID;constraint:OnDelete:CASCADE;"`
	DescendantID uint     `gorm:"primaryKey"`
	Descendant   Resource `gorm:"foreignKey:DescendantID;constraint:OnDelete:CASCADE;"`
	Depth        int      `gorm:"not null"`
}
