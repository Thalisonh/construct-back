package domain

import (
	"time"
)

type Client struct {
	ID         string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	UserID     string    `bson:"user_id" json:"user_id" datastore:"user_id"`
	CompanyID  string    `bson:"company_id" json:"company_id" datastore:"company_id" gorm:"index"`
	Name       string    `bson:"name" json:"name" datastore:"name"`
	Phone      string    `bson:"phone" json:"phone" datastore:"phone"`
	Address    string    `bson:"address" json:"address" datastore:"address"`
	Summary    string    `bson:"summary" json:"summary" datastore:"summary"`
	Comments   []Comment `bson:"comments" json:"comments" datastore:"comments" gorm:"foreignKey:ClientID"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at" datastore:"updated_at"`
	ClickCount int       `bson:"click_count" json:"click_count" datastore:"click_count"`
}

type Comment struct {
	ID        string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	ClientID  string    `bson:"client_id" json:"client_id" datastore:"client_id"`
	Content   string    `bson:"content" json:"content" datastore:"content"`
	CreatedAt time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
}
