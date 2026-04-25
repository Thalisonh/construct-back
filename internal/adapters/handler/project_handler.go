package handler

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"construct-backend/internal/core/services"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService      ports.ProjectService
	subscriptionService *services.SubscriptionService
}

func NewProjectHandler(projectService ports.ProjectService, subscriptionService *services.SubscriptionService) *ProjectHandler {
	return &ProjectHandler{
		projectService:      projectService,
		subscriptionService: subscriptionService,
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

type verifyPublicProjectPinRequest struct {
	Pin string `json:"pin" binding:"required,len=4,numeric"`
}

const publicProjectPinHeader = "X-Project-Pin"

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := c.GetString("user_id")
	companyID := c.GetString("company_id")
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verifica se o plano permite criar mais obras
	if err := h.subscriptionService.CheckProjectLimit(companyID); err != nil {
		if strings.HasPrefix(err.Error(), "limite_atingido") {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":            err.Error(),
				"upgrade_required": true,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(companyID, userID, req.Name, req.ClientID, req.Address, req.Summary, req.StartDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	companyID := c.GetString("company_id")
	clientID := c.Query("client_id")

	var projects []domain.Project
	var err error

	if clientID != "" {
		projects, err = h.projectService.ListProjectsByClient(clientID, companyID)
	} else {
		projects, err = h.projectService.ListProjects(companyID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	project, err := h.projectService.GetProject(id, companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) GetPublicProject(c *gin.Context) {
	id := c.Param("id")
	pin := c.GetHeader(publicProjectPinHeader)
	project, err := h.projectService.GetPublicProject(id, pin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) VerifyPublicProjectPin(c *gin.Context) {
	id := c.Param("id")
	var req verifyPublicProjectPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pin"})
		return
	}

	if err := h.projectService.VerifyPublicProjectPin(id, req.Pin); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid pin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.UpdateProject(id, req.Name, req.ClientID, req.Address, req.Summary, req.StartDate, req.IsPublic, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	if err := h.projectService.DeleteProject(id, companyID); err != nil {
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
	userID := c.GetString("user_id")
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.projectService.AddTask(projectID, req.Name, req.Status, req.DueDate, companyID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

type updateTaskRequest struct {
	Status string `json:"status"`
}

func (h *ProjectHandler) UpdateTask(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("taskId")
	var req updateTaskRequest
	// We use ShouldBindJSON but don't error out if it fails, to maintain compatibility with parameterless PUT (toggle)
	c.ShouldBindJSON(&req)

	task, err := h.projectService.UpdateTask(id, companyID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *ProjectHandler) DeleteTask(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("taskId")
	if err := h.projectService.DeleteTask(id, companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

type createSubtaskRequest struct {
	Name   string `json:"name" binding:"required"`
	Status string `json:"status" binding:"required"`
}

type diaryItemRequest struct {
	Type       string `json:"type" binding:"required"`
	Label      string `json:"label"`
	Content    string `json:"content" binding:"required"`
	Visibility string `json:"visibility" binding:"required"`
}

type diaryEntryRequest struct {
	EntryDate string             `json:"entry_date" binding:"required"`
	Title     string             `json:"title"`
	Items     []diaryItemRequest `json:"items" binding:"required,min=1"`
}

func validateDiaryItems(items []diaryItemRequest) error {
	for _, item := range items {
		if item.Type != "text" && item.Type != "field" {
			return fmt.Errorf("invalid diary item type")
		}
		if item.Visibility != "public" && item.Visibility != "internal" {
			return fmt.Errorf("invalid diary item visibility")
		}
	}
	return nil
}

func (h *ProjectHandler) AddSubtask(c *gin.Context) {
	userID := c.GetString("user_id")
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	taskID := c.Param("taskId")
	var req createSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subtask, err := h.projectService.AddSubtask(taskID, req.Name, req.Status, companyID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subtask)
}

func (h *ProjectHandler) CreateDiaryEntry(c *gin.Context) {
	userID := c.GetString("user_id")
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	var req diaryEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateDiaryItems(req.Items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]domain.DiaryItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.DiaryItem{
			Type:       item.Type,
			Label:      item.Label,
			Content:    item.Content,
			Visibility: item.Visibility,
		})
	}

	entry, err := h.projectService.CreateDiaryEntry(projectID, companyID, userID, req.EntryDate, req.Title, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (h *ProjectHandler) ListDiaryEntries(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	entries, err := h.projectService.ListDiaryEntries(projectID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *ProjectHandler) ListPublicDiaryEntries(c *gin.Context) {
	projectID := c.Param("id")
	pin := c.GetHeader(publicProjectPinHeader)
	entries, err := h.projectService.ListPublicDiaryEntries(projectID, pin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "diary not found"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *ProjectHandler) UpdateDiaryEntry(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	entryID := c.Param("entryId")
	var req diaryEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateDiaryItems(req.Items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]domain.DiaryItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.DiaryItem{
			Type:       item.Type,
			Label:      item.Label,
			Content:    item.Content,
			Visibility: item.Visibility,
		})
	}

	entry, err := h.projectService.UpdateDiaryEntry(entryID, projectID, companyID, req.EntryDate, req.Title, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func (h *ProjectHandler) DeleteDiaryEntry(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	entryID := c.Param("entryId")
	if err := h.projectService.DeleteDiaryEntry(entryID, projectID, companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProjectHandler) UpdateSubtask(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("subtaskId")

	subtask, err := h.projectService.UpdateSubtask(id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subtask)
}

func (h *ProjectHandler) DeleteSubtask(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("subtaskId")
	if err := h.projectService.DeleteSubtask(id, companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProjectHandler) GetTask(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("taskId")
	task, err := h.projectService.GetTask(id, companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *ProjectHandler) GetSubtask(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("subtaskId")
	subtask, err := h.projectService.GetSubtask(id, companyID)
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
