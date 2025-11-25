package domain

import (
	"time"
)

type Link struct {
	ID          string    `bson:"_id" json:"id"`
	URL         string    `bson:"url" json:"url"`
	Description string    `bson:"description" json:"description"`
	UserID      string    `bson:"user_id" json:"user_id"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}
