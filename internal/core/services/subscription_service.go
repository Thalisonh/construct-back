package services

import (
	"construct-backend/internal/core/ports"
	"fmt"
	"time"
)

const (
	PlanFree       = "free"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"

	PlanStatusActive   = "active"
	PlanStatusInactive = "inactive"

	freePlanProjectLimit = 3
)

// SubscriptionService orquestra pagamentos e controle de acesso.
// Depende APENAS da interface PaymentGateway — nunca do Mercado Pago diretamente.
type SubscriptionService struct {
	gateway     ports.PaymentGateway
	companyRepo ports.CompanyRepository
	subRepo     ports.SubscriptionRepository
	successURL  string
	failureURL  string
}

func NewSubscriptionService(
	gateway ports.PaymentGateway,
	companyRepo ports.CompanyRepository,
	subRepo ports.SubscriptionRepository,
	successURL, failureURL string,
) *SubscriptionService {
	return &SubscriptionService{
		gateway:     gateway,
		companyRepo: companyRepo,
		subRepo:     subRepo,
		successURL:  successURL,
		failureURL:  failureURL,
	}
}

// StartCheckout cria uma sessão de checkout no gateway e retorna a URL de redirect.
func (s *SubscriptionService) StartCheckout(companyID, plan string) (string, error) {
	company, err := s.companyRepo.GetCompanyByID(companyID)
	if err != nil {
		return "", fmt.Errorf("company not found: %w", err)
	}

	resp, err := s.gateway.CreateCheckout(ports.CheckoutRequest{
		Plan:        plan,
		CompanyID:   companyID,
		CompanyName: company.Name,
		Email:       company.Email,
		SuccessURL:  s.successURL,
		FailureURL:  s.failureURL,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create checkout: %w", err)
	}

	return resp.CheckoutURL, nil
}

// HandleWebhook processa um evento recebido do gateway e atualiza o plano se aprovado.
func (s *SubscriptionService) HandleWebhook(body []byte, headers map[string]string) error {
	if sig, ok := headers["x-signature"]; ok {
		if !s.gateway.ValidateWebhookSignature(body, sig) {
			return fmt.Errorf("invalid webhook signature")
		}
	}

	event, err := s.gateway.ParseWebhook(body, headers)
	if err != nil {
		return fmt.Errorf("failed to parse webhook: %w", err)
	}

	if event.EventType != "payment.approved" {
		// Outros eventos (cancelamentos, reembolsos) — ignorar por enquanto
		return nil
	}

	// Ativa o plano por 30 dias (mensal)
	expiresAt := time.Now().AddDate(0, 1, 0)
	return s.companyRepo.UpdateCompanyPlan(
		event.CompanyID,
		event.PlanName,
		PlanStatusActive,
		event.PaymentID,
		&expiresAt,
	)
}

// CheckProjectLimit verifica se a empresa pode criar mais obras no plano atual.
// Retorna erro se o plano free atingiu o limite de 3 obras.
func (s *SubscriptionService) CheckProjectLimit(companyID string) error {
	company, err := s.companyRepo.GetCompanyByID(companyID)
	if err != nil {
		return err
	}

	// Usuários dos planos pagos não têm limite
	if company.Plan != PlanFree {
		return nil
	}

	count, err := s.subRepo.CountProjectsByCompany(companyID)
	if err != nil {
		return err
	}

	if count >= freePlanProjectLimit {
		return fmt.Errorf("limite_atingido: plano gratuito permite até %d obras. Faça upgrade para criar mais", freePlanProjectLimit)
	}

	return nil
}

// GetSubscriptionStatus retorna o plano e status da empresa.
func (s *SubscriptionService) GetSubscriptionStatus(companyID string) (*SubscriptionStatus, error) {
	company, err := s.companyRepo.GetCompanyByID(companyID)
	if err != nil {
		return nil, err
	}

	count, _ := s.subRepo.CountProjectsByCompany(companyID)

	return &SubscriptionStatus{
		Plan:       company.Plan,
		Status:     company.PlanStatus,
		ExpiresAt:  company.PlanExpiresAt,
		ProjectCount: int(count),
		ProjectLimit: planProjectLimit(company.Plan),
	}, nil
}

type SubscriptionStatus struct {
	Plan         string     `json:"plan"`
	Status       string     `json:"plan_status"`
	ExpiresAt    *time.Time `json:"plan_expires_at"`
	ProjectCount int        `json:"project_count"`
	ProjectLimit int        `json:"project_limit"` // -1 = ilimitado
}

func planProjectLimit(plan string) int {
	if plan == PlanFree {
		return freePlanProjectLimit
	}
	return -1
}
