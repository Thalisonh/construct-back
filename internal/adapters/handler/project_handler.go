package handler

import (
	"construct-backend/internal/core/ports"
	"fmt"
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
	Name      string `json:"name" binding:"required"`
	ClientID  string `json:"client_id" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	Address   string `json:"address" binding:"required"`
	Summary   string `json:"summary" binding:"required"`
	IsPublic  bool   `json:"is_public" binding:"required"`
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := c.GetString("user_id")
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(userID, req.Name, req.ClientID, req.Address, req.Summary, req.StartDate)
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

func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) GetPublicProject(c *gin.Context) {
	id := c.Param("id")
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if !project.IsPublic {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.UpdateProject(id, req.Name, req.ClientID, req.Address, req.Summary, req.StartDate, req.IsPublic)
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

type createTaskRequest struct {
	Name    string `json:"name" binding:"required"`
	Status  string `json:"status" binding:"required"`
	DueDate string `json:"due_date" binding:"required"`
}

func (h *ProjectHandler) AddTask(c *gin.Context) {
	projectID := c.Param("id")
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.projectService.AddTask(projectID, req.Name, req.Status, req.DueDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *ProjectHandler) UpdateTask(c *gin.Context) {
	id := c.Param("taskId")

	task, err := h.projectService.UpdateTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *ProjectHandler) DeleteTask(c *gin.Context) {
	id := c.Param("taskId")
	if err := h.projectService.DeleteTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

type createSubtaskRequest struct {
	Name   string `json:"name" binding:"required"`
	Status string `json:"status" binding:"required"`
}

func (h *ProjectHandler) AddSubtask(c *gin.Context) {
	taskID := c.Param("taskId")
	var req createSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subtask, err := h.projectService.AddSubtask(taskID, req.Name, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subtask)
}

func (h *ProjectHandler) UpdateSubtask(c *gin.Context) {
	id := c.Param("subtaskId")

	subtask, err := h.projectService.UpdateSubtask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subtask)
}

func (h *ProjectHandler) DeleteSubtask(c *gin.Context) {
	id := c.Param("subtaskId")
	if err := h.projectService.DeleteSubtask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProjectHandler) GetTask(c *gin.Context) {
	id := c.Param("taskId")
	task, err := h.projectService.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *ProjectHandler) GetSubtask(c *gin.Context) {
	id := c.Param("subtaskId")
	subtask, err := h.projectService.GetSubtask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subtask not found"})
		return
	}

	c.JSON(http.StatusOK, subtask)
}

func (h *ProjectHandler) ListTasks(c *gin.Context) {
	projectID := c.Param("id")
	tasks, err := h.projectService.ListTasks(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
