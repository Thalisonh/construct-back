package main

import (
	"construct-backend/internal/adapters/handler"
	"construct-backend/internal/adapters/payment"
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
		userRepo      ports.UserRepository
		projectRepo   ports.ProjectRepository
		linkRepo      ports.LinkRepository
		companyRepo   ports.CompanyRepository
		subRepo       ports.SubscriptionRepository
		dashboardRepo ports.DashboardRepository
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
	db.AutoMigrate(&domain.User{}, &domain.Project{}, &domain.Link{}, &domain.Client{}, &domain.Comment{}, &domain.Task{}, &domain.Subtask{}, &domain.LinkClick{}, &domain.Company{})

	pgRepo := repository.NewPostgresRepository(db)
	userRepo = pgRepo
	projectRepo = pgRepo
	linkRepo = pgRepo
	companyRepo = pgRepo
	subRepo = pgRepo
	dashboardRepo = pgRepo
	clientRepo := pgRepo
	log.Println("Connected to PostgreSQL")

	// Services
	authService := services.NewAuthService(userRepo, companyRepo, jwtSecret)
	projectService := services.NewProjectService(projectRepo)
	linkService := services.NewLinkService(linkRepo)
	userService := services.NewUserService(userRepo, linkRepo)
	clientService := services.NewClientService(clientRepo)
	companyService := services.NewCompanyService(companyRepo)
	dashboardService := services.NewDashboardService(dashboardRepo)

	// Payment Gateway (Mercado Pago Adapter — troque aqui para mudar de provider)
	mpToken := os.Getenv("MP_ACCESS_TOKEN")
	mpSuccessURL := os.Getenv("MP_SUCCESS_URL")
	mpFailureURL := os.Getenv("MP_FAILURE_URL")
	if mpSuccessURL == "" {
		mpSuccessURL = "http://localhost:5173/dashboard"
	}
	if mpFailureURL == "" {
		mpFailureURL = "http://localhost:5173/checkout"
	}
	gateway := payment.NewMercadoPagoAdapter(mpToken, mpSuccessURL, mpFailureURL)
	subscriptionService := services.NewSubscriptionService(gateway, companyRepo, subRepo, mpSuccessURL, mpFailureURL)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	projectHandler := handler.NewProjectHandler(projectService, subscriptionService)
	linkHandler := handler.NewLinkHandler(linkService)
	userHandler := handler.NewUserHandler(userService)
	clientHandler := handler.NewClientHandler(clientService)
	companyHandler := handler.NewCompanyHandler(companyService, userService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	// Router
	r := handler.SetupRouter(authHandler, userHandler, dashboardHandler, projectHandler, linkHandler, clientHandler, companyHandler, subscriptionHandler, jwtSecret)

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
