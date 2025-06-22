package handlers

import (
	"bytes"
	"encoding/json"
	"http_api/internal/models"
	"http_api/internal/services"
	"http_api/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskHandlers(t *testing.T) {
	// Инициализируем реальные зависимости
	store := storage.NewInMemoryTaskStorage()
	service := services.NewTaskService(store)
	handler := NewTaskHandler(service)

	t.Run("Create and get task", func(t *testing.T) {
		// Создаем задачу
		createBody := bytes.NewBufferString(`{"description":"test task"}`)
		createReq := httptest.NewRequest("POST", "/tasks", createBody)
		createRec := httptest.NewRecorder()

		handler.HandleTasks(createRec, createReq)

		if createRec.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, createRec.Code)
		}

		var createdTask models.Task
		if err := json.NewDecoder(createRec.Body).Decode(&createdTask); err != nil {
			t.Fatal(err)
		}

		// Получаем задачу
		getReq := httptest.NewRequest("GET", "/tasks/"+createdTask.ID, nil)
		getRec := httptest.NewRecorder()

		handler.HandleTaskByID(getRec, getReq)

		if getRec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, getRec.Code)
		}

		var retrievedTask models.Task
		if err := json.NewDecoder(getRec.Body).Decode(&retrievedTask); err != nil {
			t.Fatal(err)
		}

		if retrievedTask.ID != createdTask.ID {
			t.Errorf("Expected task ID %s, got %s", createdTask.ID, retrievedTask.ID)
		}
	})

	t.Run("List tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tasks", nil)
		rec := httptest.NewRecorder()

		handler.HandleTasks(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var taskList models.TaskList
		if err := json.NewDecoder(rec.Body).Decode(&taskList); err != nil {
			t.Fatal(err)
		}

		if taskList.Total < 1 {
			t.Error("Expected at least one task in list")
		}
	})

	t.Run("Update task", func(t *testing.T) {
		// Сначала создаем задачу для обновления
		createBody := bytes.NewBufferString(`{"description":"to update"}`)
		createReq := httptest.NewRequest("POST", "/tasks", createBody)
		createRec := httptest.NewRecorder()
		handler.HandleTasks(createRec, createReq)

		var task models.Task
		json.NewDecoder(createRec.Body).Decode(&task)

		// Обновляем задачу
		updateBody := bytes.NewBufferString(`{"description":"updated"}`)
		updateReq := httptest.NewRequest("PUT", "/tasks/"+task.ID, updateBody)
		updateRec := httptest.NewRecorder()

		handler.HandleTaskByID(updateRec, updateReq)

		if updateRec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, updateRec.Code)
		}
	})

	t.Run("Cancel task", func(t *testing.T) {
		// Создаем задачу для отмены
		createBody := bytes.NewBufferString(`{"description":"to cancel"}`)
		createReq := httptest.NewRequest("POST", "/tasks", createBody)
		createRec := httptest.NewRecorder()
		handler.HandleTasks(createRec, createReq)

		var task models.Task
		json.NewDecoder(createRec.Body).Decode(&task)

		// Отменяем задачу
		cancelReq := httptest.NewRequest("POST", "/tasks/"+task.ID+"/cancel", nil)
		cancelRec := httptest.NewRecorder()

		handler.HandleTaskByID(cancelRec, cancelReq)

		if cancelRec.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, cancelRec.Code)
		}
	})

	t.Run("Delete task", func(t *testing.T) {
		// Создаем задачу для удаления
		createBody := bytes.NewBufferString(`{"description":"to delete"}`)
		createReq := httptest.NewRequest("POST", "/tasks", createBody)
		createRec := httptest.NewRecorder()
		handler.HandleTasks(createRec, createReq)

		var task models.Task
		json.NewDecoder(createRec.Body).Decode(&task)

		// Удаляем задачу
		deleteReq := httptest.NewRequest("DELETE", "/tasks/"+task.ID, nil)
		deleteRec := httptest.NewRecorder()

		handler.HandleTaskByID(deleteRec, deleteReq)

		if deleteRec.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, deleteRec.Code)
		}
	})

	t.Run("Invalid requests", func(t *testing.T) {
		tests := []struct {
			name   string
			method string
			url    string
			body   string
			want   int
		}{
			{"Invalid JSON", "POST", "/tasks", `{"description":}`, http.StatusBadRequest},
			{"Empty description", "POST", "/tasks", `{"description":""}`, http.StatusBadRequest},
			{"Nonexistent task", "GET", "/tasks/nonexistent", "", http.StatusNotFound},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var body *bytes.Buffer
				if tt.body != "" {
					body = bytes.NewBufferString(tt.body)
				} else {
					body = bytes.NewBufferString("")
				}

				req := httptest.NewRequest(tt.method, tt.url, body)
				rec := httptest.NewRecorder()

				if tt.method == "GET" && tt.url != "/tasks" {
					handler.HandleTaskByID(rec, req)
				} else {
					handler.HandleTasks(rec, req)
				}

				if rec.Code != tt.want {
					t.Errorf("Expected status %d, got %d", tt.want, rec.Code)
				}
			})
		}
	})
}
