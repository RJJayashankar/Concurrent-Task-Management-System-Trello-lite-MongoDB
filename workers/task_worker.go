package workers

import (
	"context"
	"fmt"
	"time"
	"trello-lite/databases"
	"trello-lite/models"

	"go.mongodb.org/mongo-driver/bson"
)

func StartOverdueScanner() {

	ticker := time.NewTicker(30 * time.Second)
	fmt.Println("Background Worker: Overdue scanner started...")

	for range ticker.C {
		fmt.Println("Background Worker: Checking for overdue tasks...")
		scanForOverdueTasks()
	}
}

func scanForOverdueTasks() {
	collection := databases.GetCollection(databases.Client, "tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"status":  bson.M{"$ne": "Done"},
		"duedate": bson.M{"$lt": time.Now()},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		fmt.Println("Worker Error:", err)
		return
	}
	defer cursor.Close(ctx)

	var overdueTasks []models.Task
	if err = cursor.All(ctx, &overdueTasks); err != nil {
		return
	}

	if len(overdueTasks) > 0 {
		fmt.Printf("ALERT: Found %d overdue tasks!\n", len(overdueTasks))
		for _, task := range overdueTasks {
			fmt.Printf("- Task '%s' was due on %v\n", task.Title, task.DueDate)
		}
	}
}
