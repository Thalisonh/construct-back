package handler

import (
	"construct-backend/internal/core/services"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService}
}

type createCheckoutRequest struct {
	Plan string `json:"plan" binding:"required"`
}

// CreateCheckout cria uma sessão de pagamento e retorna a URL do gateway.
func (h *SubscriptionHandler) CreateCheckout(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req createCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Plan == "free" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plano gratuito não requer checkout"})
		return
	}

	checkoutURL, err := h.subscriptionService.StartCheckout(companyID, req.Plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checkout_url": checkoutURL})
}

// HandleWebhook processa eventos de pagamento recebidos do gateway.
// Esta rota NÃO usa o authMiddleware — é chamada pelo Mercado Pago diretamente.
func (h *SubscriptionHandler) HandleWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	if err := h.subscriptionService.HandleWebhook(body, headers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// O MP espera 200 OK imediatamente
	c.Status(http.StatusOK)
}

// GetSubscriptionStatus retorna o plano, status e contagem de obras da empresa.
func (h *SubscriptionHandler) GetSubscriptionStatus(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	status, err := h.subscriptionService.GetSubscriptionStatus(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}
