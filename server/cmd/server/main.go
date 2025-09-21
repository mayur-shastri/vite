// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	// Your application packages
	"github.com/bhavyajaix/BalkanID-filevault/graph"
	"github.com/bhavyajaix/BalkanID-filevault/graph/generated"
	"github.com/bhavyajaix/BalkanID-filevault/internal/database"
	"github.com/bhavyajaix/BalkanID-filevault/internal/middleware"
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

	// 3. Initialize Layers (Repository -> Service)
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	// 4. Inject Dependencies into the Resolver
	// The resolver now has access to the user service.
	resolver := &graph.Resolver{
		DB:          db, // Keep DB for other features you'll build
		UserService: userService,
	}

	// --- Server Setup ---

	router := chi.NewRouter()
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	authedSrv := middleware.AuthMiddleware(srv)

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", authedSrv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
