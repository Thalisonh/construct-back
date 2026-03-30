package ports

// PaymentGateway é a Port de pagamento — não depende de nenhum provider específico.
// Para trocar de gateway, basta criar um novo Adapter que implemente esta interface.
type PaymentGateway interface {
	CreateCheckout(req CheckoutRequest) (*CheckoutResponse, error)
	ParseWebhook(body []byte, headers map[string]string) (*WebhookEvent, error)
	ValidateWebhookSignature(body []byte, signature string) bool
}

// CheckoutRequest contém os dados necessários para criar uma sessão de pagamento.
type CheckoutRequest struct {
	Plan        string
	CompanyID   string
	CompanyName string
	Email       string
	Price       float64
	SuccessURL  string
	FailureURL  string
}

// CheckoutResponse retorna a URL de redirect e o ID externo da preferência.
type CheckoutResponse struct {
	CheckoutURL string
	ExternalID  string // ID da preferência no gateway
}

// WebhookEvent é a representação normalizada de um evento de pagamento.
type WebhookEvent struct {
	Provider   string
	EventType  string // "payment.approved" | "subscription.cancelled"
	PaymentID  string
	CompanyID  string // extraído do metadata/external_reference
	PlanName   string
}
