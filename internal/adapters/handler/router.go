package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func SetupRouter(authHandler *AuthHandler, projectHandler *ProjectHandler, linkHandler *LinkHandler, jwtSecret string) *gin.Engine {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)
	r.POST("/auth/google", authHandler.GoogleLogin)
	r.POST("/signup/google", authHandler.GoogleLogin)

	api := r.Group("/")
	api.Use(authMiddleware(jwtSecret))
	{
		api.GET("/projects", projectHandler.ListProjects)
		api.POST("/projects", projectHandler.CreateProject)
		api.PUT("/projects/:id", projectHandler.UpdateProject)
		api.DELETE("/projects/:id", projectHandler.DeleteProject)

		api.GET("/links", linkHandler.ListLinks)
		api.POST("/links", linkHandler.CreateLink)
		api.DELETE("/links/:id", linkHandler.DeleteLink)
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
