package main

import (
	"log"
	"net/http"

	"practice2/handlers"
	"practice2/middleware"
	"practice2/storage"
)

func main() {
	store := storage.NewTaskStorage()
	taskHandler := handlers.NewTaskHandler(store)
	rateLimiter := middleware.NewRateLimiter()

	mux := http.NewServeMux()

	
	tasksHandler := http.HandlerFunc( func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskHandler.GetTasks(w, r)
		case http.MethodPost:
			taskHandler.CreateTask(w, r)
		case http.MethodPatch:
			taskHandler.UpdateTask(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	
	handler := middleware.RateLimitMiddleware(rateLimiter)(tasksHandler)
	handler = middleware.APIKeyMiddleware(handler)

	
	mux.Handle("/tasks", handler)

	loggingMiddleware := middleware.NewLoggingMiddleware(mux)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", loggingMiddleware); err != nil {
		log.Fatal(err)
	}
}
