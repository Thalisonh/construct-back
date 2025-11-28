package domain

import (
	"time"
)

type Project struct {
	ID          string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	Title       string    `bson:"title" json:"title" datastore:"title"`
	Description string    `bson:"description" json:"description" datastore:"description"`
	UserID      string    `bson:"user_id" json:"user_id" datastore:"user_id" gorm:"index"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at" datastore:"updated_at"`
}
