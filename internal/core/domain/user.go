package domain

import (
	"time"
)

type User struct {
	ID        string    `json:"id" datastore:"-" gorm:"primaryKey"`
	Username  string    `json:"username" datastore:"username" gorm:"uniqueIndex"`
	Email     string    `json:"email" datastore:"email" gorm:"uniqueIndex"`
	Password  string    `json:"-" datastore:"password"`
	Name      string    `json:"name" datastore:"name"`
	Phone     string    `json:"phone" datastore:"phone"`
	CompanyID string    `json:"company_id" datastore:"company_id" gorm:"index"`
	Role      string    `json:"role" datastore:"role"` // "admin" or "member"
	Bio       string    `json:"bio" datastore:"bio"`
	Avatar    string    `json:"avatar" datastore:"avatar"`
	CreatedAt time.Time `json:"created_at" datastore:"created_at"`
	UpdatedAt time.Time `json:"updated_at" datastore:"updated_at"`
}

type UsernameVerification struct {
	Username string `json:"username"`
}

type PublicProfile struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	Avatar    string `json:"avatar"`
	CompanyID string `json:"company_id"`
	Links     []Link `json:"links"`
}
