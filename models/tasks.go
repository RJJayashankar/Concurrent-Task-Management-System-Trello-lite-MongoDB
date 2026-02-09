package models

import (
	"time"
)

type Task struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description" bson:"description"`
	Status      string    `json:"status" bson:"status"`
	Priority    string    `json:"priority" bson:"priority"`
	DueDate     time.Time `json:"duedate" bson:"duedate"`
	ProjectId   string    `json:"projectid" bson:"projectid"`
	AssignedTo  string    `json:"assignedto" bson:"assignedto"`
	CreatedAt   time.Time `json:"createdat" bson:"createdat"`
	UpdatedAt   time.Time `json:"updatedat" bson:"updatedat"`
}
