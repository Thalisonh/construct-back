package handler

import (
	"construct-backend/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LinkHandler struct {
	linkService ports.LinkService
}

func NewLinkHandler(linkService ports.LinkService) *LinkHandler {
	return &LinkHandler{
		linkService: linkService,
	}
}

type createLinkRequest struct {
	ProjectID   string `json:"project_id" binding:"required"`
	URL         string `json:"url" binding:"required,url"`
	Description string `json:"description"`
}

func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.linkService.CreateLink(req.ProjectID, req.URL, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, link)
}

func (h *LinkHandler) ListLinks(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	links, err := h.linkService.ListLinks(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, links)
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	id := c.Param("id")
	if err := h.linkService.DeleteLink(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
