package storage

import (
	"internal/models"
	"sync"
)

type InMemoryTaskStorage struct {
	mu    sync.RWMutex
	tasks map[string]*models.Task
}

func NewInMemoryTaskStorage() *InMemoryTaskStorage {
	return &InMemoryTaskStorage{
		tasks: make(map[string]*models.Task),
	}
}

func (s *InMemoryTaskStorage) Create(task *models.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

func (s *InMemoryTaskStorage) Get(id string) (*models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, exists := s.tasks[id]
	return task, exists
}

func (s *InMemoryTaskStorage) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}
