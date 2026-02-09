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

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	newUser.CreatedAt = time.Now()
	collection := databases.GetCollection(databases.Client, "users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, newUser)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Signup successful!"})
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Setup connection and context
	collection := databases.GetCollection(databases.Client, "users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. Security Check using your Utils
	role := r.Header.Get("Role")
	if role != "Super Admin" && role != "Admin" {
		utils.SendError(w, http.StatusForbidden, "Access denied: Admin privileges required")
		return
	}

	// 3. Define the Aggregation Pipeline
	// We use bson.D to avoid the "missing type in composite literal" error
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{}}},                 // Match all users
		bson.D{{Key: "$addFields", Value: bson.M{"id": "$_id"}}}, // Map _id to id
	}

	// 4. Execute Aggregate
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to query users")
		return
	}
	defer cursor.Close(ctx)

	// 5. Decode results into a slice (initialized to avoid 'null')
	users := []models.User{}
	if err := cursor.All(ctx, &users); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error parsing user data")
		return
	}

	// 6. Send success response using Utils
	utils.SendSuccess(w, "User list retrieved successfully", users)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}

	// 1. Decode the email from the request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 2. Look for the user in MongoDB
	collection := databases.GetCollection(databases.Client, "users")
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"email": request.Email}).Decode(&user)

	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "User not found. Please signup first.")
		return
	}

	// 3. Instead of checking password, we use the Role found in the DB
	// This makes the Role the "Key" for the token
	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	// 4. Return the token and the role to the user
	utils.SendSuccess(w, "Authentication successful", map[string]string{
		"token": token,
		"role":  user.Role,
		"id":    user.ID,
	})
}
