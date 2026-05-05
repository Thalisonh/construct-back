package domain

import "time"

const (
	QuoteStatusDraft     = "draft"
	QuoteStatusPublished = "published"
	QuoteStatusArchived  = "archived"
)

type Quote struct {
	ID                string          `json:"id" gorm:"primaryKey"`
	CompanyID         string          `json:"company_id" gorm:"index"`
	ClientID          string          `json:"client_id" gorm:"index"`
	Title             string          `json:"title"`
	Status            string          `json:"status" gorm:"index"`
	ClientSnapshot    ClientSnapshot  `json:"client_snapshot" gorm:"serializer:json"`
	CompanySnapshot   CompanySnapshot `json:"company_snapshot" gorm:"serializer:json"`
	PaymentTerms      PaymentTerms    `json:"payment_terms" gorm:"serializer:json"`
	Items             []QuoteItem     `json:"items" gorm:"foreignKey:QuoteID"`
	Subtotal          int64           `json:"subtotal"`
	Total             int64           `json:"total"`
	ShareToken        string          `json:"share_token,omitempty" gorm:"uniqueIndex"`
	SharePasswordHash string          `json:"-" gorm:"column:share_password_hash"`
	PDFURL            string          `json:"pdf_url"`
	CreatedBy         string          `json:"created_by" gorm:"index"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type QuoteItem struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	QuoteID   string    `json:"quote_id" gorm:"index"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	Quantity  float64   `json:"quantity"`
	UnitPrice int64     `json:"unit_price"`
	Total     int64     `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ClientSnapshot struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Summary string `json:"summary"`
}

type PaymentTerms struct {
	InstallmentsEnabled bool    `json:"installments_enabled"`
	InstallmentCount    int     `json:"installment_count"`
	InterestType        string  `json:"interest_type"`
	InterestRate        float64 `json:"interest_rate"`
	InstallmentAmount   int64   `json:"installment_amount"`
	CashAmount          int64   `json:"cash_amount"`
}

type CompanySnapshot struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	LogoURL string `json:"logo_url"`
}
