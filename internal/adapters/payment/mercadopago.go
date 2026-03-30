package payment

import (
	"construct-backend/internal/core/ports"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// MercadoPagoAdapter implementa a interface PaymentGateway utilizando a API do Mercado Pago.
// Para trocar de provider, crie um novo arquivo neste pacote implementando ports.PaymentGateway.
type MercadoPagoAdapter struct {
	accessToken string
	successURL  string
	failureURL  string
}

func NewMercadoPagoAdapter(accessToken, successURL, failureURL string) *MercadoPagoAdapter {
	return &MercadoPagoAdapter{
		accessToken: accessToken,
		successURL:  successURL,
		failureURL:  failureURL,
	}
}

// CreateCheckout cria uma preferência de pagamento no Mercado Pago e retorna a URL de checkout.
func (m *MercadoPagoAdapter) CreateCheckout(req ports.CheckoutRequest) (*ports.CheckoutResponse, error) {
	priceMap := map[string]float64{
		"pro":        59.0,
		"enterprise": 149.0,
	}

	price := req.Price
	if price == 0 {
		if p, ok := priceMap[req.Plan]; ok {
			price = p
		}
	}

	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("ConstructPro — Plano %s", strings.Title(req.Plan)),
				"quantity":    1,
				"unit_price":  price,
				"currency_id": "BRL",
			},
		},
		"external_reference": req.CompanyID,
		"metadata": map[string]string{
			"company_id": req.CompanyID,
			"plan":       req.Plan,
		},
		"back_urls": map[string]string{
			"success": req.SuccessURL,
			"failure": req.FailureURL,
			"pending": req.FailureURL,
		},
		"auto_return": "approved",
		"expires":     true,
		"expiration_date_to": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest("POST", "https://api.mercadopago.com/checkout/preferences", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("mercadopago: failed to create preference, status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	checkoutURL, _ := result["init_point"].(string)
	externalID, _ := result["id"].(string)

	return &ports.CheckoutResponse{
		CheckoutURL: checkoutURL,
		ExternalID:  externalID,
	}, nil
}

// mpWebhookPayload é o payload padrão recebido nos webhooks do Mercado Pago.
type mpWebhookPayload struct {
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

// ParseWebhook interpreta o payload do webhook do Mercado Pago e retorna um WebhookEvent normalizado.
func (m *MercadoPagoAdapter) ParseWebhook(body []byte, headers map[string]string) (*ports.WebhookEvent, error) {
	var payload mpWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("mercadopago: invalid webhook payload: %w", err)
	}

	// Busca detalhes do pagamento para extrair a external_reference (company_id) e o plano
	paymentData, err := m.fetchPaymentDetails(payload.Data.ID)
	if err != nil {
		return nil, err
	}

	companyID, _ := paymentData["external_reference"].(string)

	// Mapeia o status do MP para nosso evento interno
	status, _ := paymentData["status"].(string)
	eventType := "payment.other"
	if status == "approved" {
		eventType = "payment.approved"
	} else if status == "cancelled" || status == "rejected" {
		eventType = "payment.failed"
	}

	// Tenta extrair o plano dos metadados
	planName := "pro"
	if meta, ok := paymentData["metadata"].(map[string]interface{}); ok {
		if p, ok := meta["plan"].(string); ok {
			planName = p
		}
	}

	return &ports.WebhookEvent{
		Provider:  "mercadopago",
		EventType: eventType,
		PaymentID: payload.Data.ID,
		CompanyID: companyID,
		PlanName:  planName,
	}, nil
}

// fetchPaymentDetails consulta a API do MP para obter detalhes de um pagamento.
func (m *MercadoPagoAdapter) fetchPaymentDetails(paymentID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s", paymentID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+m.accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

// ValidateWebhookSignature valida o cabeçalho x-signature do Mercado Pago.
// Em sandbox, retorna sempre true. Em produção, implemente a verificação HMAC do MP.
func (m *MercadoPagoAdapter) ValidateWebhookSignature(body []byte, signature string) bool {
	// TODO: implementar validação HMAC com MP_WEBHOOK_SECRET em produção
	// Ref: https://www.mercadopago.com.br/developers/pt/docs/your-integrations/notifications/webhooks
	return true
}
