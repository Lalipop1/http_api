package api_test

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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskAPI(t *testing.T) {
	// Setup
	storage := storage.NewInMemoryTaskStorage()
	service := services.NewTaskService(storage)
	handler := handlers.NewTaskHandler(service)

	t.Run("Full task lifecycle", func(t *testing.T) {
		// Create task
		createBody := bytes.NewBufferString(`{"description":"integration test"}`)
		createReq := httptest.NewRequest("POST", "/tasks", createBody)
		createRec := httptest.NewRecorder()

		handler.HandleTasks(createRec, createReq)

		assert.Equal(t, http.StatusCreated, createRec.Code)

		var createdTask models.Task
		err := json.NewDecoder(createRec.Body).Decode(&createdTask)
		assert.NoError(t, err)
		assert.Equal(t, models.StatusPending, createdTask.Status)

		// Get task status
		getReq := httptest.NewRequest("GET", "/tasks/"+createdTask.ID, nil)
		getRec := httptest.NewRecorder()

		handler.HandleTaskByID(getRec, getReq)

		assert.Equal(t, http.StatusOK, getRec.Code)

		var retrievedTask models.Task
		err = json.NewDecoder(getRec.Body).Decode(&retrievedTask)
		assert.NoError(t, err)
		assert.Equal(t, createdTask.ID, retrievedTask.ID)

		// Wait for task completion (simulate)
		time.Sleep(100 * time.Millisecond)

		// Verify final status
		getFinalReq := httptest.NewRequest("GET", "/tasks/"+createdTask.ID, nil)
		getFinalRec := httptest.NewRecorder()

		handler.HandleTaskByID(getFinalRec, getFinalReq)

		assert.Equal(t, http.StatusOK, getFinalRec.Code)

		var finalTask models.Task
		err = json.NewDecoder(getFinalRec.Body).Decode(&finalTask)
		assert.NoError(t, err)
		assert.Equal(t, models.StatusCompleted, finalTask.Status)
		assert.True(t, finalTask.Duration > 0)
	})
}
