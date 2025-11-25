package handler

import (
	"construct-backend/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService ports.ProjectService
}

func NewProjectHandler(projectService ports.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

type createProjectRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := c.GetString("user_id")
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(userID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID := c.GetString("user_id")
	projects, err := h.projectService.ListProjects(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.UpdateProject(id, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if err := h.projectService.DeleteProject(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
