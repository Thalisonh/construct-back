package domain

import (
	"time"
)

type Project struct {
	ID        string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	Name      string    `bson:"name" json:"name" datastore:"name"`
	ClientID  string    `bson:"client_id" json:"client_id" datastore:"client_id"`
	StartDate time.Time `bson:"start_date" json:"start_date" datastore:"start_date"`
	Address   string    `bson:"address" json:"address" datastore:"address"`
	Summary   string    `bson:"summary" json:"summary" datastore:"summary"`
	Status    string    `bson:"status" json:"status" datastore:"status"`
	UserID    string    `bson:"user_id" json:"user_id" datastore:"user_id" gorm:"index"`
	Tasks     []Task    `bson:"tasks" json:"tasks" datastore:"tasks" gorm:"foreignKey:ProjectID"`
	CreatedAt time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at" datastore:"updated_at"`
}

type Task struct {
	ID        string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	ProjectID string    `bson:"project_id" json:"project_id" datastore:"project_id"`
	Name      string    `bson:"name" json:"name" datastore:"name"`
	DueDate   time.Time `bson:"due_date" json:"due_date" datastore:"due_date"`
	Status    string    `bson:"status" json:"status" datastore:"status"`
	Subtasks  []Subtask `bson:"subtasks" json:"subtasks" datastore:"subtasks" gorm:"foreignKey:TaskID"`
	CreatedAt time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
}

type Subtask struct {
	ID        string    `bson:"_id" json:"id" datastore:"-" gorm:"primaryKey"`
	TaskID    string    `bson:"task_id" json:"task_id" datastore:"task_id"`
	Name      string    `bson:"name" json:"name" datastore:"name"`
	Status    string    `bson:"status" json:"status" datastore:"status"`
	CreatedAt time.Time `bson:"created_at" json:"created_at" datastore:"created_at"`
}
