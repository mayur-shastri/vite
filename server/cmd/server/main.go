// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	// Your application packages
	"github.com/bhavyajaix/BalkanID-filevault/graph"
	"github.com/bhavyajaix/BalkanID-filevault/graph/generated"
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/internal/file"
	"github.com/bhavyajaix/BalkanID-filevault/internal/folders"
	"github.com/bhavyajaix/BalkanID-filevault/internal/middleware"
	"github.com/bhavyajaix/BalkanID-filevault/internal/permission"
	"github.com/bhavyajaix/BalkanID-filevault/internal/search"
	"github.com/bhavyajaix/BalkanID-filevault/internal/share"
	"github.com/bhavyajaix/BalkanID-filevault/internal/tag"
	"github.com/bhavyajaix/BalkanID-filevault/internal/user"
)

const defaultPort = "8080"

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, loading from environment")
	}

	// Set up port
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// --- Dependency Injection Setup ---

	// 1. Initialize Database connection
	db := database.Connect()

	// 2. Run Migrations
	// database.Migrate(db)

	storagePath := os.Getenv("FILEVAULT_STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./testuploads"
		log.Printf("FILEVAULT_STORAGE_PATH not set, using default: %s", storagePath)
	}

	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		log.Fatalf("could not create storage directory: %v", err)
	}

	// 3. Initialize Layers (Repository -> Service)
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	fileRepo := file.NewRepository(db)
	permissionRepo := permission.NewRepository(db)
	foldersRepo := folders.NewRepository(db)
	foldersService := folders.NewService(foldersRepo)
	permissionService := permission.NewService(permissionRepo, foldersRepo, userRepo)
	fileService := file.NewService(fileRepo, db, storagePath, permissionRepo)
	shareRepo := share.NewRepository(db)
	shareService := share.NewService(shareRepo, foldersRepo, fileRepo, db)
	tagRepo := tag.NewTagRepository(db)
	tagService := tag.NewTagService(tagRepo)
	searchRepo := search.NewSearchRepository(db)
	searchService := search.NewSearchService(searchRepo)
	// 4. Inject Dependencies into the Resolver
	// The resolver now has access to the user service.
	resolver := &graph.Resolver{
		DB:                db, // Keep DB for other features you'll build
		UserService:       userService,
		FileService:       fileService,
		FolderService:     foldersService,
		PermissionService: permissionService,
		ShareService:      shareService,
		TagService:        tagService,
		SearchService:     searchService,
	}

	// --- Server Setup ---

	router := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	router.Use(corsMiddleware.Handler)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	authedSrv := middleware.AuthMiddleware(srv)

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", authedSrv)
	router.Get("/download/{resourceID}", file.DownloadFileHandler(db, permissionRepo))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
