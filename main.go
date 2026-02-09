package main

import (
	"fmt"
	"net/http"
	"trello-lite/databases"
	"trello-lite/handlers"
	"trello-lite/middleware"
	"trello-lite/utils" // 1. ADD THIS IMPORT
	"trello-lite/workers"
)

// 2. DEFINE THE FUNCTION HERE (Outside of main)
func handleNotFound(w http.ResponseWriter, r *http.Request) {
	utils.SendError(w, http.StatusNotFound, "Invalid endpoint: "+r.URL.Path)
}

func main() {
	databases.ConnectDB()

	// Background worker
	go workers.StartOverdueScanner()

	// 1. Specific Handlers
	http.HandleFunc("/signup", middleware.AuthMiddleware(handlers.SignupHandler))
	http.HandleFunc("/project/create", middleware.AuthMiddleware(handlers.CreateProjectHandler))
	http.HandleFunc("/task/create", middleware.AuthMiddleware(handlers.CreateTaskHandler))
	http.HandleFunc("/tasks", middleware.AuthMiddleware(handlers.GetTasksByProjectHandler))
	http.HandleFunc("/task/update", middleware.AuthMiddleware(handlers.UpdateTaskStatusHandler))
	http.HandleFunc("/task/delete", middleware.AuthMiddleware(handlers.DeleteTaskHandler))
	http.HandleFunc("/task/search", middleware.AuthMiddleware(handlers.SearchTaskHandler))
	http.HandleFunc("/getProject", middleware.AuthMiddleware(handlers.GetMyProjectsHandler))
	http.HandleFunc("/taskOwnerUpdate", middleware.AuthMiddleware(handlers.UpdateTaskownerHandler))
	http.HandleFunc("/getallusers", middleware.AuthMiddleware((handlers.GetAllUsersHandler)))
	http.HandleFunc("/everything", middleware.AuthMiddleware(handlers.GetEverythingAggregateHandler))
	http.HandleFunc("/login", handlers.LoginHandler)

	// 2. The Catch-All Handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleNotFound(w, r)
	})

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
