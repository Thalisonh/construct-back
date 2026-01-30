package domain

import (
	"time"
)

type Link struct {
	ID          string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	URL         string    `bson:"url" json:"url" datastore:"url"`
	Description string    `bson:"description" json:"description" datastore:"description"`
	UserID      string    `bson:"user_id" json:"user_id" datastore:"user_id" gorm:"index"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at" datastore:"updated_at"`
	Count       int       `bson:"count" json:"count" datastore:"count"`
}

type LinkClick struct {
	ID        string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	LinkID    string    `bson:"link_id" json:"link_id" datastore:"link_id" gorm:"index"`
	CreatedAt time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
}
