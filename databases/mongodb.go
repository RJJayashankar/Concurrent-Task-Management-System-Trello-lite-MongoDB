package databases

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB!")
	Client = client

	// CALL INDEX CREATION HERE
	CreateIndexes(client)

	return client
}

func CreateIndexes(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Tasks Collection Indexes
	taskColl := GetCollection(client, "tasks")
	taskIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "projectid", Value: 1}}},
		{Keys: bson.D{{Key: "assignedto", Value: 1}}},
	}
	taskColl.Indexes().CreateMany(ctx, taskIndexes)

	// 2. Users Collection Indexes
	userColl := GetCollection(client, "users")
	userIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	userColl.Indexes().CreateOne(ctx, userIndex)

	// 3. Projects Collection Indexes (ADDED)
	projColl := GetCollection(client, "projects")
	projIndexes := []mongo.IndexModel{
		// Speeds up finding projects you own
		{Keys: bson.D{{Key: "ownerId", Value: 1}}},
		// Speeds up finding projects where you are a member (Multikey Index)
		{Keys: bson.D{{Key: "memberIds", Value: 1}}},
	}
	_, err := projColl.Indexes().CreateMany(ctx, projIndexes)
	if err != nil {
		fmt.Println("Could not create project indexes:", err)
	}

	fmt.Println("Database Indexes verified/created for Users, Tasks, and Projects.")
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("Trello_lite").Collection(collectionName)
}
