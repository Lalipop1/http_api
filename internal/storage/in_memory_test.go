package storage

import (
	"errors"
	"http_api/internal/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryTaskStorage(t *testing.T) {
	t.Run("Create and Get task", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		task := &models.Task{ID: "test1", Status: models.StatusPending}

		storage.Create(task)
		retrievedTask, exists := storage.Get("test1")

		assert.True(t, exists)
		assert.Equal(t, task, retrievedTask)
	})

	t.Run("Get non-existent task", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		_, exists := storage.Get("nonexistent")

		assert.False(t, exists)
	})

	t.Run("Update task successfully", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		task := &models.Task{ID: "test2", Status: models.StatusPending}
		storage.Create(task)

		updatedTask, err := storage.Update("test2", func(t *models.Task) (*models.Task, error) {
			t.Status = models.StatusCompleted
			return t, nil
		})

		assert.NoError(t, err)
		assert.Equal(t, models.StatusCompleted, updatedTask.Status)
	})

	t.Run("Update non-existent task", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()

		_, err := storage.Update("nonexistent", func(t *models.Task) (*models.Task, error) {
			return t, nil
		})

		assert.True(t, errors.Is(err, ErrTaskNotFound))
	})

	t.Run("Delete existing task", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		task := &models.Task{ID: "test3"}
		storage.Create(task)

		deleted := storage.Delete("test3")
		assert.True(t, deleted)

		_, exists := storage.Get("test3")
		assert.False(t, exists)
	})

	t.Run("Delete non-existent task", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()

		deleted := storage.Delete("nonexistent")
		assert.False(t, deleted)
	})

	t.Run("Concurrent access", func(t *testing.T) {
		storage := NewInMemoryTaskStorage()
		var wg sync.WaitGroup
		count := 100

		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(id int) {
				defer wg.Done()
				task := &models.Task{ID: string(rune(id))}
				storage.Create(task)
				storage.Get(string(rune(id)))
			}(i)
		}

		wg.Wait()
		assert.Equal(t, count, len(storage.tasks))
	})
}
