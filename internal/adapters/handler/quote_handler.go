package handler

import (
	"construct-backend/internal/core/domain"
	"construct-backend/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QuoteHandler struct {
	quoteService ports.QuoteService
}

func NewQuoteHandler(quoteService ports.QuoteService) *QuoteHandler {
	return &QuoteHandler{quoteService: quoteService}
}

type quoteItemRequest struct {
	ID        string  `json:"id"`
	Name      string  `json:"name" binding:"required"`
	Unit      string  `json:"unit"`
	Quantity  float64 `json:"quantity" binding:"required"`
	UnitPrice int64   `json:"unit_price"`
}

type quoteRequest struct {
	ClientID       string                `json:"client_id"`
	Title          string                `json:"title" binding:"required"`
	ClientSnapshot domain.ClientSnapshot `json:"client_snapshot"`
	PaymentTerms   domain.PaymentTerms   `json:"payment_terms"`
	Items          []quoteItemRequest    `json:"items" binding:"required"`
}

type publicQuotePasswordRequest struct {
	Password string `json:"password" binding:"required,len=6"`
}

func (h *QuoteHandler) CreateQuote(c *gin.Context) {
	companyID := c.GetString("company_id")
	userID := c.GetString("user_id")
	if companyID == "" || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req quoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quote, err := h.quoteService.CreateQuote(companyID, userID, req.ClientID, req.Title, req.ClientSnapshot, req.PaymentTerms, toQuoteItems(req.Items))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, quote)
}

func (h *QuoteHandler) ListQuotes(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	quotes, err := h.quoteService.ListQuotes(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quotes)
}

func (h *QuoteHandler) GetQuote(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	quote, err := h.quoteService.GetQuote(c.Param("id"), companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quote not found"})
		return
	}
	c.JSON(http.StatusOK, quote)
}

func (h *QuoteHandler) UpdateQuote(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req quoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quote, err := h.quoteService.UpdateQuote(c.Param("id"), companyID, req.ClientID, req.Title, req.ClientSnapshot, req.PaymentTerms, toQuoteItems(req.Items))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quote)
}

func (h *QuoteHandler) PublishQuote(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	quote, password, err := h.quoteService.PublishQuote(c.Param("id"), companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"quote":       quote,
		"password":    password,
		"share_url":   "/quotes/" + quote.ShareToken,
		"share_token": quote.ShareToken,
	})
}

func (h *QuoteHandler) ArchiveQuote(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	quote, err := h.quoteService.ArchiveQuote(c.Param("id"), companyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, quote)
}

func (h *QuoteHandler) DownloadQuotePDF(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	pdf, filename, err := h.quoteService.GenerateQuotePDF(c.Param("id"), companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", pdf)
}

func (h *QuoteHandler) GetPublicQuote(c *gin.Context) {
	quote, err := h.quoteService.GetPublicQuote(c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quote not found"})
		return
	}
	c.JSON(http.StatusOK, quote)
}

func (h *QuoteHandler) VerifyPublicQuotePassword(c *gin.Context) {
	var req publicQuotePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quote, err := h.quoteService.VerifyPublicQuotePassword(c.Param("token"), req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha inválida"})
		return
	}
	c.JSON(http.StatusOK, quote)
}

func toQuoteItems(items []quoteItemRequest) []domain.QuoteItem {
	result := make([]domain.QuoteItem, 0, len(items))
	for _, item := range items {
		result = append(result, domain.QuoteItem{
			ID:        item.ID,
			Name:      item.Name,
			Unit:      item.Unit,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	return result
}
