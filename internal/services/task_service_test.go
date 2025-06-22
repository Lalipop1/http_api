package services

import (
	"context"
	"errors"
	"http_api/internal/models"
	"http_api/internal/storage"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Create(task *models.Task) {
	m.Called(task)
}

func (m *MockStorage) Get(id string) (*models.Task, bool) {
	args := m.Called(id)
	return args.Get(0).(*models.Task), args.Bool(1)
}

func (m *MockStorage) GetAll() ([]models.Task, error) {
	args := m.Called()
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockStorage) Update(id string, updateFn func(*models.Task) (*models.Task, error)) (*models.Task, error) {
	args := m.Called(id, updateFn)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockStorage) Delete(id string) bool {
	args := m.Called(id)
	return args.Bool(0)
}

func TestTaskService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create task successfully", func(t *testing.T) {
		mockStorage := new(MockStorage)
		service := NewTaskService(mockStorage)

		mockStorage.On("Create", mock.AnythingOfType("*models.Task")).Once()

		task, err := service.CreateTask(ctx, "test desc")

		assert.NoError(t, err)
		assert.Equal(t, "test desc", task.Description)
		assert.Equal(t, models.StatusPending, task.Status)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Get existing task", func(t *testing.T) {
		mockStorage := new(MockStorage)
		service := NewTaskService(mockStorage)
		expectedTask := &models.Task{ID: "test1"}

		mockStorage.On("Get", "test1").Return(expectedTask, true)

		task, err := service.GetTask(ctx, "test1")

		assert.NoError(t, err)
		assert.Equal(t, expectedTask, task)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Get non-existent task", func(t *testing.T) {
		mockStorage := new(MockStorage)
		service := NewTaskService(mockStorage)

		mockStorage.On("Get", "nonexistent").Return(&models.Task{}, false)

		_, err := service.GetTask(ctx, "nonexistent")

		assert.True(t, errors.Is(err, storage.ErrTaskNotFound))
		mockStorage.AssertExpectations(t)
	})

	t.Run("Cancel pending task", func(t *testing.T) {
		mockStorage := new(MockStorage)
		service := NewTaskService(mockStorage)
		task := &models.Task{ID: "test2", Status: models.StatusPending}

		mockStorage.On("Update", "test2", mock.Anything).Run(func(args mock.Arguments) {
			updateFn := args.Get(1).(func(*models.Task) (*models.Task, error))
			updateFn(task)
		}).Return(task, nil)

		result, err := service.CancelTask(ctx, "test2")

		assert.NoError(t, err)
		assert.Equal(t, models.StatusCancelled, result.Status)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Fail to cancel completed task", func(t *testing.T) {
		mockStorage := new(MockStorage)
		service := NewTaskService(mockStorage)

		mockStorage.On("Update", "test3", mock.Anything).Run(func(args mock.Arguments) {
			updateFn := args.Get(1).(func(*models.Task) (*models.Task, error))
			task := &models.Task{
				ID:     "test3",
				Status: models.StatusCompleted,
			}
			_, err := updateFn(task)
			assert.Error(t, err)
		}).Return(nil, storage.ErrInvalidState)

		_, err := service.CancelTask(ctx, "test3")

		assert.True(t, errors.Is(err, storage.ErrInvalidState))
		mockStorage.AssertExpectations(t)
	})

	t.Run("Delete existing task", func(t *testing.T) {
		mockStorage := new(MockStorage)
		service := NewTaskService(mockStorage)

		mockStorage.On("Delete", "test4").Return(true)

		err := service.DeleteTask(ctx, "test4")

		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Delete non-existent task", func(t *testing.T) {
		mockStorage := new(MockStorage)
		service := NewTaskService(mockStorage)

		mockStorage.On("Delete", "nonexistent").Return(false)

		err := service.DeleteTask(ctx, "nonexistent")

		assert.True(t, errors.Is(err, storage.ErrTaskNotFound))
		mockStorage.AssertExpectations(t)
	})
}
