package domain

import (
	"time"
)

type User struct {
	ID        string    `bson:"_id" json:"id" datastore:"-"`
	Email     string    `bson:"email" json:"email" datastore:"email"`
	Password  string    `bson:"password" json:"-" datastore:"password"`
	Name      string    `bson:"name" json:"name" datastore:"name"`
	CreatedAt time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at" datastore:"updated_at"`
}

type UsernameVerification struct {
	Username string `bson:"username" json:"username"`
}

type PublicProfile struct {
	ID       string `bson:"_id" json:"id"`
	Username string `bson:"username" json:"username"`
	Name     string `bson:"name" json:"name"`
	Bio      string `bson:"bio" json:"bio"`
	Avatar   string `bson:"avatar" json:"avatar"`
	Links    []Link `bson:"links" json:"links"`
}
