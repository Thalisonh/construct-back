package main

import (
	"construct-backend/internal/adapters/handler"
	"construct-backend/internal/adapters/repository"
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"construct-backend/internal/core/services"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret"
	}

	var (
		userRepo    ports.UserRepository
		projectRepo ports.ProjectRepository
		linkRepo    ports.LinkRepository
	)

	dsn := os.Getenv("POSTGRES_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		log.Fatal("Failed to connect to Postgres:", err)

		return
	}

	// Auto Migrate
	// Auto Migrate
	db.AutoMigrate(&domain.User{}, &domain.Project{}, &domain.Link{}, &domain.Client{}, &domain.Comment{}, &domain.Task{}, &domain.Subtask{}, &domain.LinkClick{})

	pgRepo := repository.NewPostgresRepository(db)
	userRepo = pgRepo
	projectRepo = pgRepo
	linkRepo = pgRepo
	clientRepo := pgRepo
	log.Println("Connected to PostgreSQL")

	// Services
	authService := services.NewAuthService(userRepo, jwtSecret)
	projectService := services.NewProjectService(projectRepo)
	linkService := services.NewLinkService(linkRepo)
	userService := services.NewUserService(userRepo, linkRepo)
	clientService := services.NewClientService(clientRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	linkHandler := handler.NewLinkHandler(linkService)
	userHandler := handler.NewUserHandler(userService)
	clientHandler := handler.NewClientHandler(clientService)

	// Router
	r := handler.SetupRouter(authHandler, userHandler, projectHandler, linkHandler, clientHandler, jwtSecret)

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
