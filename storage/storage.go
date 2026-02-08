package storage

import (
	"sync"

	"practice2/models"
)

type TaskStorage struct {
	mu     sync.Mutex
	tasks  map[int]models.Task
	nextID int
}

func NewTaskStorage() *TaskStorage {
	return &TaskStorage{
		tasks:  make(map[int]models.Task),
		nextID: 1,
	}
}

func (s *TaskStorage) Create(title string) models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := models.Task{
		ID:    s.nextID,
		Title: title,
		Done:  false,
	}
	s.tasks[s.nextID] = task
	s.nextID++

	return task
}

func (s *TaskStorage) GetAll() []models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		result = append(result, task)
	}

	return result
}

func (s *TaskStorage) GetByID(id int) (models.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	return task, exists
}

func (s *TaskStorage) Update(id int, done bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return false
	}

	task.Done = done
	s.tasks[id] = task
	return true
}