package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func SetupRouter(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	projectHandler *ProjectHandler,
	linkHandler *LinkHandler,
	clientHandler *ClientHandler,
	jwtSecret string,
) *gin.Engine {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)
	r.POST("/auth/google", authHandler.GoogleLogin)
	r.POST("/signup/google", authHandler.GoogleLogin)
	r.POST("/auth/verify", authHandler.TokenVerify)
	r.GET("/public/profile/:username", userHandler.GetPublicProfile)
	r.GET("/public/projects/:id", projectHandler.GetPublicProject)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	api := r.Group("/")
	api.Use(authMiddleware(jwtSecret))
	{
		api.GET("/username", userHandler.VerifyUserName)
		api.POST("/username", userHandler.UpdateUsername)

		api.GET("/projects", projectHandler.ListProjects)
		api.GET("/projects/:id", projectHandler.GetProject)
		api.POST("/projects", projectHandler.CreateProject)
		api.PUT("/projects/:id", projectHandler.UpdateProject)
		api.DELETE("/projects/:id", projectHandler.DeleteProject)

		api.POST("/projects/:id/tasks", projectHandler.AddTask)
		api.GET("/projects/:id/tasks", projectHandler.ListTasks)
		api.GET("/tasks/:taskId", projectHandler.GetTask)
		api.PUT("/tasks/:taskId", projectHandler.UpdateTask)
		api.DELETE("/tasks/:taskId", projectHandler.DeleteTask)

		api.POST("/tasks/:taskId/subtasks", projectHandler.AddSubtask)
		api.GET("/subtasks/:subtaskId", projectHandler.GetSubtask)
		api.PUT("/subtasks/:subtaskId", projectHandler.UpdateSubtask)
		api.DELETE("/subtasks/:subtaskId", projectHandler.DeleteSubtask)

		api.GET("/links", linkHandler.ListLinks)
		api.POST("/links", linkHandler.CreateLink)
		api.DELETE("/links/:id", linkHandler.DeleteLink)
		api.PUT("/links/:id", linkHandler.UpdateLink)

		api.GET("/user/username", userHandler.GetUsername)

		api.POST("/clients", clientHandler.CreateClient)
		api.GET("/clients", clientHandler.ListClients)
		api.GET("/clients/:id", clientHandler.GetClient)
		api.PUT("/clients/:id", clientHandler.UpdateClient)
		api.DELETE("/clients/:id", clientHandler.DeleteClient)
		api.POST("/clients/:id/comments", clientHandler.AddComment)
	}

	return r
}

func authMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		c.Next()
	}
}
