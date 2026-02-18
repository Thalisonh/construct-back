package handler

import (
	"construct-backend/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	companyService ports.CompanyService
	userService    ports.UserService
}

func NewCompanyHandler(companyService ports.CompanyService, userService ports.UserService) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
		userService:    userService,
	}
}

func (h *CompanyHandler) GetCompany(c *gin.Context) {
	companyID := c.GetString("company_id")
	role := c.GetString("role")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
		return
	}

	company, err := h.companyService.GetCompany(companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
		return
	}

	c.JSON(http.StatusOK, company)
}

type updateCompanyRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Phone   string `json:"phone" binding:"required"`
	Address string `json:"address" binding:"required"`
}

func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	companyID := c.GetString("company_id")
	role := c.GetString("role")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
		return
	}

	// Only admin should update? Role is in context too from middleware?
	// Not yet, I only set user_id and company_id in middleware.
	// Let's assume for now, and maybe add role check later.

	var req updateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company, err := h.companyService.UpdateCompany(companyID, req.Name, req.Email, req.Phone, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, company)
}

func (h *CompanyHandler) ListMembers(c *gin.Context) {
	companyID := c.GetString("company_id")
	role := c.GetString("role")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
		return
	}

	members, err := h.userService.GetCompanyMembers(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}

type addMemberRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

func (h *CompanyHandler) AddMember(c *gin.Context) {
	companyID := c.GetString("company_id")
	role := c.GetString("role")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
		return
	}

	var req addMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.userService.AddCompanyMember(companyID, req.Email, req.Name, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}
