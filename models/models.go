package models

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type CreateTaskRequest struct {
	Title string `json:"title"`
}

type UpdateTaskRequest struct {
	Done bool `json:"done"`
}

type UpdateTaskResponse struct {
	Updated bool `json:"updated"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}