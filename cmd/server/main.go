package main

import (
	"construct-backend/internal/adapters/handler"
	"construct-backend/internal/adapters/repository"
	"construct-backend/internal/core/services"
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func main() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://localhost:27017").SetServerAPIOptions(serverAPI)

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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret"
	}

	db := client.Database("construct")

	// Repositories
	mongoRepo := repository.NewMongoDBRepository(db)

	// Services
	authService := services.NewAuthService(mongoRepo, jwtSecret)
	projectService := services.NewProjectService(mongoRepo)
	linkService := services.NewLinkService(mongoRepo)
	userService := services.NewUserService(mongoRepo)

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
