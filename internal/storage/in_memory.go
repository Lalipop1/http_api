package storage

import (
	"errors"
	"internal/models"
	"sync"
)

type TaskStorage interface {
	Create(task *models.Task)
	Get(id string) (*models.Task, bool)
	GetAll() ([]models.Task, error)
	Update(id string, updateFn func(*models.Task) (*models.Task, error)) (*models.Task, error)
	Delete(id string) bool
}

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

func (s *InMemoryTaskStorage) GetAll() ([]models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, *task)
	}
	return tasks, nil
}

func (s *InMemoryTaskStorage) Update(id string, updateFn func(*models.Task) (*models.Task, error)) (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	updatedTask, err := updateFn(task)
	if err != nil {
		return nil, err
	}

	s.tasks[id] = updatedTask
	return updatedTask, nil
}

func (s *InMemoryTaskStorage) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return false
	}

	delete(s.tasks, id)
	return true
}

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidState = errors.New("invalid task state for operation")
)
