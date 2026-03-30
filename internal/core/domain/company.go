package domain

import (
	"time"
)

type Company struct {
	ID             string     `json:"id" datastore:"-" gorm:"primaryKey"`
	Name           string     `json:"name" datastore:"name"`
	CNPJ           string     `json:"cnpj" datastore:"cnpj" gorm:"uniqueIndex"`
	Email          string     `json:"email" datastore:"email"`
	Phone          string     `json:"phone" datastore:"phone"`
	Address        string     `json:"address" datastore:"address"`
	// Subscription fields
	Plan           string     `json:"plan" gorm:"default:free"`          // free | pro | enterprise
	PlanStatus     string     `json:"plan_status" gorm:"default:active"` // active | inactive
	PlanExpiresAt  *time.Time `json:"plan_expires_at"`
	SubscriptionID string     `json:"subscription_id"`
	CreatedAt      time.Time  `json:"created_at" datastore:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" datastore:"updated_at"`
}
