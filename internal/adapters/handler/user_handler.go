package handler

import (
	"construct-backend/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService ports.UserService
}

func NewUserHandler(userService ports.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
func (h *UserHandler) VerifyUserName(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	if err := h.userService.VerifyUserName(username); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

type usernameRequest struct {
	Username string `json:"username" binding:"required"`
}

func (h *UserHandler) UpdateUsername(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == "" || userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req usernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.UpdateUsername(userID.(string), req.Username); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) GetUsername(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if userID == "" || userID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	username, err := h.userService.GetUsername(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"username": username})
}

func (h *UserHandler) GetPublicProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	user, err := h.userService.GetPublicProfile(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	_, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	c.Status(http.StatusOK)
}
