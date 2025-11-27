package main

import (
	"construct-backend/internal/adapters/handler"
	"construct-backend/internal/adapters/repository"
	"construct-backend/internal/core/services"
	"context"
	"fmt"
	"log"
	"os"

	"construct-backend/internal/core/ports"

	"cloud.google.com/go/datastore"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
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

	dbType := os.Getenv("DB_TYPE")
	if dbType == "datastore" {
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		if projectID == "" {
			log.Fatal("GOOGLE_CLOUD_PROJECT environment variable is required for Datastore")
		}
		ctx := context.Background()
		client, err := datastore.NewClient(ctx, projectID)
		if err != nil {
			log.Fatal("Failed to create Datastore client:", err)
		}
		defer client.Close()

		dsRepo := repository.NewDatastoreRepository(client)
		userRepo = dsRepo
		projectRepo = dsRepo
		linkRepo = dsRepo
		log.Println("Connected to Google Datastore")
	} else {
		mongoURI := os.Getenv("MONGO_URI")
		if mongoURI == "" {
			mongoURI = "mongodb://localhost:27017"
		}

		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

		client, err := mongo.Connect(opts)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()

		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
			panic(err)
		}
		fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

		db := client.Database("construct")
		mongoRepo := repository.NewMongoDBRepository(db)
		userRepo = mongoRepo
		projectRepo = mongoRepo
		linkRepo = mongoRepo
	}

	// Services
	authService := services.NewAuthService(userRepo, jwtSecret)
	projectService := services.NewProjectService(projectRepo)
	linkService := services.NewLinkService(linkRepo)
	userService := services.NewUserService(userRepo, linkRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	linkHandler := handler.NewLinkHandler(linkService)
	userHandler := handler.NewUserHandler(userService)

	// Router
	r := handler.SetupRouter(authHandler, userHandler, projectHandler, linkHandler, jwtSecret)

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
