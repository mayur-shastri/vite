package graph

import (
	"github.com/bhavyajaix/BalkanID-filevault/internal/file"
	"github.com/bhavyajaix/BalkanID-filevault/internal/folders"
	"github.com/bhavyajaix/BalkanID-filevault/internal/user"

	"gorm.io/gorm"
)

type Resolver struct {
	DB            *gorm.DB
	UserService   user.Service // Add the user service
	FileService   file.Service
	FolderService folders.Service
}
