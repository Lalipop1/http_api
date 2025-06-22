package main

import (
	"bytes"
	"encoding/json"
	"http_api/internal/handlers"
	"http_api/internal/models"
	"http_api/internal/services"
	"http_api/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerRoutes(t *testing.T) {
	// Инициализация сервера
	taskStorage := storage.NewInMemoryTaskStorage()
	taskService := services.NewTaskService(taskStorage)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Создаем новый роутер
	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", taskHandler.HandleTasks)
	mux.HandleFunc("/tasks/", taskHandler.HandleTaskByID)

	// Переменная для хранения ID созданной задачи
	var createdTaskID string

	t.Run("GET /tasks - пустой список", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.TaskList
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.Total)
	})

	t.Run("POST /tasks - создание задачи", func(t *testing.T) {
		body := bytes.NewBufferString(`{"description":"test task"}`)
		req := httptest.NewRequest("POST", "/tasks", body)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var task models.Task
		err := json.NewDecoder(w.Body).Decode(&task)
		assert.NoError(t, err)
		assert.Equal(t, "test task", task.Description)
		assert.Equal(t, models.StatusPending, task.Status)

		createdTaskID = task.ID // Сохраняем для последующих тестов
	})

	t.Run("GET /tasks - непустой список", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.TaskList
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, 1, response.Total)
	})

	t.Run("GET /tasks/{id} - получение задачи", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks/"+createdTaskID, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var task models.Task
		err := json.NewDecoder(w.Body).Decode(&task)
		assert.NoError(t, err)
		assert.Equal(t, createdTaskID, task.ID)
	})

	t.Run("PUT /tasks/{id} - обновление задачи", func(t *testing.T) {
		body := bytes.NewBufferString(`{"description":"updated description"}`)
		req := httptest.NewRequest("PUT", "/tasks/"+createdTaskID, body)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var task models.Task
		err := json.NewDecoder(w.Body).Decode(&task)
		assert.NoError(t, err)
		assert.Equal(t, "updated description", task.Description)
	})

	t.Run("POST /tasks/{id}/cancel - отмена задачи", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/tasks/"+createdTaskID+"/cancel", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var task models.Task
		err := json.NewDecoder(w.Body).Decode(&task)
		assert.NoError(t, err)
		assert.Equal(t, models.StatusCancelled, task.Status)
	})

	t.Run("DELETE /tasks/{id} - удаление задачи", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/tasks/"+createdTaskID, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	// Тесты для ошибочных сценариев
	t.Run("Невалидные запросы", func(t *testing.T) {
		tests := []struct {
			name         string
			method       string
			path         string
			body         string
			expectedCode int
		}{
			{"Невалидный JSON", "POST", "/tasks", `{"description":}`, http.StatusBadRequest},
			{"Пустое описание", "POST", "/tasks", `{"description":""}`, http.StatusBadRequest},
			{"Несуществующая задача", "GET", "/tasks/nonexistent", "", http.StatusNotFound},
			{"Отмена несуществующей задачи", "POST", "/tasks/nonexistent/cancel", "", http.StatusNotFound},
			{"Удаление несуществующей задачи", "DELETE", "/tasks/nonexistent", "", http.StatusNotFound},
			{"Неподдерживаемый метод", "PATCH", "/tasks/123", "", http.StatusMethodNotAllowed},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var body *bytes.Buffer
				if tt.body != "" {
					body = bytes.NewBufferString(tt.body)
				} else {
					body = bytes.NewBufferString("")
				}

				req := httptest.NewRequest(tt.method, tt.path, body)
				w := httptest.NewRecorder()

				mux.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedCode, w.Code)
			})
		}
	})

	// Тест для несуществующего эндпоинта
	t.Run("Несуществующий эндпоинт", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
