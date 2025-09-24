package file

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/internal/middleware"
	"github.com/bhavyajaix/BalkanID-filevault/internal/permission"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// func DownloadFileHandler(db *gorm.DB, permissionRepo permission.Repository) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// 1. Get resourceID from URL
// 		resourceIDStr := chi.URLParam(r, "resourceID")
// 		resourceID, err := strconv.ParseUint(resourceIDStr, 10, 64)
// 		if err != nil {
// 			http.Error(w, "invalid resource ID", http.StatusBadRequest)
// 			return
// 		}

// 		// 3. Fetch resource with PhysicalFile
// 		var resource database.Resource
// 		if err := db.WithContext(r.Context()).
// 			Preload("PhysicalFile").
// 			Where("id = ?", resourceID).
// 			First(&resource).Error; err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				http.Error(w, "resource not found", http.StatusNotFound)
// 				return
// 			}
// 			http.Error(w, "internal server error", http.StatusInternalServerError)
// 			return
// 		}

// 		// 5. Make sure physical file exists
// 		if resource.PhysicalFile == nil {
// 			http.Error(w, "file missing", http.StatusInternalServerError)
// 			return
// 		}

// 		filePath := resource.PhysicalFile.FilePath
// 		mimeType := resource.PhysicalFile.MimeType
// 		filename := resource.Name

// 		// Check if the resource itself is public or if any of its ancestors are public.
// 		isPubliclyAccessible := resource.IsPublic
// 		if !isPubliclyAccessible {
// 			var publicAncestorCount int64
// 			subQuery := db.Model(&database.ResourceAncestor{}).Select("ancestor_id").Where("descendant_id = ?", resource.ID)
// 			if err := db.Model(&database.Resource{}).
// 				Where("id IN (?) AND is_public = ?", subQuery, true).
// 				Count(&publicAncestorCount).Error; err != nil {
// 				// This is an internal query error, not an access error
// 				http.Error(w, "internal server error while checking public access", http.StatusInternalServerError)
// 				return
// 			}
// 			if publicAncestorCount > 0 {
// 				isPubliclyAccessible = true
// 			}
// 		}

// 		if isPubliclyAccessible {
// 			// 6. Stream the file like res.sendFile
// 			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
// 			w.Header().Set("Content-Type", mimeType)
// 			http.ServeFile(w, r, filePath)
// 			return
// 		}

// 		// 2. Get user ID from context (populated by AuthMiddleware)
// 		userID, ok := r.Context().Value(middleware.UserContextKey).(uint)
// 		if !ok {
// 			http.Error(w, "unauthorized", http.StatusUnauthorized)
// 			return
// 		}

// 		// 4. Check user permission
// 		_, err = permissionRepo.FindPermission(resource.ID, userID)
// 		if err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				http.Error(w, "access denied", http.StatusForbidden)
// 				return
// 			}
// 			http.Error(w, "internal server error", http.StatusInternalServerError)
// 			return
// 		}

// 		// If we reach here, the user has permission (either direct or via a public ancestor)

// 		// 6. Stream the file like res.sendFile
// 		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
// 		w.Header().Set("Content-Type", mimeType)
// 		http.ServeFile(w, r, filePath)
// 	}
// }

func DownloadFileHandler(db *gorm.DB, permissionRepo permission.Repository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Get resourceID from URL
        resourceIDStr := chi.URLParam(r, "resourceID")
        resourceID, err := strconv.ParseUint(resourceIDStr, 10, 64)
        if err != nil {
            http.Error(w, "invalid resource ID", http.StatusBadRequest)
            return
        }

        // 3. Fetch resource with PhysicalFile
        var resource database.Resource
        if err := db.WithContext(r.Context()).
            Preload("PhysicalFile").
            Where("id = ?", resourceID).
            First(&resource).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                http.Error(w, "resource not found", http.StatusNotFound)
                return
            }
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }

        // 5. Make sure physical file exists
        if resource.PhysicalFile == nil {
            http.Error(w, "file missing", http.StatusInternalServerError)
            return
        }

        filePath := resource.PhysicalFile.FilePath
        mimeType := resource.PhysicalFile.MimeType
        filename := resource.Name

        // Check if the resource itself is public or if any of its ancestors are public.
        isPubliclyAccessible := resource.IsPublic
        if !isPubliclyAccessible {
            var publicAncestorCount int64
            subQuery := db.Model(&database.ResourceAncestor{}).
                Select("ancestor_id").
                Where("descendant_id = ?", resource.ID)

            if err := db.Model(&database.Resource{}).
                Where("id IN (?)", subQuery).
                Where("is_public = ?", true).
                Count(&publicAncestorCount).Error; err != nil {
                // This is an internal query error, not an access error
                http.Error(w, "internal server error while checking public access", http.StatusInternalServerError)
                return
            }
            if publicAncestorCount > 0 {
                isPubliclyAccessible = true
            }
        }

		fmt.Println("**********", isPubliclyAccessible);

        if isPubliclyAccessible {
            // 6. Stream the file
            w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
            w.Header().Set("Content-Type", mimeType)
            http.ServeFile(w, r, filePath)
            return
        }

        // 2. Get user ID from context (populated by AuthMiddleware)
        userIDVal := r.Context().Value(middleware.UserContextKey)
        if userIDVal == nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        var userID uint
        switch v := userIDVal.(type) {
        case uint:
            userID = v
        case uint64:
            userID = uint(v)
        case int:
            userID = uint(v)
        default:
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        // 4. Check user permission
        _, err = permissionRepo.FindPermission(resource.ID, userID)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                http.Error(w, "access denied", http.StatusForbidden)
                return
            }
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }

        // If we reach here, the user has permission
        w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
        w.Header().Set("Content-Type", mimeType)
        http.ServeFile(w, r, filePath)
    }
}
