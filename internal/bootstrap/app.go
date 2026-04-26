package bootstrap

import (
	"construct-backend/internal/adapters/handler"
	"construct-backend/internal/adapters/payment"
	"construct-backend/internal/adapters/repository"
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"construct-backend/internal/core/services"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewRouter() (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	var (
		userRepo      ports.UserRepository
		projectRepo   ports.ProjectRepository
		linkRepo      ports.LinkRepository
		companyRepo   ports.CompanyRepository
		subRepo       ports.SubscriptionRepository
		dashboardRepo ports.DashboardRepository
		clientRepo    ports.ClientRepository
	)

	repositoryDriver := os.Getenv("REPOSITORY_DRIVER")
	if repositoryDriver == "" {
		repositoryDriver = "dynamodb"
	}

	switch repositoryDriver {
	case "postgres":
		pgRepo, err := newPostgresRepository()
		if err != nil {
			return nil, err
		}

		userRepo = pgRepo
		projectRepo = pgRepo
		linkRepo = pgRepo
		companyRepo = pgRepo
		subRepo = pgRepo
		dashboardRepo = pgRepo
		clientRepo = pgRepo
	case "dynamodb":
		dynamoRepo, err := repository.NewDynamoRepositoryFromEnv(context.Background())
		if err != nil {
			return nil, err
		}

		userRepo = dynamoRepo
		projectRepo = dynamoRepo
		linkRepo = dynamoRepo
		companyRepo = dynamoRepo
		subRepo = dynamoRepo
		dashboardRepo = dynamoRepo
		clientRepo = dynamoRepo
	default:
		return nil, fmt.Errorf("unsupported repository driver %q", repositoryDriver)
	}

	authService := services.NewAuthService(userRepo, companyRepo, jwtSecret)
	projectService := services.NewProjectService(projectRepo)
	linkService := services.NewLinkService(linkRepo)
	userService := services.NewUserService(userRepo, linkRepo)
	clientService := services.NewClientService(clientRepo)
	companyService := services.NewCompanyService(companyRepo, linkRepo)
	dashboardService := services.NewDashboardService(dashboardRepo)

	mpToken := os.Getenv("MP_ACCESS_TOKEN")
	mpSuccessURL := os.Getenv("MP_SUCCESS_URL")
	mpFailureURL := os.Getenv("MP_FAILURE_URL")
	if mpSuccessURL == "" {
		return nil, fmt.Errorf("MP_SUCCESS_URL is required")
	}
	if mpFailureURL == "" {
		return nil, fmt.Errorf("MP_FAILURE_URL is required")
	}

	gateway := payment.NewMercadoPagoAdapter(mpToken, mpSuccessURL, mpFailureURL)
	subscriptionService := services.NewSubscriptionService(gateway, companyRepo, subRepo, mpSuccessURL, mpFailureURL)

	authHandler := handler.NewAuthHandler(authService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	projectHandler := handler.NewProjectHandler(projectService, subscriptionService)
	linkHandler := handler.NewLinkHandler(linkService)
	userHandler := handler.NewUserHandler(userService)
	clientHandler := handler.NewClientHandler(clientService)
	companyHandler := handler.NewCompanyHandler(companyService, userService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	return handler.SetupRouter(authHandler, userHandler, dashboardHandler, projectHandler, linkHandler, clientHandler, companyHandler, subscriptionHandler, jwtSecret), nil
}

func newPostgresRepository() (*repository.PostgresRepository, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("POSTGRES_DSN is required when REPOSITORY_DRIVER=postgres")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to Postgres: %w", err)
	}

	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := db.AutoMigrate(&domain.User{}, &domain.Project{}, &domain.Link{}, &domain.Client{}, &domain.Comment{}, &domain.Task{}, &domain.Subtask{}, &domain.LinkClick{}, &domain.Company{}, &domain.DiaryEntry{}, &domain.DiaryItem{}); err != nil {
			return nil, fmt.Errorf("auto migrate Postgres: %w", err)
		}
		log.Println("Postgres auto migration completed")
	}

	log.Println("Connected to PostgreSQL")
	return repository.NewPostgresRepository(db), nil
}
