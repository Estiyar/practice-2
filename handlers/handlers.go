package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"practice2/models"
	"practice2/storage"
)

type TaskHandler struct {
	storage *storage.TaskStorage
}

func NewTaskHandler(storage *storage.TaskStorage) *TaskHandler {
	return &TaskHandler{storage: storage}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	doneStr := r.URL.Query().Get("done")

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			h.sendError(w, "invalid id", http.StatusBadRequest)
			return
		}

		task, exists := h.storage.GetByID(id)
		if !exists {
			h.sendError(w, "task not found", http.StatusNotFound)
			return
		}

		if doneStr != "" {
			done, err := parseBool(doneStr)
			if err != nil {
				h.sendError(w, "invalid done parameter", http.StatusBadRequest)
				return
			}
			if task.Done != done {
				h.sendError(w, "task not found", http.StatusNotFound)
				return
			}
		}

		h.sendJSON(w, task, http.StatusOK)
		return
	}

	tasks := h.storage.GetAll()

	if doneStr != "" {
		done, err := parseBool(doneStr)
		if err != nil {
			h.sendError(w, "invalid done parameter", http.StatusBadRequest)
			return
		}

		filtered := make([]models.Task, 0)
		for _, task := range tasks {
			if task.Done == done {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	h.sendJSON(w, tasks, http.StatusOK)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		h.sendError(w, "title cannot be empty", http.StatusBadRequest)
		return
	}

	task := h.storage.Create(req.Title)
	h.sendJSON(w, task, http.StatusCreated)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		h.sendError(w, "id parameter required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !h.storage.Update(id, req.Done) {
		h.sendError(w, "task not found", http.StatusNotFound)
		return
	}

	h.sendJSON(w, models.UpdateTaskResponse{Updated: true}, http.StatusOK)
}

func (h *TaskHandler) sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TaskHandler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: message})
}

func parseBool(s string) (bool, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}
	return false, &strconv.NumError{Func: "ParseBool", Num: s, Err: strconv.ErrSyntax}
}