package domain

import (
	"time"
)

type Link struct {
	ID          string    `bson:"_id" json:"id"`
	URL         string    `bson:"url" json:"url"`
	Description string    `bson:"description" json:"description"`
	ProjectID   string    `bson:"project_id" json:"project_id"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}
