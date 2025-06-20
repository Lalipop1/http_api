package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"http_api/internal/models"
	"http_api/internal/storage"
	"math/rand"
	"time"
)

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// В случае ошибки используем timestamp как fallback
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

type TaskService struct {
	storage storage.TaskStorage
}

func NewTaskService(storage storage.TaskStorage) *TaskService {
	return &TaskService{storage: storage}
}

func (s *TaskService) CreateTask(ctx context.Context, description string) (*models.Task, error) {
	task := &models.Task{
		ID:          generateID(),
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
		Description: description,
	}

	s.storage.Create(task)

	go s.processTask(task.ID)

	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id string) (*models.Task, error) {
	task, exists := s.storage.Get(id)
	if !exists {
		return nil, storage.ErrTaskNotFound
	}
	return task, nil
}

func (s *TaskService) ListTasks(ctx context.Context) (*models.TaskList, error) {
	tasks, err := s.storage.GetAll()
	if err != nil {
		return nil, err
	}

	return &models.TaskList{
		Tasks: tasks,
		Total: len(tasks),
	}, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, update models.TaskUpdate) (*models.Task, error) {
	updatedTask, err := s.storage.Update(id, func(task *models.Task) (*models.Task, error) {
		if update.Description != nil {
			task.Description = *update.Description
		}
		return task, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return updatedTask, nil
}

func (s *TaskService) CancelTask(ctx context.Context, id string) (*models.Task, error) {
	updatedTask, err := s.storage.Update(id, func(task *models.Task) (*models.Task, error) {
		if task.Status != models.StatusPending && task.Status != models.StatusProcessing {
			return nil, storage.ErrInvalidState
		}

		now := time.Now()
		task.Status = models.StatusCancelled
		task.CancelledAt = &now
		if task.StartedAt != nil {
			task.Duration = now.Sub(*task.StartedAt).Seconds()
		}
		return task, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	return updatedTask, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	if !s.storage.Delete(id) {
		return storage.ErrTaskNotFound
	}
	return nil
}

func (s *TaskService) processTask(id string) {
	_, err := s.storage.Update(id, func(task *models.Task) (*models.Task, error) {
		if task.Status != models.StatusPending {
			return nil, storage.ErrInvalidState
		}

		now := time.Now()
		task.Status = models.StatusProcessing
		task.StartedAt = &now
		return task, nil
	})

	if err != nil {
		return
	}

	// Имитация длительной задачи
	processingTime := 3*time.Minute + time.Duration(rand.Intn(120))*time.Second
	time.Sleep(processingTime)

	// Завершение задачи
	_, err = s.storage.Update(id, func(task *models.Task) (*models.Task, error) {
		if task.Status != models.StatusProcessing {
			return nil, storage.ErrInvalidState
		}

		now := time.Now()
		task.Status = models.StatusCompleted
		task.CompletedAt = &now
		task.Duration = now.Sub(*task.StartedAt).Seconds()
		task.Result = fmt.Sprintf("Processed for %.2f seconds", task.Duration)
		return task, nil
	})
}
