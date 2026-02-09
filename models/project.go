package models

import (
	"time"
)

type Project struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	OwnerID     string    `json:"ownerId" bson:"ownerId"`
	MemberIDs   []string  `json:"memberIds" bson:"memberIds"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

type ProjectDetailResponse struct {
	ID          string    `json:"id" bson:"id"`
	Name        string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	OwnerID     string    `json:"ownerId" bson:"ownerId"`
	MemberIDs   []string  `json:"memberIds" bson:"memberIds"`
	Members     []User    `json:"members" bson:"members"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}
