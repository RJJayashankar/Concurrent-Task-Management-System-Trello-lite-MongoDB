package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"trello-lite/databases"
	"trello-lite/models"
	"trello-lite/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	if task.Status == "" {
		task.Status = "Todo"
	}

	collection := databases.GetCollection(databases.Client, "tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, task)
	if err != nil {

		fmt.Println("DB Insert Error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Task created",
		"id":      result.InsertedID,
	})
}

func GetTasksByProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("projectId")
	role := r.Header.Get("Role")
	userID := r.Header.Get("User-ID")

	if projectID == "" {
		utils.SendError(w, http.StatusBadRequest, "Missing projectId")
		return
	}

	collection := databases.GetCollection(databases.Client, "tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Build the Match Criteria based on RBAC
	matchCriteria := bson.M{"projectid": projectID}

	if role == "User" {
		// Regular users only see tasks assigned to them
		matchCriteria["assignedto"] = userID
	} else if role != "Super Admin" && role != "Admin" {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized Role")
		return
	}

	// 2. Define the Aggregation Pipeline
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchCriteria}},
		bson.D{{Key: "$addFields", Value: bson.M{"id": "$_id"}}},
	}

	// 3. Execute Aggregation
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error fetching tasks")
		return
	}
	defer cursor.Close(ctx)

	// 4. Initialize as empty slice to avoid 'null' in JSON
	tasks := []models.Task{}
	if err := cursor.All(ctx, &tasks); err != nil {
		fmt.Println("MONGODB DECODE ERROR:", err)
		utils.SendError(w, http.StatusInternalServerError, "Data format error")
		return
	}

	// 5. Send Success Response using your Utils
	utils.SendSuccess(w, "Tasks retrieved successfully", tasks)
}

func UpdateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	role := r.Header.Get("Role")
	userID := r.Header.Get("User-ID")
	collection := databases.GetCollection(databases.Client, "tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Try to find the task by String ID (like "0001")
	filter := bson.M{"_id": data.ID}

	// 2. Security for regular Users
	if role == "User" {
		filter["assignedto"] = userID
	}

	update := bson.M{"$set": bson.M{
		"status":    data.Status,
		"updatedat": time.Now(),
	}}

	result, err := collection.UpdateOne(ctx, filter, update)

	// 3. FALLBACK: If string ID fails, try converting to Hex ObjectId
	if err == nil && result.MatchedCount == 0 && len(data.ID) == 24 {
		objID, _ := primitive.ObjectIDFromHex(data.ID)
		filter = bson.M{"_id": objID}
		if role == "User" {
			filter["assignedto"] = userID
		}
		result, err = collection.UpdateOne(ctx, filter, update)
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Task not found. Check if ID "+data.ID+" exists in the 'tasks' collection.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("id")
	role := r.Header.Get("Role")

	collection := databases.GetCollection(databases.Client, "tasks")

	// DIAGNOSTIC: Count how many tasks exist in total
	count, _ := collection.CountDocuments(context.TODO(), bson.M{})
	fmt.Printf("DEBUG: Total tasks in collection: %d | Trying to delete: '%s'\n", count, taskID)

	if role == "User" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Attempt the delete
	result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": taskID})

	if err != nil {
		fmt.Println("DB ERROR:", err)
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		fmt.Println("LOG: No document matched this ID.")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Deleted " + taskID})
}
func SearchTaskHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the 'title' from the URL query: /task/search?title=Fix
	queryTitle := r.URL.Query().Get("title")
	if queryTitle == "" {
		http.Error(w, "Query parameter 'title' is required", http.StatusBadRequest)
		return
	}

	collection := databases.GetCollection(databases.Client, "tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. Create a Regex filter
	// "i" means case-insensitive (finds 'fix' or 'Fix')
	filter := bson.M{
		"title": bson.M{"$regex": queryTitle, "$options": "i"},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// 3. Decode the results into a slice (list) of Tasks
	var results []models.Task
	if err = cursor.All(ctx, &results); err != nil {
		http.Error(w, "Error parsing results", http.StatusInternalServerError)
		return
	}

	// 4. Send back the results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func UpdateTaskownerHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ID         string `json:"id"`
		AssignedTo string `json:"assignedto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	role := r.Header.Get("Role")
	userID := r.Header.Get("User-ID")
	collection := databases.GetCollection(databases.Client, "tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Try to find the task by String ID (like "0001")
	filter := bson.M{"_id": data.ID}

	// 2. Security for regular Users
	if role == "User" {
		filter["assignedto"] = userID
	}

	update := bson.M{"$set": bson.M{
		"AssignedTo": data.AssignedTo,
		"updatedat":  time.Now(),
	}}

	result, err := collection.UpdateOne(ctx, filter, update)

	// 3. FALLBACK: If string ID fails, try converting to Hex ObjectId
	if err == nil && result.MatchedCount == 0 && len(data.ID) == 24 {
		objID, _ := primitive.ObjectIDFromHex(data.ID)
		filter = bson.M{"_id": objID}
		if role == "User" {
			filter["assignedto"] = userID
		}
		result, err = collection.UpdateOne(ctx, filter, update)
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Task not found. Check if ID "+data.ID+" exists in the 'tasks' collection.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
}
