package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"trello-lite/databases"
	"trello-lite/models"
	"trello-lite/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newProj models.Project
	if err := json.NewDecoder(r.Body).Decode(&newProj); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	newProj.CreatedAt = time.Now()
	collection := databases.GetCollection(databases.Client, "projects")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, newProj)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Project created!"})
}

func GetMyProjectsHandler(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get("Role")
	userID := r.Header.Get("User-ID")

	projectColl := databases.GetCollection(databases.Client, "projects")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	matchCriteria := bson.M{}
	if role != "Super Admin" {
		matchCriteria = bson.M{
			"$or": []bson.M{
				{"ownerId": userID},
				{"memberIds": userID},
			},
		}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchCriteria}},

		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "memberIds",
			"foreignField": "id",
			"as":           "members",
		}}},

		{{Key: "$addFields", Value: bson.M{
			"id": "$_id",
		}}},
	}

	cursor, err := projectColl.Aggregate(ctx, pipeline)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	responseList := make([]models.ProjectDetailResponse, 0)
	if err := cursor.All(ctx, &responseList); err != nil {
		http.Error(w, "Decoding error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseList)
}

func GetEverythingAggregateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Initialize with empty slices so they show as [] in JSON, not null
	type DataContent struct {
		Users    []models.User    `json:"users"`
		Projects []models.Project `json:"projects"`
		Tasks    []models.Task    `json:"tasks"`
	}

	content := DataContent{
		Users:    []models.User{},
		Projects: []models.Project{},
		Tasks:    []models.Task{},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{}}},
		{{Key: "$addFields", Value: bson.M{"id": "$_id"}}},
	}

	// --- USERS ---
	uColl := databases.GetCollection(databases.Client, "users")
	if cur, err := uColl.Aggregate(ctx, pipeline); err == nil {
		cur.All(ctx, &content.Users)
	}

	// --- PROJECTS ---
	pColl := databases.GetCollection(databases.Client, "projects")
	if cur, err := pColl.Aggregate(ctx, pipeline); err == nil {
		cur.All(ctx, &content.Projects)
	}

	// --- TASKS ---
	tColl := databases.GetCollection(databases.Client, "tasks")
	if cur, err := tColl.Aggregate(ctx, pipeline); err == nil {
		cur.All(ctx, &content.Tasks)
	}

	// 2. Use your utils to send the data
	utils.SendSuccess(w, "All system data retrieved successfully", content)
}
